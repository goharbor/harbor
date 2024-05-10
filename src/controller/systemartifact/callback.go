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

func cleanupCallBack(ctx context.Context, _ string) error {
	err := cleanupController.Start(ctx, true, task.ExecutionTriggerSchedule)
	if err != nil {
		logger.Errorf("System artifact cleanup job encountered errors: %v", err)
	}
	return err
}
