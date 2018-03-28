// Copyright 2018 The Harbor Authors. All rights reserved.
package period

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/tests"
	"github.com/vmware/harbor/src/jobservice_v2/utils"
)

var redisPool = tests.GiveMeRedisPool()

func TestScheduler(t *testing.T) {
	scheduler := myPeriodicScheduler()
	params := make(map[string]interface{})
	params["image"] = "testing:v1"
	id, runAt, err := scheduler.Schedule("fake_job", params, "5 * * * * *")
	if err != nil {
		t.Error(err)
	}

	if time.Now().Unix() >= runAt {
		t.Error("the running at time of scheduled job should be after now, but seems not")
	}

	if err := scheduler.Load(); err != nil {
		t.Error(err)
	}

	if scheduler.pstore.size() != 1 {
		t.Errorf("expect 1 item in pstore but got '%d'\n", scheduler.pstore.size())
	}

	if err := scheduler.UnSchedule(id); err != nil {
		t.Error(err)
	}
	if err := scheduler.Clear(); err != nil {
		t.Error(err)
	}

	err = tests.Clear(utils.KeyPeriodicPolicy(tests.GiveMeTestNamespace()), redisPool.Get())
	err = tests.Clear(utils.KeyPeriodicPolicyScore(tests.GiveMeTestNamespace()), redisPool.Get())
	err = tests.Clear(utils.KeyPeriodicNotification(tests.GiveMeTestNamespace()), redisPool.Get())
	if err != nil {
		t.Error(err)
	}
}

func TestPubFunc(t *testing.T) {
	scheduler := myPeriodicScheduler()
	p := &PeriodicJobPolicy{
		PolicyID: "fake_ID",
		JobName:  "fake_job",
		CronSpec: "5 * * * * *",
	}
	if err := scheduler.AcceptPeriodicPolicy(p); err != nil {
		t.Error(err)
	}
	if scheduler.pstore.size() != 1 {
		t.Errorf("expect 1 item in pstore but got '%d' after accepting \n", scheduler.pstore.size())
	}
	if rmp := scheduler.RemovePeriodicPolicy("fake_ID"); rmp == nil {
		t.Error("expect none nil object returned after removing but got nil")
	}
	if scheduler.pstore.size() != 0 {
		t.Errorf("expect 0 item in pstore but got '%d' \n", scheduler.pstore.size())
	}
}

func myPeriodicScheduler() *RedisPeriodicScheduler {
	sysCtx := context.Background()
	ctx := &env.Context{
		SystemContext: sysCtx,
		WG:            new(sync.WaitGroup),
		ErrorChan:     make(chan error, 1),
	}

	return NewRedisPeriodicScheduler(ctx, tests.GiveMeTestNamespace(), redisPool)
}
