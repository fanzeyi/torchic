package redis

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
	"github.com/youtube/vitess/go/pools"
)

type ResourceConn struct {
	redis.Conn
}

func (r ResourceConn) Close() {
	r.Conn.Close()
}

// pool is a global pool of Redis connections.
// See https://godoc.org/github.com/garyburd/redigo/redis#Pool
var pool *pools.ResourcePool

func Init(host, port, password string) {
	glog.Info("Initializing redis connection pool")
	glog.Infof("Redis server: %s:%s", host, port)

	pool = pools.NewResourcePool(func() (pools.Resource, error) {
		c, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
		return ResourceConn{c}, err
	}, 20, 40, 3*time.Second)
}

func GetConn() ResourceConn {
	r, err := pool.Get(context.Background())
	if err != nil {
		glog.Fatalf("Error when getting connection %s", err)
	}

	return r.(ResourceConn)
}

func ReturnConn(conn ResourceConn) {
	pool.Put(conn)
}

func Stats() string {
	return pool.StatsJSON()
}

func BuildKey(prefix, format string, a ...interface{}) string {
	return fmt.Sprintf(fmt.Sprintf("%s:%s", prefix, format), a...)
}
