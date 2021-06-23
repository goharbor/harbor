package gc

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/lib/config"

	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/jobservice/job"
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
	_, err := Ctl.Start(orm.Context(), *param, task.ExecutionTriggerSchedule)
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
