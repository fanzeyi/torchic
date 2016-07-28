package redis

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/golang/glog"
)

// pool is a global pool of Redis connections.
// See https://godoc.org/github.com/garyburd/redigo/redis#Pool
var pool *redis.Pool

func Init(host, port, password string) {
	glog.Info("Initializing redis connection pool")
	glog.Infof("Redis server: %s:%s", host, port)

	pool = &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
			if err != nil {
				return nil, err
			}

			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func GetConn() redis.Conn {
	return pool.Get()
}
