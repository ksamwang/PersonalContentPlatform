package application

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/ai/infrastructure/anthropic"
	"github.com/ksamwang/PersonalContentPlatform/internal/ai/infrastructure/openai"
	"github.com/ksamwang/PersonalContentPlatform/internal/ai/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/document"
	settingsports "github.com/ksamwang/PersonalContentPlatform/internal/settings/ports"
)

type Service struct {
	db       *pgxpool.Pool
	settings settingsports.Repository
}
type Hit struct {
	ContentID uuid.UUID `json:"content_id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	Locale    string    `json:"locale"`
	Type      string    `json:"type"`
	Slug      string    `json:"slug"`
	Excerpt   string    `json:"excerpt"`
	Score     float64   `json:"score"`
}

type indexSource struct {
	content, localization, revision uuid.UUID
	locale                          string
	texts                           []string
}

func New(db *pgxpool.Pool, settings settingsports.Repository) *Service {
	return &Service{db: db, settings: settings}
}

func chunks(value string) []string {
	r := []rune(strings.TrimSpace(value))
	if len(r) == 0 {
		return nil
	}
	const size = 900
	const overlap = 120
	out := []string{}
	for start := 0; start < len(r); start += size - overlap {
		end := start + size
		if end > len(r) {
			end = len(r)
		}
		out = append(out, string(r[start:end]))
		if end == len(r) {
			break
		}
	}
	return out
}
func vectorLiteral(values []float32) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.FormatFloat(float64(v), 'f', -1, 32)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
func (s *Service) embedder(ctx context.Context, ws uuid.UUID) (*openai.Provider, string, error) {
	cfg, err := s.settings.Embedding(ctx, ws)
	if err != nil {
		return nil, "", err
	}
	if cfg.BaseURL == "" || cfg.APIKey == "" || cfg.Model == "" {
		return nil, "", fmt.Errorf("embedding provider configuration is incomplete")
	}
	return openai.New(cfg.BaseURL, cfg.APIKey, cfg.Model), cfg.Model, nil
}

func (s *Service) Index(ctx context.Context, ws uuid.UUID) (int, error) {
	sources, err := s.indexSources(ctx, ws, uuid.Nil)
	if err != nil {
		return 0, err
	}
	return s.storeSources(ctx, ws, uuid.Nil, sources)
}

func (s *Service) IndexContent(ctx context.Context, ws, contentID uuid.UUID) (int, error) {
	sources, err := s.indexSources(ctx, ws, contentID)
	if err != nil {
		return 0, err
	}
	return s.storeSources(ctx, ws, contentID, sources)
}

func (s *Service) indexSources(ctx context.Context, ws, contentID uuid.UUID) ([]indexSource, error) {
	query := `SELECT c.object_id,l.id,rv.id,l.locale,rv.title,rv.summary,rv.body_json FROM contents c JOIN content_localizations l ON l.content_id=c.object_id JOIN content_revisions rv ON rv.id=l.current_revision_id WHERE c.workspace_id=$1 AND c.deleted_at IS NULL AND l.deleted_at IS NULL AND l.state<>'archived'`
	args := []any{ws}
	if contentID != uuid.Nil {
		query += ` AND c.object_id=$2`
		args = append(args, contentID)
	}
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	sources := []indexSource{}
	for rows.Next() {
		var source indexSource
		var title, summary string
		var body []byte
		if err = rows.Scan(&source.content, &source.localization, &source.revision, &source.locale, &title, &summary, &body); err != nil {
			return nil, err
		}
		source.texts = chunks(title + "\n" + summary + "\n" + document.PlainText(body))
		sources = append(sources, source)
	}
	return sources, rows.Err()
}

func (s *Service) storeSources(ctx context.Context, ws, contentID uuid.UUID, sources []indexSource) (int, error) {
	all := []string{}
	for _, source := range sources {
		all = append(all, source.texts...)
	}
	if len(all) == 0 {
		if contentID == uuid.Nil {
			_, err := s.db.Exec(ctx, `DELETE FROM content_chunks WHERE workspace_id=$1`, ws)
			return 0, err
		}
		_, err := s.db.Exec(ctx, `DELETE FROM content_chunks WHERE workspace_id=$1 AND content_id=$2`, ws, contentID)
		return 0, err
	}
	provider, model, err := s.embedder(ctx, ws)
	if err != nil {
		return 0, err
	}
	vectors := [][]float32{}
	for start := 0; start < len(all); start += 32 {
		end := start + 32
		if end > len(all) {
			end = len(all)
		}
		batch, _, embedErr := provider.Embed(ctx, all[start:end], model)
		if embedErr != nil {
			return 0, embedErr
		}
		vectors = append(vectors, batch...)
	}
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)
	if contentID == uuid.Nil {
		_, err = tx.Exec(ctx, `DELETE FROM content_chunks WHERE workspace_id=$1`, ws)
	} else {
		_, err = tx.Exec(ctx, `DELETE FROM content_chunks WHERE workspace_id=$1 AND content_id=$2`, ws, contentID)
	}
	if err != nil {
		return 0, err
	}
	offset := 0
	for _, src := range sources {
		for i, text := range src.texts {
			if _, err = tx.Exec(ctx, `INSERT INTO content_chunks(id,workspace_id,content_id,localization_id,revision_id,locale,chunk_index,text,embedding) VALUES(gen_random_uuid(),$1,$2,$3,$4,$5,$6,$7,$8::vector)`, ws, src.content, src.localization, src.revision, src.locale, i, text, vectorLiteral(vectors[offset])); err != nil {
				return 0, err
			}
			offset++
		}
	}
	return len(all), tx.Commit(ctx)
}

func (s *Service) Search(ctx context.Context, ws uuid.UUID, q, locale string, limit int) ([]Hit, error) {
	if limit < 1 || limit > 50 {
		limit = 20
	}
	provider, model, err := s.embedder(ctx, ws)
	if err != nil {
		return s.fullTextSearch(ctx, ws, q, locale, limit)
	}
	vectors, _, err := provider.Embed(ctx, []string{q}, model)
	if err != nil || len(vectors) == 0 {
		return s.fullTextSearch(ctx, ws, q, locale, limit)
	}
	rows, err := s.db.Query(ctx, `SELECT ch.content_id,rv.title,rv.summary,ch.locale,c.type,l.slug,ch.text,0.35*ts_rank_cd(ch.tsv,websearch_to_tsquery('simple',$2))+0.65*(1-(ch.embedding <=> $3::vector)) score FROM content_chunks ch JOIN contents c ON c.object_id=ch.content_id JOIN content_localizations l ON l.id=ch.localization_id JOIN content_revisions rv ON rv.id=ch.revision_id WHERE ch.workspace_id=$1 AND ($4='' OR ch.locale=$4) ORDER BY score DESC LIMIT $5`, ws, q, vectorLiteral(vectors[0]), locale, limit)
	if err != nil {
		return s.fullTextSearch(ctx, ws, q, locale, limit)
	}
	defer rows.Close()
	hits := []Hit{}
	seen := map[uuid.UUID]bool{}
	for rows.Next() {
		var h Hit
		if err = rows.Scan(&h.ContentID, &h.Title, &h.Summary, &h.Locale, &h.Type, &h.Slug, &h.Excerpt, &h.Score); err != nil {
			return nil, err
		}
		if !seen[h.ContentID] {
			hits = append(hits, h)
			seen[h.ContentID] = true
		}
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(hits) == 0 {
		return s.fullTextSearch(ctx, ws, q, locale, limit)
	}
	return hits, nil
}

func (s *Service) fullTextSearch(ctx context.Context, ws uuid.UUID, q, locale string, limit int) ([]Hit, error) {
	rows, err := s.db.Query(ctx, `SELECT s.object_id,s.title,s.summary,s.locale,c.type,l.slug,left(s.body_text,600),ts_rank_cd(s.tsv,websearch_to_tsquery('simple',$2)) FROM search_documents s JOIN contents c ON c.object_id=s.object_id JOIN content_localizations l ON l.content_id=s.object_id AND l.locale=s.locale WHERE s.workspace_id=$1 AND ($3='' OR s.locale=$3) AND s.tsv @@ websearch_to_tsquery('simple',$2) ORDER BY ts_rank_cd(s.tsv,websearch_to_tsquery('simple',$2)) DESC LIMIT $4`, ws, q, locale, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	hits := []Hit{}
	for rows.Next() {
		var hit Hit
		if err = rows.Scan(&hit.ContentID, &hit.Title, &hit.Summary, &hit.Locale, &hit.Type, &hit.Slug, &hit.Excerpt, &hit.Score); err != nil {
			return nil, err
		}
		hits = append(hits, hit)
	}
	return hits, rows.Err()
}

func (s *Service) RAG(ctx context.Context, ws uuid.UUID, q, locale string) (string, []Hit, error) {
	hits, err := s.Search(ctx, ws, q, locale, 8)
	if err != nil {
		return "", nil, err
	}
	raw, _ := json.Marshal(hits)
	cfg, err := s.settings.AI(ctx, ws)
	if err != nil {
		return "", nil, err
	}
	var provider ports.Provider = openai.New(cfg.BaseURL, cfg.APIKey, cfg.Model)
	if cfg.Provider == "anthropic-compatible" {
		provider = anthropic.New(cfg.BaseURL, cfg.APIKey, cfg.Model)
	}
	result, err := provider.Generate(ctx, ports.Request{System: "Answer only from the supplied sources. If the sources are insufficient, say so. Cite sources inline as [1], [2].", User: fmt.Sprintf("Question: %s\nSources: %s", q, raw), Model: cfg.Model})
	if err != nil {
		return "", nil, err
	}
	return result.Text, hits, nil
}
