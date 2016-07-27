package crawler

import "github.com/golang/glog"

const (
	DefaultUserAgent = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"
)

type Crawler struct {
	enqueue chan []*URLContext

	// map [host] -> worker
	workers map[string]*Worker
	visited map[string]bool
}

//func StartWorker() {
//incoming := make(chan *url.URL)
//enqueue := make(chan []*url.URL)

//w := &Worker{
//id:       "1",
//host:     "en.wikipedia.org",
//incoming: incoming,
//enqueue:  enqueue,
//opts: &Options{
//UserAgent: "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
//},
//}

//go w.run()

//u, _ := url.Parse("https://en.wikipedia.org/wiki/SS_Washingtonian_(1913)")

//incoming <- u

//for {
//select {
//case links := <-enqueue:
//glog.Infof("Received links: %v", links)
//}
//}
//}

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
			glog.Info("received from enqueue")
			glog.Infof("received %v", links)
			for _, link := range links {
				c.dispatch(link)
			}

			glog.Info("Dispatch end")
		}

		glog.Infof("%d", len(c.enqueue))
	}
}

func (c *Crawler) dispatch(link *URLContext) {
	glog.Infof("Dispatching %s", link.normalizedURL.String())
	worker, ok := c.workers[link.NormalizedURL().Host]

	if !ok {
		worker = c.launchWorker(link)
	}

	worker.push(link)
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
