// Worker is working unit of crawler. There will be a worker for each different
// host. Worker has a TTL (time to live) that guarantee there will not be idle
// workers.
//
// Code is based on: https://github.com/PuerkitoBio/gocrawl/blob/master/worker.go
package crawler

import (
	"backend/redis"
	"backend/utils"
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
	redigo "github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
)

const (
	CrawlDelayPrefix  = "CrawlDelay"
	DefaultCrawlDelay = 1
	LastCrawlPrefix   = "LastCrawl"
)

const (
	workerQueue = "workerQueue"
)

type Options struct {
	UserAgent string
}

type Worker struct {
	id string

	// incomming channel is where the jobs coming in
	incoming utils.PopChannel

	outgoing *utils.PopChannel

	// stop channel is where the worker receives its stop signal
	stop chan int

	// enqueue is where the worker pushs links retrieved from current page
	// to coordinator
	enqueue chan<- []*url.URL

	opts *Options
}

var (
	ErrEnqueueRedirect = errors.New("Redirection not followed")
)

var HttpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return ErrEnqueueRedirect
	},
}

func (w *Worker) run() {
	defer func() {
		glog.Infof("Worker#%s done.", w.id)
	}()

	for {
		conn := redis.GetConn()

		key := redis.BuildKey(workerQueue, "%s", w.id)
		workKey := redis.BuildKey(workerQueue, "%s:working", w.id)

		reply, err := redigo.String(conn.Do("RPOP", workKey))

		if err != nil {
			// no current working
			glog.Infof("Reading from %s", key)
			reply, err = redigo.String(conn.Do("BRPOPLPUSH", key, workKey, 10))

			if err != nil {
				// no work received
				glog.Errorf("Error: %s", err)
				conn.Close()
				time.Sleep(1 * time.Second)
				continue
			}
		}

		// release conn to redis while crawling
		conn.Close()

		ctx := deserializeURLContext(reply)

		w.crawl(ctx)

		conn = redis.GetConn()
		defer conn.Close()

		// Work done.
		conn.Do("RPOP", workKey)
	}
}

// crawl function calls fetch to fetch remote web page then process the data
func (w *Worker) crawl(target *URLContext) {
	//glog.Infof("Crawling URL: %s", target.normalizedURL.String())
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
	var err error

	if res, err = w._fetch(target); err != nil {
		slient := false

		if ue, ok := err.(*url.Error); ok {
			// We do not let http client to handle redirection.
			// Manually handling redirection would make sure all requests
			// are following crawler's policy
			if ue.Err == ErrEnqueueRedirect {
				// CONCERN: ue.URL might be relative? Need confirm
				w.enqueueSingleString(ue.URL, target)
				glog.Warningf("Enqueuing redirection: %s", ue.URL)
				slient = true
			}
		}

		if !slient {
			glog.Errorf("#%s Error while fetching %s: %s", w.id, target.normalizedURL, err)
		}

		return nil, false
	}

	ok = true

	return
}

func (w *Worker) checkCrawlFrequency(target *URLContext) int64 {
	conn := redis.GetConn()
	defer conn.Close()

	var key string

	key = redis.BuildKey(CrawlDelayPrefix, "%s", target.NormalizedURL().Host)
	delay, err := redigo.Int64(conn.Do("GET", key))

	if err != nil && err != redigo.ErrNil {
		glog.Errorf("Error while retrieving redis key %s, %s", key, err)
		return 0
	} else if err == redigo.ErrNil {
		delay = 1
	}

	key = redis.BuildKey(LastCrawlPrefix, "%s", target.NormalizedURL().Host)
	last, err := redigo.Int64(conn.Do("GET", key))

	if err != nil && err != redigo.ErrNil {
		glog.Errorf("Error while retrieving redis key %s, %s", key, err)
		return 0
	} else if err == redigo.ErrNil {
		return 0
	}

	// UNIX ts, unit: second
	current := time.Now().Unix()

	//glog.Infof("current: %d last: %d delay: %d", current, last, delay)

	return (last + delay) - current
}

func (w *Worker) markCrawlTime(target *URLContext) {
	conn := redis.GetConn()

	key := redis.BuildKey(LastCrawlPrefix, "%s", target.NormalizedURL().Host)
	_, err := conn.Do("SET", key, time.Now().Unix())

	if err != nil {
		glog.Errorf("Error while setting redis key %s, %s", key, err)
	}
}

func (w *Worker) _fetch(target *URLContext) (*http.Response, error) {
	if diff := w.checkCrawlFrequency(target); diff > 0 {
		wait := time.After(time.Duration(diff) * time.Second)
		glog.Infof("Wait for %d seconds", diff)
		<-wait
	}

	defer w.markCrawlTime(target)

	glog.Infof("#%s Fetching URL: %s", w.id, target.normalizedURL.String())

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

func (w *Worker) visitURL(target *URLContext, res *http.Response) {
	var doc *goquery.Document

	if body, err := ioutil.ReadAll(res.Body); err != nil {
		glog.Errorf("Error reading body %s: %s", target.url, err)
		return
	} else if node, err := html.Parse(bytes.NewBuffer(body)); err != nil {
		glog.Errorf("Error parsing %s: %s", target.url, err)
		return
	} else {
		doc = goquery.NewDocumentFromNode(node)
		doc.Url = target.url
	}

	// res.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if doc != nil {
		// skip page explicity states non-English
		if lang, exist := doc.Find("html").Attr("lang"); exist && !strings.HasPrefix(lang, "en") {
			return
		}
		links := w.processLinks(target, doc)
		glog.Infof("Sending to enqueue, length=%d", len(links))
		w.enqueue <- links
	}

	w.visited(target)

	w.sendResponse(target, doc)

	return
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

func (w *Worker) enqueueSingleString(raw string, source *URLContext) {
	ctx, err := url.Parse(raw)

	if err != nil {
		return
	}

	w.enqueueSingle(ctx, source.URL())
}

func (w *Worker) enqueueSingle(u, src *url.URL) {
	w.enqueue <- []*url.URL{src, u}
}

func (w *Worker) push(link *URLContext) {
	w.incoming.Stack(link)
}

func (w *Worker) sendResponse(link *URLContext, document *goquery.Document) {
	resp := &CrawlResponse{
		Document: document,
		Link:     link,
	}

	w.outgoing.Stack(resp)
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
