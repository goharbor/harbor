package systemartifact

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	SystemArtifactCleanupCallback = "SYSTEM_ARTIFACT_CLEANUP"
)

func init() {
	if err := scheduler.RegisterCallbackFunc(SystemArtifactCleanupCallback, cleanupCallBack); err != nil {
		log.Fatalf("failed to register the callback for the scan all schedule, error %v", err)
	}
}

func cleanupCallBack(ctx context.Context, param string) error {

	err := Ctl.Start(ctx, true, task.ExecutionTriggerSchedule)
	logger.Errorf("System artifact cleanup job encountered errors: %v", err)
	return err
}
