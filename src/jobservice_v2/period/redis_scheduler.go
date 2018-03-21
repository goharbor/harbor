// Copyright 2018 The Harbor Authors. All rights reserved.

package period

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron"

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
	defer conn.Close()

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
					break
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
func (rps *RedisPeriodicScheduler) Schedule(jobName string, params models.Parameters, cronSpec string) (string, int64, error) {
	if utils.IsEmptyStr(jobName) {
		return "", 0, errors.New("empty job name is not allowed")
	}
	if utils.IsEmptyStr(cronSpec) {
		return "", 0, errors.New("cron spec is not set")
	}

	//Get next run time
	schedule, err := cron.Parse(cronSpec)
	if err != nil {
		return "", 0, err
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
		return "", 0, nil
	}

	//Check existing
	//If existing, treat as a succeed submitting and return the exitsing id
	if score, ok := rps.exists(string(rawJSON)); ok {
		id, err := rps.getIDByScore(score)
		return id, 0, err
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
		return "", 0, err
	}

	//Save to redis db and publish notification via redis transaction
	conn := rps.redisPool.Get()
	defer conn.Close()

	err = conn.Send("MULTI")
	if err != nil {
		return "", 0, err
	}
	err = conn.Send("ZADD", utils.KeyPeriodicPolicy(rps.namespace), score, rawJSON)
	if err != nil {
		return "", 0, err
	}
	err = conn.Send("ZADD", utils.KeyPeriodicPolicyScore(rps.namespace), score, uuid)
	if err != nil {
		return "", 0, err
	}
	err = conn.Send("PUBLISH", utils.KeyPeriodicNotification(rps.namespace), rawJSON2)
	if err != nil {
		return "", 0, err
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return "", 0, err
	}

	return uuid, schedule.Next(time.Now()).Unix(), nil
}

//UnSchedule is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) UnSchedule(cronJobPolicyID string) error {
	if utils.IsEmptyStr(cronJobPolicyID) {
		return errors.New("cron job policy ID is empty")
	}

	score, err := rps.getScoreByID(cronJobPolicyID)
	if err != nil {
		return err
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
	defer conn.Close()

	err = conn.Send("MULTI")
	if err != nil {
		return err
	}
	err = conn.Send("ZREMRANGEBYSCORE", utils.KeyPeriodicPolicy(rps.namespace), score, score) //Accurately remove the item with the specified score
	if err != nil {
		return err
	}
	err = conn.Send("ZREMRANGEBYSCORE", utils.KeyPeriodicPolicyScore(rps.namespace), score, score) //Remove key score mapping
	if err != nil {
		return err
	}
	err = conn.Send("PUBLISH", utils.KeyPeriodicNotification(rps.namespace), rawJSON)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXEC")

	return err
}

//Load data from zset
func (rps *RedisPeriodicScheduler) Load() error {
	conn := rps.redisPool.Get()
	defer conn.Close()

	//Let's build key score mapping locally first
	bytes, err := redis.MultiBulk(conn.Do("ZRANGE", utils.KeyPeriodicPolicyScore(rps.namespace), 0, -1, "WITHSCORES"))
	if err != nil {
		return err
	}
	keyScoreMap := make(map[int64]string)
	for i, l := 0, len(bytes); i < l; i = i + 2 {
		pid := string(bytes[i].([]byte))
		rawScore := bytes[i+1].([]byte)
		score, err := strconv.ParseInt(string(rawScore), 10, 64)
		if err != nil {
			//Ignore
			continue
		}
		keyScoreMap[score] = pid
	}

	bytes, err = redis.MultiBulk(conn.Do("ZRANGE", utils.KeyPeriodicPolicy(rps.namespace), 0, -1, "WITHSCORES"))
	if err != nil {
		return err
	}

	allPeriodicPolicies := make([]*periodicJobPolicy, 0, len(bytes)/2)
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
		if pid, ok := keyScoreMap[score]; ok {
			policy.PolicyID = pid
		} else {
			//Something wrong, should not be happended
			//ignore here
			continue
		}

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
	defer conn.Close()

	_, err := conn.Do("ZREMRANGEBYRANK", utils.KeyPeriodicPolicy(rps.namespace), 0, -1)

	return err
}

func (rps *RedisPeriodicScheduler) exists(rawPolicy string) (int64, bool) {
	if utils.IsEmptyStr(rawPolicy) {
		return 0, false
	}

	conn := rps.redisPool.Get()
	defer conn.Close()

	count, err := redis.Int64(conn.Do("ZSCORE", utils.KeyPeriodicPolicy(rps.namespace), rawPolicy))
	return count, err == nil
}

func (rps *RedisPeriodicScheduler) getScoreByID(id string) (int64, error) {
	conn := rps.redisPool.Get()
	defer conn.Close()

	return redis.Int64(conn.Do("ZSCORE", utils.KeyPeriodicPolicyScore(rps.namespace), id))
}

func (rps *RedisPeriodicScheduler) getIDByScore(score int64) (string, error) {
	conn := rps.redisPool.Get()
	defer conn.Close()

	ids, err := redis.Strings(conn.Do("ZRANGEBYSCORE", utils.KeyPeriodicPolicyScore(rps.namespace), score, score))
	if err != nil {
		return "", err
	}

	return ids[0], nil
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
