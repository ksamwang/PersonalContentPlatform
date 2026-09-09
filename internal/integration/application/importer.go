package application

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	assetdomain "github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
)

type AssetMigrator interface {
	Prepare(context.Context, uuid.UUID, uuid.UUID, string, string, int64) (assetdomain.UploadPlan, error)
	Upload(context.Context, uuid.UUID, uuid.UUID, io.Reader, int64) error
	Complete(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (assetdomain.Asset, error)
}
type ImportService struct {
	db     *pgxpool.Pool
	assets AssetMigrator
}

func NewImportService(db *pgxpool.Pool, assets AssetMigrator) *ImportService {
	return &ImportService{db, assets}
}

type ImportReport struct {
	Format        string   `json:"format"`
	Contents      int      `json:"contents"`
	Localizations int      `json:"localizations"`
	Assets        int      `json:"assets"`
	Entities      int      `json:"entities"`
	Relations     int      `json:"relations"`
	Conflicts     []string `json:"conflicts"`
	Renamed       []string `json:"renamed"`
	Applied       bool     `json:"applied"`
}
type importPayload struct {
	manifest domain.Manifest
	files    map[uuid.UUID][]byte
}

var htmlTag = regexp.MustCompile(`(?s)<[^>]+>`)
var htmlTitle = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

func textBody(text string) json.RawMessage {
	paragraphs := []map[string]any{}
	for _, part := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n\n") {
		part = strings.TrimSpace(part)
		if part != "" {
			paragraphs = append(paragraphs, map[string]any{"type": "paragraph", "content": []map[string]string{{"type": "text", "text": part}}})
		}
	}
	raw, _ := json.Marshal(map[string]any{"schemaVersion": 1, "type": "doc", "content": paragraphs})
	return raw
}
func portableManifest(format string, data []byte) (domain.Manifest, error) {
	text := string(data)
	title := "Imported content"
	locale := "zh-CN"
	slug := "imported-content"
	if format == "html" {
		if match := htmlTitle.FindStringSubmatch(text); len(match) > 1 {
			title = strings.TrimSpace(htmlTag.ReplaceAllString(match[1], ""))
		}
		text = htmlTag.ReplaceAllString(text, " ")
	} else {
		if strings.HasPrefix(text, "---\n") {
			parts := strings.SplitN(text, "---\n", 3)
			if len(parts) == 3 {
				for _, line := range strings.Split(parts[1], "\n") {
					pair := strings.SplitN(line, ":", 2)
					if len(pair) == 2 {
						switch strings.TrimSpace(pair[0]) {
						case "title":
							title = strings.TrimSpace(pair[1])
						case "locale":
							locale = strings.TrimSpace(pair[1])
						case "slug":
							slug = strings.TrimSpace(pair[1])
						}
					}
				}
				text = parts[2]
			}
		}
	}
	contentID, localID, revisionID := id.New(), id.New(), id.New()
	body := textBody(text)
	sum := sha256.Sum256(append([]byte(title), body...))
	return domain.Manifest{SchemaVersion: 1, Contents: []domain.Content{{ID: contentID, Type: "article", DefaultLocale: locale, Visibility: "private"}}, Localizations: []domain.Localization{{ID: localID, ContentID: contentID, Locale: locale, Slug: slug, State: "draft", TranslationStatus: "draft", CurrentRevisionID: &revisionID}}, Revisions: []domain.Revision{{ID: revisionID, LocalizationID: localID, Sequence: 1, SchemaVersion: 1, Title: title, Body: body, Metadata: json.RawMessage(`{}`), ContentHash: hex.EncodeToString(sum[:])}}}, nil
}
func parseImport(format string, data []byte) (importPayload, error) {
	payload := importPayload{files: map[uuid.UUID][]byte{}}
	switch format {
	case "markdown", "html":
		manifest, err := portableManifest(format, data)
		payload.manifest = manifest
		return payload, err
	case "json":
		return payload, json.Unmarshal(data, &payload.manifest)
	case "bundle":
		reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return payload, err
		}
		for _, file := range reader.File {
			body, readErr := func() ([]byte, error) {
				r, e := file.Open()
				if e != nil {
					return nil, e
				}
				defer r.Close()
				return io.ReadAll(r)
			}()
			if readErr != nil {
				return payload, readErr
			}
			if file.Name == "manifest.json" {
				if err = json.Unmarshal(body, &payload.manifest); err != nil {
					return payload, err
				}
			} else if strings.HasPrefix(file.Name, "assets/") {
				parts := strings.Split(file.Name, "/")
				if len(parts) >= 3 {
					if assetID, e := uuid.Parse(parts[1]); e == nil {
						payload.files[assetID] = body
					}
				}
			}
		}
		return payload, nil
	default:
		return payload, fmt.Errorf("unsupported import format")
	}
}
func (s *ImportService) Preview(ctx context.Context, ws uuid.UUID, format string, data []byte) (ImportReport, error) {
	payload, err := parseImport(format, data)
	if err != nil {
		return ImportReport{}, err
	}
	report := ImportReport{Format: format, Contents: len(payload.manifest.Contents), Localizations: len(payload.manifest.Localizations), Assets: len(payload.files), Entities: len(payload.manifest.Entities), Relations: len(payload.manifest.Relations), Conflicts: []string{}, Renamed: []string{}}
	for _, local := range payload.manifest.Localizations {
		var exists bool
		_ = s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM content_localizations WHERE workspace_id=$1 AND locale=$2 AND slug=$3)`, ws, local.Locale, local.Slug).Scan(&exists)
		if exists {
			report.Conflicts = append(report.Conflicts, local.Locale+"/"+local.Slug)
		}
	}
	return report, nil
}
func (s *ImportService) Apply(ctx context.Context, ws, user uuid.UUID, format string, data []byte) (ImportReport, error) {
	payload, err := parseImport(format, data)
	if err != nil {
		return ImportReport{}, err
	}
	report, err := s.Preview(ctx, ws, format, data)
	if err != nil {
		return report, err
	}
	assetMap := map[uuid.UUID]uuid.UUID{}
	if s.assets != nil {
		for _, asset := range payload.manifest.Assets {
			body, ok := payload.files[asset.ID]
			if !ok {
				continue
			}
			plan, e := s.assets.Prepare(ctx, ws, user, filepath.Base(asset.Filename), asset.MIME, int64(len(body)))
			if e != nil {
				return report, e
			}
			if e = s.assets.Upload(ctx, ws, plan.UploadID, bytes.NewReader(body), int64(len(body))); e != nil {
				return report, e
			}
			imported, e := s.assets.Complete(ctx, ws, user, plan.UploadID)
			if e != nil {
				return report, e
			}
			assetMap[asset.ID] = imported.ID
		}
	}
	contentMap := map[uuid.UUID]uuid.UUID{}
	localMap := map[uuid.UUID]uuid.UUID{}
	revisionByLocal := map[uuid.UUID]domain.Revision{}
	for _, revision := range payload.manifest.Revisions {
		revisionByLocal[revision.LocalizationID] = revision
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return report, err
	}
	defer tx.Rollback(ctx)
	for _, content := range payload.manifest.Contents {
		newID := id.New()
		contentMap[content.ID] = newID
		if _, err = tx.Exec(ctx, `INSERT INTO objects(id,workspace_id,kind) VALUES($1,$2,'content')`, newID, ws); err != nil {
			return report, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO contents(object_id,workspace_id,type,default_locale,visibility) VALUES($1,$2,$3,$4,$5)`, newID, ws, content.Type, content.DefaultLocale, content.Visibility); err != nil {
			return report, err
		}
	}
	for _, local := range payload.manifest.Localizations {
		contentID, ok := contentMap[local.ContentID]
		if !ok {
			continue
		}
		slug := safeName(local.Slug)
		base := slug
		for n := 2; ; n++ {
			var exists bool
			if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM content_localizations WHERE workspace_id=$1 AND locale=$2 AND slug=$3)`, ws, local.Locale, slug).Scan(&exists); err != nil {
				return report, err
			}
			if !exists {
				break
			}
			slug = fmt.Sprintf("%s-import-%d", base, n)
		}
		if slug != local.Slug {
			report.Renamed = append(report.Renamed, local.Locale+"/"+local.Slug+" -> "+slug)
		}
		localID, revisionID := id.New(), id.New()
		localMap[local.ID] = localID
		revision := revisionByLocal[local.ID]
		for oldID, newID := range assetMap {
			revision.Body = bytes.ReplaceAll(revision.Body, []byte(oldID.String()), []byte(newID.String()))
			revision.Metadata = bytes.ReplaceAll(revision.Metadata, []byte(oldID.String()), []byte(newID.String()))
		}
		if len(revision.Body) == 0 {
			revision.Body = json.RawMessage(`{"schemaVersion":1,"type":"doc","content":[]}`)
		}
		if len(revision.Metadata) == 0 {
			revision.Metadata = json.RawMessage(`{}`)
		}
		hash := revision.ContentHash
		if hash == "" {
			sum := sha256.Sum256(append([]byte(revision.Title), revision.Body...))
			hash = hex.EncodeToString(sum[:])
		}
		if _, err = tx.Exec(ctx, `INSERT INTO content_localizations(id,workspace_id,content_id,locale,state,slug,current_revision_id,translation_status) VALUES($1,$2,$3,$4,'draft',$5,NULL,'draft')`, localID, ws, contentID, local.Locale, slug); err != nil {
			return report, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO content_revisions(id,workspace_id,localization_id,seq,schema_version,title,summary,body_json,metadata_json,content_hash,created_by) VALUES($1,$2,$3,1,$4,$5,$6,$7,$8,$9,$10)`, revisionID, ws, localID, revision.SchemaVersion, revision.Title, revision.Summary, revision.Body, revision.Metadata, hash, user); err != nil {
			return report, err
		}
		if _, err = tx.Exec(ctx, `UPDATE content_localizations SET current_revision_id=$1 WHERE id=$2`, revisionID, localID); err != nil {
			return report, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO draft_buffers(localization_id,workspace_id,version,title,summary,body_json,metadata_json,updated_by) VALUES($1,$2,1,$3,$4,$5,$6,$7)`, localID, ws, revision.Title, revision.Summary, revision.Body, revision.Metadata, user); err != nil {
			return report, err
		}
	}
	entityMap := map[uuid.UUID]uuid.UUID{}
	for _, entity := range payload.manifest.Entities {
		newID := id.New()
		entityMap[entity.ID] = newID
		if _, err = tx.Exec(ctx, `INSERT INTO objects(id,workspace_id,kind) VALUES($1,$2,'entity')`, newID, ws); err != nil {
			return report, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO entities(object_id,workspace_id,type,canonical_name,description) VALUES($1,$2,$3,$4,$5)`, newID, ws, entity.Type, entity.CanonicalName, entity.Description); err != nil {
			return report, err
		}
	}
	for _, alias := range payload.manifest.EntityAliases {
		if entityID, ok := entityMap[alias.EntityID]; ok {
			_, err = tx.Exec(ctx, `INSERT INTO entity_aliases(id,workspace_id,entity_id,alias,locale) VALUES($1,$2,$3,$4,$5) ON CONFLICT DO NOTHING`, id.New(), ws, entityID, alias.Alias, alias.Locale)
			if err != nil {
				return report, err
			}
		}
	}
	predicateMap := map[uuid.UUID]uuid.UUID{}
	for _, predicate := range payload.manifest.Predicates {
		newID := id.New()
		if err = tx.QueryRow(ctx, `INSERT INTO relation_predicates(id,workspace_id,key,label_zh,label_en) VALUES($1,$2,$3,$4,$5) ON CONFLICT(workspace_id,key) DO UPDATE SET label_zh=EXCLUDED.label_zh RETURNING id`, newID, ws, predicate.Key, predicate.LabelZH, predicate.LabelEN).Scan(&newID); err != nil {
			return report, err
		}
		predicateMap[predicate.ID] = newID
	}
	for _, relation := range payload.manifest.Relations {
		source := contentMap[relation.SourceID]
		if source == uuid.Nil {
			source = entityMap[relation.SourceID]
		}
		target := contentMap[relation.TargetID]
		if target == uuid.Nil {
			target = entityMap[relation.TargetID]
		}
		predicate := predicateMap[relation.PredicateID]
		if source != uuid.Nil && target != uuid.Nil && predicate != uuid.Nil {
			_, err = tx.Exec(ctx, `INSERT INTO relations(id,workspace_id,source_object_id,target_object_id,predicate_id,confirmed,provenance) VALUES($1,$2,$3,$4,$5,$6,$7)`, id.New(), ws, source, target, predicate, relation.Confirmed, relation.Provenance)
			if err != nil {
				return report, err
			}
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return report, err
	}
	for _, usage := range payload.manifest.AssetUsages {
		assetID, assetOK := assetMap[usage.AssetID]
		ownerID, ownerOK := contentMap[usage.OwnerObjectID]
		if !ownerOK {
			ownerID, ownerOK = entityMap[usage.OwnerObjectID]
		}
		if assetOK && ownerOK {
			_, _ = s.db.Exec(ctx, `INSERT INTO asset_usages(id,workspace_id,asset_id,owner_object_id,role,locator_json) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, id.New(), ws, assetID, ownerID, usage.Role, usage.Locator)
		}
	}
	report.Applied = true
	return report, nil
}
