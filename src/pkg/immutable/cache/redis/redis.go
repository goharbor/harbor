package redis

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/immutable/cache"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
)

// Cache ...
type Cache struct {
	pool *redis.Pool
}

// NewRedisCache ...
func NewRedisCache(pool *redis.Pool) Cache {
	return Cache{
		pool: pool,
	}
}

// Set ...
func (imrc *Cache) Set(pid int64, imc cache.IMCandidate) error {
	conn := imrc.pool.Get()
	defer conn.Close()

	return imrc.set(conn, pid, imc)
}

// SetMultiple ...
func (imrc *Cache) SetMultiple(pid int64, imcs []cache.IMCandidate) error {
	conn := imrc.pool.Get()
	defer conn.Close()

	for _, c := range imcs {
		if err := imrc.set(conn, pid, c); err != nil {
			log.Warning(err.Error())
			continue
		}
	}

	return nil
}

// set ...
func (imrc *Cache) set(conn redis.Conn, pid int64, imc cache.IMCandidate) error {
	if _, err := conn.Do("SADD", imrc.repositoryKey(pid), imc.Repository+"::"+imc.Tag); err != nil {
		return err
	}

	if _, err := conn.Do("HSET", imrc.tagKey(pid, imc.Repository, imc.Tag), "immutable", imc.Immutable); err != nil {
		return err
	}

	return nil
}

// Stat ...
func (imrc *Cache) Stat(pid int64, repository string, tag string) (bool, error) {
	conn := imrc.pool.Get()
	defer conn.Close()

	member, err := redis.Bool(conn.Do("SISMEMBER", imrc.repositoryKey(pid), repository+"::"+tag))
	if err != nil {
		return false, err
	}

	if !member {
		return false, cache.ErrRepoUnknown
	}

	isImmutable, err := redis.Bool(conn.Do("HGET", imrc.tagKey(pid, repository, tag), "immutable"))
	if err != nil {
		return false, err
	}

	return isImmutable, nil
}

// Clear ...
func (imrc *Cache) Clear(pid int64, imc cache.IMCandidate) error {
	conn := imrc.pool.Get()
	defer conn.Close()

	// Check membership to repository first
	member, err := redis.Bool(conn.Do("SISMEMBER", imrc.repositoryKey(pid), imc.Repository+"::"+imc.Tag))
	if err != nil {
		return err
	}

	if !member {
		return cache.ErrRepoUnknown
	}

	remRepo, err := redis.Bool(conn.Do("SREM", imrc.repositoryKey(pid), imc.Repository+"::"+imc.Tag))
	if err != nil {
		return err
	}

	if !remRepo {
		return errors.New("failed to remove repository from project repo list")
	}

	remTag, err := conn.Do("DEL", imrc.tagKey(pid, imc.Repository, imc.Tag))
	if err != nil {
		return err
	}

	if remTag == 0 {
		return cache.ErrTagUnknown
	}

	return nil
}

// Flush ...
func (imrc *Cache) Flush(pid int64) error {
	conn := imrc.pool.Get()
	defer conn.Close()

	reply, err := conn.Do("DEL", imrc.repositoryKey(pid))
	if err != nil {
		return err
	}

	if reply == 0 {
		return cache.ErrRepoUnknown
	}

	return delKeys(conn, imrc.tagKeyPrfix(pid))
}

func (imrc *Cache) tagKeyPrfix(pid int64) string {
	return fmt.Sprintf("immutable::project::tags::%d*", pid)
}

func (imrc *Cache) tagKey(pid int64, repository string, tag string) string {
	return fmt.Sprintf("immutable::project::tags::%d::repo::%s::tag::%s", pid, repository, tag)
}

func (imrc *Cache) repositoryKey(pid int64) string {
	return fmt.Sprintf("immutable::project::%d::repositories", pid)
}

func delKeys(con redis.Conn, pattern string) error {
	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.Values(con.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return fmt.Errorf("error retrieving '%s' keys", pattern)
		}
		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return fmt.Errorf("unexpected type for Int, got type %T", err)
		}
		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return fmt.Errorf("converts an array command reply to a []string %v", err)
		}
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}
	for _, key := range keys {
		_, err := con.Do("DEL", key)
		if err != nil {
			return fmt.Errorf("failed to clean registry cache %v", err)
		}
	}
	return nil
}
