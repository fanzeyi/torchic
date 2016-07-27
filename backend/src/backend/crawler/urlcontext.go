package crawler

import (
	"net/url"
	"strings"

	"github.com/PuerkitoBio/purell"
)

const (
	robotsTxtPath = "/robots.txt"
)

const (
	URLNormalizationFlags = purell.FlagsAllGreedy
)

type URLContext struct {
	url                 *url.URL
	normalizedURL       *url.URL
	sourceURL           *url.URL
	normalizedSourceURL *url.URL
}

// URL returns the URL.
func (uc *URLContext) URL() *url.URL {
	return uc.url
}

// NormalizedURL returns the normalized URL (using Options.URLNormalizationFlags)
// of the URL.
func (uc *URLContext) NormalizedURL() *url.URL {
	return uc.normalizedURL
}

// SourceURL returns the source URL, if any (the URL that enqueued this
// URL).
func (uc *URLContext) SourceURL() *url.URL {
	return uc.sourceURL
}

// NormalizedSourceURL returns the normalized form of the source URL,
// if any (using Options.URLNormalizationFlags).
func (uc *URLContext) NormalizedSourceURL() *url.URL {
	return uc.normalizedSourceURL
}

// IsRobotsURL indicates if the URL is a robots.txt URL.
func (uc *URLContext) IsRobotsURL() bool {
	return isRobotsURL(uc.normalizedURL)
}

func isRobotsURL(u *url.URL) bool {
	if u == nil {
		return false
	}
	return strings.ToLower(u.Path) == robotsTxtPath
}

func urlToURLContext(u, src *url.URL) *URLContext {
	var rawSrc *url.URL

	rawU := *u
	purell.NormalizeURL(u, URLNormalizationFlags)
	if src != nil {
		rawSrc = &url.URL{}
		*rawSrc = *src
		purell.NormalizeURL(src, URLNormalizationFlags)
	}

	return &URLContext{
		url:                 &rawU,
		normalizedURL:       u,
		sourceURL:           rawSrc,
		normalizedSourceURL: src,
	}
}

func stringToURLContext(str string, src *url.URL) (*URLContext, error) {
	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}
	return urlToURLContext(u, src), nil
}

func urlsToURLContexts(urls []*url.URL, source *url.URL) (result []*URLContext) {
	for _, u := range urls {
		result = append(result, urlToURLContext(u, source))
	}

	return
}
