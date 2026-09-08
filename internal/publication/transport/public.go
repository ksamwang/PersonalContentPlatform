package transport

import (
	"encoding/xml"
	"github.com/go-chi/chi/v5"
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
}

func (h *HTTP) list(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(r.Context(), `SELECT v.locale,v.type,v.slug,v.title,v.summary,v.rendered_html,v.published_at FROM publication_views v JOIN workspaces ws ON ws.id=v.workspace_id WHERE ws.slug=$1 AND v.locale=$2 AND v.visibility='public' ORDER BY v.published_at DESC LIMIT 50`, chi.URLParam(r, "workspace"), chi.URLParam(r, "locale"))
	if err != nil {
		httpx.Error(w, 500, "public_list_failed", "public content is unavailable")
		return
	}
	defer rows.Close()
	items := []Page{}
	for rows.Next() {
		var p Page
		if err = rows.Scan(&p.Locale, &p.Type, &p.Slug, &p.Title, &p.Summary, &p.HTML, &p.PublishedAt); err != nil {
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
	Locale      string    `json:"locale"`
	Type        string    `json:"type"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Summary     string    `json:"summary"`
	HTML        string    `json:"html"`
	PublishedAt time.Time `json:"published_at"`
}

func (h *HTTP) page(w http.ResponseWriter, r *http.Request) {
	var p Page
	err := h.db.QueryRow(r.Context(), `SELECT v.locale,v.type,v.slug,v.title,v.summary,v.rendered_html,v.published_at FROM publication_views v JOIN workspaces w ON w.id=v.workspace_id WHERE w.slug=$1 AND v.locale=$2 AND v.type=$3 AND v.slug=$4 AND v.visibility IN ('public','unlisted')`, chi.URLParam(r, "workspace"), chi.URLParam(r, "locale"), chi.URLParam(r, "type"), chi.URLParam(r, "slug")).Scan(&p.Locale, &p.Type, &p.Slug, &p.Title, &p.Summary, &p.HTML, &p.PublishedAt)
	if err != nil {
		httpx.Error(w, 404, "publication_not_found", "published content was not found")
		return
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
	rows, err := h.db.Query(r.Context(), `SELECT v.type,v.slug,v.title,v.summary,v.published_at FROM publication_views v JOIN workspaces w ON w.id=v.workspace_id WHERE w.slug=$1 AND v.locale=$2 AND v.visibility='public' ORDER BY v.published_at DESC LIMIT 50`, workspace, locale)
	if err != nil {
		httpx.Error(w, 500, "rss_failed", "feed is unavailable")
		return
	}
	defer rows.Close()
	feed := rssDoc{Version: "2.0", Channel: rssChannel{Title: "Personal Content Platform", Link: "/" + locale, Description: "Latest publications"}}
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
