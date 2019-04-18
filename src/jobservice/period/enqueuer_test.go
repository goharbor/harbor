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
	"github.com/goharbor/harbor/src/jobservice/tests"
)

func TestPeriodicEnqueuerStartStop(t *testing.T) {
	ns := tests.GiveMeTestNamespace()
	ps := &periodicJobPolicyStore{
		lock:     new(sync.RWMutex),
		policies: make(map[string]*PeriodicJobPolicy),
	}
	enqueuer := newEnqueuer(ns, redisPool, ps, nil)
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

	statsManager := opm.NewRedisJobStatsManager(context.Background(), ns, redisPool)
	statsManager.Start()
	defer statsManager.Shutdown()

	enqueuer := newEnqueuer(ns, redisPool, ps, statsManager)
	if err := enqueuer.enqueue(); err != nil {
		t.Error(err)
	}

	if err := clear(ns); err != nil {
		t.Error(err)
	}
}

func clear(ns string) error {
	err := tests.Clear(utils.RedisKeyScheduled(ns), redisPool.Get())
	err = tests.Clear(utils.KeyJobStats(ns, "fake_ID"), redisPool.Get())
	err = tests.Clear(utils.RedisKeyLastPeriodicEnqueue(ns), redisPool.Get())
	if err != nil {
		return err
	}

	return nil
}
