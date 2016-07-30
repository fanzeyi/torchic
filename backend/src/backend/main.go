package main

import (
	"backend/crawler"
	"backend/index"
	"backend/redis"
	"backend/utils"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/golang/glog"
)

var id = flag.Uint("id", 1, "Running instance ID. Must be unique")

func main() {

	sigChan := make(chan os.Signal)
	go func() {
		stacktrace := make([]byte, 32768)
		for _ = range sigChan {
			length := runtime.Stack(stacktrace, true)
			fmt.Println(string(stacktrace[:length]))

			fmt.Println(redis.Stats())
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)

	exitSignal := make(chan os.Signal)
	signal.Notify(exitSignal, syscall.SIGINT, syscall.SIGTERM)

	flag.Parse()

	redis.Init(REDIS_HOST, REDIS_PORT, REDIS_PASSWORD)

	crawlRespChan := utils.NewPopChannel()

	worker := crawler.NewCrawler(uint32(*id), 10, &crawlRespChan)
	worker.Run()
	worker.Push("https://en.wikipedia.org/")

	indexer := index.NewIndexer(&crawlRespChan)
	go indexer.Run()

	<-exitSignal
	glog.Flush()
}
