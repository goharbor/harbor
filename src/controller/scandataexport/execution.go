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

package scandataexport

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	q2 "github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	"github.com/goharbor/harbor/src/pkg/systemartifact"
	"github.com/goharbor/harbor/src/pkg/task"
)

var Ctl = NewController()

type Controller interface {
	Start(ctx context.Context, criteria export.Request) (executionID int64, err error)
	GetExecution(ctx context.Context, executionID int64) (*export.Execution, error)
	ListExecutions(ctx context.Context, userName string) ([]*export.Execution, error)
	GetTask(ctx context.Context, executionID int64) (*task.Task, error)
	DeleteExecution(ctx context.Context, executionID int64) error
}

func NewController() Controller {
	return &controller{
		execMgr:        task.ExecMgr,
		taskMgr:        task.Mgr,
		makeCtx:        orm.Context,
		sysArtifactMgr: systemartifact.Mgr,
	}
}

type controller struct {
	execMgr        task.ExecutionManager
	taskMgr        task.Manager
	makeCtx        func() context.Context
	sysArtifactMgr systemartifact.Manager
}

func (c *controller) ListExecutions(ctx context.Context, userName string) ([]*export.Execution, error) {
	keywords := make(map[string]interface{})
	keywords["VendorType"] = job.ScanDataExportVendorType
	keywords[fmt.Sprintf("ExtraAttrs.%s", export.UserNameAttribute)] = userName

	q := q2.New(q2.KeyWords{})
	q.Keywords = keywords
	execsForUser, err := c.execMgr.List(ctx, q)
	if err != nil {
		return nil, err
	}
	execs := make([]*export.Execution, 0)
	for _, execForUser := range execsForUser {
		execs = append(execs, c.convertToExportExecStatus(ctx, execForUser))
	}
	return execs, nil
}

func (c *controller) GetTask(ctx context.Context, executionID int64) (*task.Task, error) {
	logger := log.GetLogger(ctx)
	query := q2.New(q2.KeyWords{})

	keywords := make(map[string]interface{})
	keywords["VendorType"] = job.ScanDataExportVendorType
	keywords["ExecutionID"] = executionID
	query.Keywords = keywords
	query.Sorts = append(query.Sorts, &q2.Sort{
		Key:  "ID",
		DESC: true,
	})
	tasks, err := c.taskMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.Errorf("No task found for execution Id : %d", executionID)
	}
	// for the export JOB there would be a single instance of the task corresponding to the execution
	// we will hence return the latest instance of the task associated with this execution
	logger.Infof("Returning task instance with ID : %d", tasks[0].ID)
	return tasks[0], nil
}

func (c *controller) GetExecution(ctx context.Context, executionID int64) (*export.Execution, error) {
	logger := log.GetLogger(ctx)
	exec, err := c.execMgr.Get(ctx, executionID)
	if err != nil {
		logger.Errorf("Error when fetching execution status for ExecutionId: %d error : %v", executionID, err)
		return nil, err
	}
	if exec == nil {
		logger.Infof("No execution found for ExecutionId: %d", executionID)
		return nil, nil
	}
	return c.convertToExportExecStatus(ctx, exec), nil
}

func (c *controller) DeleteExecution(ctx context.Context, executionID int64) error {
	logger := log.GetLogger(ctx)
	err := c.execMgr.Delete(ctx, executionID)
	if err != nil {
		logger.Errorf("Error when deleting execution  for ExecutionId: %d, error : %v", executionID, err)
	}
	return err
}

func (c *controller) Start(ctx context.Context, request export.Request) (executionID int64, err error) {
	logger := log.GetLogger(ctx)
	vendorID := int64(ctx.Value(export.CsvJobVendorIDKey).(int))
	extraAttrs := make(map[string]interface{})
	extraAttrs[export.ProjectIDsAttribute] = request.Projects
	extraAttrs[export.JobNameAttribute] = request.JobName
	extraAttrs[export.UserNameAttribute] = request.UserName
	id, err := c.execMgr.Create(ctx, job.ScanDataExportVendorType, vendorID, task.ExecutionTriggerManual, extraAttrs)
	logger.Infof("Created an execution record with id : %d for vendorID: %d", id, vendorID)
	if err != nil {
		logger.Errorf("Encountered error when creating job : %v", err)
		return 0, err
	}

	// create a job object and fill with metadata and parameters
	params := make(map[string]interface{})
	params[export.JobID] = fmt.Sprintf("%d", id)
	params[export.JobRequest] = request
	params[export.JobModeKey] = export.JobModeExport

	j := &task.Job{
		Name: job.ScanDataExportVendorType,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
	}

	_, err = c.taskMgr.Create(ctx, id, j)

	if err != nil {
		logger.Errorf("Unable to create a scan data export job: %v", err)
		c.markError(ctx, id, err)
		return 0, err
	}

	logger.Info("Created job for scan data export successfully")
	return id, nil
}

func (c *controller) markError(ctx context.Context, executionID int64, err error) {
	logger := log.GetLogger(ctx)
	// try to stop the execution first in case that some tasks are already created
	if err := c.execMgr.StopAndWait(ctx, executionID, 10*time.Second); err != nil {
		logger.Errorf("failed to stop the execution %d: %v", executionID, err)
	}
	if err := c.execMgr.MarkError(ctx, executionID, err.Error()); err != nil {
		logger.Errorf("failed to mark error for the execution %d: %v", executionID, err)
	}
}

func (c *controller) convertToExportExecStatus(ctx context.Context, exec *task.Execution) *export.Execution {
	execStatus := &export.Execution{
		ID:            exec.ID,
		UserID:        exec.VendorID,
		Status:        exec.Status,
		StatusMessage: exec.StatusMessage,
		Trigger:       exec.Trigger,
		StartTime:     exec.StartTime,
		EndTime:       exec.EndTime,
	}
	if pids, ok := exec.ExtraAttrs[export.ProjectIDsAttribute]; ok {
		for _, pid := range pids.([]interface{}) {
			execStatus.ProjectIDs = append(execStatus.ProjectIDs, int64(pid.(float64)))
		}
	}
	if digest, ok := exec.ExtraAttrs[export.DigestKey]; ok {
		execStatus.ExportDataDigest = digest.(string)
	}
	if jobName, ok := exec.ExtraAttrs[export.JobNameAttribute]; ok {
		execStatus.JobName = jobName.(string)
	}
	if userName, ok := exec.ExtraAttrs[export.UserNameAttribute]; ok {
		execStatus.UserName = userName.(string)
	}
	if statusMessage, ok := exec.ExtraAttrs[export.StatusMessageAttribute]; ok {
		execStatus.StatusMessage = statusMessage.(string)
	}

	if len(execStatus.ExportDataDigest) > 0 {
		artifactExists := c.isCsvArtifactPresent(ctx, exec.ID, execStatus.ExportDataDigest)
		execStatus.FilePresent = artifactExists
	}

	return execStatus
}

func (c *controller) isCsvArtifactPresent(ctx context.Context, execID int64, digest string) bool {
	logger := log.GetLogger(ctx)
	repositoryName := fmt.Sprintf("scandata_export_%v", execID)
	exists, err := c.sysArtifactMgr.Exists(ctx, strings.ToLower(export.Vendor), repositoryName, digest)
	if err != nil {
		logger.Errorf("failed to check existence of csv artifact for vendor: %s repository: %s digest: %s",
			strings.ToLower(export.Vendor), repositoryName, digest)
		exists = false
	}
	return exists
}
