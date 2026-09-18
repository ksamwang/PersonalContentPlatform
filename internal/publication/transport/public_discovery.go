package transport

import (
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
	"time"
)

func (h *HTTP) tags(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT min(btrim(tag.value)),count(DISTINCT v.content_id) FROM publication_views v JOIN workspaces ws ON ws.id=v.workspace_id CROSS JOIN LATERAL jsonb_array_elements_text(CASE WHEN jsonb_typeof(v.metadata_json->'tags')='array' THEN v.metadata_json->'tags' ELSE '[]'::jsonb END) raw_tag(value) CROSS JOIN LATERAL regexp_split_to_table(raw_tag.value,'[,，、;；]+') tag(value) WHERE ws.slug=$1 AND v.locale=$2 AND v.visibility='public' AND btrim(tag.value)<>'' GROUP BY lower(btrim(tag.value)) ORDER BY count(DISTINCT v.content_id) DESC,min(btrim(tag.value))`, chi.URLParam(r, "workspace"), chi.URLParam(r, "locale"))
	if err != nil {
		httpx.Error(w, 500, "tags_failed", "tags are unavailable")
		return
	}
	defer rows.Close()
	items := []map[string]any{}
	for rows.Next() {
		var name string
		var count int
		if rows.Scan(&name, &count) == nil {
			items = append(items, map[string]any{"name": name, "count": count})
		}
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}

type PublicCollection struct {
	Title    string          `json:"title"`
	Slug     string          `json:"slug"`
	Sections []PublicSection `json:"sections"`
}
type PublicSection struct {
	Title string `json:"title"`
	Items []Page `json:"items"`
}

func (h *HTTP) collections(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT c.title,c.slug FROM collections c JOIN workspaces w ON w.id=c.workspace_id WHERE w.slug=$1 AND c.visibility='public' ORDER BY c.title`, chi.URLParam(r, "workspace"))
	if err != nil {
		httpx.Error(w, 500, "collections_failed", "collections are unavailable")
		return
	}
	defer rows.Close()
	items := []PublicCollection{}
	for rows.Next() {
		var item PublicCollection
		if rows.Scan(&item.Title, &item.Slug) == nil {
			item.Sections = []PublicSection{}
			items = append(items, item)
		}
	}
	httpx.JSON(w, 200, map[string]any{"items": items})
}
func (h *HTTP) collection(w http.ResponseWriter, r *http.Request) {
	var result PublicCollection
	if err := h.db.QueryRow(r.Context(), `SELECT c.title,c.slug FROM collections c JOIN workspaces w ON w.id=c.workspace_id WHERE w.slug=$1 AND c.slug=$2 AND c.visibility='public'`, chi.URLParam(r, "workspace"), chi.URLParam(r, "slug")).Scan(&result.Title, &result.Slug); err != nil {
		httpx.Error(w, 404, "collection_not_found", "collection was not found")
		return
	}
	rows, err := h.db.Query(r.Context(), `SELECT s.id,s.title,v.content_id,v.locale,v.type,v.slug,v.title,v.summary,v.metadata_json,v.published_at FROM collection_sections s JOIN collections c ON c.object_id=s.collection_id JOIN workspaces w ON w.id=c.workspace_id LEFT JOIN collection_items i ON i.section_id=s.id LEFT JOIN publication_views v ON v.content_id=i.object_id AND v.locale=$3 AND v.visibility='public' WHERE w.slug=$1 AND c.slug=$2 ORDER BY s.sort_key,i.sort_key`, chi.URLParam(r, "workspace"), chi.URLParam(r, "slug"), chi.URLParam(r, "locale"))
	if err != nil {
		httpx.Error(w, 500, "collection_failed", "collection is unavailable")
		return
	}
	defer rows.Close()
	result.Sections = []PublicSection{}
	index := map[string]int{}
	for rows.Next() {
		var sid, title string
		var contentID *uuid.UUID
		var locale, kind, slug, itemTitle, summary *string
		var metadata []byte
		var published *time.Time
		if rows.Scan(&sid, &title, &contentID, &locale, &kind, &slug, &itemTitle, &summary, &metadata, &published) != nil {
			continue
		}
		idx, ok := index[sid]
		if !ok {
			result.Sections = append(result.Sections, PublicSection{Title: title, Items: []Page{}})
			idx = len(result.Sections) - 1
			index[sid] = idx
		}
		if contentID != nil {
			result.Sections[idx].Items = append(result.Sections[idx].Items, Page{ContentID: *contentID, Locale: *locale, Type: *kind, Slug: *slug, Title: *itemTitle, Summary: *summary, Metadata: metadata, PublishedAt: *published})
		}
	}
	httpx.JSON(w, 200, result)
}
