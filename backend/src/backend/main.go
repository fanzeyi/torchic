package main

import (
	"backend/redis"
	"flag"

	"github.com/golang/glog"
)

func main() {
	flag.Parse()

	glog.Info("Prepare to repel boarders")

	redis.Init(REDIS_HOST, REDIS_PORT, REDIS_PASSWORD)

	//crawler.Start(true)
}
