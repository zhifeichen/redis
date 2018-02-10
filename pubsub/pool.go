package pubsub

import (
  "github.com/garyburd/redigo/redis"
  "time"
)

var pool *redis.Pool

func InitPool(addr string) {
  pool = &redis.Pool{
    MaxIdle: 3,
    IdleTimeout: 240 * time.Second,
    Dial: func () (redis.Conn, error) { return redis.Dial("tcp", addr) },
  }
}

func ClosePool() error {
  return pool.Close()
}

func Pool() *redis.Pool {
  return pool
}
