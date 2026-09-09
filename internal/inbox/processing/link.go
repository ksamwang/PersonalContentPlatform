package processing

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type LinkResult struct {
	Title, Text string
	Cover       *string
}
type LinkFetcher struct{ client *http.Client }

func NewLinkFetcher() *LinkFetcher {
	return &LinkFetcher{client: &http.Client{Timeout: 15 * time.Second}}
}

var titleRE = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
var ogImageRE = regexp.MustCompile(`(?is)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']|<meta[^>]+content=["']([^"']+)["'][^>]+property=["']og:image["']`)
var scriptRE = regexp.MustCompile(`(?is)<(?:script|style|nav|footer)[^>]*>.*?</(?:script|style|nav|footer)>`)
var tagRE = regexp.MustCompile(`(?s)<[^>]+>`)
var spaceRE = regexp.MustCompile(`\s+`)

func (f *LinkFetcher) Fetch(ctx context.Context, url string) (LinkResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return LinkResult{}, err
	}
	req.Header.Set("User-Agent", "PersonalContentPlatform/1.1")
	resp, err := f.client.Do(req)
	if err != nil {
		return LinkResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return LinkResult{}, fmt.Errorf("source returned HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return LinkResult{}, err
	}
	raw := string(body)
	result := LinkResult{}
	if match := titleRE.FindStringSubmatch(raw); len(match) > 1 {
		result.Title = strings.TrimSpace(html.UnescapeString(tagRE.ReplaceAllString(match[1], " ")))
	}
	if match := ogImageRE.FindStringSubmatch(raw); len(match) > 1 {
		value := match[1]
		if value == "" && len(match) > 2 {
			value = match[2]
		}
		if value != "" {
			result.Cover = &value
		}
	}
	clean := scriptRE.ReplaceAllString(raw, " ")
	result.Text = strings.TrimSpace(html.UnescapeString(spaceRE.ReplaceAllString(tagRE.ReplaceAllString(clean, " "), " ")))
	if len([]rune(result.Text)) > 12000 {
		result.Text = string([]rune(result.Text)[:12000])
	}
	return result, nil
}
