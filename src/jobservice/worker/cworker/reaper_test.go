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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/list"

	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/jobservice/job"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/suite"
)

// ReaperTestSuite is used to test reaper
type ReaperTestSuite struct {
	suite.Suite

	namespace string
	pool      *redis.Pool
	r         *reaper
	ctl       *mockLcmCtl
	jid       string
}

// TestReaper is the entry func of the ReaperTestSuite
func TestReaper(t *testing.T) {
	suite.Run(t, &ReaperTestSuite{})
}

// SetupSuite prepares the test suite environment
func (suite *ReaperTestSuite) SetupSuite() {
	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()

	ctx := context.TODO()
	suite.ctl = &mockLcmCtl{}
	suite.r = &reaper{
		context:   ctx,
		namespace: suite.namespace,
		pool:      suite.pool,
		jobTypes:  []string{job.SampleJob},
		lcmCtl:    suite.ctl,
	}

	conn := suite.pool.Get()
	defer func() {
		err := conn.Close()
		suite.NoError(err, "close redis connection error")
	}()

	// Mock data in the redis DB
	cwp := utils.MakeIdentifier()
	wpk := fmt.Sprintf("%s%s", rds.KeyNamespacePrefix(suite.namespace), "worker_pools")
	_ = conn.Send("SADD", wpk, cwp)
	_ = conn.Send("HSET", fmt.Sprintf("%s:%s", wpk, cwp), "heartbeat_at", time.Now().Unix())
	err := conn.Flush()
	suite.NoError(err, "mock current pool error")

	// Mock lock info of job DEMO
	lk := rds.KeyJobLock(suite.namespace, job.SampleJob)
	_, err = conn.Do("INCR", lk)
	suite.NoError(err, "set lock data error")

	wp := utils.MakeIdentifier()
	lik := rds.KeyJobLockInfo(suite.namespace, job.SampleJob)
	_, err = conn.Do("HINCRBY", lik, wp, 1)
	suite.NoError(err, "set lock_info data error")

	// Mock in-progress job
	ipk := rds.KeyInProgressQueue(suite.namespace, job.SampleJob, wp)
	j, err := mockJobData()
	suite.NoError(err, "mock job")

	_, err = conn.Do("LPUSH", ipk, j)
	suite.NoError(err, "push mock job to queue")

	// Mock job stats
	suite.jid = utils.MakeIdentifier()
	err = mockJobStats(conn, suite.namespace, suite.jid)
	suite.NoError(err, "mock job stats")

	// Mock in-progress job track
	tk := rds.KeyJobTrackInProgress(suite.namespace)
	_, err = conn.Do("HSET", tk, suite.jid, 1)
	suite.NoError(err, "mock in-progress track")
}

// TearDownSuite clears down the test suite environment
func (suite *ReaperTestSuite) TearDownSuite() {
	conn := suite.pool.Get()
	defer func() {
		err := conn.Close()
		suite.NoError(err, "close redis connection error")
	}()

	_ = tests.ClearAll(suite.namespace, conn)
}

func (suite *ReaperTestSuite) TestRequeueInProgressJobs() {
	err := suite.r.reEnqueueInProgressJobs()
	suite.NoError(err, "requeue in-progress jobs")

	conn := suite.pool.Get()
	defer func() {
		err := conn.Close()
		suite.NoError(err, "close redis connection error")
	}()

	v, err := redis.Int(conn.Do("GET", rds.KeyJobLock(suite.namespace, job.SampleJob)))
	suite.NoError(err, "get job lock info")
	suite.Equal(0, v, "lock should be 0")
}

func (suite *ReaperTestSuite) TestSyncOutdatedStats() {
	// Use real track to test
	mt := job.NewBasicTrackerWithID(
		context.TODO(),
		suite.jid,
		suite.namespace,
		suite.pool,
		func(hookURL string, change *job.StatusChange) error {
			return nil
		},
		list.New())
	err := mt.Load()
	suite.NoError(err, "track job stats")
	suite.ctl.On("Track", suite.jid).Return(mt, nil)

	err = suite.r.syncOutdatedStats()
	suite.NoError(err, "sync outdated stats")

	// Check result
	conn := suite.pool.Get()
	defer func() {
		err := conn.Close()
		suite.NoError(err, "close redis connection error")
	}()

	status, err := redis.String(conn.Do("HGET", rds.KeyJobStats(suite.namespace, suite.jid), "status"))
	suite.NoError(err, "get status")
	suite.Equal(job.SuccessStatus.String(), status, "check status")
}

func mockJobData() (string, error) {
	j := make(map[string]interface{})
	j["name"] = job.SampleJob
	j["id"] = utils.MakeIdentifier()
	j["t"] = time.Now().Unix()
	args := make(map[string]interface{})
	j["args"] = args
	args["image"] = "test suite"

	b, err := json.Marshal(&j)
	if err != nil {
		return "", nil
	}

	return string(b), nil
}

func mockJobStats(conn redis.Conn, ns string, jid string) error {
	rev := time.Now().Unix()
	sk := rds.KeyJobStats(ns, jid)

	ack := &job.ACK{
		Revision: rev,
		Status:   job.SuccessStatus.String(),
	}

	b, err := json.Marshal(ack)
	if err != nil {
		return err
	}

	args := []interface{}{
		sk,
		"id", jid,
		"status", job.RunningStatus.String(),
		"name", job.SampleJob,
		"kind", job.KindGeneric,
		"unique", 0,
		"ref_link", fmt.Sprintf("/api/v1/jobs/%s", jid),
		"enqueue_time", time.Now().Unix(),
		"update_time", time.Now().Unix(),
		"revision", rev,
		"ack", string(b),
	}

	_, err = conn.Do("HMSET", args...)

	return err
}

type mockLcmCtl struct {
	mock.Mock
}

func (m *mockLcmCtl) Serve() error {
	return nil
}

// New tracker from the new provided stats
func (m *mockLcmCtl) New(stats *job.Stats) (job.Tracker, error) {
	args := m.Called(stats)
	if args.Get(0) != nil {
		return args.Get(0).(job.Tracker), nil
	}

	return nil, args.Error(1)
}

// Track the life cycle of the specified existing job
func (m *mockLcmCtl) Track(jobID string) (job.Tracker, error) {
	args := m.Called(jobID)
	if args.Get(0) != nil {
		return args.Get(0).(job.Tracker), nil
	}

	return nil, args.Error(1)
}
