package pool

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/gomodule/redigo/redis"
)

// DeDuplicator is designed to handle the uniqueness of the job.
// Once a job is declared to be unique, the job can be enqueued only if
// no same job (same job name and parameters) in the queue or running in progress.
// Adopt the same unique mechanism with the upstream framework.
type DeDuplicator struct {
	// Redis namespace
	namespace string
	// Redis conn pool
	pool *redis.Pool
}

// NewDeDuplicator is constructor of DeDuplicator
func NewDeDuplicator(ns string, pool *redis.Pool) *DeDuplicator {
	return &DeDuplicator{
		namespace: ns,
		pool:      pool,
	}
}

// Unique checks if the job is unique and set unique flag if it is not set yet.
func (dd *DeDuplicator) Unique(jobName string, params models.Parameters) error {
	uniqueKey, err := redisKeyUniqueJob(dd.namespace, jobName, params)
	if err != nil {
		return fmt.Errorf("unique job error: %s", err)
	}

	conn := dd.pool.Get()
	defer conn.Close()

	args := []interface{}{
		uniqueKey,
		1,
		"NX",
		"EX",
		86400,
	}

	res, err := redis.String(conn.Do("SET", args...))
	if err == nil && strings.ToUpper(res) == "OK" {
		return nil
	}

	return errors.New("unique job error: duplicated")
}

// DelUniqueSign delete the job unique sign
func (dd *DeDuplicator) DelUniqueSign(jobName string, params models.Parameters) error {
	uniqueKey, err := redisKeyUniqueJob(dd.namespace, jobName, params)
	if err != nil {
		return fmt.Errorf("delete unique job error: %s", err)
	}

	conn := dd.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("DEL", uniqueKey); err != nil {
		return fmt.Errorf("delete unique job error: %s", err)
	}

	return nil
}

// Same key with upstream framework
func redisKeyUniqueJob(namespace, jobName string, args map[string]interface{}) (string, error) {
	var buf bytes.Buffer

	buf.WriteString(utils.KeyNamespacePrefix(namespace))
	buf.WriteString("unique:running:")
	buf.WriteString(jobName)
	buf.WriteRune(':')

	if args != nil {
		err := json.NewEncoder(&buf).Encode(args)
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}
