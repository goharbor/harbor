package main

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/albrow/jobs"
	"github.com/gocraft/health"
	"github.com/gomodule/redigo/redis"
)

var namespace = "jobs"
var pool = newPool(":6379")

func epsilonHandler(i int) error {
	atomic.AddInt64(&totcount, 1)
	return nil
}

func main() {
	stream := health.NewStream().AddSink(&health.WriterSink{os.Stdout})
	cleanKeyspace()

	queueNames := []string{"myqueue", "myqueue2", "myqueue3", "myqueue4", "myqueue5"}
	queues := []*jobs.Type{}

	for _, qn := range queueNames {
		q, err := jobs.RegisterType(qn, 3, epsilonHandler)
		if err != nil {
			panic(err)
		}
		queues = append(queues, q)
	}

	job := stream.NewJob("enqueue_all")

	numJobs := 40000 / len(queues)
	for _, q := range queues {
		for i := 0; i < numJobs; i++ {
			_, err := q.Schedule(100, time.Now(), i)
			if err != nil {
				panic(err)
			}
		}
	}

	job.Complete(health.Success)

	go monitor()

	job = stream.NewJob("run_all")
	pool, err := jobs.NewPool(&jobs.PoolConfig{
		// NumWorkers: 1000,
		// BatchSize:  3000,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		pool.Close()
		if err := pool.Wait(); err != nil {
			panic(err)
		}
	}()
	if err := pool.Start(); err != nil {
		panic(err)
	}
	job.Complete(health.Success)
	select {}
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
