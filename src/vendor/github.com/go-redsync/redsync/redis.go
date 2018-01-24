package redsync

import "github.com/garyburd/redigo/redis"

// A Pool maintains a pool of Redis connections.
type Pool interface {
	Get() redis.Conn
}
