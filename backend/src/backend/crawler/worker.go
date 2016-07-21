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
	id   string
	host string

	// == channels
	// incomming channel is where the jobs coming in
	incoming chan url.URL

	// stop channel is where the worker receives its stop signal
	stop chan int

	// enqueue
	enqueue chan<- interface{}

	timeout *time.Time

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
		glog.Infof("Worker #%s done.", w.id)
	}()

	for {

		select {
		case <-w.stop:
			// clean up and exit
			glog.Info("Stop signal received.")
			return
		case target := <-w.incoming:
			w.crawl(target)
		}
	}
}

func (w *Worker) processLinks(target url.URL, doc *goquery.Document) (result []*url.URL) {
	// <base> html tag for relative URLs
	baseUrl, _ := doc.Find("base[href]").Attr("href")

	urls := doc.Find("a[href]").Map(func(_ int, s *goquery.Selection) string {
		val, _ := s.Attr("href")

		// nofollow
		// https://en.wikipedia.org/wiki/Nofollow
		if rel, _ := s.Attr("rel"); strings.Contains(rel, "nofollow") {
			return ""
		}

		if baseUrl != "" {
			val = handleBaseTag(&target, baseUrl, val)
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

func (w *Worker) visited(target url.URL) {
}

func (w *Worker) visitURL(target url.URL, res *http.Response) interface{} {
	var doc *goquery.Document

	if body, err := ioutil.ReadAll(res.Body); err != nil {
		glog.Errorf("Error reading body %s: %s", target, err)
		return nil
	} else if node, err := html.Parse(bytes.NewBuffer(body)); err != nil {
		glog.Errorf("Error parsing %s: %s", target, err)
		return nil
	} else {
		doc = goquery.NewDocumentFromNode(node)
		doc.Url = &target
	}

	// res.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if doc != nil {
		w.processLinks(target, doc)
	}

	w.visited(target)

	return doc
}

// crawl function calls fetch to fetch remote web page then process the data
func (w *Worker) crawl(target url.URL) {
	if res, ok := w.fetch(target); ok {
		// handle fetched web page
		defer res.Body.Close()

		// success
		if res.StatusCode >= 200 && res.StatusCode < 300 {
		} else {
			// Error
			glog.Errorf("Error status code for %s: %s", target, res.Status)
		}
	}
}

func (w *Worker) _fetch(target url.URL) (*http.Response, error) {
	req, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", w.opts.UserAgent)
	return HttpClient.Do(req)
}

func (w *Worker) fetch(target url.URL) (res *http.Response, ok bool) {
	for {
		if _, err := w._fetch(target); err != nil {
			if ue, ok := err.(*url.Error); ok {
				// We do not let http client to handle redirection.
				// Manually handling redirection would make sure all requests
				// are following crawler's policy
				if ue.Err == ErrEnqueueRedirect {
					w.enqueue <- ue.URL
				}
			}

			glog.Errorf("Error while fetching %s: %s", target, err)

			return nil, false
		}

		ok = true
	}

	return
}

func handleBaseTag(root *url.URL, baseHref string, aHref string) string {
	resolvedBase, err := root.Parse(baseHref)
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
