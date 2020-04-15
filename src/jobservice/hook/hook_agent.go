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
	"net/url"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/gomodule/redigo/redis"
)

const (
	// The max time for expiring the retrying events
	// 1 day
	maxEventExpireTime = 3600 * 24
	// Waiting a short while if any errors occurred
	shortLoopInterval = 5 * time.Second
	// Waiting for long while if no retrying elements found
	longLoopInterval = 5 * time.Minute
)

// Agent is designed to handle the hook events with reasonable numbers of concurrent threads
type Agent interface {
	// Trigger hooks
	Trigger(evt *Event) error

	// Serves retry loop now
	Serve() error
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
	redisPool *redis.Pool
	wg        *sync.WaitGroup
}

// NewAgent is constructor of basic agent
func NewAgent(ctx *env.Context, ns string, redisPool *redis.Pool) Agent {
	return &basicAgent{
		context:   ctx.SystemContext,
		namespace: ns,
		client:    NewClient(ctx.SystemContext),
		redisPool: redisPool,
		wg:        ctx.WG,
	}
}

// Trigger implements the same method of interface @Agent
func (ba *basicAgent) Trigger(evt *Event) error {
	if evt == nil {
		return errors.New("nil web hook event")
	}

	if err := evt.Validate(); err != nil {
		return errors.Wrap(err, "trigger error")
	}

	// Treat hook event is success if it is successfully sent or cached in the retry queue.
	if err := ba.client.SendEvent(evt); err != nil {
		// Push event to the retry queue
		if er := ba.pushForRetry(evt); er != nil {
			// Failed to push to the hook event retry queue, return error with all context
			return errors.Wrap(er, err.Error())
		}

		logger.Warningf("Send hook event '%s' to '%s' failed with error: %s; push hook event to the queue for retrying later", evt.Message, evt.URL, err)
		// Treat as successful hook event as the event has been put into the retry queue for future resending.
		return nil
	}

	// Mark event hook ACK including "revision", "status" and "check_in_at" in the job stats to indicate
	// the related hook event has been successfully fired.
	// The ACK can be used by the reaper to justify if the hook event should be resent again.
	// The failure of persisting this ACK may cause duplicated hook event resending issue, which
	// can be ignored.
	if err := ba.ack(evt); err != nil {
		// Just log error
		logger.Error(errors.Wrap(err, "trigger"))
	}

	return nil
}

// Start the basic agent
// Termination depends on the system context
// Blocking call
func (ba *basicAgent) Serve() error {
	ba.wg.Add(1)

	go ba.loopRetry()
	logger.Info("Hook event retrying loop is started")

	return nil
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

	for {
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
	}
}

func (ba *basicAgent) reSend() error {
	conn := ba.redisPool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error(errors.Wrap(err, "resend"))
		}
	}()

	// Pick up one queued event for resending
	evt, err := ba.popMinOne(conn)
	if err != nil {
		return err
	}

	// Args for executing script
	args := []interface{}{
		rds.KeyJobStats(ba.namespace, evt.Data.JobID),
		evt.Data.Status,
		evt.Data.Metadata.Revision,
		evt.Data.Metadata.CheckInAt,
	}

	// If failed to check the status matching, just ignore it, continue the resending
	reply, err := redis.String(rds.CheckStatusMatchScript.Do(conn, args...))
	if err != nil {
		// Log error
		logger.Error(errors.Wrap(err, "resend"))
	} else {
		if reply != "ok" {
			return errors.Errorf("outdated hook event: %s", evt.Message)
		}
	}

	return ba.Trigger(evt)
}

// popMinOne picks up one event for retrying
func (ba *basicAgent) popMinOne(conn redis.Conn) (*Event, error) {
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

// ack hook event
func (ba *basicAgent) ack(evt *Event) error {
	conn := ba.redisPool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Error(errors.Wrap(err, "ack"))
		}
	}()

	k := rds.KeyJobStats(ba.namespace, evt.Data.JobID)
	k2 := rds.KeyJobTrackInProgress(ba.namespace)
	reply, err := redis.String(rds.HookAckScript.Do(
		conn,
		k,
		k2,
		evt.Data.Status,
		evt.Data.Metadata.Revision,
		evt.Data.Metadata.CheckInAt,
		evt.Data.JobID,
	))
	if err != nil {
		return errors.Wrap(err, "ack")
	}

	if reply != "ok" {
		return errors.Errorf("no ack done for event: %s", evt.Message)
	}

	return nil
}
