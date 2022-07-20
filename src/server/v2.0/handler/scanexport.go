package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/scandataexport"
	"github.com/goharbor/harbor/src/jobservice/logger"
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
		sysArtifactMgr:    systemartifact.Mgr,
		userMgr:           user.Mgr,
	}
}

type scanDataExportAPI struct {
	BaseAPI
	scanDataExportCtl scandataexport.Controller
	sysArtifactMgr    systemartifact.Manager
	userMgr           user.Manager
}

func (se *scanDataExportAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

func (se *scanDataExportAPI) ExportScanData(ctx context.Context, params operation.ExportScanDataParams) middleware.Responder {
	if err := se.RequireAuthenticated(ctx); err != nil {
		return se.SendError(ctx, err)
	}

	// check if the MIME type for the export is the Generic vulnerability data
	if params.XScanDataType != v1.MimeTypeGenericVulnerabilityReport {
		error := &models.Error{Message: fmt.Sprintf("Unsupported MIME type : %s", params.XScanDataType)}
		errors := &models.Errors{Errors: []*models.Error{error}}
		return operation.NewExportScanDataBadRequest().WithPayload(errors)
	}

	// loop through the list of projects and validate that scan privilege and create privilege
	// is available for all projects
	// TODO : Should we just ignore projects that do not have the required level of access?

	projects := params.Criteria.Projects
	for _, project := range projects {
		if err := se.RequireProjectAccess(ctx, project, rbac.ActionCreate, rbac.ResourceScan); err != nil {
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
		error := &models.Error{Message: fmt.Sprintf("User : %s not found", secContext.GetUsername())}
		errors := &models.Errors{Errors: []*models.Error{error}}
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
	err := se.RequireAuthenticated(ctx)
	if err != nil {
		return se.SendError(ctx, err)
	}
	execution, err := se.scanDataExportCtl.GetExecution(ctx, params.ExecutionID)
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
		JobName:     execution.JobName,
		UserName:    execution.UserName,
		FilePresent: execution.FilePresent,
	}

	return operation.NewGetScanDataExportExecutionOK().WithPayload(&sdeExec)
}

func (se *scanDataExportAPI) DownloadScanData(ctx context.Context, params operation.DownloadScanDataParams) middleware.Responder {
	err := se.RequireAuthenticated(ctx)
	if err != nil {
		return se.SendError(ctx, err)
	}
	execution, err := se.scanDataExportCtl.GetExecution(ctx, params.ExecutionID)

	if err != nil {
		if notFound := orm.AsNotFoundError(err, "execution with id: %d not found", params.ExecutionID); notFound != nil {
			return middleware.ResponderFunc(func(writer http.ResponseWriter, producer runtime.Producer) {
				writer.WriteHeader(http.StatusNotFound)
			})
		}
		return se.SendError(ctx, err)
	}

	// check if the CSV artifact for the execution exists
	if !execution.FilePresent {
		return middleware.ResponderFunc(func(writer http.ResponseWriter, producer runtime.Producer) {
			writer.WriteHeader(http.StatusNotFound)
		})
	}

	// check if the execution being downloaded is owned by the current user
	secContext, err := se.GetSecurityContext(ctx)
	if err != nil {
		return se.SendError(ctx, err)
	}

	if secContext.GetUsername() != execution.UserName {
		return middleware.ResponderFunc(func(writer http.ResponseWriter, producer runtime.Producer) {
			writer.WriteHeader(http.StatusForbidden)
		})
	}

	repositoryName := fmt.Sprintf("scandata_export_%v", params.ExecutionID)
	file, err := se.sysArtifactMgr.Read(ctx, strings.ToLower(export.Vendor), repositoryName, execution.ExportDataDigest)
	if err != nil {
		return se.SendError(ctx, err)
	}
	logger.Infof("reading data from file : %s", repositoryName)

	return middleware.ResponderFunc(func(writer http.ResponseWriter, producer runtime.Producer) {
		defer se.cleanUpArtifact(ctx, repositoryName, execution.ExportDataDigest, params.ExecutionID, file)

		writer.Header().Set("Content-Type", "text/csv")
		writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fmt.Sprintf("%s.csv", repositoryName)))
		nbytes, err := io.Copy(writer, file)
		if err != nil {
			logger.Errorf("Encountered error while copying data: %v", err)
		} else {
			logger.Debugf("Copied %v bytes from file to client", nbytes)
		}
	})
}

func (se *scanDataExportAPI) GetScanDataExportExecutionList(ctx context.Context, params operation.GetScanDataExportExecutionListParams) middleware.Responder {
	err := se.RequireAuthenticated(ctx)
	if err != nil {
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
			JobName:     execution.JobName,
			FilePresent: execution.FilePresent,
		}
		execs = append(execs, sdeExec)
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
	logger.Infof("Deleting report artifact : %v:%v", repositoryName, digest)

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
