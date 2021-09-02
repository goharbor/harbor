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

package gc

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

func init() {
	err := scheduler.RegisterCallbackFunc(SchedulerCallback, gcCallback)
	if err != nil {
		log.Fatalf("failed to registry GC call back, %v", err)
	}

	if err := task.RegisterTaskStatusChangePostFunc(GCVendorType, gcTaskStatusChange); err != nil {
		log.Fatalf("failed to register the task status change post for the gc job, error %v", err)
	}
}

func gcCallback(ctx context.Context, p string) error {
	param := &Policy{}
	if err := json.Unmarshal([]byte(p), param); err != nil {
		return fmt.Errorf("failed to unmarshal the param: %v", err)
	}
	_, err := Ctl.Start(ctx, *param, task.ExecutionTriggerSchedule)
	return err
}

func gcTaskStatusChange(ctx context.Context, taskID int64, status string) error {
	if status == job.SuccessStatus.String() && config.QuotaPerProjectEnable(ctx) {
		go func() {
			quota.RefreshForProjects(orm.Context())
		}()
	}

	return nil
}
