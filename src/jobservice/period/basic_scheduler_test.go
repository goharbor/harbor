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
package period

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/opm"

	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/tests"
)

var redisPool = tests.GiveMeRedisPool()

func TestScheduler(t *testing.T) {
	statsManager := opm.NewRedisJobStatsManager(context.Background(), tests.GiveMeTestNamespace(), redisPool)
	statsManager.Start()
	defer statsManager.Shutdown()

	scheduler := myPeriodicScheduler(statsManager)
	params := make(map[string]interface{})
	params["image"] = "testing:v1"
	id, runAt, err := scheduler.Schedule("fake_job", params, "5 * * * * *")
	if err != nil {
		t.Fatal(err)
	}

	if time.Now().Unix() >= runAt {
		t.Fatal("the running at time of scheduled job should be after now, but seems not")
	}

	if err := scheduler.Load(); err != nil {
		t.Fatal(err)
	}

	if scheduler.pstore.size() != 1 {
		t.Fatalf("expect 1 item in pstore but got '%d'\n", scheduler.pstore.size())
	}

	if err := scheduler.UnSchedule(id); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.Clear(); err != nil {
		t.Fatal(err)
	}

	err = tests.Clear(utils.KeyPeriodicPolicy(tests.GiveMeTestNamespace()), redisPool.Get())
	err = tests.Clear(utils.KeyPeriodicPolicyScore(tests.GiveMeTestNamespace()), redisPool.Get())
	err = tests.Clear(utils.KeyPeriodicNotification(tests.GiveMeTestNamespace()), redisPool.Get())
	if err != nil {
		t.Fatal(err)
	}
}

func TestPubFunc(t *testing.T) {
	statsManager := opm.NewRedisJobStatsManager(context.Background(), tests.GiveMeTestNamespace(), redisPool)
	statsManager.Start()
	defer statsManager.Shutdown()

	scheduler := myPeriodicScheduler(statsManager)
	p := &PeriodicJobPolicy{
		PolicyID: "fake_ID",
		JobName:  "fake_job",
		CronSpec: "5 * * * * *",
	}
	if err := scheduler.AcceptPeriodicPolicy(p); err != nil {
		t.Fatal(err)
	}
	if scheduler.pstore.size() != 1 {
		t.Fatalf("expect 1 item in pstore but got '%d' after accepting \n", scheduler.pstore.size())
	}
	if rmp := scheduler.RemovePeriodicPolicy("fake_ID"); rmp == nil {
		t.Fatal("expect none nil object returned after removing but got nil")
	}
	if scheduler.pstore.size() != 0 {
		t.Fatalf("expect 0 item in pstore but got '%d' \n", scheduler.pstore.size())
	}
}

func myPeriodicScheduler(statsManager opm.JobStatsManager) *basicScheduler {
	sysCtx := context.Background()
	ctx := &env.Context{
		SystemContext: sysCtx,
		WG:            new(sync.WaitGroup),
		ErrorChan:     make(chan error, 1),
	}

	return NewScheduler(ctx, tests.GiveMeTestNamespace(), redisPool, statsManager)
}
