package crawler

import (
	"backend/redis"
	"backend/utils"
	"fmt"
	"hash/fnv"

	"github.com/PuerkitoBio/goquery"
	"github.com/golang/glog"
)

const (
	DefaultUserAgent = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
)

const (
	visitedPrefix = "Visited"
	// (unit: second) 7200 seconds = 2 hour
	visitedExpireTime = 7200
)

type CrawlResponse struct {
	Document *goquery.Document
	Link     *URLContext
}

type Crawler struct {
	enqueue  chan []*URLContext
	outgoing *utils.PopChannel

	visited    map[string]bool
	workers    []*Worker
	maxWorkers uint32
}

func NewCrawler(numOfWorkers uint32, outgoing *utils.PopChannel) *Crawler {
	crawler := new(Crawler)
	crawler.outgoing = outgoing
	crawler.init(numOfWorkers)

	return crawler
}

func (c *Crawler) init(numOfWorkers uint32) {
	c.enqueue = make(chan []*URLContext, 10)
	c.workers = make([]*Worker, numOfWorkers)
	c.maxWorkers = numOfWorkers

	var i uint32

	for i = 0; i < numOfWorkers; i++ {
		c.workers[i] = c.launchWorker(i)
	}
}

// Crawler runloop
func (c *Crawler) Run() {
	for {
		select {
		case links := <-c.enqueue:
			c.enqueueUrls(links)
		}
	}
}

func (c *Crawler) enqueueUrls(links []*URLContext) {
	count := 0

	for _, link := range links {
		if c.hasVisited(link) {
			//glog.Infof("Ignored on visited: %s", link.normalizedURL)
			continue
		}
		c.dispatch(link)
		count += 1
	}

	glog.Infof("Enqueue received %d links", count)
}

func (c *Crawler) hasVisited(link *URLContext) bool {
	conn := redis.GetConn()
	defer conn.Close()

	res, err := conn.Do("GET", fmt.Sprintf("%s:%s", visitedPrefix, link.normalizedURL.String()))

	if err != nil {
		glog.Errorf("Error when redis GET: %s", err)
		return false
	}

	return res != nil // || link.NormalizedURL().Host != "en.wikipedia.org"
}

func (c *Crawler) setVisited(link *URLContext) {
	conn := redis.GetConn()
	defer conn.Close()

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

func (c *Crawler) dispatch(link *URLContext) {
	//glog.Infof("Dispatching %s", link.normalizedURL.String())
	dest := hash(link.NormalizedURL().Host) % c.maxWorkers

	c.workers[dest].push(link)
	c.setVisited(link)
}

func (c *Crawler) launchWorker(no uint32) *Worker {
	glog.Infof("Launching worker #%d", no)
	//incoming := make(chan *URLContext)

	w := &Worker{
		id:       no,
		incoming: utils.NewPopChannel(),
		outgoing: c.outgoing,
		enqueue:  c.enqueue,
		opts: &Options{
			UserAgent: DefaultUserAgent,
		},
	}

	go w.run()

	return w
}

func (c *Crawler) Push(link string) {
	ctx, _ := stringToURLContext(link, nil)
	c.enqueue <- []*URLContext{ctx}
}
