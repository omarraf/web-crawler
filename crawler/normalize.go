package crawler

import (
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// Extracted holds the results of a single HTML parse pass.
type Extracted struct {
	Links []string // internal same-domain page links (<a href>)
	Feeds []string // RSS/Atom feed URLs (<link rel="alternate">)
}

// Extract parses HTML from body in one pass, returning internal page links
// and any RSS/Atom feed URLs discovered via <link rel="alternate">.
func Extract(base *url.URL, body io.Reader) Extracted {
	seenLinks := make(map[string]struct{})
	seenFeeds := make(map[string]struct{})
	var links, feeds []string

	z := html.NewTokenizer(body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return Extracted{Links: links, Feeds: feeds}
		case html.StartTagToken, html.SelfClosingTagToken:
			tok := z.Token()
			switch tok.Data {
			case "a":
				href := attrVal(tok, "href")
				if href == "" {
					continue
				}
				normalized := NormalizeURL(base, href)
				if normalized == "" {
					continue
				}
				parsed, err := url.Parse(normalized)
				if err != nil || !IsSameDomain(base, parsed) {
					continue
				}
				if _, ok := seenLinks[normalized]; !ok {
					seenLinks[normalized] = struct{}{}
					links = append(links, normalized)
				}
			case "link":
				rel := attrVal(tok, "rel")
				typ := attrVal(tok, "type")
				if rel != "alternate" {
					continue
				}
				if typ != "application/rss+xml" && typ != "application/atom+xml" {
					continue
				}
				href := attrVal(tok, "href")
				if href == "" {
					continue
				}
				feedURL := NormalizeURL(base, href)
				if feedURL == "" {
					continue
				}
				if _, ok := seenFeeds[feedURL]; !ok {
					seenFeeds[feedURL] = struct{}{}
					feeds = append(feeds, feedURL)
				}
			}
		}
	}
}

// NormalizeURL resolves href against base and returns a clean absolute URL string.
// Returns "" on error or if the result is not http/https.
func NormalizeURL(base *url.URL, href string) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	resolved := base.ResolveReference(ref)
	resolved.Fragment = ""
	scheme := strings.ToLower(resolved.Scheme)
	if scheme != "http" && scheme != "https" {
		return ""
	}
	return resolved.String()
}

// IsSameDomain returns true if target has the same host as base (ignoring www prefix).
func IsSameDomain(base *url.URL, target *url.URL) bool {
	return normalizeHost(base.Host) == normalizeHost(target.Host)
}

func normalizeHost(h string) string {
	h = strings.ToLower(h)
	h = strings.TrimPrefix(h, "www.")
	if idx := strings.LastIndex(h, ":"); idx != -1 {
		h = h[:idx]
	}
	return h
}

func attrVal(tok html.Token, key string) string {
	for _, a := range tok.Attr {
		if a.Key == key {
			return strings.TrimSpace(a.Val)
		}
	}
	return ""
}
