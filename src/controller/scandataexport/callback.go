package scandataexport

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	ExportDataCleanupCallback = "SCAN_EXPORT_CLEANUP"
)

type TriggerParam struct {
	TimeWindowMinutes int
	PageSize          int
}

func init() {
	if err := scheduler.RegisterCallbackFunc(ExportDataCleanupCallback, cleanupCallBack); err != nil {
		log.Fatalf("failed to register the callback for the scan all schedule, error %v", err)
	}
}

func cleanupCallBack(ctx context.Context, param string) error {
	triggerParams := TriggerParam{}
	if param != "" {
		err := json.Unmarshal([]byte(param), triggerParams)
		if err != nil {
			return err
		}
	}

	err := Ctl.StartCleanup(ctx, task.ExecutionTriggerSchedule, triggerParams, true)
	logger.Errorf("Export data artifact cleanup job encountered errors: %v", err)
	return err
}
