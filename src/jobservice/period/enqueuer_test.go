// Copyright 2018 The Harbor Authors. All rights reserved.
package period

import (
	"sync"
	"testing"
	"time"

	"github.com/vmware/harbor/src/jobservice/tests"
	"github.com/vmware/harbor/src/jobservice/utils"
)

func TestPeriodicEnqueuerStartStop(t *testing.T) {
	ns := tests.GiveMeTestNamespace()
	ps := &periodicJobPolicyStore{
		lock:     new(sync.RWMutex),
		policies: make(map[string]*PeriodicJobPolicy),
	}
	enqueuer := newPeriodicEnqueuer(ns, redisPool, ps)
	enqueuer.start()
	<-time.After(100 * time.Millisecond)
	enqueuer.stop()
}

func TestEnqueue(t *testing.T) {
	ns := tests.GiveMeTestNamespace()

	pl := &PeriodicJobPolicy{
		PolicyID: "fake_ID",
		JobName:  "fake_name",
		CronSpec: "5 * * * * *",
	}
	ps := &periodicJobPolicyStore{
		lock:     new(sync.RWMutex),
		policies: make(map[string]*PeriodicJobPolicy),
	}
	ps.add(pl)

	enqueuer := newPeriodicEnqueuer(ns, redisPool, ps)
	if err := enqueuer.enqueue(); err != nil {
		t.Error(err)
	}

	err := tests.Clear(utils.RedisKeyScheduled(ns), redisPool.Get())
	err = tests.Clear(utils.KeyJobStats(ns, "fake_ID"), redisPool.Get())
	err = tests.Clear(utils.RedisKeyLastPeriodicEnqueue(ns), redisPool.Get())
	if err != nil {
		t.Error(err)
	}
}
