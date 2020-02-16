package dao

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
)

// RawJSON represents the raw JSON text.
type RawJSON string

// RedisBase provides basic object storage capabilities based on redis backend.
// Object is stored as a whole raw json text.
// Objects are stored in a ZSET structure.
type RedisBase struct {
	pool      *redis.Pool
	namespace string
}

// NewRedisBase is constructor of RedisBase
func NewRedisBase(pool *redis.Pool, namespace string) *RedisBase {
	if pool == nil || len(namespace) == 0 {
		return nil
	}

	return &RedisBase{
		pool:      pool,
		namespace: namespace,
	}
}

// Get the object with the specified key
func (rb *RedisBase) Get(key string) (RawJSON, error) {
	if len(key) == 0 {
		return "", errors.New("nil key")
	}

	conn := rb.pool.Get()
	defer conn.Close()

	score, err := redis.Int64(conn.Do("HGET", storageIndexKey(rb.namespace), key))
	if err != nil {
		return "", err
	}

	args := []interface{}{
		storageListKey(rb.namespace),
		score,
		score,
	}

	rawJSONs, err := redis.Strings(conn.Do("ZRANGEBYSCORE", args...))
	if err != nil {
		if err == redis.ErrNil {
			return "", nil // not existing
		}

		return "", err
	}

	if len(rawJSONs) == 0 {
		return "", nil
	}

	return RawJSON(rawJSONs[0]), nil
}

// Save or update the object
func (rb *RedisBase) Save(key string, object interface{}) error {
	if object == nil {
		return errors.New("object is nil")
	}

	rawJSON, err := toJSON(object)
	if err != nil {
		return err
	}
	score := time.Now().UnixNano()

	conn := rb.pool.Get()
	defer conn.Close()

	if err := conn.Send("MULTI"); err != nil {
		return err
	}
	if err := conn.Send("ZADD", storageListKey(rb.namespace), score, rawJSON); err != nil {
		return err
	}
	if err := conn.Send("HSET", storageIndexKey(rb.namespace), key, score); err != nil {
		return err
	}
	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}

	return nil
}

// Update the object
func (rb *RedisBase) Update(key string, object interface{}) error {
	if object == nil {
		return errors.New("object is nil")
	}

	rawJSON, err := toJSON(object)
	if err != nil {
		return err
	}
	score := time.Now().UnixNano()

	conn := rb.pool.Get()
	defer conn.Close()

	// Check exists
	ret, err := redis.Int(conn.Do("HEXISTS", storageIndexKey(rb.namespace), key))
	if err != nil {
		return err
	}

	if ret == 0 {
		return fmt.Errorf("%s not found", key)
	}

	// Get original score
	oldScore, err := redis.Int64(conn.Do("HGET", storageIndexKey(rb.namespace), key))
	if err != nil {
		return err
	}

	// Delete and save new in one transaction
	if err := conn.Send("MULTI"); err != nil {
		return err
	}
	if err := conn.Send("ZREMRANGEBYSCORE", storageListKey(rb.namespace), oldScore, oldScore); err != nil {
		return err
	}

	if err := conn.Send("HDEL", storageIndexKey(rb.namespace), key); err != nil {
		return err
	}
	if err := conn.Send("ZADD", storageListKey(rb.namespace), score, rawJSON); err != nil {
		return err
	}
	if err := conn.Send("HSET", storageIndexKey(rb.namespace), key, score); err != nil {
		return err
	}
	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}

	return nil
}

// Delete the specified object with the key
func (rb *RedisBase) Delete(key string) error {
	if len(key) == 0 {
		return errors.New("nil key")
	}

	conn := rb.pool.Get()
	defer conn.Close()

	score, err := redis.Int64(conn.Do("HGET", storageIndexKey(rb.namespace), key))
	if err != nil {
		return err
	}
	args := []interface{}{
		storageListKey(rb.namespace),
		score,
		score,
	}

	if err := conn.Send("MULTI"); err != nil {
		return err
	}

	if err := conn.Send("ZREMRANGEBYSCORE", args...); err != nil {
		return err
	}

	if err := conn.Send("HDEL", storageIndexKey(rb.namespace), key); err != nil {
		return err
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}

	return err
}

// List the objects
func (rb *RedisBase) List(queryParam *models.QueryParam) ([]RawJSON, error) {
	var page, pageSize uint = 1, 25

	if queryParam != nil {
		if queryParam.Page > 0 {
			page = queryParam.Page
		}

		if queryParam.PageSize > 0 {
			pageSize = queryParam.PageSize
		}
	}

	// Total
	size, err := rb.Size()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return []RawJSON{}, nil
	}

	pages := (uint)(size / 25)
	if size%25 != 0 {
		pages++
	}

	if page > pages {
		return []RawJSON{}, nil
	}

	// Desc ordered
	end := (size - 1) - (int64)((page-1)*pageSize)
	start := end - (int64)(pageSize) + 1
	if start < 0 {
		start = 0
	}

	conn := rb.pool.Get()
	defer conn.Close()

	args := []interface{}{
		storageListKey(rb.namespace),
		start,
		end,
	}

	rawJSONs, err := redis.Strings(conn.Do("ZREVRANGE", args...))
	if err != nil {
		return nil, err
	}

	results := []RawJSON{}
	for _, jsonText := range rawJSONs {
		shouldApppend := true

		if queryParam != nil && len(queryParam.Keyword) > 0 {
			if !strings.Contains(jsonText, queryParam.Keyword) {
				shouldApppend = false
			}
		}

		if shouldApppend {
			results = append(results, RawJSON(jsonText))
		}
	}

	return results, nil
}

// Size return the len of the list
func (rb *RedisBase) Size() (int64, error) {
	conn := rb.pool.Get()
	defer conn.Close()

	return redis.Int64(conn.Do("ZCARD", storageListKey(rb.namespace)))
}

// Exists checks the existence of the object by key
func (rb *RedisBase) Exists(key string) bool {
	if len(key) == 0 {
		return false
	}

	conn := rb.pool.Get()
	defer conn.Close()

	ret, err := redis.Int(conn.Do("HEXISTS", storageIndexKey(rb.namespace), key))
	if err != nil {
		return false
	}

	return ret == 1
}

func toJSON(object interface{}) (RawJSON, error) {
	jsonData, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	return RawJSON(jsonData), nil
}

func storageListKey(ns string) string {
	return fmt.Sprintf("%s:store", ns)
}

func storageIndexKey(ns string) string {
	return fmt.Sprintf("%s:index", ns)
}
