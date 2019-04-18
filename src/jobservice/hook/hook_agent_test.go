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
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
)

func TestEventSending(t *testing.T) {
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

	ctx, cancel := context.WithCancel(context.Background())

	ns := tests.GiveMeTestNamespace()
	pool := tests.GiveMeRedisPool()

	conn := pool.Get()
	defer tests.ClearAll(ns, conn)

	agent := NewAgent(ctx, ns, pool)
	agent.Serve()

	go func() {
		defer func() {
			cancel()
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

			if err := agent.Trigger(evt); err != nil {
				t.Fatal(err)
			}
		}

		// Check results
		<-done
		if count != expected {
			t.Fatalf("expected %d hook events but only got %d", expected, count)
		}
	}()

	// Wait
	<-ctx.Done()
}

func TestRetryAndPopMin(t *testing.T) {
	ctx := context.Background()
	ns := tests.GiveMeTestNamespace()
	pool := tests.GiveMeRedisPool()

	conn := pool.Get()
	defer tests.ClearAll(ns, conn)

	tks := make(chan bool, maxHandlers)
	// Put tokens
	for i := 0; i < maxHandlers; i++ {
		tks <- true
	}

	agent := &basicAgent{
		context:   ctx,
		namespace: ns,
		client:    NewClient(),
		events:    make(chan *Event, maxEventChanBuffer),
		tokens:    tks,
		redisPool: pool,
	}

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
	conn = pool.Get()
	defer conn.Close()

	key := rds.KeyJobStats(ns, "fake_job_ID")
	_, err := conn.Do("HSET", key, "status", job.SuccessStatus.String())
	if err != nil {
		t.Fatal(err)
	}

	if err := agent.pushForRetry(evt); err != nil {
		t.Fatal(err)
	}

	if err := agent.popMinOnes(); err != nil {
		t.Fatal(err)
	}

	// Check results
	if len(agent.events) > 0 {
		t.Error("the hook event should be discard but actually not")
	}

	// Change status
	_, err = conn.Do("HSET", key, "status", job.PendingStatus.String())
	if err != nil {
		t.Fatal(err)
	}

	if err := agent.pushForRetry(evt); err != nil {
		t.Fatal(err)
	}

	if err := agent.popMinOnes(); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Duration(1) * time.Second)

	if len(agent.events) != 1 {
		t.Errorf("the hook event should be requeued but actually not: %d", len(agent.events))
	}
}
