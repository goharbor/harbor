// Copyright 2018 The Harbor Authors. All rights reserved.
package opm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/vmware/harbor/src/jobservice/job"
	"github.com/vmware/harbor/src/jobservice/models"
	"github.com/vmware/harbor/src/jobservice/utils"
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
	//make sure data existing
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

	//make sure data existing
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

	//make sure data existing
	testingStats := createFakeStats()
	mgr.Save(testingStats)
	<-time.After(200 * time.Millisecond)

	//Start http server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func getRedisHost() string {
	redisHost := os.Getenv(testingRedisHost)
	if redisHost == "" {
		redisHost = "10.160.178.186" //for local test
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
