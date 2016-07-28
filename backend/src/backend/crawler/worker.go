// Worker is working unit of crawler. There will be a worker for each different
// host. Worker has a TTL (time to live) that guarantee there will not be idle
// workers.
//
// Code is based on: https://github.com/PuerkitoBio/gocrawl/blob/master/worker.go
package crawler

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
)

type Options struct {
	UserAgent string
}

type Worker struct {
	id uint32

	// == channels
	// incomming channel is where the jobs coming in
	incoming popChannel

	// stop channel is where the worker receives its stop signal
	stop chan int

	// enqueue
	enqueue chan<- []*URLContext

	opts *Options
}

var (
	ErrEnqueueRedirect = errors.New("redirection not followed")
)

var HttpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return ErrEnqueueRedirect
	},
}

func (w *Worker) run() {
	defer func() {
		glog.Infof("Worker#%d done.", w.id)
	}()

	for {

		select {
		case <-w.stop:
			// clean up and exit
			glog.Info("Stop signal received.")
			return
		case jobs := <-w.incoming:
			for _, ctx := range jobs {
				w.crawl(ctx)
			}
		}
	}
}

// crawl function calls fetch to fetch remote web page then process the data
func (w *Worker) crawl(target *URLContext) {
	glog.Infof("Crawling URL: %s", target.normalizedURL.String())
	if res, ok := w.fetchURL(target); ok {
		// handle fetched web page
		defer res.Body.Close()

		// success
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			w.visitURL(target, res)
		} else {
			// Error
			glog.Errorf("Error status code for %s: %s", target.normalizedURL, res.Status)
		}
	}
}

func (w *Worker) fetchURL(target *URLContext) (res *http.Response, ok bool) {
	glog.Infof("#%d Fetching URL: %s", w.id, target.normalizedURL.String())

	var err error

	if res, err = w._fetch(target); err != nil {
		if ue, ok := err.(*url.Error); ok {
			// We do not let http client to handle redirection.
			// Manually handling redirection would make sure all requests
			// are following crawler's policy
			if ue.Err == ErrEnqueueRedirect {
				w.enqueueSingle(ue.URL, target)
				glog.Warningf("Enqueuing redirection: %s", ue.URL)
			}
		}

		glog.Errorf("#%d Error while fetching %s: %s", w.id, target.normalizedURL, err)

		return nil, false
	}

	ok = true

	return
}

func (w *Worker) _fetch(target *URLContext) (*http.Response, error) {
	time.Sleep(1 * time.Second)
	req, err := http.NewRequest("GET", target.url.String(), nil)
	if err != nil {
		return nil, err
	}

	if target.sourceURL != nil {
		req.Header.Add("Referer", target.sourceURL.String())
	}

	req.Header.Add("User-Agent", w.opts.UserAgent)
	return HttpClient.Do(req)
}

func (w *Worker) visitURL(target *URLContext, res *http.Response) interface{} {
	var doc *goquery.Document

	if body, err := ioutil.ReadAll(res.Body); err != nil {
		glog.Errorf("Error reading body %s: %s", target.url, err)
		return nil
	} else if node, err := html.Parse(bytes.NewBuffer(body)); err != nil {
		glog.Errorf("Error parsing %s: %s", target.url, err)
		return nil
	} else {
		doc = goquery.NewDocumentFromNode(node)
		doc.Url = target.url
	}

	// res.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if doc != nil {
		links := w.processLinks(target, doc)
		glog.Infof("Sending to enqueue, length=%d", len(links))
		w.enqueue <- urlsToURLContexts(links, target.url)
	}

	w.visited(target)

	return doc
}

func (w *Worker) processLinks(target *URLContext, doc *goquery.Document) (result []*url.URL) {
	// <base> html tag for relative URLs.
	baseUrl, _ := doc.Find("base[href]").Attr("href")

	urls := doc.Find("a[href]").Map(func(_ int, s *goquery.Selection) string {
		val, _ := s.Attr("href")

		// nofollow
		// https://en.wikipedia.org/wiki/Nofollow
		if rel, _ := s.Attr("rel"); strings.Contains(rel, "nofollow") {
			return ""
		}

		if baseUrl != "" {
			val = handleBaseTag(target, baseUrl, val)
		}

		return val
	})

	for _, s := range urls {
		if len(s) > 0 && !strings.HasPrefix(s, "#") {
			if parsed, err := url.Parse(s); err == nil {
				parsed = doc.Url.ResolveReference(parsed)
				result = append(result, parsed)
			} else {
				glog.Warningf("ignored on unparsable policy %s: %s", s, err)
			}
		}
	}

	return
}

func (w *Worker) visited(target *URLContext) {
}

func (w *Worker) enqueueSingle(raw string, source *URLContext) {
	ctx, err := stringToURLContext(raw, source.url)

	if err != nil {
		return
	}

	w.enqueue <- []*URLContext{ctx}
}

func (w *Worker) push(link *URLContext) {
	w.incoming.stack(link)
}

func handleBaseTag(root *URLContext, baseHref string, aHref string) string {
	resolvedBase, err := root.url.Parse(baseHref)
	if err != nil {
		return ""
	}

	parsedURL, err := url.Parse(aHref)
	if err != nil {
		return ""
	}
	// If a[href] starts with a /, it overrides the base[href]
	if parsedURL.Host == "" && !strings.HasPrefix(aHref, "/") {
		aHref = path.Join(resolvedBase.Path, aHref)
	}

	resolvedURL, err := resolvedBase.Parse(aHref)
	if err != nil {
		return ""
	}
	return resolvedURL.String()
}
