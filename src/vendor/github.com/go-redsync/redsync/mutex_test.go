package redsync

import (
	"strconv"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/stvp/tempredis"
)

func TestMutex(t *testing.T) {
	pools := newMockPools(8)
	mutexes := newTestMutexes(pools, "test-mutex", 8)
	orderCh := make(chan int)
	for i, mutex := range mutexes {
		go func(i int, mutex *Mutex) {
			err := mutex.Lock()
			if err != nil {
				t.Fatalf("Expected err == nil, got %q", err)
			}
			defer mutex.Unlock()

			assertAcquired(t, pools, mutex)

			orderCh <- i
		}(i, mutex)
	}
	for range mutexes {
		<-orderCh
	}
}

func TestMutexExtend(t *testing.T) {
	pools := newMockPools(8)
	mutexes := newTestMutexes(pools, "test-mutex-extend", 1)
	mutex := mutexes[0]

	err := mutex.Lock()
	if err != nil {
		t.Fatalf("Expected err == nil, got %q", err)
	}
	defer mutex.Unlock()

	time.Sleep(1 * time.Second)

	expiries := getPoolExpiries(pools, mutex.name)
	ok := mutex.Extend()
	if !ok {
		t.Fatalf("Expected ok == true, got %v", ok)
	}
	expiries2 := getPoolExpiries(pools, mutex.name)

	for i, expiry := range expiries {
		if expiry >= expiries2[i] {
			t.Fatalf("Expected expiries[%d] > expiry, got %d %d", i, expiries2[i], expiry)
		}
	}
}

func TestMutexQuorum(t *testing.T) {
	pools := newMockPools(4)
	for mask := 0; mask < 1<<uint(len(pools)); mask++ {
		mutexes := newTestMutexes(pools, "test-mutex-partial-"+strconv.Itoa(mask), 1)
		mutex := mutexes[0]
		mutex.tries = 1

		n := clogPools(pools, mask, mutex)

		if n >= len(pools)/2+1 {
			err := mutex.Lock()
			if err != nil {
				t.Fatalf("Expected err == nil, got %q", err)
			}
			assertAcquired(t, pools, mutex)
		} else {
			err := mutex.Lock()
			if err != ErrFailed {
				t.Fatalf("Expected err == %q, got %q", ErrFailed, err)
			}
		}
	}
}

func newMockPools(n int) []Pool {
	pools := []Pool{}
	for _, server := range servers {
		func(server *tempredis.Server) {
			pools = append(pools, &redis.Pool{
				MaxIdle:     3,
				IdleTimeout: 240 * time.Second,
				Dial: func() (redis.Conn, error) {
					return redis.Dial("unix", server.Socket())
				},
				TestOnBorrow: func(c redis.Conn, t time.Time) error {
					_, err := c.Do("PING")
					return err
				},
			})
		}(server)
		if len(pools) == n {
			break
		}
	}
	return pools
}

func getPoolValues(pools []Pool, name string) []string {
	values := []string{}
	for _, pool := range pools {
		conn := pool.Get()
		value, err := redis.String(conn.Do("GET", name))
		conn.Close()
		if err != nil && err != redis.ErrNil {
			panic(err)
		}
		values = append(values, value)
	}
	return values
}

func getPoolExpiries(pools []Pool, name string) []int {
	expiries := []int{}
	for _, pool := range pools {
		conn := pool.Get()
		expiry, err := redis.Int(conn.Do("PTTL", name))
		conn.Close()
		if err != nil && err != redis.ErrNil {
			panic(err)
		}
		expiries = append(expiries, expiry)
	}
	return expiries
}

func clogPools(pools []Pool, mask int, mutex *Mutex) int {
	n := 0
	for i, pool := range pools {
		if mask&(1<<uint(i)) == 0 {
			n++
			continue
		}
		conn := pool.Get()
		_, err := conn.Do("SET", mutex.name, "foobar")
		conn.Close()
		if err != nil {
			panic(err)
		}
	}
	return n
}

func newTestMutexes(pools []Pool, name string, n int) []*Mutex {
	mutexes := []*Mutex{}
	for i := 0; i < n; i++ {
		mutexes = append(mutexes, &Mutex{
			name:   name,
			expiry: 8 * time.Second,
			tries:  32,
			delay:  500 * time.Millisecond,
			factor: 0.01,
			quorum: len(pools)/2 + 1,
			pools:  pools,
		})
	}
	return mutexes
}

func assertAcquired(t *testing.T, pools []Pool, mutex *Mutex) {
	n := 0
	values := getPoolValues(pools, mutex.name)
	for _, value := range values {
		if value == mutex.value {
			n++
		}
	}
	if n < mutex.quorum {
		t.Fatalf("Expected n >= %d, got %d", mutex.quorum, n)
	}
}
