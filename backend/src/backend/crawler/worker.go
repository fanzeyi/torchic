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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/goquery"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
	"github.com/temoto/robotstxt"
)

const (
	CrawlDelayPrefix  = "CrawlDelay"
	DefaultCrawlDelay = 1
	LastCrawlPrefix   = "LastCrawl"
	RobotsTxtPrefix   = "RobotsTxt"
)

const (
	workerQueue = "workerQueue"
)

type Options struct {
	UserAgent string
}

type Worker struct {
	id string

	outgoing *utils.PopChannel

	// stop channel is where the worker receives its stop signal
	stop chan int

	// enqueue is where the worker pushs links retrieved from current page
	// to coordinator
	enqueue chan<- []*url.URL

	opts *Options

	robots map[string]*robotstxt.Group
}

var (
	ErrEnqueueRedirect = errors.New("Redirection not followed")
)

var HttpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if isRobotsURL(req.URL) {
			if len(via) >= 10 {
				return errors.New("stopped after 10 redirects")
			}
			if len(via) > 0 {
				req.Header.Set("User-Agent", via[0].Header.Get("User-Agent"))
			}
			return nil
		}
		glog.Warningf("Redirection detected: %v", req)
		return ErrEnqueueRedirect
	},
	Timeout: 10 * time.Second,
}

func (w *Worker) run() {
	w.robots = make(map[string]*robotstxt.Group)

	defer func() {
		glog.Infof("Worker#%s done.", w.id)
	}()

	for {
		conn := redis.GetQueueConn()

		key := redis.BuildKey(workerQueue, "%s", w.id)
		workKey := redis.BuildKey(workerQueue, "%s:working", w.id)

		reply, err := redigo.String(conn.Do("RPOP", workKey))

		if err != nil {
			// no current working
			if err != redigo.ErrNil {
				glog.Errorf("[%s] Error while getting work: %s", w.id, err)
				redis.ReturnQueueConn(conn)

				continue
			}
			reply, err = redigo.String(conn.Do("BRPOPLPUSH", key, workKey, 10))

			if err != nil {
				// no work received
				if err != redigo.ErrNil {
					glog.Errorf("[%s] Error while getting work: %s", w.id, err)
					redis.ReturnQueueConn(conn)
					continue
				}
				redis.ReturnQueueConn(conn)
				time.Sleep(1 * time.Second)
				continue
			}
		}

		// release conn to redis while crawling
		redis.ReturnQueueConn(conn)

		ctx := deserializeURLContext(reply)

		if ctx.IsRobotsURL() {
			w.requestRobotsTxt(ctx)
		} else if w.isAllowedPerRobotsPolicies(ctx) {
			w.requestURL(ctx)
		} else {
			glog.Infof("Disallowed by robots policy: %s", ctx.URL().String())
		}

		conn = redis.GetQueueConn()

		// Work done.
		conn.Do("RPOP", workKey)

		redis.ReturnQueueConn(conn)
	}
}

func (w *Worker) isAllowedPerRobotsPolicies(ctx *URLContext) bool {
	group, ok := w.robots[ctx.NormalizedURL().Host]

	if !ok {
		w.loadRobotsTxtFromRedis(ctx)

		group, ok = w.robots[ctx.NormalizedURL().Host]

		if !ok {
			// No robotstxt found. Always true
			return true
		}
	}

	return group.Test(ctx.NormalizedURL().Path)
}

func (w *Worker) requestRobotsTxt(ctx *URLContext) {
	if res, ok := w.fetchURL(ctx); ok {
		defer res.Body.Close()

		buf, err := ioutil.ReadAll(res.Body)

		if err != nil {
			glog.Errorf("Error while reading robots.txt body: %s", err)
			return
		}

		robot, err := robotstxt.FromStatusAndBytes(res.StatusCode, buf)

		if err != nil {
			glog.Errorf("Error while parsing robots file: %s", err)
			return
		}

		w.loadRobotsTxt(ctx, robot)
		w.saveRobotsTxt(ctx, res.StatusCode, buf)
	} else {
		// special case. the robots.txt request is failed because of any reason.
		// treat it as allow all
		w.saveRobotsTxt(ctx, 400, []byte{})
	}
}

func (w *Worker) loadRobotsTxt(ctx *URLContext, robot *robotstxt.RobotsData) {
	w.robots[ctx.NormalizedURL().Host] = robot.FindGroup(w.opts.UserAgent)
}

func (w *Worker) saveRobotsTxt(ctx *URLContext, statusCode int, data []byte) {
	conn := redis.GetConn()
	defer redis.ReturnConn(conn)

	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("%3d", statusCode))
	buf.Write(data)

	_, err := conn.Do("SET", redis.BuildKey(RobotsTxtPrefix, "%s", ctx.NormalizedURL().Host), buf.Bytes())

	if err != nil {
		glog.Errorf("%s", err)
	}
}

func (w *Worker) loadRobotsTxtFromRedis(ctx *URLContext) {
	conn := redis.GetConn()
	defer redis.ReturnConn(conn)

	res, err := redigo.Bytes(conn.Do("GET", redis.BuildKey(RobotsTxtPrefix, "%s", ctx.NormalizedURL().Host)))

	if err != nil {
		glog.Errorf("Error while getting robotstxt %s from redis: %s", ctx.NormalizedURL().Host, err)
		return
	}

	statusCode, err := strconv.Atoi(string(res[:3]))

	robot, err := robotstxt.FromStatusAndBytes(statusCode, res[3:])

	if err != nil {
		glog.Errorf("Error while parsing robotstxt %s from redis: %s", ctx.NormalizedURL().Host, err)
		return
	}

	w.loadRobotsTxt(ctx, robot)
}

// crawl function calls fetch to fetch remote web page then process the data
func (w *Worker) requestURL(target *URLContext) {
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
				glog.Warningf("[%s] Enqueuing redirection: %s", w.id, ue.URL)
				slient = true
			}
		}

		if !slient {
			glog.Errorf("#%s Error while fetching %s: %s", w.id, target.normalizedURL, err)
		}

		ok = false

		return
	}

	ok = true

	return
}

func (w *Worker) checkCrawlFrequency(target *URLContext) int64 {
	conn := redis.GetConn()
	defer redis.ReturnConn(conn)

	var key string

	group, ok := w.robots[target.NormalizedURL().Host]

	var delay int64

	if ok {
		delay = int64(group.CrawlDelay / time.Second)
	}

	if delay == 0 {
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
	defer redis.ReturnConn(conn)

	key := redis.BuildKey(LastCrawlPrefix, "%s", target.NormalizedURL().Host)
	_, err := conn.Do("SET", key, time.Now().Unix())

	if err != nil {
		glog.Errorf("Error while setting redis key %s, %s", key, err)
	}
}

func (w *Worker) _fetch(target *URLContext) (*http.Response, error) {
	if diff := w.checkCrawlFrequency(target); diff > 0 {
		wait := time.After(time.Duration(diff) * time.Second)
		glog.Infof("[%s] Wait %s for %d seconds", w.id, target.NormalizedURL().Host, diff)
		<-wait
	}

	defer w.markCrawlTime(target)

	glog.Infof("[%s] Fetching: %s", w.id, target.url.String())

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

	contentType := res.Header.Get("Content-Type")

	// skip non text pages
	if !strings.HasPrefix(contentType, "text") {
		return
	}

	if body, err := ioutil.ReadAll(res.Body); err != nil {
		glog.Errorf("[%s] Error reading body %s: %s", w.id, target.url, err)
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

		robotMeta, robotMetaExist := doc.Find("meta[name=\"robots\"]").Attr("content")

		if robotMetaExist && strings.Index(robotMeta, "noindex") != -1 {
			return
		}

		canonical, canonicalExist := doc.Find("link[rel=\"canonical\"]").Attr("href")

		if canonicalExist {

			canonicalCtx, err := stringToURLContext(canonical, nil)

			if err != nil {
				glog.Errorf("Error while converting canonical URL: %s", err)
			} else if !target.CompareTo(canonicalCtx) {
				// it is not the original page

				w.enqueueSingle(canonicalCtx.url, target.url)
				return
			}
		}

		if !(robotMetaExist && strings.Index(robotMeta, "nofollow") != -1) {
			links := w.processLinks(target, doc)
			glog.Infof("[%s] Sending to enqueue, length=%d", w.id, len(links))
			w.enqueue <- links
		}
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

		if hreflang, _ := s.Attr("hreflang"); hreflang != "" && !strings.HasPrefix(hreflang, "en") {
			// ignore non-English links
			//glog.Infof("ignored link %s.", val)
			return ""
		}

		if baseUrl != "" {
			val = handleBaseTag(target, baseUrl, val)
		}

		return val
	})

	result = append(result, target.URL())

	for _, s := range urls {
		if len(s) > 0 && !strings.HasPrefix(s, "#") {
			if parsed, err := url.Parse(s); err == nil {
				parsed = doc.Url.ResolveReference(parsed)

				if parsed.Scheme == "http" || parsed.Scheme == "https" {
					result = append(result, parsed)
				}
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
