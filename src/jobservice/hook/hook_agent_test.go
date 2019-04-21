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
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"sync"
)

// HookAgentTestSuite tests functions of hook agent
type HookAgentTestSuite struct {
	suite.Suite

	pool      *redis.Pool
	namespace string
	lcmCtl    lcm.Controller

	envContext *env.Context
	cancel     context.CancelFunc
}

// TestHookAgentTestSuite is entry of go test
func TestHookAgentTestSuite(t *testing.T) {
	suite.Run(t, new(HookAgentTestSuite))
}

// SetupSuite prepares test suites
func (suite *HookAgentTestSuite) SetupSuite() {
	suite.pool = tests.GiveMeRedisPool()
	suite.namespace = tests.GiveMeTestNamespace()

	ctx, cancel := context.WithCancel(context.Background())
	suite.envContext = &env.Context{
		SystemContext: ctx,
		WG:            new(sync.WaitGroup),
	}
	suite.cancel = cancel

	suite.lcmCtl = lcm.NewController(suite.envContext, suite.namespace, suite.pool, func(hookURL string, change *job.StatusChange) error { return nil })
}

// TearDownSuite prepares test suites
func (suite *HookAgentTestSuite) TearDownSuite() {
	conn := suite.pool.Get()
	defer conn.Close()

	tests.ClearAll(suite.namespace, conn)
}

// TestEventSending ...
func (suite *HookAgentTestSuite) TestEventSending() {
	done := make(chan bool, 1)

	expected := uint32(1300) // >1024 max
	count := uint32(0)
	counter := &count

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			c := atomic.AddUint32(counter, 1)
			if c == expected {
				done <- true
			}
		}()
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	// in case test failed and avoid dead lock
	go func() {
		<-time.After(time.Duration(10) * time.Second)
		done <- true // time out
	}()

	agent := NewAgent(suite.envContext, suite.namespace, suite.pool)
	agent.Attach(suite.lcmCtl)
	agent.Serve()

	go func() {
		defer func() {
			suite.cancel()
		}()

		for i := uint32(0); i < expected; i++ {
			changeData := &job.StatusChange{
				JobID:  fmt.Sprintf("job-%d", i),
				Status: "running",
			}

			evt := &Event{
				URL:       ts.URL,
				Message:   fmt.Sprintf("status of job %s change to %s", changeData.JobID, changeData.Status),
				Data:      changeData,
				Timestamp: time.Now().Unix(),
			}

			err := agent.Trigger(evt)
			require.Nil(suite.T(), err, "agent trigger: nil error expected but got %s", err)
		}

		// Check results
		<-done
		require.Equal(suite.T(), expected, count, "expected %d hook events but only got %d", expected, count)
	}()

	// Wait
	suite.envContext.WG.Wait()
}

// TestRetryAndPopMin ...
func (suite *HookAgentTestSuite) TestRetryAndPopMin() {
	ctx := context.Background()

	tks := make(chan bool, maxHandlers)
	// Put tokens
	for i := 0; i < maxHandlers; i++ {
		tks <- true
	}

	agent := &basicAgent{
		context:   ctx,
		namespace: suite.namespace,
		client:    NewClient(ctx),
		events:    make(chan *Event, maxEventChanBuffer),
		tokens:    tks,
		redisPool: suite.pool,
	}
	agent.Attach(suite.lcmCtl)

	changeData := &job.StatusChange{
		JobID:  "fake_job_ID",
		Status: job.RunningStatus.String(),
	}

	evt := &Event{
		URL:       "https://fake.js.com",
		Message:   fmt.Sprintf("status of job %s change to %s", changeData.JobID, changeData.Status),
		Data:      changeData,
		Timestamp: time.Now().Unix(),
	}

	// Mock job stats
	conn := suite.pool.Get()
	defer conn.Close()

	key := rds.KeyJobStats(suite.namespace, "fake_job_ID")
	_, err := conn.Do("HSET", key, "status", job.SuccessStatus.String())
	require.Nil(suite.T(), err, "prepare job stats: nil error returned but got %s", err)

	err = agent.pushForRetry(evt)
	require.Nil(suite.T(), err, "push for retry: nil error expected but got %s", err)

	err = agent.reSend()
	require.Error(suite.T(), err, "resend: non nil error expected but got nil")
	assert.Equal(suite.T(), 0, len(agent.events), "the hook event should be discard but actually not")

	// Change status
	_, err = conn.Do("HSET", key, "status", job.PendingStatus.String())
	require.Nil(suite.T(), err, "prepare job stats: nil error returned but got %s", err)

	err = agent.pushForRetry(evt)
	require.Nil(suite.T(), err, "push for retry: nil error expected but got %s", err)
	err = agent.reSend()
	require.Nil(suite.T(), err, "resend: nil error should be returned but got %s", err)

	<-time.After(time.Duration(1) * time.Second)

	assert.Equal(suite.T(), 1, len(agent.events), "the hook event should be requeued but actually not: %d", len(agent.events))
}
