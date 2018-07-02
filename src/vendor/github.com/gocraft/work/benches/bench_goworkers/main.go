package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/gocraft/health"
	"github.com/gomodule/redigo/redis"
	"github.com/jrallison/go-workers"
)

func myJob(m *workers.Msg) {
	atomic.AddInt64(&totcount, 1)
}

var namespace = "bench_test"
var pool = newPool(":6379")

func main() {

	stream := health.NewStream().AddSink(&health.WriterSink{os.Stdout})
	stream.Event("wat")
	cleanKeyspace()

	workers.Configure(map[string]string{
		// location of redis instance
		"server": "localhost:6379",
		// instance of the database
		"database": "0",
		// number of connections to keep open with redis
		"pool": "10",
		// unique process id for this instance of workers (for proper recovery of inprogress jobs on crash)
		"process":   "1",
		"namespace": namespace,
	})
	workers.Middleware = &workers.Middlewares{}

	queues := []string{"myqueue", "myqueue2", "myqueue3", "myqueue4", "myqueue5"}
	numJobs := 100000 / len(queues)

	job := stream.NewJob("enqueue_all")
	for _, q := range queues {
		enqueueJobs(q, numJobs)
	}
	job.Complete(health.Success)

	for _, q := range queues {
		workers.Process(q, myJob, 10)
	}

	go monitor()

	// Blocks until process is told to exit via unix signal
	workers.Run()
}

var totcount int64

func monitor() {
	t := time.Tick(1 * time.Second)

	curT := 0
	c1 := int64(0)
	c2 := int64(0)
	prev := int64(0)

DALOOP:
	for {
		select {
		case <-t:
			curT++
			v := atomic.AddInt64(&totcount, 0)
			fmt.Printf("after %d seconds, count is %d\n", curT, v)
			if curT == 1 {
				c1 = v
			} else if curT == 3 {
				c2 = v
			}
			if v == prev {
				break DALOOP
			}
			prev = v
		}
	}
	fmt.Println("Jobs/sec: ", float64(c2-c1)/2.0)
	os.Exit(0)
}

func enqueueJobs(queue string, count int) {
	for i := 0; i < count; i++ {
		workers.Enqueue(queue, "Foo", []int{i})
	}
}

func cleanKeyspace() {
	conn := pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", namespace+"*"))
	if err != nil {
		panic("could not get keys: " + err.Error())
	}
	for _, k := range keys {
		//fmt.Println("deleting ", k)
		if _, err := conn.Do("DEL", k); err != nil {
			panic("could not del: " + err.Error())
		}
	}
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxActive:   3,
		MaxIdle:     3,
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
