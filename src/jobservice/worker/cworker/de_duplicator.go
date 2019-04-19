package cworker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/gomodule/redigo/redis"
)

// DeDuplicator is designed to handle the uniqueness of the job.
// Once a job is declared to be unique, the job can be enqueued only if
// no same job (same job name and parameters) in the queue or running in progress.
// Adopt the same unique mechanism with the upstream framework.
type DeDuplicator interface {
	// Check the uniqueness of the unique job and set the unique flag if it is not set yet.
	//
	// Parameters:
	//  jobName string           : name of the job
	//  params models.Parameters : parameters of the job
	//
	// Returns:
	//  If no unique flag and successfully set it, a nil error is returned;
	//  otherwise, a non nil error is returned.
	MustUnique(jobName string, params job.Parameters) error

	// Remove the unique flag after job exiting
	// Parameters:
	//  jobName string           : name of the job
	//  params models.Parameters : parameters of the job
	//
	// Returns:
	//  If unique flag is successfully removed, a nil error is returned;
	//  otherwise, a non nil error is returned.
	DelUniqueSign(jobName string, params job.Parameters) error
}

// redisDeDuplicator implements the redisDeDuplicator interface based on redis.
type redisDeDuplicator struct {
	// Redis namespace
	namespace string
	// Redis conn worker
	pool *redis.Pool
}

// NewDeDuplicator is constructor of redisDeDuplicator
func NewDeDuplicator(ns string, pool *redis.Pool) DeDuplicator {
	return &redisDeDuplicator{
		namespace: ns,
		pool:      pool,
	}
}

// MustUnique checks if the job is unique and set unique flag if it is not set yet.
func (rdd *redisDeDuplicator) MustUnique(jobName string, params job.Parameters) error {
	uniqueKey, err := redisKeyUniqueJob(rdd.namespace, jobName, params)
	if err != nil {
		return fmt.Errorf("job unique key generated error: %s", err)
	}

	conn := rdd.pool.Get()
	defer conn.Close()

	args := []interface{}{
		uniqueKey,
		1,
		"NX",
		"EX",
		86400,
	}

	res, err := redis.String(conn.Do("SET", args...))
	if err == redis.ErrNil {
		return errs.ConflictError(uniqueKey)
	}

	if err == nil {
		if strings.ToUpper(res) == "OK" {
			return nil
		}

		return errors.New("unique job error: missing 'OK' reply")
	}

	return err
}

// DelUniqueSign delete the job unique sign
func (rdd *redisDeDuplicator) DelUniqueSign(jobName string, params job.Parameters) error {
	uniqueKey, err := redisKeyUniqueJob(rdd.namespace, jobName, params)
	if err != nil {
		return fmt.Errorf("delete unique job error: %s", err)
	}

	conn := rdd.pool.Get()
	defer conn.Close()

	if _, err := conn.Do("DEL", uniqueKey); err != nil {
		return fmt.Errorf("delete unique job error: %s", err)
	}

	return nil
}

// Same key with upstream framework
func redisKeyUniqueJob(namespace, jobName string, args map[string]interface{}) (string, error) {
	var buf bytes.Buffer

	buf.WriteString(rds.KeyNamespacePrefix(namespace))
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
