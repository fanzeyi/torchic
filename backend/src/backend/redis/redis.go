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
var index *pools.ResourcePool
var queue *pools.ResourcePool

func initRedis(host, port, password string, db int) *pools.ResourcePool {
	glog.Info("Initializing redis connection pool")
	glog.Infof("Redis server: %s:%s", host, port)

	options := make([]redis.DialOption, 0)

	if password != "" {
		options = append(options, redis.DialPassword(password))
	}

	if db != -1 {
		options = append(options, redis.DialDatabase(db))
	}

	return pools.NewResourcePool(func() (pools.Resource, error) {
		c, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", host, port), options...)
		return ResourceConn{c}, err
	}, 15, 25, 10*time.Second)
}

func Init(host, port, password string, db int) {
	index = initRedis(host, port, password, db)
}

func InitQueue(host, port, password string, db int) {
	queue = initRedis(host, port, password, db)
}

func getConn(pool *pools.ResourcePool) ResourceConn {
	r, err := pool.Get(context.Background())
	if err != nil {
		glog.Fatalf("Error when getting connection %s", err)
	}

	return r.(ResourceConn)
}

func GetConn() ResourceConn {
	return getConn(index)
}

func GetQueueConn() ResourceConn {
	return getConn(queue)
}

func ReturnConn(conn ResourceConn) {
	index.Put(conn)
}

func ReturnQueueConn(conn ResourceConn) {
	queue.Put(conn)
}

func Stats() string {
	return index.StatsJSON()
}

func BuildKey(prefix, format string, a ...interface{}) string {
	return fmt.Sprintf(fmt.Sprintf("%s:%s", prefix, format), a...)
}
