package utils

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"

	redis "github.com/garyburd/redigo/redis"
)

var pool *redis.Pool

func OpenRedisPool() redis.Conn {
	if pool != nil {
		return pool.Get()
	}

	mutex := &sync.Mutex{}
	mutex.Lock()
	InitCache()
	defer mutex.Unlock()

	return pool.Get()
}

func initConn() (redis.Conn, error) {
	host, port := redisConfig()
	addr := fmt.Sprintf("%s:%d", host, port)
	c, err := redis.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return c, err
}

func InitCache() {
	pool = redis.NewPool(initConn, 2)
	conn := OpenRedisPool()
	defer conn.Close()
	pong, err := conn.Do("ping")
	if err != nil {
		log.Panic("got err", err)
		panic(-1)
	}
	log.Println("reach cache server ", pong)
}

func DestroyCache() {
	log.Println("destroying Cache")
	if pool != nil {
		pool.Close()
		log.Println("cache was closed")
	}
}

func redisConfig() (string, int) {
	addr := os.Getenv("REDIS_HOST")
	port, _ := strconv.Atoi(os.Getenv("REDIS_PORT"))

	return addr, port
}
