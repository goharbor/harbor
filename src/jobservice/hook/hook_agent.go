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
	"github.com/goharbor/harbor/src/lib/log"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/gomodule/redigo/redis"
)

const (
	// Backoff duration of direct retrying.
	errRetryBackoff = 5 * time.Minute
	// Max concurrency of retrying goroutines.
	maxConcurrency = 512
)

// Agent is designed to handle the hook events with reasonable numbers of concurrent threads.
type Agent interface {
	// Trigger hooks
	Trigger(evt *Event) error
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

// Basic agent for usage
type basicAgent struct {
	context   context.Context
	namespace string
	client    Client
	redisPool *redis.Pool
	tokens    chan struct{}
}

// NewAgent is constructor of basic agent
func NewAgent(ctx *env.Context, ns string, redisPool *redis.Pool) Agent {
	return &basicAgent{
		context:   ctx.SystemContext,
		namespace: ns,
		client:    NewClient(ctx.SystemContext),
		redisPool: redisPool,
		tokens:    make(chan struct{}, maxConcurrency),
	}
}

// Trigger implements the same method of interface @Agent
func (ba *basicAgent) Trigger(evt *Event) error {
	if evt == nil {
		return errors.New("nil hook event")
	}

	if err := evt.Validate(); err != nil {
		return errors.Wrap(err, "trigger error")
	}

	// Send hook event with retry supported.
	// Exponential backoff is used and the max elapsed time is 5m.
	// If it is still failed to send hook event after all tries, the reaper may help to fix the inconsistent status.
	if err := ba.client.SendEvent(evt); err != nil {
		// Start retry at background.
		go ba.retry(evt)

		return errors.Wrap(err, "trigger hook event error")
	}

	// Mark event hook ACK including "revision", "status" and "check_in_at" in the job stats to indicate
	// the related hook event has been successfully fired.
	// The ACK can be used by the reaper to justify if the hook event should be resent again.
	// The failure of persisting this ACK may cause duplicated hook event resending issue, which
	// can be ignored.
	if err := ba.ack(evt); err != nil {
		// Just log error
		logger.Error(errors.Wrap(err, "hook event ack error"))
	}

	// For debugging
	logger.Debugf("Hook event is successfully sent: %s->%s", evt.Message, evt.URL)

	return nil
}

// retry event with exponential backoff.
// Limit the max concurrency (defined by maxConcurrency) of retrying goroutines.
func (ba *basicAgent) retry(evt *Event) {
	// Apply for a running token.
	// If no token is available, then hold until token is released.
	ba.tokens <- struct{}{}
	// Release token
	defer func() {
		<-ba.tokens
	}()

	// Resend hook event
	bf := newBackoff(errRetryBackoff)
	bf.Reset()

	err := backoff.RetryNotify(func() error {
		logger.Debugf("Retry: sending hook event: %s->%s", evt.Message, evt.URL)

		// Try to avoid sending outdated events, just a try-best operation.
		ot, err := ba.isOutdated(evt)
		if err != nil {
			// Log error and continue.
			logger.Error(err)
		} else {
			if ot {
				logger.Infof("Hook event is abandoned as it's outdated: %s->%s", evt.Message, evt.URL)
				return nil
			}
		}

		return ba.client.SendEvent(evt)
	}, bf, func(e error, d time.Duration) {
		logger.Errorf("Retry: sending hook event error: %s, evt=%s->%s, duration=%v", e.Error(), evt.Message, evt.URL, d)
	})

	if err != nil {
		logger.Errorf("Retry: still failed after all retries: %s, evt=%s->%s", err.Error(), evt.Message, evt.URL)
	}
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

// Check if the event has been outdated.
func (ba *basicAgent) isOutdated(evt *Event) (bool, error) {
	if evt == nil || evt.Data == nil {
		return false, nil
	}

	conn := ba.redisPool.Get()
	defer func() {
		_ = conn.Close()
	}()

	key := rds.KeyJobStats(ba.namespace, evt.Data.JobID)
	values, err := rds.HmGet(conn, key, "ack")
	if err != nil {
		return false, errors.Wrap(err, "check outdated event error")
	}

	// Parse ack
	if ab, ok := values[0].([]byte); ok && len(ab) > 0 {
		ack := &job.ACK{}
		if err := json.Unmarshal(ab, ack); err != nil {
			return false, errors.Wrap(err, "parse ack error")
		}

		// Revision
		diff := ack.Revision - evt.Data.Metadata.Revision
		switch {
		// Revision of the hook event has left behind the current acked revision.
		case diff > 0:
			return true, nil
		case diff < 0:
			return false, nil
		case diff == 0:
			// Continue to compare the status.
		}

		// Status
		st := job.Status(ack.Status)
		if err := st.Validate(); err != nil {
			return false, errors.Wrap(err, "validate acked job status error")
		}

		est := job.Status(evt.Data.Status)
		if err := est.Validate(); err != nil {
			return false, errors.Wrap(err, "validate job status error")
		}

		switch {
		case st.Before(est):
			return false, nil
		case st.After(est):
			return true, nil
		case st.Equal(est):
			log.Debugf("ignore the consistent status: %v", est)
			// Continue to compare check in at timestamp
		}

		// Check in timestamp
		if ack.CheckInAt >= evt.Data.Metadata.CheckInAt {
			return true, nil
		}
	}

	return false, nil
}

func newBackoff(maxElapsedTime time.Duration) backoff.BackOff {
	bf := backoff.NewExponentialBackOff()
	bf.InitialInterval = 2 * time.Second
	bf.RandomizationFactor = 0.5
	bf.Multiplier = 2
	bf.MaxElapsedTime = maxElapsedTime

	return bf
}
