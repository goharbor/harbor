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
	"encoding/json"
	"testing"
	"time"

	"github.com/gocraft/work"

	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/goharbor/harbor/src/jobservice/utils"
)

func TestSweeper(t *testing.T) {
	epoch := time.Now().Unix() - 1000
	if err := createFakeScheduledJob(epoch); err != nil {
		t.Fatal(err)
	}
	ns := tests.GiveMeTestNamespace()
	sweeper := NewSweeper(ns, redisPool, work.NewClient(ns, redisPool))
	if err := sweeper.ClearOutdatedScheduledJobs(); err != nil {
		t.Fatal(err)
	}
	err := tests.Clear(utils.RedisKeyScheduled(ns), redisPool.Get())
	if err != nil {
		t.Fatal(err)
	}
}

func createFakeScheduledJob(runAt int64) error {
	fakeJob := make(map[string]interface{})
	fakeJob["name"] = "fake_periodic_job"
	fakeJob["id"] = "fake_job_id"
	fakeJob["t"] = runAt
	fakeJob["args"] = make(map[string]interface{})

	rawJSON, err := json.Marshal(&fakeJob)
	if err != nil {
		return err
	}

	conn := redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("ZADD", utils.RedisKeyScheduled(tests.GiveMeTestNamespace()), runAt, rawJSON)
	return err
}
