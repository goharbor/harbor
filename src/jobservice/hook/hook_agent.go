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
	"math/rand"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/goharbor/harbor/src/jobservice/job"

	"sync"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/gomodule/redigo/redis"
)

const (
	// Influenced by the worker number setting
	maxEventChanBuffer = 1024
	// Max concurrent client handlers
	maxHandlers = 5
	// The max time for expiring the retrying events
	// 180 days
	maxEventExpireTime = 3600 * 24 * 180
	// Waiting a short while if any errors occurred
	shortLoopInterval = 5 * time.Second
	// Waiting for long while if no retrying elements found
	longLoopInterval = 5 * time.Minute
)

// Agent is designed to handle the hook events with reasonable numbers of concurrent threads
type Agent interface {
	// Trigger hooks
	Trigger(evt *Event) error
	// Serves events now
	Serve() error
	// Attach a job life cycle controller
	Attach(ctl lcm.Controller)
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
	ctl       lcm.Controller
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

// Attach a job life cycle controller
func (ba *basicAgent) Attach(ctl lcm.Controller) {
	ba.ctl = ctl
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
func (ba *basicAgent) Serve() error {
	if ba.ctl == nil {
		return errors.New("nil life cycle controller of hook agent")
	}

	ba.wg.Add(1)
	go ba.loopRetry()
	logger.Info("Hook event retrying loop is started")

	ba.wg.Add(1)
	go ba.serve()
	logger.Info("Basic hook agent is started")

	return nil
}

func (ba *basicAgent) serve() {
	defer func() {
		logger.Info("Basic hook agent is stopped")
		ba.wg.Done()
	}()

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

						// Put to the event chan after waiting for a reasonable while,
						// waiting is important, it can avoid sending large scale failure expecting
						// requests in a short while.
						// As 'pushForRetry' has checked the timestamp and expired event
						// will be directly discarded and nil error is returned, no need to
						// check it again here.
						<-time.After(time.Duration(rand.Int31n(55)+5) * time.Second)
						ba.events <- evt
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
		logger.Warningf("Event is expired: %s", rawJSON)

		return nil
	}

	conn := ba.redisPool.Get()
	defer func() {
		_ = conn.Close()
	}()

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

func (ba *basicAgent) loopRetry() {
	defer func() {
		logger.Info("Hook event retrying loop exit")
		ba.wg.Done()
	}()

	token := make(chan bool, 1)
	token <- true

	for {
		<-token
		if err := ba.reSend(); err != nil {
			waitInterval := shortLoopInterval
			if err == rds.ErrNoElements {
				// No elements
				waitInterval = longLoopInterval
			} else {
				logger.Errorf("Resend hook event error: %s", err.Error())
			}

			select {
			case <-time.After(waitInterval):
				// Just wait, do nothing
			case <-ba.context.Done():
				// Terminated
				return
			}
		}

		// Put token back
		token <- true
	}
}

func (ba *basicAgent) reSend() error {
	evt, err := ba.popMinOne()
	if err != nil {
		return err
	}

	jobID, status, err := extractJobID(evt.Data)
	if err != nil {
		return err
	}

	t, err := ba.ctl.Track(jobID)
	if err != nil {
		return err
	}

	diff := status.Compare(job.Status(t.Job().Info.Status))
	if diff > 0 ||
		(diff == 0 && t.Job().Info.CheckIn != evt.Data.CheckIn) {
		ba.events <- evt
		return nil
	}

	return errors.Errorf("outdated hook event: %s, latest job status: %s", evt.Message, t.Job().Info.Status)
}

func (ba *basicAgent) popMinOne() (*Event, error) {
	conn := ba.redisPool.Get()
	defer func() {
		_ = conn.Close()
	}()

	key := rds.KeyHookEventRetryQueue(ba.namespace)
	minOne, err := rds.ZPopMin(conn, key)
	if err != nil {
		return nil, err
	}

	rawEvent, ok := minOne.([]byte)
	if !ok {
		return nil, errors.New("bad request: non bytes slice for raw event")
	}

	evt := &Event{}
	if err := evt.Deserialize(rawEvent); err != nil {
		return nil, err
	}

	return evt, nil
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

	return "", "", errors.New("malform job status change data")
}
