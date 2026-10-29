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

package optimizer

import (
	"context"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	dockerfileoptdao "github.com/goharbor/harbor/src/pkg/dockerfileoptimization/dao"
	"github.com/goharbor/harbor/src/pkg/task"
)

var (
	robotCtl = robot.Ctl
	taskMgr  = task.Mgr
	execMgr  = task.ExecMgr
	optDAO   = dockerfileoptdao.New()
)

func init() {
	if err := task.RegisterTaskStatusChangePostFunc(job.OptimizeArtifactVendorType, optimizeTaskStatusChange); err != nil {
		log.Fatalf("failed to register the task status change post for the optimize job, error %v", err)
	}
}

// optimizeTaskStatusChange cleans up the single-use robot account when the optimize
// task reaches a final status and, as a safety net, marks the dockerfile_optimization
// row as errored if the job died before persisting a terminal state.
func optimizeTaskStatusChange(ctx context.Context, taskID int64, status string) (err error) {
	logger := log.G(ctx).WithFields(log.Fields{"task_id": taskID, "status": status})

	js := job.Status(status)
	if !js.Final() {
		return nil
	}

	t, err := taskMgr.Get(ctx, taskID)
	if err != nil {
		return err
	}

	exec, err := execMgr.Get(ctx, t.ExecutionID)
	if err != nil {
		return err
	}

	if robotID := getRobotID(t.ExtraAttrs); robotID > 0 {
		if err := robotCtl.Delete(ctx, robotID, &robot.Option{Operator: secret.JobserviceUser}); err != nil {
			// Should not block the main flow, just logged
			logger.WithFields(log.Fields{"robot_id": robotID, "error": err}).Error("delete robot account failed")
		} else {
			logger.WithField("robot_id", robotID).Debug("Robot account for the optimize task is removed")
		}
	}

	// Safety net: the job normally persists a terminal status itself. If the task
	// ended without the row leaving Pending/Running (job panicked, pod died, ...),
	// mark it errored so the portal does not poll forever.
	if js == job.ErrorStatus || js == job.StoppedStatus {
		repo, digest := getArtifactCoords(exec.ExtraAttrs)
		if repo != "" && digest != "" {
			rec, err := optDAO.GetByArtifact(ctx, repo, digest)
			if err == nil && (rec.Status == dockerfileoptdao.StatusPending || rec.Status == dockerfileoptdao.StatusRunning) {
				if err := optDAO.UpdateStatus(ctx, repo, digest, dockerfileoptdao.StatusError,
					"optimize job finished with status "+status); err != nil {
					logger.WithField("error", err).Error("failed to mark optimization record errored")
				}
			}
		}
	}

	return nil
}

func getRobotID(extraAttrs map[string]any) int64 {
	if extraAttrs == nil {
		return 0
	}
	if v, ok := extraAttrs[robotIDKey]; ok {
		if f, ok := v.(float64); ok {
			return int64(f)
		}
		if i, ok := v.(int64); ok {
			return i
		}
	}
	return 0
}

func getArtifactCoords(extraAttrs map[string]any) (repo string, digest string) {
	if extraAttrs == nil {
		return "", ""
	}
	art, ok := extraAttrs["artifact"].(map[string]any)
	if !ok {
		return "", ""
	}
	repo, _ = art["repository_name"].(string)
	digest, _ = art["digest"].(string)
	return repo, digest
}
