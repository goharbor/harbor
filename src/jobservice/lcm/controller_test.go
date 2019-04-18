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

package lcm

import (
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"testing"
)

var (
	ns   = tests.GiveMeTestNamespace()
	pool = tests.GiveMeRedisPool()
	ctl  = NewController(ns, pool)
)

func TestLifeCycleController(t *testing.T) {
	conn := pool.Get()
	defer tests.ClearAll(ns, conn)

	// only mock status data
	jobID := "fake_job_ID_lcm_ctl"
	key := rds.KeyJobStats(ns, jobID)
	if err := setStatus(conn, key, job.PendingStatus); err != nil {
		t.Fatalf("mock data failed: %s\n", err.Error())
	}

	// Switch status one by one
	tk := ctl.Track(jobID)

	current, err := tk.Current()
	nilError(t, err)
	expect(t, job.PendingStatus, current)

	nilError(t, tk.Run())
	current, err = tk.Current()
	nilError(t, err)
	expect(t, job.RunningStatus, current)

	nilError(t, tk.Succeed())
	current, err = tk.Current()
	nilError(t, err)
	expect(t, job.SuccessStatus, current)

	if err := tk.Fail(); err == nil {
		t.Fatalf("expect non nil error but got nil when switch status from %s to %s", current, job.ErrorStatus)
	}
}

func expect(t *testing.T, expected job.Status, current job.Status) {
	if expected != current {
		t.Fatalf("expect status %s but got %s", expected, current)
	}
}

func nilError(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}
