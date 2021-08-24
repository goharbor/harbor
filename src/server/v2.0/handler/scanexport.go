package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/scandataexport"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/scan_data_export"
	"io"
	"net/http"
)

func newScanDataExportApi() *scanDataExportAPI {
	return &scanDataExportAPI{
		scanDataExportCtl: scandataexport.Ctl,
		regCli:            registry.Cli,
	}
}

type scanDataExportAPI struct {
	BaseAPI
	scanDataExportCtl scandataexport.Controller
	regCli            registry.Client
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

	/*
		projects := params.Criteria.Projects
		for _, project := range projects {
			if err := se.RequireProjectAccess(ctx, project, rbac.ActionCreate, rbac.ResourceScan); err != nil {
				return se.SendError(ctx, err)
			}
		}*/

	// TODO: Obtain an execution Id and create a URL based on the actual execution ID

	scanDataExportJob := new(models.ScanDataExportJob)
	// create a random vendorId since the export job is not tied to
	// is not associated with the state of any other scan report entity

	userContext := context.WithValue(ctx, "vendorId", -1)
	jobId, err := se.scanDataExportCtl.Start(userContext, se.convertToCriteria(params.Criteria))
	if err != nil {
		return se.SendError(ctx, err)
	}
	scanDataExportJob.ID = jobId
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
		EndTime:    strfmt.DateTime(execution.EndTime),
		ID:         execution.ID,
		StartTime:  strfmt.DateTime(execution.StartTime),
		Status:     execution.Status,
		StatusText: execution.StatusMessage,
		Trigger:    execution.Trigger,
		UserID:     execution.UserID,
	}

	return operation.NewGetScanDataExportExecutionOK().WithPayload(&sdeExec)
}

func (se *scanDataExportAPI) DownloadScanData(ctx context.Context, params operation.DownloadScanDataParams) middleware.Responder {
	err := se.RequireAuthenticated(ctx)
	if err != nil {
		return se.SendError(ctx, err)
	}
	execution, err := se.scanDataExportCtl.GetExecution(ctx, params.ExecutionID)

	repositoryName := fmt.Sprintf("scandata_export_%v", params.ExecutionID)
	size, file, err := se.regCli.PullBlob(repositoryName, execution.ExportDataDigest)
	// file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return se.SendError(ctx, err)
	}
	logger.Infof("reading data from file : %s with size : %d", repositoryName, size)

	return middleware.ResponderFunc(func(writer http.ResponseWriter, producer runtime.Producer) {
		defer se.cleanUpArtifact(repositoryName, execution.ExportDataDigest)
		defer file.Close()
		writer.Header().Set("Content-Type", "text/csv")
		writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fmt.Sprintf("%s.csv", repositoryName)))
		nbytes, err := io.Copy(writer, file)
		if err != nil {
			logger.Errorf("Encountered error while copying data: %v", err)
		} else {
			logger.Infof("Copied %v bytes from file to client", nbytes)
		}

	})
}

func (se *scanDataExportAPI) convertToCriteria(requestCriteria *models.ScanDataExportCriteria) export.Criteria {
	return export.Criteria{
		CVEIds:       requestCriteria.CVEIds,
		Labels:       requestCriteria.Labels,
		Projects:     requestCriteria.Projects,
		Repositories: requestCriteria.Repositories,
	}
}

func (se *scanDataExportAPI) cleanUpArtifact(repositoryName string, digest string) error {
	logger.Infof("Deleting report artifact : %v:%v", repositoryName, digest)
	return se.regCli.DeleteBlob(repositoryName, digest)
}
