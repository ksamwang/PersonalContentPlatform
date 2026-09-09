package transport

import (
	"encoding/json"
	"encoding/xml"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/httpx"
	"net/http"
	"time"
)

type HTTP struct{ db *pgxpool.Pool }

func NewHTTP(db *pgxpool.Pool) *HTTP { return &HTTP{db: db} }
func (h *HTTP) Register(r chi.Router) {
	r.Get("/v1/public/{workspace}/{locale}/contents", h.list)
	r.Get("/v1/public/{workspace}/{locale}/{type}/{slug}", h.page)
	r.Get("/v1/public/{workspace}/{locale}/rss.xml", h.rss)
	r.Get("/v1/public/{workspace}/{locale}/tags", h.tags)
	r.Get("/v1/public/{workspace}/{locale}/collections", h.collections)
	r.Get("/v1/public/{workspace}/{locale}/collections/{slug}", h.collection)
}

func (h *HTTP) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT v.content_id,v.locale,v.type,v.slug,v.title,v.summary,v.rendered_html,v.metadata_json,v.published_at FROM publication_views v JOIN workspaces ws ON ws.id=v.workspace_id WHERE ws.slug=$1 AND v.locale=$2 AND v.visibility='public' AND ($3='' OR v.type=$3) AND ($4='' OR v.title ILIKE '%'||$4||'%' OR v.summary ILIKE '%'||$4||'%' OR v.rendered_html ILIKE '%'||$4||'%') AND ($5='' OR EXISTS(SELECT 1 FROM jsonb_array_elements_text(CASE WHEN jsonb_typeof(v.metadata_json->'tags')='array' THEN v.metadata_json->'tags' ELSE '[]'::jsonb END) tag(value) WHERE lower(tag.value)=lower($5))) ORDER BY v.published_at DESC LIMIT 50`, chi.URLParam(r, "workspace"), chi.URLParam(r, "locale"), r.URL.Query().Get("type"), r.URL.Query().Get("q"), r.URL.Query().Get("tag"))
	if err != nil {
		httpx.Error(w, 500, "public_list_failed", "public content is unavailable")
		return
	}
	defer rows.Close()
	items := []Page{}
	for rows.Next() {
		var p Page
		if err = rows.Scan(&p.ContentID, &p.Locale, &p.Type, &p.Slug, &p.Title, &p.Summary, &p.HTML, &p.Metadata, &p.PublishedAt); err != nil {
			httpx.Error(w, 500, "public_list_failed", "public content is unavailable")
			return
		}
		p.HTML = ""
		items = append(items, p)
	}
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	httpx.JSON(w, 200, map[string]any{"items": items})
}

type Page struct {
	ContentID   uuid.UUID       `json:"content_id"`
	Locale      string          `json:"locale"`
	Type        string          `json:"type"`
	Slug        string          `json:"slug"`
	Title       string          `json:"title"`
	Summary     string          `json:"summary"`
	HTML        string          `json:"html"`
	PublishedAt time.Time       `json:"published_at"`
	Metadata    json.RawMessage `json:"metadata"`
	Alternates  []Alternate     `json:"alternates,omitempty"`
}
type Alternate struct {
	Locale string `json:"locale"`
	Type   string `json:"type"`
	Slug   string `json:"slug"`
}

func (h *HTTP) page(w http.ResponseWriter, r *http.Request) {
	var p Page
	err := h.db.QueryRow(r.Context(), `SELECT v.content_id,v.locale,v.type,v.slug,v.title,v.summary,v.rendered_html,v.metadata_json,v.published_at FROM publication_views v JOIN workspaces w ON w.id=v.workspace_id WHERE w.slug=$1 AND v.locale=$2 AND v.type=$3 AND v.slug=$4 AND v.visibility IN ('public','unlisted')`, chi.URLParam(r, "workspace"), chi.URLParam(r, "locale"), chi.URLParam(r, "type"), chi.URLParam(r, "slug")).Scan(&p.ContentID, &p.Locale, &p.Type, &p.Slug, &p.Title, &p.Summary, &p.HTML, &p.Metadata, &p.PublishedAt)
	if err != nil {
		httpx.Error(w, 404, "publication_not_found", "published content was not found")
		return
	}
	rows, _ := h.db.Query(r.Context(), `SELECT locale,type,slug FROM publication_views WHERE content_id=$1 AND visibility IN ('public','unlisted') ORDER BY locale`, p.ContentID)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var alt Alternate
			if rows.Scan(&alt.Locale, &alt.Type, &alt.Slug) == nil {
				p.Alternates = append(p.Alternates, alt)
			}
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=60, stale-while-revalidate=300")
	httpx.JSON(w, 200, p)
}

type rssDoc struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Channel rssChannel `xml:"channel"`
}
type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []rssItem `xml:"item"`
}
type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Published   string `xml:"pubDate"`
}

func (h *HTTP) rss(w http.ResponseWriter, r *http.Request) {
	workspace, locale := chi.URLParam(r, "workspace"), chi.URLParam(r, "locale")
	var siteName, publicURL, description string
	var enabled bool
	if err := h.db.QueryRow(r.Context(), `SELECT COALESCE(NULLIF(settings_json#>>'{site,name}',''),name),COALESCE(settings_json#>>'{site,public_url}',''),COALESCE(settings_json#>>'{site,description}','Latest publications'),COALESCE((settings_json#>>'{site,rss_enabled}')::boolean,true) FROM workspaces WHERE slug=$1`, workspace).Scan(&siteName, &publicURL, &description, &enabled); err != nil || !enabled {
		httpx.Error(w, 404, "rss_disabled", "feed is disabled")
		return
	}
	rows, err := h.db.Query(r.Context(), `SELECT v.type,v.slug,v.title,v.summary,v.published_at FROM publication_views v JOIN workspaces w ON w.id=v.workspace_id WHERE w.slug=$1 AND v.locale=$2 AND v.visibility='public' ORDER BY v.published_at DESC LIMIT 50`, workspace, locale)
	if err != nil {
		httpx.Error(w, 500, "rss_failed", "feed is unavailable")
		return
	}
	defer rows.Close()
	link := publicURL
	if link == "" {
		link = "/" + locale
	}
	feed := rssDoc{Version: "2.0", Channel: rssChannel{Title: siteName, Link: link, Description: description}}
	for rows.Next() {
		var kind, slug, title, summary string
		var published time.Time
		if err = rows.Scan(&kind, &slug, &title, &summary, &published); err != nil {
			break
		}
		link := "/" + locale + "/" + kind + "/" + slug
		feed.Channel.Items = append(feed.Channel.Items, rssItem{Title: title, Link: link, Description: summary, Published: published.Format(time.RFC1123Z)})
	}
	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	_ = xml.NewEncoder(w).Encode(feed)
}
