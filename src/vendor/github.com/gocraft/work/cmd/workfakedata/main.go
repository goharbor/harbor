package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
)

var redisHostPort = flag.String("redis", ":6379", "redis hostport")
var redisNamespace = flag.String("ns", "work", "redis namespace")

func epsilonHandler(job *work.Job) error {
	fmt.Println("epsilon")
	time.Sleep(time.Second)

	if rand.Intn(2) == 0 {
		return fmt.Errorf("random error")
	}
	return nil
}

type context struct{}

func main() {
	flag.Parse()
	fmt.Println("Installing some fake data")

	pool := newPool(*redisHostPort)
	cleanKeyspace(pool, *redisNamespace)

	// Enqueue some jobs:
	go func() {
		conn := pool.Get()
		defer conn.Close()
		conn.Do("SADD", *redisNamespace+":known_jobs", "foobar")
	}()

	go func() {
		for {
			en := work.NewEnqueuer(*redisNamespace, pool)
			for i := 0; i < 20; i++ {
				en.Enqueue("foobar", work.Q{"i": i})
			}

			time.Sleep(1 * time.Second)
		}
	}()

	wp := work.NewWorkerPool(context{}, 5, *redisNamespace, pool)
	wp.Job("foobar", epsilonHandler)
	wp.Start()

	select {}
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxActive:   20,
		MaxIdle:     20,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", addr)
			if err != nil {
				return nil, err
			}
			return c, nil
			//return redis.NewLoggingConn(c, log.New(os.Stdout, "", 0), "redis"), err
		},
		Wait: true,
		//TestOnBorrow: func(c redis.Conn, t time.Time) error {
		//	_, err := c.Do("PING")
		//	return err
		//},
	}
}

func cleanKeyspace(pool *redis.Pool, namespace string) {
	conn := pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", namespace+"*"))
	if err != nil {
		panic("could not get keys: " + err.Error())
	}
	for _, k := range keys {
		if _, err := conn.Do("DEL", k); err != nil {
			panic("could not del: " + err.Error())
		}
	}
}
