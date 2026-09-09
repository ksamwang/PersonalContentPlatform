package application

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	assetdomain "github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/content/document"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
)

type AssetIO interface {
	Open(context.Context, uuid.UUID, uuid.UUID) (io.ReadCloser, assetdomain.StoredObject, error)
}
type ArchiveService struct {
	exports *ExportService
	assets  AssetIO
}

func NewArchiveService(exports *ExportService, assets AssetIO) *ArchiveService {
	return &ArchiveService{exports, assets}
}
func safeName(value string) string {
	value = regexp.MustCompile(`[^a-zA-Z0-9._-]+`).ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "content"
	}
	return value
}
func (s *ArchiveService) Build(ctx context.Context, ws, actor uuid.UUID, format string) ([]byte, string, string, error) {
	manifest, err := s.exports.Build(ctx, ws, actor)
	if err != nil {
		return nil, "", "", err
	}
	if format == "json" {
		raw, err := json.MarshalIndent(manifest, "", "  ")
		return raw, "workspace-manifest.json", "application/json", err
	}
	var output bytes.Buffer
	archive := zip.NewWriter(&output)
	manifestRaw, _ := json.MarshalIndent(manifest, "", "  ")
	entry, _ := archive.Create("manifest.json")
	_, _ = entry.Write(manifestRaw)
	revisions := map[uuid.UUID]domain.Revision{}
	for _, revision := range manifest.Revisions {
		revisions[revision.ID] = revision
	}
	for _, localization := range manifest.Localizations {
		if localization.CurrentRevisionID == nil {
			continue
		}
		revision, ok := revisions[*localization.CurrentRevisionID]
		if !ok {
			continue
		}
		base := "contents/" + localization.Locale + "/" + safeName(localization.Slug)
		if format == "markdown" || format == "bundle" {
			file, _ := archive.Create(base + ".md")
			_, _ = fmt.Fprintf(file, "---\ntitle: %s\nsummary: %s\nlocale: %s\nslug: %s\n---\n\n%s\n", revision.Title, revision.Summary, localization.Locale, localization.Slug, document.PlainText(revision.Body))
		}
		if format == "html" || format == "bundle" {
			file, _ := archive.Create(base + ".html")
			_, _ = fmt.Fprintf(file, "<!doctype html><html lang=%q><head><meta charset=\"utf-8\"><title>%s</title><meta name=\"description\" content=%q></head><body><h1>%s</h1>%s</body></html>", localization.Locale, revision.Title, revision.Summary, revision.Title, document.HTML(revision.Body))
		}
	}
	if format == "bundle" && s.assets != nil {
		for _, asset := range manifest.Assets {
			reader, _, openErr := s.assets.Open(ctx, ws, asset.ID)
			if openErr != nil {
				continue
			}
			ext := filepath.Ext(asset.Filename)
			file, _ := archive.Create("assets/" + asset.ID.String() + "/original" + ext)
			_, _ = io.Copy(file, reader)
			_ = reader.Close()
		}
	}
	if err = archive.Close(); err != nil {
		return nil, "", "", err
	}
	name := "workspace-" + format + ".zip"
	return output.Bytes(), name, "application/zip", nil
}
