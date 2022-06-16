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

var (
	cleanupController = Ctl
)

func init() {
	if err := scheduler.RegisterCallbackFunc(SystemArtifactCleanupCallback, cleanupCallBack); err != nil {
		log.Fatalf("failed to register the callback for the system artifact cleanup schedule, error %v", err)
	}
}

func cleanupCallBack(ctx context.Context, param string) error {
	err := cleanupController.Start(ctx, true, task.ExecutionTriggerSchedule)
	if err != nil {
		logger.Errorf("System artifact cleanup job encountered errors: %v", err)
	}
	return err
}
