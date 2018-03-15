// Copyright 2018 The Harbor Authors. All rights reserved.

package opm

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/job"
	"github.com/vmware/harbor/src/jobservice_v2/models"
	"github.com/vmware/harbor/src/jobservice_v2/utils"
)

const (
	processBufferSize = 1024
	opSaveStats       = "save_job_stats"
	opUpdateStatus    = "update_job_status"
	maxFails          = 3
)

type queueItem struct {
	op    string
	fails uint
	data  interface{}
}

//RedisJobStatsManager implements JobStatsManager based on redis.
type RedisJobStatsManager struct {
	namespace string
	redisPool *redis.Pool
	context   context.Context

	stopChan    chan struct{}
	doneChan    chan struct{}
	processChan chan *queueItem
	isRunning   bool //no need to sync
}

//NewRedisJobStatsManager is constructor of RedisJobStatsManager
func NewRedisJobStatsManager(ctx context.Context, namespace string, redisPool *redis.Pool) *RedisJobStatsManager {
	return &RedisJobStatsManager{
		namespace:   namespace,
		context:     ctx,
		redisPool:   redisPool,
		stopChan:    make(chan struct{}, 1),
		doneChan:    make(chan struct{}, 1),
		processChan: make(chan *queueItem, processBufferSize),
	}
}

//Start is implementation of same method in JobStatsManager interface.
func (rjs *RedisJobStatsManager) Start() {
	if rjs.isRunning {
		return
	}
	go rjs.loop()
	rjs.isRunning = true
}

//Stop is implementation of same method in JobStatsManager interface.
func (rjs *RedisJobStatsManager) Stop() {
	if !rjs.isRunning {
		return
	}
	rjs.stopChan <- struct{}{}
	<-rjs.doneChan
}

//Save is implementation of same method in JobStatsManager interface.
func (rjs *RedisJobStatsManager) Save(jobStats models.JobStats) {
	item := &queueItem{
		op:   opSaveStats,
		data: jobStats,
	}

	rjs.processChan <- item
}

//Retrieve is implementation of same method in JobStatsManager interface.
func (rjs *RedisJobStatsManager) Retrieve(jobID string) (models.JobStats, error) {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobID)
	vals, err := redis.Strings(conn.Do("HGETALL", key))
	if err != nil {
		return models.JobStats{}, err
	}

	res := models.JobStats{
		Stats: &models.JobStatData{},
	}
	for i, l := 0, len(vals); i < l; i = i + 2 {
		prop := vals[i]
		value := vals[i+1]
		switch prop {
		case "id":
			res.Stats.JobID = value
			break
		case "name":
			res.Stats.JobName = value
			break
		case "kind":
			res.Stats.JobKind = value
		case "unique":
			v, err := strconv.ParseBool(value)
			if err != nil {
				v = false
			}
			res.Stats.IsUnique = v
		case "status":
			res.Stats.Status = value
			break
		case "ref_link":
			res.Stats.RefLink = value
			break
		case "enqueue_time":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.EnqueueTime = v
			break
		case "update_time":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.UpdateTime = v
			break
		case "run_at":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.RunAt = v
			break
		case "check_in_at":
			v, _ := strconv.ParseInt(value, 10, 64)
			res.Stats.CheckInAt = v
			break
		case "check_in":
			res.Stats.CheckIn = value
			break
		default:
		}
	}

	return res, nil
}

//SetJobStatus is implementation of same method in JobStatsManager interface.
func (rjs *RedisJobStatsManager) SetJobStatus(jobID string, status string) {
	item := &queueItem{
		op:   opUpdateStatus,
		data: []string{jobID, status},
	}

	rjs.processChan <- item
}

func (rjs *RedisJobStatsManager) loop() {
	controlChan := make(chan struct{})

	defer func() {
		rjs.isRunning = false
		//Notify other sub goroutines
		close(controlChan)
		log.Info("Redis job stats manager is stopped")
	}()

	for {
		select {
		case item := <-rjs.processChan:
			if err := rjs.process(item); err != nil {
				item.fails++
				if item.fails < maxFails {
					//Retry after a random interval
					go func() {
						timer := time.NewTimer(time.Duration(rand.Intn(5)) * time.Second)
						defer timer.Stop()

						select {
						case <-timer.C:
							rjs.processChan <- item
							return
						case <-controlChan:
						}
					}()
				} else {
					log.Warningf("Failed to process '%s' request with error: %s (%d times tried)\n", item.op, err, maxFails)
				}
			}
			break
		case <-rjs.stopChan:
			rjs.doneChan <- struct{}{}
			return
		case <-rjs.context.Done():
			return
		}
	}
}

func (rjs *RedisJobStatsManager) updateJobStatus(jobID string, status string) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobID)
	args := make([]interface{}, 0, 3)
	args = append(args, key, "status", status, "update_time", time.Now().Unix())
	_, err := conn.Do("HMSET", args...)

	return err
}

func (rjs *RedisJobStatsManager) saveJobStats(jobStats models.JobStats) error {
	conn := rjs.redisPool.Get()
	defer conn.Close()

	key := utils.KeyJobStats(rjs.namespace, jobStats.Stats.JobID)
	args := make([]interface{}, 0)
	args = append(args, key)
	args = append(args,
		"id", jobStats.Stats.JobID,
		"name", jobStats.Stats.JobName,
		"kind", jobStats.Stats.JobKind,
		"unique", jobStats.Stats.IsUnique,
		"status", jobStats.Stats.Status,
		"ref_link", jobStats.Stats.RefLink,
		"enqueue_time", jobStats.Stats.EnqueueTime,
		"update_time", jobStats.Stats.UpdateTime,
		"run_at", jobStats.Stats.RunAt,
	)
	if jobStats.Stats.CheckInAt > 0 && !utils.IsEmptyStr(jobStats.Stats.CheckIn) {
		args = append(args,
			"check_in", jobStats.Stats.CheckIn,
			"check_in_at", jobStats.Stats.CheckInAt,
		)
	}

	conn.Send("HMSET", args...)
	//If job kind is periodic job, expire time should not be set
	//If job kind is scheduled job, expire time should be runAt+1day
	if jobStats.Stats.JobKind != job.JobKindPeriodic {
		var expireTime int64 = 60 * 60 * 24
		if jobStats.Stats.JobKind == job.JobKindScheduled {
			nowTime := time.Now().Unix()
			future := jobStats.Stats.RunAt - nowTime
			if future > 0 {
				expireTime += future
			}
		}
		conn.Send("EXPIRE", key, expireTime)
	}

	return conn.Flush()
}

func (rjs *RedisJobStatsManager) process(item *queueItem) error {
	switch item.op {
	case opSaveStats:
		jobStats := item.data.(models.JobStats)
		return rjs.saveJobStats(jobStats)
	case opUpdateStatus:
		data := item.data.([]string)
		return rjs.updateJobStatus(data[0], data[1])
	default:
		break
	}

	return nil
}
