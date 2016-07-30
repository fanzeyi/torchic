package crawler

import (
	"backend/redis"
	"backend/utils"
	"bytes"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
)

const (
	visitedPrefix = "visited"
	// (unit: second) 7200 seconds = 2 hour
	visitedExpireTime = 7200
)

const (
	crawlQueue = "crawlQueue"
)

type CrawlResponse struct {
	Document *goquery.Document
	Link     *URLContext
}

type Crawler struct {
	id       uint32
	enqueue  chan []*url.URL
	outgoing *utils.PopChannel

	visited    map[string]bool
	workers    []*Worker
	maxWorkers uint32
}

func NewCrawler(id, numOfWorkers uint32, outgoing *utils.PopChannel) *Crawler {
	crawler := new(Crawler)
	crawler.id = id
	crawler.outgoing = outgoing
	crawler.init(numOfWorkers)

	return crawler
}

func (c *Crawler) init(numOfWorkers uint32) {
	c.enqueue = make(chan []*url.URL)
	c.workers = make([]*Worker, numOfWorkers)
	c.maxWorkers = numOfWorkers

	var i uint32

	for i = 0; i < numOfWorkers; i++ {
		c.workers[i] = c.launchWorker(i)
	}
}

// Crawler runloop
func (c *Crawler) Run() {
	go c.coordinatorRun()
	go c.pusherRun()
}

func (c *Crawler) coordinatorRun() {
	conn := redis.GetConn()
	defer redis.ReturnConn(conn)

	for {
		key := redis.BuildKey(crawlQueue, "%d", c.id)
		reply, err := redigo.String(conn.Do("BRPOPLPUSH", crawlQueue, key, 10))

		if reply == "" || err != nil {
			// no work received
			time.Sleep(1 * time.Second)
			continue
		}

		ctx := deserializeURLContext(reply)

		if c.hasVisited(conn, ctx) {
			// visited, drop task
			conn.Do("RPOP", key)
			continue
		}

		// dispatch
		dest := hash(ctx.NormalizedURL().Host) % c.maxWorkers

		destKey := redis.BuildKey(workerQueue, "%s", c.workers[dest].id)
		_, err = conn.Do("RPOPLPUSH", key, destKey)

		// set visited
		c.setVisited(conn, ctx)
	}
}

func (c *Crawler) pusherRun() {
	for {
		select {
		case links := <-c.enqueue:
			c.enqueueUrls(links)
		}
	}
}

func serializeUrl(link, source *url.URL) string {
	var buffer bytes.Buffer
	buffer.WriteString(base64.StdEncoding.EncodeToString([]byte(link.String())))
	buffer.WriteRune(':')

	if source != nil {
		buffer.WriteString(base64.StdEncoding.EncodeToString([]byte(source.String())))
	}

	return buffer.String()
}

func (c *Crawler) enqueueUrls(links []*url.URL) {
	count := 0

	var source *url.URL

	if len(links) > 1 {
		source, links = links[0], links[1:]
	}

	result := make([]interface{}, 0)

	result = append(result, crawlQueue)

	for _, link := range links {
		result = append(result, serializeUrl(link, source))
		count += 1
	}

	conn := redis.GetConn()
	defer redis.ReturnConn(conn)

	conn.Do("LPUSH", result...)

	glog.Infof("Enqueued %d links", count)
}

func (c *Crawler) hasVisited(conn redis.ResourceConn, link *URLContext) bool {
	res, err := conn.Do("GET", fmt.Sprintf("%s:%s", visitedPrefix, link.normalizedURL.String()))

	if err != nil {
		glog.Errorf("Error when redis GET: %s", err)
		return false
	}

	return res != nil // || link.NormalizedURL().Host != "en.wikipedia.org"
}

func (c *Crawler) setVisited(conn redis.ResourceConn, link *URLContext) {
	_, err := conn.Do("SETEX", fmt.Sprintf("%s:%s", visitedPrefix, link.normalizedURL.String()), visitedExpireTime, true)

	if err != nil {
		glog.Errorf("Error when redis SETNX: %s", err)
	}
}

func hash(raw string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(raw))
	return h.Sum32()
}

func (c *Crawler) launchWorker(id uint32) *Worker {
	w := &Worker{
		id:       fmt.Sprintf("%d:%d", c.id, id),
		incoming: utils.NewPopChannel(),
		outgoing: c.outgoing,
		enqueue:  c.enqueue,
		opts: &Options{
			UserAgent: DefaultUserAgent,
		},
	}

	glog.Infof("Launching worker #%s", w.id)

	go w.run()

	return w
}

func (c *Crawler) Push(link string) {
	ctx, _ := url.Parse(link)
	c.enqueue <- []*url.URL{ctx}
}
