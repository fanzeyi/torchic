package crawler

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/purell"
	"github.com/golang/glog"
)

const (
	robotsTxtPath = "/robots.txt"
)

const (
	URLNormalizationFlags = purell.FlagsUsuallySafeGreedy | purell.FlagSortQuery | purell.FlagRemoveFragment | purell.FlagRemoveDirectoryIndex | purell.FlagForceHTTP | purell.FlagRemoveDuplicateSlashes | purell.FlagRemoveUnnecessaryHostDots | purell.FlagRemoveEmptyPortSeparator
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

func (uc *URLContext) getRobotsURLCtx() (*URLContext, error) {
	robURL, err := uc.normalizedURL.Parse(robotsTxtPath)
	if err != nil {
		return nil, err
	}
	return &URLContext{
		robURL,
		robURL,       // Normalized is same as raw
		uc.sourceURL, // Source and normalized source is same as for current context
		uc.normalizedSourceURL,
	}, nil
}

func (uc *URLContext) serialize() string {
	return serializeUrl(uc.url, uc.sourceURL)
}

func (uc *URLContext) hash(u *url.URL) string {
	hasher := md5.New()
	hasher.Write([]byte(u.String()))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (uc *URLContext) Hash() string {
	return uc.hash(uc.url)
}

func (uc *URLContext) NormalizedHash() string {
	return uc.hash(uc.normalizedURL)
}

func (uc *URLContext) CompareTo(ctx *URLContext) bool {
	return uc.normalizedURL.String() == ctx.normalizedURL.String()
}

func deserializeURLContext(data string) *URLContext {
	parts := strings.Split(data, ":")
	link, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		glog.Errorf("Deserialize failed: %s", err)
		return nil
	}

	src, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		glog.Errorf("Deserialize failed: %s", err)
		return nil
	}

	u, err := url.Parse(string(link))
	if err != nil {
		glog.Errorf("Deserialize failed: %s", err)
		return nil
	}

	srcUrl, err := url.Parse(string(src))
	if err != nil {
		glog.Errorf("Deserialize failed: %s", err)
		return urlToURLContext(u, nil)
	}

	return urlToURLContext(u, srcUrl)
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
