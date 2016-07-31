package main

import (
	"backend/crawler"
	"backend/index"
	"backend/mysql"
	"backend/redis"
	"backend/utils"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"net/http"
	_ "net/http/pprof"

	"github.com/golang/glog"
)

var id = flag.Uint("id", 1, "Running instance ID. Must be unique")
var seed = flag.String("seed", "", "seed url to start")
var prof = flag.Bool("prof", false, "Enable profiling web server")

//var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	if *prof {
		go func() {
			glog.Infof("%s", http.ListenAndServe("localhost:6060", nil))
		}()
	}

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
	mysql.Init(MYSQL_ADDRESS, MYSQL_USERNAME, MYSQL_PASSWORD, MYSQL_DBNAME)

	crawlRespChan := utils.NewPopChannel()

	worker := crawler.NewCrawler(uint32(*id), 10, &crawlRespChan)
	worker.Run()

	if *seed != "" {
		worker.Push(*seed)
	}

	indexer := index.NewIndexer(&crawlRespChan)
	go indexer.Run()

	<-exitSignal

	glog.Flush()
}
