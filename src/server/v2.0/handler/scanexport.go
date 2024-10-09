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

package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/scandataexport"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/systemartifact"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/scan_data_export"
)

func newScanDataExportAPI() *scanDataExportAPI {
	return &scanDataExportAPI{
		scanDataExportCtl: scandataexport.Ctl,
		proCtl:            project.Ctl,
		sysArtifactMgr:    systemartifact.Mgr,
		userMgr:           user.Mgr,
	}
}

type scanDataExportAPI struct {
	BaseAPI
	scanDataExportCtl scandataexport.Controller
	proCtl            project.Controller
	sysArtifactMgr    systemartifact.Manager
	userMgr           user.Manager
}

func (se *scanDataExportAPI) Prepare(_ context.Context, _ string, _ interface{}) middleware.Responder {
	return nil
}

func (se *scanDataExportAPI) ExportScanData(ctx context.Context, params operation.ExportScanDataParams) middleware.Responder {
	// validate the request params
	if err := se.validateScanExportParams(ctx, params); err != nil {
		return se.SendError(ctx, err)
	}

	for _, pid := range params.Criteria.Projects {
		if err := se.RequireProjectAccess(ctx, pid, rbac.ActionCreate, rbac.ResourceExportCVE); err != nil {
			return se.SendError(ctx, err)
		}
	}

	scanDataExportJob := new(models.ScanDataExportJob)

	secContext, err := se.GetSecurityContext(ctx)

	if err != nil {
		return se.SendError(ctx, err)
	}

	// vendor id associated with the job == the user id
	usr, err := se.userMgr.GetByName(ctx, secContext.GetUsername())

	if err != nil {
		return se.SendError(ctx, err)
	}

	if usr == nil {
		err := &models.Error{Message: fmt.Sprintf("User : %s not found", secContext.GetUsername())}
		errors := &models.Errors{Errors: []*models.Error{err}}
		return operation.NewExportScanDataForbidden().WithPayload(errors)
	}

	userContext := context.WithValue(ctx, export.CsvJobVendorIDKey, usr.UserID)

	if err != nil {
		return se.SendError(ctx, err)
	}

	jobID, err := se.scanDataExportCtl.Start(userContext, se.convertToCriteria(params.Criteria, secContext.GetUsername(), usr.UserID))
	if err != nil {
		return se.SendError(ctx, err)
	}
	scanDataExportJob.ID = jobID
	return operation.NewExportScanDataOK().WithPayload(scanDataExportJob)
}

func (se *scanDataExportAPI) GetScanDataExportExecution(ctx context.Context, params operation.GetScanDataExportExecutionParams) middleware.Responder {
	if err := se.RequireAuthenticated(ctx); err != nil {
		return se.SendError(ctx, err)
	}

	execution, err := se.scanDataExportCtl.GetExecution(ctx, params.ExecutionID)
	if err != nil {
		return se.SendError(ctx, err)
	}

	// check the permission by project ids in execution
	if err = se.requireProjectsAccess(ctx, execution.ProjectIDs, rbac.ActionRead, rbac.ResourceExportCVE); err != nil {
		return se.SendError(ctx, err)
	}

	// check if the execution being fetched is owned by the current user
	secContext, err := se.GetSecurityContext(ctx)
	if err != nil {
		return se.SendError(ctx, err)
	}
	if secContext.GetUsername() != execution.UserName {
		err = errors.New(nil).WithCode(errors.ForbiddenCode)
		return se.SendError(ctx, err)
	}

	if err != nil {
		return se.SendError(ctx, err)
	}
	sdeExec := models.ScanDataExportExecution{
		EndTime:     strfmt.DateTime(execution.EndTime),
		ID:          execution.ID,
		StartTime:   strfmt.DateTime(execution.StartTime),
		Status:      execution.Status,
		StatusText:  execution.StatusMessage,
		Trigger:     execution.Trigger,
		UserID:      execution.UserID,
		UserName:    execution.UserName,
		FilePresent: execution.FilePresent,
	}
	// add human friendly message when status is error
	if sdeExec.Status == job.ErrorStatus.String() && sdeExec.StatusText == "" {
		sdeExec.StatusText = "Please contact the system administrator to check the logs of jobservice."
	}

	return operation.NewGetScanDataExportExecutionOK().WithPayload(&sdeExec)
}

func (se *scanDataExportAPI) DownloadScanData(ctx context.Context, params operation.DownloadScanDataParams) middleware.Responder {
	if err := se.RequireAuthenticated(ctx); err != nil {
		return se.SendError(ctx, err)
	}

	execution, err := se.scanDataExportCtl.GetExecution(ctx, params.ExecutionID)
	if err != nil {
		if notFound := orm.AsNotFoundError(err, "execution with id: %d not found", params.ExecutionID); notFound != nil {
			return middleware.ResponderFunc(func(writer http.ResponseWriter, _ runtime.Producer) {
				writer.WriteHeader(http.StatusNotFound)
			})
		}
		return se.SendError(ctx, err)
	}

	// check the permission by project ids in execution
	if err = se.requireProjectsAccess(ctx, execution.ProjectIDs, rbac.ActionRead, rbac.ResourceExportCVE); err != nil {
		return se.SendError(ctx, err)
	}

	// check if the execution being downloaded is owned by the current user
	secContext, err := se.GetSecurityContext(ctx)
	if err != nil {
		return se.SendError(ctx, err)
	}

	if secContext.GetUsername() != execution.UserName {
		return middleware.ResponderFunc(func(writer http.ResponseWriter, _ runtime.Producer) {
			writer.WriteHeader(http.StatusForbidden)
		})
	}

	// check if the CSV artifact for the execution exists
	if !execution.FilePresent {
		return middleware.ResponderFunc(func(writer http.ResponseWriter, _ runtime.Producer) {
			writer.WriteHeader(http.StatusNotFound)
		})
	}

	repositoryName := fmt.Sprintf("scandata_export_%v", params.ExecutionID)
	file, err := se.sysArtifactMgr.Read(ctx, strings.ToLower(export.Vendor), repositoryName, execution.ExportDataDigest)
	if err != nil {
		return se.SendError(ctx, err)
	}
	log.Infof("reading data from file : %s", repositoryName)

	return middleware.ResponderFunc(func(writer http.ResponseWriter, _ runtime.Producer) {
		defer se.cleanUpArtifact(ctx, repositoryName, execution.ExportDataDigest, params.ExecutionID, file)

		writer.Header().Set("Content-Type", "text/csv")
		writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fmt.Sprintf("%s.csv", repositoryName)))
		nbytes, err := io.Copy(writer, file)
		if err != nil {
			log.Errorf("Encountered error while copying data: %v", err)
		} else {
			log.Debugf("Copied %v bytes from file to client", nbytes)
		}
	})
}

func (se *scanDataExportAPI) GetScanDataExportExecutionList(ctx context.Context, _ operation.GetScanDataExportExecutionListParams) middleware.Responder {
	if err := se.RequireAuthenticated(ctx); err != nil {
		return se.SendError(ctx, err)
	}

	secContext, err := se.GetSecurityContext(ctx)
	if err != nil {
		return se.SendError(ctx, err)
	}

	executions, err := se.scanDataExportCtl.ListExecutions(ctx, secContext.GetUsername())
	if err != nil {
		return se.SendError(ctx, err)
	}
	// projectSet store the unrepeated project ids
	projectSet := make(map[int64]struct{})
	execs := make([]*models.ScanDataExportExecution, 0)
	for _, execution := range executions {
		sdeExec := &models.ScanDataExportExecution{
			EndTime:     strfmt.DateTime(execution.EndTime),
			ID:          execution.ID,
			StartTime:   strfmt.DateTime(execution.StartTime),
			Status:      execution.Status,
			StatusText:  execution.StatusMessage,
			Trigger:     execution.Trigger,
			UserID:      execution.UserID,
			UserName:    execution.UserName,
			FilePresent: execution.FilePresent,
		}
		// add human friendly message when status is error
		if sdeExec.Status == job.ErrorStatus.String() && sdeExec.StatusText == "" {
			sdeExec.StatusText = "Please contact the system administrator to check the logs of jobservice."
		}
		// store project ids
		for _, pid := range execution.ProjectIDs {
			projectSet[pid] = struct{}{}
		}

		execs = append(execs, sdeExec)
	}

	// convert projectSet to pids
	var pids []int64
	for pid := range projectSet {
		pids = append(pids, pid)
	}
	// check the permission by project ids in execution
	if err = se.requireProjectsAccess(ctx, pids, rbac.ActionRead, rbac.ResourceExportCVE); err != nil {
		return se.SendError(ctx, err)
	}

	sdeExecList := models.ScanDataExportExecutionList{Items: execs}
	return operation.NewGetScanDataExportExecutionListOK().WithPayload(&sdeExecList)
}

func (se *scanDataExportAPI) convertToCriteria(requestCriteria *models.ScanDataExportRequest, userName string, userID int) export.Request {
	return export.Request{
		UserID:       userID,
		UserName:     userName,
		JobName:      requestCriteria.JobName,
		CVEIds:       requestCriteria.CVEIds,
		Labels:       requestCriteria.Labels,
		Projects:     requestCriteria.Projects,
		Repositories: requestCriteria.Repositories,
		Tags:         requestCriteria.Tags,
	}
}

func (se *scanDataExportAPI) cleanUpArtifact(ctx context.Context, repositoryName, digest string, execID int64, file io.ReadCloser) {
	file.Close()
	log.Infof("Deleting report artifact : %v:%v", repositoryName, digest)

	// the entire delete operation is executed within a transaction to ensure that any failures
	// during the blob creation or tracking record creation result in a rollback of the transaction
	vendor := strings.ToLower(export.Vendor)
	err := orm.WithTransaction(func(ctx context.Context) error {
		err := se.sysArtifactMgr.Delete(ctx, vendor, repositoryName, digest)
		if err != nil {
			log.Errorf("Error deleting system artifact record for %s/%s/%s: %v", vendor, repositoryName, digest, err)
			return err
		}
		// delete the underlying execution
		err = se.scanDataExportCtl.DeleteExecution(ctx, execID)
		if err != nil {
			log.Errorf("Error deleting csv export job execution for %s/%s/%s: %v", vendor, repositoryName, digest, err)
		}
		return err
	})(ctx)

	if err != nil {
		log.Errorf("Error deleting system artifact record for %s/%s/%s: %v", vendor, repositoryName, digest, err)
	}
}

func (se *scanDataExportAPI) requireProjectsAccess(ctx context.Context, pids []int64, action rbac.Action, subresource ...rbac.Resource) error {
	// check project permission one by one, return error if any project cannot
	// access permission.
	for _, pid := range pids {
		if err := se.RequireProjectAccess(ctx, pid, action, subresource...); err != nil {
			return err
		}
	}

	return nil
}

// validateScanExportParams validates scan data export request parameters by
// following policies.
// rules:
//  1. check the scan data type
//  2. the criteria should not be empty
//  3. currently only the export of single project is open
//  4. check the existence of project
//  5. do not allow to input space in the repo/tag/cve_id (space will lead to misjudge for doublestar filter)
func (se *scanDataExportAPI) validateScanExportParams(ctx context.Context, params operation.ExportScanDataParams) error {
	// check if the MIME type for the export is the Generic vulnerability data
	if params.XScanDataType != v1.MimeTypeGenericVulnerabilityReport {
		return errors.BadRequestError(errors.Errorf("Unsupported MIME type : %s", params.XScanDataType))
	}

	criteria := params.Criteria
	if criteria == nil {
		return errors.BadRequestError(errors.Errorf("criteria is invalid: %v", criteria))
	}

	// validate project id, currently we only support single project
	if len(criteria.Projects) != 1 {
		return errors.BadRequestError(errors.Errorf("only support export single project, invalid value: %v", criteria.Projects))
	}

	// check whether the project exists
	exist, err := se.proCtl.Exists(ctx, criteria.Projects[0])
	if err != nil {
		return errors.UnknownError(errors.Errorf("check the existence of project error: %v", err))
	}

	if !exist {
		return errors.NotFoundError(errors.Errorf("project %d not found", criteria.Projects[0]))
	}

	// check spaces
	space := " "
	inspectList := []string{criteria.Repositories, criteria.Tags, criteria.CVEIds}
	for _, s := range inspectList {
		if strings.Contains(s, space) {
			return errors.BadRequestError(errors.Errorf("invalid criteria value, please remove additional spaces for input: %s", s))
		}
	}

	return nil
}
