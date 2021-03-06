package crawler

import (
	"backend/redis"
	"backend/utils"
	"bytes"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	redigo "github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (compatible; Googlebot/2.1)"
)

const (
	visitedPrefix      = "visited"
	visitedPrefixSplit = len(visitedPrefix) + 3
	// (unit: second) 7200 seconds = 2 hour
	visitedExpireTime = 86400
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
	conn := redis.GetQueueConn()
	defer redis.ReturnQueueConn(conn)

	for {
		key := redis.BuildKey(crawlQueue, "%d", c.id)

		reply, err := redigo.String(conn.Do("BRPOPLPUSH", crawlQueue, key, 10))

		if reply == "" || err != nil {
			// no work received
			time.Sleep(1 * time.Second)
			continue
		}

		ctx := deserializeURLContext(reply)

		if c.hasVisited(conn, ctx) || strings.HasPrefix(ctx.NormalizedURL().Host, "en.m") {
			// visited, drop task
			conn.Do("RPOP", key)
			continue
		}

		// dispatch
		dest := hash(ctx.NormalizedURL().Host) % c.maxWorkers
		destKey := redis.BuildKey(workerQueue, "%s", c.workers[dest].id)

		robot, err := redigo.Bool(conn.Do("EXISTS", redis.BuildKey(RobotsTxtPrefix, "%s", ctx.NormalizedURL().Host)))

		if err != nil {
			glog.Errorf("Error while EXISTS robotsfile: %s", err)
			robot = false
		}

		if !robot {
			// need to fetch robot.txt first
			robotCtx, err := ctx.getRobotsURLCtx()

			_, err = conn.Do("LPUSH", destKey, robotCtx.serialize())

			if err != nil {
				glog.Errorf("Error while LPUSH robotsfile: %s", err)
				continue
			}
		}

		_, err = conn.Do("RPOPLPUSH", key, destKey)

		if err != nil {
			glog.Errorf("Error while RPOPLPUSH: %s", err)
			continue
		}

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

	conn := redis.GetQueueConn()
	defer redis.ReturnQueueConn(conn)

	conn.Do("LPUSH", result...)

	glog.Infof("Enqueued %d links", count)
}

func (c *Crawler) hasVisited(conn redis.ResourceConn, link *URLContext) bool {
	key := redis.BuildKey(visitedPrefix, "%s", link.NormalizedHash())

	lastTs, err := redigo.Int64(conn.Do("HGET", key[:visitedPrefixSplit], key[visitedPrefixSplit:]))

	if err != nil && err != redigo.ErrNil {
		glog.Errorf("Error when redis EXISTS: %s", err)
		return false
	} else if err == redigo.ErrNil {
		return false
	}

	now := time.Now().Unix()

	return now < lastTs+visitedExpireTime
}

func (c *Crawler) setVisited(conn redis.ResourceConn, link *URLContext) {
	key := redis.BuildKey(visitedPrefix, "%s", link.NormalizedHash())

	_, err := conn.Do("HSET", key[:visitedPrefixSplit], key[visitedPrefixSplit:], time.Now().Unix())

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
