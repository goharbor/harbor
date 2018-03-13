// Copyright 2018 The Harbor Authors. All rights reserved.

package period

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/models"
	"github.com/vmware/harbor/src/jobservice_v2/utils"
)

//RedisPeriodicScheduler manages the periodic scheduling policies.
type RedisPeriodicScheduler struct {
	context   context.Context
	redisPool *redis.Pool
	namespace string
	pstore    *periodicJobPolicyStore
	enqueuer  *periodicEnqueuer
}

//NewRedisPeriodicScheduler is constructor of RedisPeriodicScheduler
func NewRedisPeriodicScheduler(ctx context.Context, namespace string, redisPool *redis.Pool) *RedisPeriodicScheduler {
	pstore := &periodicJobPolicyStore{
		lock:     new(sync.RWMutex),
		policies: make(map[string]*periodicJobPolicy),
	}
	enqueuer := newPeriodicEnqueuer(namespace, redisPool, pstore)

	return &RedisPeriodicScheduler{
		context:   ctx,
		redisPool: redisPool,
		namespace: namespace,
		pstore:    pstore,
		enqueuer:  enqueuer,
	}
}

//Start to serve
//Enable PUB/SUB
func (rps *RedisPeriodicScheduler) Start() error {
	defer func() {
		log.Info("Redis scheduler is stopped")
	}()

	//Load existing periodic job policies
	if err := rps.Load(); err != nil {
		return err
	}

	//As we get one connection from the pool, don't try to close it.
	conn := rps.redisPool.Get()
	psc := redis.PubSubConn{
		Conn: conn,
	}

	err := psc.Subscribe(redis.Args{}.AddFlat(utils.KeyPeriodicNotification(rps.namespace))...)
	if err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		for {
			switch res := psc.Receive().(type) {
			case error:
				done <- res
				return
			case redis.Message:
				if notification := readMessage(res.Data); notification != nil {
					log.Infof("Got periodic job policy change notification: %s:%s\n", notification.Event, notification.PeriodicJobPolicy.PolicyID)

					switch notification.Event {
					case periodicJobPolicyChangeEventSchedule:
						rps.pstore.add(notification.PeriodicJobPolicy)
					case periodicJobPolicyChangeEventUnSchedule:
						if notification.PeriodicJobPolicy != nil {
							rps.pstore.remove(notification.PeriodicJobPolicy.PolicyID)
						}
					default:
						//do nothing
					}
				}
			case redis.Subscription:
				switch res.Kind {
				case "subscribe":
					log.Infof("Subscribe redis channel %s\n", res.Channel)
				case "unsubscribe":
					//Unsubscribe all, means main goroutine is exiting
					log.Infof("Unsubscribe redis channel %s\n", res.Channel)
					done <- nil
					return
				}
			}
		}
	}()

	//start enqueuer
	rps.enqueuer.start()
	defer rps.enqueuer.stop()
	log.Info("Redis scheduler is started")

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	//blocking here
	for err == nil {
		select {
		case <-ticker.C:
			err = psc.Ping("ping!")
		case <-rps.context.Done():
			err = errors.New("context exit")
		case err = <-done:
			return err
		}
	}

	//Unsubscribe all
	psc.Unsubscribe()
	return <-done
}

//Schedule is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) Schedule(jobName string, params models.Parameters, cronSpec string) (string, error) {
	if utils.IsEmptyStr(jobName) {
		return "", errors.New("empty job name is not allowed")
	}
	if utils.IsEmptyStr(cronSpec) {
		return "", errors.New("cron spec is not set")
	}

	//Although the ZSET can guarantee no duplicated items, we still need to check the existing
	//of the job policy to avoid publish duplicated ones to other nodes as we
	//use transaction commands.
	jobPolicy := &periodicJobPolicy{
		JobName:       jobName,
		JobParameters: params,
		CronSpec:      cronSpec,
	}
	//Serialize data
	rawJSON, err := jobPolicy.serialize()
	if err != nil {
		return "", nil
	}

	//Check existing
	//If existing, treat as a succeed submitting and return the exitsing id
	if score, ok := rps.exists(string(rawJSON)); ok {
		return utils.MakePeriodicPolicyUUIDWithScore(score), nil
	}

	uuid, score := utils.MakePeriodicPolicyUUID()
	//Set back policy ID
	jobPolicy.PolicyID = uuid
	notification := &periodicJobPolicyEvent{
		Event:             periodicJobPolicyChangeEventSchedule,
		PeriodicJobPolicy: jobPolicy,
	}
	rawJSON2, err := notification.serialize()
	if err != nil {
		return "", err
	}

	//Save to redis db and publish notification via redis transaction
	conn := rps.redisPool.Get()
	conn.Send("MULTI")
	conn.Send("ZADD", utils.KeyPeriodicPolicy(rps.namespace), score, rawJSON)
	conn.Send("PUBLISH", utils.KeyPeriodicNotification(rps.namespace), rawJSON2)
	if _, err := conn.Do("EXEC"); err != nil {
		return "", err
	}

	return uuid, nil
}

//UnSchedule is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) UnSchedule(cronJobPolicyID string) error {
	if utils.IsEmptyStr(cronJobPolicyID) {
		return errors.New("cron job policy ID is empty")
	}

	score := utils.ExtractScoreFromUUID(cronJobPolicyID)
	if score == 0 {
		return fmt.Errorf("The ID '%s' is not valid", cronJobPolicyID)
	}

	notification := &periodicJobPolicyEvent{
		Event: periodicJobPolicyChangeEventUnSchedule,
		PeriodicJobPolicy: &periodicJobPolicy{
			PolicyID: cronJobPolicyID, //Only ID required
		},
	}

	rawJSON, err := notification.serialize()
	if err != nil {
		return err
	}

	//REM from redis db
	conn := rps.redisPool.Get()
	conn.Send("MULTI")
	conn.Send("ZREMRANGEBYSCORE", utils.KeyPeriodicPolicy(rps.namespace), score, score) //Accurately remove the item with the specified score
	conn.Send("PUBLISH", utils.KeyPeriodicNotification(rps.namespace), rawJSON)
	_, err = conn.Do("EXEC")

	return err
}

//Load data from zset
func (rps *RedisPeriodicScheduler) Load() error {
	conn := rps.redisPool.Get()
	bytes, err := redis.MultiBulk(conn.Do("ZRANGE", utils.KeyPeriodicPolicy(rps.namespace), 0, -1, "WITHSCORES"))
	if err != nil {
		return err
	}

	allPeriodicPolicies := make([]*periodicJobPolicy, 0)
	for i, l := 0, len(bytes); i < l; i = i + 2 {
		rawPolicy := bytes[i].([]byte)
		rawScore := bytes[i+1].([]byte)
		policy := &periodicJobPolicy{}

		if err := policy.deSerialize(rawPolicy); err != nil {
			//Ignore error which means the policy data is not valid
			//Only logged
			log.Warningf("failed to deserialize periodic policy with error:%s; raw data: %s\n", err, rawPolicy)
			continue
		}
		score, err := strconv.ParseInt(string(rawScore), 10, 64)
		if err != nil {
			//Ignore error which means the policy data is not valid
			//Only logged
			log.Warningf("failed to parse the score of the periodic policy with error:%s\n", err)
			continue
		}

		//Set back the policy ID
		policy.PolicyID = utils.MakePeriodicPolicyUUIDWithScore(score)

		allPeriodicPolicies = append(allPeriodicPolicies, policy)
	}

	if len(allPeriodicPolicies) > 0 {
		rps.pstore.addAll(allPeriodicPolicies)
	}

	log.Infof("Load %d periodic job policies", len(allPeriodicPolicies))
	return nil
}

//Clear is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) Clear() error {
	conn := rps.redisPool.Get()
	_, err := conn.Do("ZREMRANGEBYRANK", utils.KeyPeriodicPolicy(rps.namespace), 0, -1)

	return err
}

func (rps *RedisPeriodicScheduler) exists(rawPolicy string) (int64, bool) {
	if utils.IsEmptyStr(rawPolicy) {
		return 0, false
	}

	conn := rps.redisPool.Get()
	count, err := redis.Int64(conn.Do("ZSCORE", utils.KeyPeriodicPolicy(rps.namespace), rawPolicy))
	return count, err == nil
}

func readMessage(data []byte) *periodicJobPolicyEvent {
	if data == nil || len(data) == 0 {
		return nil
	}

	notification := &periodicJobPolicyEvent{}
	err := json.Unmarshal(data, notification)
	if err != nil {
		return nil
	}

	return notification
}
