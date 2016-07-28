package crawler

import (
	"backend/redis"
	"fmt"

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

type Crawler struct {
	enqueue chan []*URLContext

	// map [host] -> worker
	workers map[string]*Worker
	visited map[string]bool
}

func (c *Crawler) Run() {
	c.init()
	go c.run()
}

func (c *Crawler) init() {
	c.enqueue = make(chan []*URLContext, 10)
	c.workers = make(map[string]*Worker)
}

// Crawler runloop
func (c *Crawler) run() {
	for {
		select {
		case links := <-c.enqueue:
			c.enqueueUrls(links)
		}
	}
}

func (c *Crawler) enqueueUrls(links []*URLContext) {
	glog.Info("received from enqueue")
	glog.Infof("received %v", links)

	for _, link := range links {
		if c.hasVisited(link) {
			glog.Infof("Ignored on visited: %s", link.normalizedURL)
			continue
		}
		c.dispatch(link)
	}

	glog.Info("Dispatch end")
}

func (c *Crawler) hasVisited(link *URLContext) bool {
	conn := redis.GetConn()
	defer conn.Close()

	res, err := conn.Do("GET", fmt.Sprintf("%s:%s", visitedPrefix, link.normalizedURL.String()))

	if err != nil {
		glog.Errorf("Error when redis GET: %s", err)
		return false
	}

	return res != nil
}

func (c *Crawler) setVisited(link *URLContext) {
	conn := redis.GetConn()
	defer conn.Close()

	_, err := conn.Do("SETEX", fmt.Sprintf("%s:%s", visitedPrefix, link.normalizedURL.String()), visitedExpireTime, true)

	if err != nil {
		glog.Errorf("Error when redis SETNX: %s", err)
	}
}

func (c *Crawler) dispatch(link *URLContext) {
	//glog.Infof("Dispatching %s", link.normalizedURL.String())
	worker, ok := c.workers[link.NormalizedURL().Host]

	if !ok {
		worker = c.launchWorker(link)
	}

	worker.push(link)

	c.setVisited(link)
}

func (c *Crawler) launchWorker(link *URLContext) *Worker {
	glog.Infof("Launching worker for %s", link.normalizedURL.Host)
	//incoming := make(chan *URLContext)

	w := &Worker{
		host:     link.normalizedURL.Host,
		incoming: newPopChannel(),
		enqueue:  c.enqueue,
		opts: &Options{
			UserAgent: DefaultUserAgent,
		},
	}

	c.workers[link.normalizedURL.Host] = w

	go w.run()

	return w
}

func (c *Crawler) Push(link string) {
	ctx, _ := stringToURLContext(link, nil)
	c.enqueue <- []*URLContext{ctx}
}
