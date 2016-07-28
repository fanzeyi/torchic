package main

import (
	"backend/crawler"
	"backend/redis"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/golang/glog"
)

func main() {
	sigChan := make(chan os.Signal)
	go func() {
		stacktrace := make([]byte, 8192)
		for _ = range sigChan {
			length := runtime.Stack(stacktrace, true)
			fmt.Println(string(stacktrace[:length]))
		}
	}()
	signal.Notify(sigChan, syscall.SIGQUIT)

	flag.Parse()

	glog.Info("Prepare to repel boarders")

	redis.Init(REDIS_HOST, REDIS_PORT, REDIS_PASSWORD)

	//crawler.Start(true)
	worker := crawler.Crawler{}
	worker.Run()
	worker.Push("https://catsgobark:nichijou@_.zr.is/")

	for {
		time.Sleep(100)
	}
}
