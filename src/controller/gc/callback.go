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
	err := scheduler.RegisterCallbackFunc(job.GarbageCollectionVendorType, gcCallback)
	if err != nil {
		log.Fatalf("failed to registry GC call back, %v", err)
	}

	if err := task.RegisterTaskStatusChangePostFunc(job.GarbageCollectionVendorType, gcTaskStatusChange); err != nil {
		log.Fatalf("failed to register the task status change post for the garbage collection job, error %v", err)
	}

	if err := task.RegisterCheckInProcessor(job.GarbageCollectionVendorType, gcCheckIn); err != nil {
		log.Fatalf("failed to register the checkin processor for the garbage collection job, error %v", err)
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

func gcTaskStatusChange(ctx context.Context, _ int64, status string) error {
	if status == job.SuccessStatus.String() && config.QuotaPerProjectEnable(ctx) {
		go func() {
			err := quota.RefreshForProjects(orm.Context())
			if err != nil {
				log.Warningf("failed to refresh project quota, error: %v", err)
			}
		}()
	}

	return nil
}

func gcCheckIn(ctx context.Context, t *task.Task, sc *job.StatusChange) error {
	taskID := t.ID
	status := t.Status

	log.Infof("received garbage collection task status update event: task-%d, status-%s", taskID, status)
	if sc.CheckIn != "" {
		var gcObj struct {
			SweepSize int64 `json:"freed_space"`
			Blobs     int64 `json:"purged_blobs"`
			Manifests int64 `json:"purged_manifests"`
		}
		if err := json.Unmarshal([]byte(sc.CheckIn), &gcObj); err != nil {
			log.Errorf("failed to resolve checkin of garbage collection task %d: %v", taskID, err)

			return err
		}
		t, err := task.Mgr.Get(ctx, taskID)
		if err != nil {
			return err
		}

		e, err := task.ExecMgr.Get(ctx, t.ExecutionID)
		if err != nil {
			return err
		}

		e.ExtraAttrs["freed_space"] = gcObj.SweepSize
		e.ExtraAttrs["purged_blobs"] = gcObj.Blobs
		e.ExtraAttrs["purged_manifests"] = gcObj.Manifests

		err = task.ExecMgr.UpdateExtraAttrs(ctx, e.ID, e.ExtraAttrs)
		if err != nil {
			log.G(ctx).WithField("error", err).Errorf("failed to update of garbage collection task %d", taskID)
			return err
		}
	}
	return nil
}
