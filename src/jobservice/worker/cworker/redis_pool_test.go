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
package cworker

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/opm"

	"github.com/goharbor/harbor/src/jobservice/tests"

	"github.com/goharbor/harbor/src/jobservice/env"
)

var rPool = tests.GiveMeRedisPool()

func TestRegisterJob(t *testing.T) {
	wp, _, _ := createRedisWorkerPool()
	defer func() {
		if err := tests.ClearAll(tests.GiveMeTestNamespace(), redisPool.Get()); err != nil {
			t.Error(err)
		}
	}()

	if err := wp.RegisterJob("fake_job", (*fakeJob)(nil)); err != nil {
		t.Error(err)
	}

	if _, ok := wp.IsKnownJob("fake_job"); !ok {
		t.Error("expected known job but registering 'fake_job' appears to have failed")
	}

	delete(wp.knownJobs, "fake_job")

	jobs := make(map[string]interface{})
	jobs["fake_job_1st"] = (*fakeJob)(nil)
	if err := wp.RegisterJobs(jobs); err != nil {
		t.Error(err)
	}

	params := make(map[string]interface{})
	params["name"] = "testing:v1"
	if err := wp.ValidateJobParameters((*fakeJob)(nil), params); err != nil {
		t.Error(err)
	}
}

func TestEnqueueJob(t *testing.T) {
	wp, sysCtx, cancel := createRedisWorkerPool()
	defer func() {
		if err := tests.ClearAll(tests.GiveMeTestNamespace(), redisPool.Get()); err != nil {
			t.Error(err)
		}
	}()
	defer cancel()

	if err := wp.RegisterJob("fake_job", (*fakeJob)(nil)); err != nil {
		t.Error(err)
	}
	if err := wp.RegisterJob("fake_unique_job", (*fakeUniqueJob)(nil)); err != nil {
		t.Error(err)
	}

	go wp.Start()
	time.Sleep(1 * time.Second)

	params := make(map[string]interface{})
	params["name"] = "testing:v1"
	stats, err := wp.Enqueue("fake_job", params, false)
	if err != nil {
		t.Error(err)
	}
	if stats.Stats.JobID == "" {
		t.Error("expect none nil job stats but got nil")
	}

	runAt := time.Now().Unix() + 20
	stats, err = wp.Schedule("fake_job", params, 20, false)
	if err != nil {
		t.Error(err)
	}

	if stats.Stats.RunAt > 0 && stats.Stats.RunAt < runAt {
		t.Errorf("expect returned 'RunAt' should be >= '%d' but seems not", runAt)
	}

	stats, err = wp.Enqueue("fake_unique_job", params, true)
	if err != nil {
		t.Error(err)
	}
	if stats.Stats.JobID == "" {
		t.Error("expect none nil job stats but got nil")
	}

	cancel()
	sysCtx.WG.Wait()
}

func TestEnqueuePeriodicJob(t *testing.T) {
	wp, _, cancel := createRedisWorkerPool()
	defer func() {
		if err := tests.ClearAll(tests.GiveMeTestNamespace(), redisPool.Get()); err != nil {
			t.Error(err)
		}
	}()
	defer cancel()

	if err := wp.RegisterJob("fake_job", (*fakeJob)(nil)); err != nil {
		t.Error(err)
	}

	go wp.Start()
	time.Sleep(1 * time.Second)

	params := make(map[string]interface{})
	params["name"] = "testing:v1"
	jobStats, err := wp.PeriodicallyEnqueue("fake_job", params, "10 * * * * *")
	if err != nil {
		t.Error(err)
	}
	<-time.After(1 * time.Second)

	jStats, err := wp.GetJobStats(jobStats.Stats.JobID)
	if err != nil {
		t.Error(err)
	}

	if jobStats.Stats.JobName != jStats.Stats.JobName {
		t.Error("expect same job stats but got different ones")
	}

	if err := wp.StopJob(jStats.Stats.JobID); err != nil {
		t.Error(err)
	}

	// cancel()
	// <-time.After(1 * time.Second)
}

func TestPoolStats(t *testing.T) {
	wp, _, cancel := createRedisWorkerPool()
	defer func() {
		if err := tests.ClearAll(tests.GiveMeTestNamespace(), redisPool.Get()); err != nil {
			t.Error(err)
		}
	}()
	defer cancel()

	go wp.Start()
	time.Sleep(1 * time.Second)

	_, err := wp.Stats()
	if err != nil {
		t.Fatal(err)
	}
}

func TestStopJob(t *testing.T) {
	wp, _, cancel := createRedisWorkerPool()
	defer func() {
		if err := tests.ClearAll(tests.GiveMeTestNamespace(), redisPool.Get()); err != nil {
			t.Error(err)
		}
	}()
	defer cancel()

	if err := wp.RegisterJob("fake_long_run_job", (*fakeRunnableJob)(nil)); err != nil {
		t.Error(err)
	}

	go wp.Start()
	time.Sleep(1 * time.Second)

	// Stop generic job
	params := make(map[string]interface{})
	params["name"] = "testing:v1"

	genericJob, err := wp.Enqueue("fake_long_run_job", params, false)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	stats, err := wp.GetJobStats(genericJob.Stats.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Stats.Status != job.RunningStatus {
		t.Fatalf("expect job running but got %s", stats.Stats.Status)
	}
	if err := wp.StopJob(genericJob.Stats.JobID); err != nil {
		t.Fatal(err)
	}

	// Stop scheduled job
	scheduledJob, err := wp.Schedule("fake_long_run_job", params, 120, false)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := wp.StopJob(scheduledJob.Stats.JobID); err != nil {
		t.Fatal(err)
	}
}

func TestCancelJob(t *testing.T) {
	wp, _, cancel := createRedisWorkerPool()
	defer func() {
		if err := tests.ClearAll(tests.GiveMeTestNamespace(), redisPool.Get()); err != nil {
			t.Error(err)
		}
	}()
	defer cancel()

	if err := wp.RegisterJob("fake_long_run_job", (*fakeRunnableJob)(nil)); err != nil {
		t.Error(err)
	}

	go wp.Start()
	time.Sleep(1 * time.Second)

	// Cancel job
	params := make(map[string]interface{})
	params["name"] = "testing:v1"

	genericJob, err := wp.Enqueue("fake_long_run_job", params, false)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)
	stats, err := wp.GetJobStats(genericJob.Stats.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Stats.Status != job.RunningStatus {
		t.Fatalf("expect job running but got %s", stats.Stats.Status)
	}

	if err := wp.CancelJob(genericJob.Stats.JobID); err != nil {
		t.Fatal(err)
	}
	time.Sleep(3 * time.Second)

	stats, err = wp.GetJobStats(genericJob.Stats.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Stats.Status != job.JobStatusCancelled {
		t.Fatalf("expect job cancelled but got %s", stats.Stats.Status)
	}

	if err := wp.RetryJob(genericJob.Stats.JobID); err != nil {
		t.Fatal(err)
	}
}

/*func TestCancelAndRetryJobWithHook(t *testing.T) {
	wp, _, cancel := createRedisWorkerPool()
	defer func() {
		if err := tests.ClearAll(tests.GiveMeTestNamespace(), redisPool.Get()); err != nil {
			t.Fatal(err)
		}
	}()
	defer cancel()

	if err := wp.RegisterJob("fake_runnable_job", (*fakeRunnableJob)(nil)); err != nil {
		t.Fatal(err)
	}

	go wp.Start()
	time.Sleep(1 * time.Second)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	params := make(map[string]interface{})
	params["name"] = "testing:v1"
	res, err := wp.Enqueue("fake_runnable_job", params, false)
	if err != nil {
		t.Fatal(err)
	}
	if err := wp.RegisterHook(res.Info.JobID, ts.URL); err != nil {
		t.Fatal(err)
	}
	// make sure it's running
	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()

CHECK:
	<-timer.C
	if check, err := wp.GetJobStats(res.Info.JobID); err != nil {
		t.Fatal(err)
	} else {
		if check.Info.Status != job.RunningStatus {
			timer.Reset(1 * time.Second)
			goto CHECK
		}
	}

	// cancel
	if err := wp.CancelJob(res.Info.JobID); err != nil {
		t.Fatal(err)
	}
	<-time.After(5 * time.Second)
	updatedRes, err := wp.GetJobStats(res.Info.JobID)
	if err != nil {
		t.Fatal(err)
	}
	if updatedRes.Info.Status != job.JobStatusCancelled {
		t.Fatalf("expect job staus '%s' but got '%s'\n", job.JobStatusCancelled, updatedRes.Info.Status)
	}
	if updatedRes.Info.DieAt == 0 {
		t.Fatalf("expect none zero 'DieAt' but got 0 value")
	}

	// retry
	if err := wp.RetryJob(updatedRes.Info.JobID); err != nil {
		t.Fatal(err)
	}
}*/

func createRedisWorkerPool() (*worker, *env.Context, context.CancelFunc) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	envCtx := &env.Context{
		SystemContext: ctx,
		WG:            new(sync.WaitGroup),
		ErrorChan:     make(chan error, 1),
		JobContext:    newContext(ctx),
	}

	return NewWorker(envCtx, tests.GiveMeTestNamespace(), 3, rPool), envCtx, cancel
}

type fakeJob struct{}

func (j *fakeJob) MaxFails() uint {
	return 3
}

func (j *fakeJob) ShouldRetry() bool {
	return true
}

func (j *fakeJob) Validate(params map[string]interface{}) error {
	if p, ok := params["name"]; ok {
		if p == "testing:v1" {
			return nil
		}
	}

	return errors.New("testing error")
}

func (j *fakeJob) Run(ctx env.JobContext, params map[string]interface{}) error {
	return nil
}

type fakeUniqueJob struct{}

func (j *fakeUniqueJob) MaxFails() uint {
	return 3
}

func (j *fakeUniqueJob) ShouldRetry() bool {
	return true
}

func (j *fakeUniqueJob) Validate(params map[string]interface{}) error {
	if p, ok := params["name"]; ok {
		if p == "testing:v1" {
			return nil
		}
	}

	return errors.New("testing error")
}

func (j *fakeUniqueJob) Run(ctx env.JobContext, params map[string]interface{}) error {
	return nil
}

type fakeRunnableJob struct{}

func (j *fakeRunnableJob) MaxFails() uint {
	return 2
}

func (j *fakeRunnableJob) ShouldRetry() bool {
	return true
}

func (j *fakeRunnableJob) Validate(params map[string]interface{}) error {
	if p, ok := params["name"]; ok {
		if p == "testing:v1" {
			return nil
		}
	}

	return errors.New("testing error")
}

func (j *fakeRunnableJob) Run(ctx env.JobContext, params map[string]interface{}) error {
	tk := time.NewTicker(200 * time.Millisecond)
	defer tk.Stop()

	for {
		select {
		case <-tk.C:
			cmd, ok := ctx.OPCommand()
			if ok {
				if cmd == opm.CtlCommandStop {
					return errs.JobStoppedError()
				}

				return errs.JobCancelledError()
			}
		case <-ctx.SystemContext().Done():
			return nil
		case <-time.After(1 * time.Minute):
			return errors.New("fake job timeout")
		}
	}
}

type fakeContext struct {
	// System context
	sysContext context.Context

	// op command func
	opCommandFunc job.CheckOPCmdFunc

	// checkin func
	checkInFunc job.CheckInFunc

	// launch job
	launchJobFunc job.LaunchJobFunc

	// other required information
	properties map[string]interface{}
}

func newContext(sysCtx context.Context) *fakeContext {
	return &fakeContext{
		sysContext: sysCtx,
		properties: make(map[string]interface{}),
	}
}

// Build implements the same method in env.Context interface
// This func will build the job execution context before running
func (c *fakeContext) Build(dep env.JobData) (env.JobContext, error) {
	jContext := &fakeContext{
		sysContext: c.sysContext,
		properties: make(map[string]interface{}),
	}

	// Copy properties
	if len(c.properties) > 0 {
		for k, v := range c.properties {
			jContext.properties[k] = v
		}
	}

	if opCommandFunc, ok := dep.ExtraData["opCommandFunc"]; ok {
		if reflect.TypeOf(opCommandFunc).Kind() == reflect.Func {
			if funcRef, ok := opCommandFunc.(job.CheckOPCmdFunc); ok {
				jContext.opCommandFunc = funcRef
			}
		}
	}
	if jContext.opCommandFunc == nil {
		return nil, errors.New("failed to inject opCommandFunc")
	}

	if checkInFunc, ok := dep.ExtraData["checkInFunc"]; ok {
		if reflect.TypeOf(checkInFunc).Kind() == reflect.Func {
			if funcRef, ok := checkInFunc.(job.CheckInFunc); ok {
				jContext.checkInFunc = funcRef
			}
		}
	}

	if jContext.checkInFunc == nil {
		return nil, errors.New("failed to inject checkInFunc")
	}

	if launchJobFunc, ok := dep.ExtraData["launchJobFunc"]; ok {
		if reflect.TypeOf(launchJobFunc).Kind() == reflect.Func {
			if funcRef, ok := launchJobFunc.(job.LaunchJobFunc); ok {
				jContext.launchJobFunc = funcRef
			}
		}
	}

	if jContext.launchJobFunc == nil {
		return nil, errors.New("failed to inject launchJobFunc")
	}

	return jContext, nil
}

// Get implements the same method in env.Context interface
func (c *fakeContext) Get(prop string) (interface{}, bool) {
	v, ok := c.properties[prop]
	return v, ok
}

// SystemContext implements the same method in env.Context interface
func (c *fakeContext) SystemContext() context.Context {
	return c.sysContext
}

// Checkin is bridge func for reporting detailed status
func (c *fakeContext) Checkin(status string) error {
	if c.checkInFunc != nil {
		c.checkInFunc(status)
	} else {
		return errors.New("nil check in function")
	}

	return nil
}

// OPCommand return the control operational command like stop/cancel if have
func (c *fakeContext) OPCommand() (string, bool) {
	if c.opCommandFunc != nil {
		return c.opCommandFunc()
	}

	return "", false
}

// GetLogger returns the logger
func (c *fakeContext) GetLogger() logger.Interface {
	return nil
}

// LaunchJob launches sub jobs
func (c *fakeContext) LaunchJob(req models.JobRequest) (models.JobStats, error) {
	if c.launchJobFunc == nil {
		return models.JobStats{}, errors.New("nil launch job function")
	}

	return c.launchJobFunc(req)
}
