// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hook

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"net/url"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/gomodule/redigo/redis"
	"math"
	"sync"
)

const (
	// Influenced by the worker number setting
	maxEventChanBuffer = 1024
	// Max concurrent client handlers
	maxHandlers = 5
	// The max time for expiring the retrying events
	// 180 days
	maxEventExpireTime = 3600 * 24 * 180
	// Interval for retrying loop
	retryInterval = 2 * time.Minute
	// Number for splitting the event list to sub set for popping out
	defaultShardNum = 3
)

// Agent is designed to handle the hook events with reasonable numbers of concurrent threads
type Agent interface {
	// Trigger hooks
	Trigger(evt *Event) error
	// Serves events now
	Serve()
}

// Event contains the hook URL and the data
type Event struct {
	URL       string            `json:"url"`
	Message   string            `json:"message"`   // meaningful text for event
	Data      *job.StatusChange `json:"data"`      // generic data
	Timestamp int64             `json:"timestamp"` // Use as time threshold of discarding the event (unit: second)
}

// Validate event
func (e *Event) Validate() error {
	_, err := url.Parse(e.URL)
	if err != nil {
		return err
	}

	if e.Data == nil {
		return errors.New("nil hook data")
	}

	return nil
}

// Serialize event to bytes
func (e *Event) Serialize() ([]byte, error) {
	return json.Marshal(e)
}

// Deserialize the bytes to event
func (e *Event) Deserialize(bytes []byte) error {
	return json.Unmarshal(bytes, e)
}

// Basic agent for usage
type basicAgent struct {
	context   context.Context
	namespace string
	client    Client
	events    chan *Event
	tokens    chan bool
	redisPool *redis.Pool
	wg        *sync.WaitGroup
}

// NewAgent is constructor of basic agent
func NewAgent(ctx *env.Context, ns string, redisPool *redis.Pool) Agent {
	tks := make(chan bool, maxHandlers)
	// Put tokens
	for i := 0; i < maxHandlers; i++ {
		tks <- true
	}
	return &basicAgent{
		context:   ctx.SystemContext,
		namespace: ns,
		client:    NewClient(ctx.SystemContext),
		events:    make(chan *Event, maxEventChanBuffer),
		tokens:    tks,
		redisPool: redisPool,
		wg:        ctx.WG,
	}
}

// Trigger implements the same method of interface @Agent
func (ba *basicAgent) Trigger(evt *Event) error {
	if evt == nil {
		return errors.New("nil event")
	}

	if err := evt.Validate(); err != nil {
		return err
	}

	ba.events <- evt

	return nil
}

// Start the basic agent
// Termination depends on the system context
// Blocking call
func (ba *basicAgent) Serve() {
	go ba.looplyRetry()
	logger.Info("Hook event retrying loop is started")
	go ba.serve()
	logger.Info("Basic hook agent is started")

}

func (ba *basicAgent) serve() {
	defer func() {
		logger.Info("Basic hook agent is stopped")
		ba.wg.Done()
	}()

	ba.wg.Add(1)
	for {
		select {
		case evt := <-ba.events:
			// if exceed, wait here
			// avoid too many request connections at the same time
			<-ba.tokens
			go func(evt *Event) {
				defer func() {
					ba.tokens <- true // return token
				}()

				if err := ba.client.SendEvent(evt); err != nil {
					logger.Errorf("Send hook event '%s' to '%s' failed with error: %s; push to the queue for retrying later", evt.Message, evt.URL, err)
					// Push event to the retry queue
					if err := ba.pushForRetry(evt); err != nil {
						// Failed to push to the retry queue, let's directly push it
						// to the event channel of this node with reasonable backoff time.
						logger.Errorf("Failed to push hook event to the retry queue: %s", err)

						// Put to the event chan now
						// In a separate goroutine to avoid occupying the token long time
						go func() {
							// As 'pushForRetry' has checked the timestamp and expired event
							// will be directly discarded and nil error is returned, no need to
							// check it again here.
							<-time.After(time.Duration((rand.Int31n(60) + 5)) * time.Second)
							ba.events <- evt
						}()
					}
				}
			}(evt)

		case <-ba.context.Done():
			return
		}
	}
}

func (ba *basicAgent) pushForRetry(evt *Event) error {
	if evt == nil {
		// do nothing
		return nil
	}

	// Anyway we'll need the raw JSON, let's try to serialize it here
	rawJSON, err := evt.Serialize()
	if err != nil {
		return err
	}

	now := time.Now().Unix()
	if evt.Timestamp > 0 && now-evt.Timestamp >= maxEventExpireTime {
		// Expired, do not need to push back to the retry queue
		logger.Warningf("Event is expired: %s\n", rawJSON)

		return nil
	}

	conn := ba.redisPool.Get()
	defer conn.Close()

	key := rds.KeyHookEventRetryQueue(ba.namespace)
	args := make([]interface{}, 0)

	// Use nano time to get more accurate timestamp
	score := time.Now().UnixNano()
	args = append(args, key, "NX", score, rawJSON)

	_, err = conn.Do("ZADD", args...)
	if err != nil {
		return err
	}

	return nil
}

func (ba *basicAgent) looplyRetry() {
	defer func() {
		logger.Info("Hook event retrying loop exit")
		ba.wg.Done()
	}()

	ba.wg.Add(1)

	// Append random seconds to avoid working in the same time slot
	tk := time.NewTicker(retryInterval + time.Duration(rand.Int31n(13)+3)*time.Second)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			if err := ba.popMinOnes(); err != nil {
				logger.Errorf("Retrying to send hook events failed with error: %s", err.Error())
			}
		case <-ba.context.Done():
			return
		}
	}
}

func (ba *basicAgent) popMinOnes() error {
	conn := ba.redisPool.Get()
	defer conn.Close()

	key := rds.KeyHookEventRetryQueue(ba.namespace)
	// Get total events
	total, err := redis.Int(conn.Do("ZCARD", key))
	if err != nil {
		return err
	}

	// Get sharding ones
	poppedNum := math.Ceil(float64(total) / float64(defaultShardNum))
	rawContent, err := redis.Values(conn.Do("ZPOPMIN", key, poppedNum))
	if err != nil {
		return err
	}

	for i, l := 0, len(rawContent); i < l; i = i + 2 {
		rawEvent := rawContent[i].([]byte)
		evt := &Event{}

		if err := evt.Deserialize(rawEvent); err != nil {
			// Partially failed
			logger.Warningf("Invalid event data when retrying to send hook event: %s", err.Error())
			continue
		}

		// Compare with current job status if it is still valid hook events
		// If it is already out of date, then directly discard it
		// If it is still valid, then retry to send it
		// Get the current status of job
		jobID, status, err := extractJobID(evt.Data)
		if err != nil {
			logger.Warning(err.Error())
			continue
		}

		latestStatus, err := ba.getJobStatus(jobID)
		if err != nil {
			logger.Warning(err.Error())
			continue
		}

		if status.Compare(latestStatus) < 0 {
			// Already out of date
			logger.Debugf("Abandon out dated status update retrying action: %s", evt.Message)
			continue
		}

		// Put to the event chan for sending with a separate goroutine to avoid long time
		// waiting
		go func(evt *Event) {
			ba.events <- evt
		}(evt)
	}

	return nil
}

func (ba *basicAgent) getJobStatus(jobID string) (job.Status, error) {
	conn := ba.redisPool.Get()
	defer conn.Close()

	key := rds.KeyJobStats(ba.namespace, jobID)
	status, err := redis.String(conn.Do("HGET", key, "status"))
	if err != nil {
		return job.PendingStatus, err
	}

	return job.Status(status), nil
}

// Extract the job ID and status from the event data field
// First return is job ID
// Second return is job status
// Last one is error
func extractJobID(data *job.StatusChange) (string, job.Status, error) {
	if data != nil && len(data.JobID) > 0 {
		status := job.Status(data.Status)
		if status.Validate() == nil {
			return data.JobID, status, nil
		}
	}

	return "", "", errors.New("invalid job status change data to extract job ID")
}
