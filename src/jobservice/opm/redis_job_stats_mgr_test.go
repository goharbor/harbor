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
package opm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/gomodule/redigo/redis"
)

const (
	dialConnectionTimeout = 30 * time.Second
	healthCheckPeriod     = time.Minute
	dialReadTimeout       = healthCheckPeriod + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
	testingRedisHost      = "REDIS_HOST"
	testingNamespace      = "testing_job_service_v2"
)

var redisHost = getRedisHost()
var redisPool = &redis.Pool{
	MaxActive: 2,
	MaxIdle:   2,
	Wait:      true,
	Dial: func() (redis.Conn, error) {
		return redis.Dial(
			"tcp",
			fmt.Sprintf("%s:%d", redisHost, 6379),
			redis.DialConnectTimeout(dialConnectionTimeout),
			redis.DialReadTimeout(dialReadTimeout),
			redis.DialWriteTimeout(dialWriteTimeout),
		)
	},
}

func TestSetJobStatus(t *testing.T) {
	mgr := createStatsManager(redisPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)
	// make sure data existing
	testingStats := createFakeStats()
	mgr.Save(testingStats)
	<-time.After(200 * time.Millisecond)

	mgr.SetJobStatus("fake_job_ID", "running")
	<-time.After(100 * time.Millisecond)
	stats, err := mgr.Retrieve("fake_job_ID")
	if err != nil {
		t.Fatal(err)
	}

	if stats.Stats.Status != "running" {
		t.Fatalf("expect job status 'running' but got '%s'\n", stats.Stats.Status)
	}

	key := utils.KeyJobStats(testingNamespace, "fake_job_ID")
	if err := clear(key, redisPool.Get()); err != nil {
		t.Fatal(err)
	}
}

func TestCommand(t *testing.T) {
	mgr := createStatsManager(redisPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)

	if err := mgr.SendCommand("fake_job_ID", CtlCommandStop); err != nil {
		t.Fatal(err)
	}

	if cmd, err := mgr.CtlCommand("fake_job_ID"); err != nil {
		t.Fatal(err)
	} else {
		if cmd != CtlCommandStop {
			t.Fatalf("expect '%s' but got '%s'", CtlCommandStop, cmd)
		}
	}
}

func TestDieAt(t *testing.T) {
	mgr := createStatsManager(redisPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)

	testingStats := createFakeStats()
	mgr.Save(testingStats)

	dieAt := time.Now().Unix()
	if err := createDeadJob(redisPool.Get(), dieAt); err != nil {
		t.Fatal(err)
	}
	<-time.After(200 * time.Millisecond)
	mgr.DieAt("fake_job_ID", dieAt)
	<-time.After(300 * time.Millisecond)

	stats, err := mgr.Retrieve("fake_job_ID")
	if err != nil {
		t.Fatal(err)
	}

	if stats.Stats.DieAt != dieAt {
		t.Fatalf("expect die at '%d' but got '%d'\n", dieAt, stats.Stats.DieAt)
	}

	key := utils.KeyJobStats(testingNamespace, "fake_job_ID")
	if err := clear(key, redisPool.Get()); err != nil {
		t.Fatal(err)
	}
	key2 := utils.RedisKeyDead(testingNamespace)
	if err := clear(key2, redisPool.Get()); err != nil {
		t.Fatal(err)
	}
}

func TestRegisterHook(t *testing.T) {
	mgr := createStatsManager(redisPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)

	if err := mgr.RegisterHook("fake_job_ID", "http://localhost:9999", false); err != nil {
		t.Fatal(err)
	}

	key := utils.KeyJobStats(testingNamespace, "fake_job_ID")
	if err := clear(key, redisPool.Get()); err != nil {
		t.Fatal(err)
	}
}

func TestExpireJobStats(t *testing.T) {
	mgr := createStatsManager(redisPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)

	// make sure data existing
	testingStats := createFakeStats()
	mgr.Save(testingStats)
	<-time.After(200 * time.Millisecond)

	if err := mgr.ExpirePeriodicJobStats("fake_job_ID"); err != nil {
		t.Fatal(err)
	}

	key := utils.KeyJobStats(testingNamespace, "fake_job_ID")
	if err := clear(key, redisPool.Get()); err != nil {
		t.Fatal(err)
	}
}

func TestCheckIn(t *testing.T) {
	mgr := createStatsManager(redisPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)

	// make sure data existing
	testingStats := createFakeStats()
	mgr.Save(testingStats)
	<-time.After(200 * time.Millisecond)

	// Start http server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		statusReport := &models.JobStatusChange{}
		if err := json.Unmarshal(data, statusReport); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if statusReport.Metadata == nil || statusReport.Metadata.JobID != "fake_job_ID" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		fmt.Fprintln(w, "ok")
	}))
	defer ts.Close()

	if err := mgr.RegisterHook("fake_job_ID", ts.URL, false); err != nil {
		t.Fatal(err)
	}

	mgr.CheckIn("fake_job_ID", "checkin")
	<-time.After(200 * time.Millisecond)

	stats, err := mgr.Retrieve("fake_job_ID")
	if err != nil {
		t.Fatal(err)
	}

	if stats.Stats.CheckIn != "checkin" {
		t.Fatalf("expect check in info 'checkin' but got '%s'\n", stats.Stats.CheckIn)
	}

	key := utils.KeyJobStats(testingNamespace, "fake_job_ID")
	if err := clear(key, redisPool.Get()); err != nil {
		t.Fatal(err)
	}
}

func TestExecutionRelated(t *testing.T) {
	mgr := createStatsManager(redisPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)

	if err := mgr.AttachExecution("upstream_id", "id1", "id2", "id3"); err != nil {
		t.Fatal(err)
	}

	// Wait for data is stable
	<-time.After(200 * time.Millisecond)
	ids, err := mgr.GetExecutions("upstream_id")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Join(ids, "/") != "id1/id2/id3" {
		t.Fatalf("expect 'id1/id2/id3' but got %s", strings.Join(ids, " / "))
	}
}

func TestUpdateJobStats(t *testing.T) {
	mgr := createStatsManager(redisPool)
	mgr.Start()
	defer mgr.Shutdown()
	<-time.After(200 * time.Millisecond)

	// make sure data existing
	testingStats := createFakeStats()
	mgr.Save(testingStats)
	<-time.After(200 * time.Millisecond)

	mgr.Update("fake_job_ID", "status", "Error")
	<-time.After(200 * time.Millisecond)

	updatedStats, err := mgr.Retrieve("fake_job_ID")
	if err != nil {
		t.Fatal(err)
	}

	if updatedStats.Stats.Status != "Error" {
		t.Fatalf("expect status to be '%s' but got '%s'", "Error", updatedStats.Stats.Status)
	}
}

func getRedisHost() string {
	redisHost := os.Getenv(testingRedisHost)
	if redisHost == "" {
		redisHost = "localhost" // for local test
	}

	return redisHost
}

func createStatsManager(redisPool *redis.Pool) JobStatsManager {
	ctx := context.Background()
	return NewRedisJobStatsManager(ctx, testingNamespace, redisPool)
}

func clear(key string, conn redis.Conn) error {
	if conn != nil {
		defer conn.Close()
		_, err := conn.Do("DEL", key)
		return err
	}

	return errors.New("failed to clear")
}

func createFakeStats() models.JobStats {
	testingStats := models.JobStats{
		Stats: &models.JobStatData{
			JobID:       "fake_job_ID",
			JobKind:     job.JobKindPeriodic,
			JobName:     "fake_job",
			Status:      "Pending",
			IsUnique:    false,
			RefLink:     "/api/v1/jobs/fake_job_ID",
			CronSpec:    "5 * * * * *",
			EnqueueTime: time.Now().Unix(),
			UpdateTime:  time.Now().Unix(),
		},
	}

	return testingStats
}

func createDeadJob(conn redis.Conn, dieAt int64) error {
	dead := make(map[string]interface{})
	dead["name"] = "fake_job"
	dead["id"] = "fake_job_ID"
	dead["args"] = make(map[string]interface{})
	dead["fails"] = 3
	dead["err"] = "testing error"
	dead["failed_at"] = dieAt

	rawJSON, err := json.Marshal(&dead)
	if err != nil {
		return err
	}

	defer conn.Close()
	key := utils.RedisKeyDead(testingNamespace)
	_, err = conn.Do("ZADD", key, dieAt, rawJSON)
	return err
}
