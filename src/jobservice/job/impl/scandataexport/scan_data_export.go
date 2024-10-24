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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
	"github.com/opencontainers/go-digest"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	"github.com/goharbor/harbor/src/pkg/systemartifact"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
	"github.com/goharbor/harbor/src/pkg/task"
)

// ScanDataExport is the struct to implement the scan data export.
// implements the Job interface
type ScanDataExport struct {
	execMgr               task.ExecutionManager
	scanDataExportDirPath string
	exportMgr             export.Manager
	digestCalculator      export.ArtifactDigestCalculator
	filterProcessor       export.FilterProcessor
	vulnDataSelector      export.VulnerabilityDataSelector
	projectMgr            project.Manager
	sysArtifactMgr        systemartifact.Manager
}

func (sde *ScanDataExport) MaxFails() uint {
	return 1
}

// MaxCurrency of the job. Unlike the WorkerPool concurrency, it controls the limit on the number jobs of that type
// that can be active at one time by within a single redis instance.
// The default value is 0, which means "no limit on job concurrency".
func (sde *ScanDataExport) MaxCurrency() uint {
	return 1
}

// ShouldRetry tells worker if retry the failed job when the fails is
// still less that the number declared by the method 'MaxFails'.
//
// Returns:
//
//	true for retry and false for none-retry
func (sde *ScanDataExport) ShouldRetry() bool {
	return true
}

// Validate Indicate whether the parameters of job are valid.
// Return:
// error if parameters are not valid. NOTES: If no parameters needed, directly return nil.
func (sde *ScanDataExport) Validate(_ job.Parameters) error {
	return nil
}

// Run the business logic here.
// The related arguments will be injected by the workerpool.
//
// ctx Context                   : Job execution context.
// params map[string]interface{} : parameters with key-pair style for the job execution.
//
// Returns:
//
//	error if failed to run. NOTES: If job is stopped or cancelled, a specified error should be returned
func (sde *ScanDataExport) Run(ctx job.Context, params job.Parameters) error {
	if _, ok := params[export.JobModeKey]; !ok {
		return errors.Errorf("no mode specified for scan data export execution")
	}

	mode := params[export.JobModeKey].(string)
	logger := ctx.GetLogger()
	logger.Infof("Scan data export job started in mode : %v", mode)
	sde.init()
	fileName := fmt.Sprintf("%s/scandata_export_%s.csv", sde.scanDataExportDirPath, params[export.JobID])

	// ensure that CSV files are cleared post the completion of the Run.
	defer sde.cleanupCsvFile(ctx, fileName, params)
	err := sde.writeCsvFile(ctx, params, fileName)
	if err != nil {
		logger.Errorf("error when writing data to CSV: %v", err)
		return err
	}

	hash, err := sde.calculateFileHash(fileName)
	if err != nil {
		logger.Errorf("Error when calculating checksum for generated file: %v", err)
		return err
	}
	logger.Infof("Export Job Id = %s, FileName = %s, Hash = %v", params[export.JobID], fileName, hash)

	csvFile, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		logger.Errorf(
			"Export Job Id = %s. Error when moving report file %s to persistent storage: %v", params[export.JobID], fileName, err)
		return err
	}
	baseFileName := filepath.Base(fileName)
	repositoryName := strings.TrimSuffix(baseFileName, filepath.Ext(baseFileName))
	logger.Infof("Creating repository for CSV file with blob : %s", repositoryName)
	stat, err := os.Stat(fileName)
	if err != nil {
		logger.Errorf("Error when fetching file size: %v", err)
		return err
	}
	logger.Infof("Export Job Id = %s. CSV file size: %d", params[export.JobID], stat.Size())
	// earlier return and update status message if the file size is 0, unnecessary to push a empty system artifact.
	if stat.Size() == 0 {
		extra := map[string]interface{}{
			export.StatusMessageAttribute: "No vulnerabilities found or matched",
		}
		updateErr := sde.updateExecAttributes(ctx, params, extra)
		if updateErr != nil {
			logger.Errorf("Export Job Id = %s. Error when updating the exec extra attributes 'status_message' to 'No vulnerabilities found or matched': %v", params[export.JobID], updateErr)
		}

		logger.Infof("Export Job Id = %s. Exported CSV file is empty, skip to push system artifact, exit job", params[export.JobID])
		return nil
	}

	csvExportArtifactRecord := model.SystemArtifact{Repository: repositoryName, Digest: hash.String(), Size: stat.Size(), Type: "ScanData_CSV", Vendor: strings.ToLower(export.Vendor)}
	artID, err := sde.sysArtifactMgr.Create(ctx.SystemContext(), &csvExportArtifactRecord, csvFile)
	if err != nil {
		logger.Errorf(
			"Export Job Id = %s. Error when persisting report file %s to persistent storage: %v", params[export.JobID], fileName, err)
		return err
	}

	logger.Infof("Export Job Id = %s. Created system artifact: %v for report file %s to persistent storage: %v", params[export.JobID], artID, fileName, err)
	err = sde.updateExecAttributes(ctx, params, map[string]interface{}{export.DigestKey: hash.String()})
	if err != nil {
		logger.Errorf("Export Job Id = %s. Error when updating execution record : %v", params[export.JobID], err)
		return err
	}
	logger.Info("Scan data export job completed")

	return nil
}

func (sde *ScanDataExport) updateExecAttributes(ctx job.Context, params job.Parameters, attrs map[string]interface{}) error {
	logger := ctx.GetLogger()
	execID, err := strconv.ParseInt(params[export.JobID].(string), 10, 64)
	if err != nil {
		logger.Errorf("Export Job Id = %s. Error when parse execution id from params: %v", params[export.JobID], err)
		return err
	}

	exec, err := sde.execMgr.Get(ctx.SystemContext(), execID)
	if err != nil {
		logger.Errorf("Export Job Id = %s. Error when fetching execution record for update : %v", params[export.JobID], err)
		return err
	}
	// copy old extra
	attrsToUpdate := exec.ExtraAttrs
	for k, v := range attrs {
		attrsToUpdate[k] = v
	}
	return sde.execMgr.UpdateExtraAttrs(ctx.SystemContext(), execID, attrsToUpdate)
}

func (sde *ScanDataExport) writeCsvFile(ctx job.Context, params job.Parameters, fileName string) error {
	logger := ctx.GetLogger()
	csvFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)
	if err != nil {
		logger.Errorf("Failed to create CSV export file %s. Error : %v", fileName, err)
		return err
	}
	defer csvFile.Close()

	logger.Infof("Created CSV export file %s", csvFile.Name())

	systemContext := ctx.SystemContext()
	var exportParams export.Params
	var artIDGroups [][]int64

	if criteria, ok := params[export.JobRequest]; ok {
		logger.Infof("Request for export : %v", criteria)
		filterCriteria, err := sde.extractCriteria(params)
		if err != nil {
			return err
		}

		projectIDs := filterCriteria.Projects
		if len(projectIDs) == 0 {
			return nil
		}

		// extract the repository ids if any repositories have been specified
		repoIDs, err := sde.filterProcessor.ProcessRepositoryFilter(systemContext, filterCriteria.Repositories, projectIDs)
		if err != nil {
			return err
		}

		if len(repoIDs) == 0 {
			logger.Infof("No repositories found with specified names: %v", filterCriteria.Repositories)
			return nil
		}

		// filter artifacts by tags
		arts, err := sde.filterProcessor.ProcessTagFilter(systemContext, filterCriteria.Tags, repoIDs)
		if err != nil {
			return err
		}

		if len(arts) == 0 {
			logger.Infof("No artifacts found with specified names: %v and tags: %v", filterCriteria.Repositories, filterCriteria.Tags)
			return nil
		}

		// filter artifacts by labels
		arts, err = sde.filterProcessor.ProcessLabelFilter(systemContext, filterCriteria.Labels, arts)
		if err != nil {
			return err
		}

		if len(arts) == 0 {
			logger.Infof("No artifacts found with specified labels: %v", filterCriteria.Labels)
			return nil
		}

		size := export.ArtifactGroupSize
		artIDGroups = make([][]int64, len(arts)/size+1)
		for i, art := range arts {
			// group artIDs to improve performance and avoid spliced sql over
			// max length
			artIDGroups[i/size] = append(artIDGroups[i/size], art.ID)
		}

		exportParams = export.Params{
			CVEIds: filterCriteria.CVEIds,
		}
	}

	for groupID, artIDGroup := range artIDGroups {
		// fetch data by group
		if len(artIDGroup) == 0 {
			continue
		}

		exportParams.ArtifactIDs = artIDGroup
		exportParams.PageNumber = 1
		exportParams.PageSize = export.QueryPageSize

		for {
			data, err := sde.exportMgr.Fetch(systemContext, exportParams)
			if err != nil {
				logger.Error("Encountered error reading from the report table", err)
				return err
			}
			if len(data) == 0 {
				logger.Infof("No more data to fetch. Exiting...")
				break
			}
			logger.Infof("Export Group Id = %d, Job Id = %s, Page Number = %d, Page Size = %d Num Records = %d", groupID, params[export.JobID], exportParams.PageNumber, exportParams.PageSize, len(data))

			// for the first page write the CSV with the headers
			if exportParams.PageNumber == 1 && groupID == 0 {
				err = gocsv.Marshal(data, csvFile)
			} else {
				err = gocsv.MarshalWithoutHeaders(data, csvFile)
			}
			if err != nil {
				return nil
			}

			exportParams.PageNumber = exportParams.PageNumber + 1
			// break earlier if this is last page
			if len(data) < int(exportParams.PageSize) {
				break
			}
		}
	}
	return nil
}

func (sde *ScanDataExport) extractCriteria(params job.Parameters) (*export.Request, error) {
	filterMap, ok := params[export.JobRequest].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("malformed criteria '%v'", params[export.JobRequest])
	}
	jsonData, err := json.Marshal(filterMap)
	if err != nil {
		return nil, err
	}
	criteria := &export.Request{}
	err = criteria.FromJSON(string(jsonData))

	if err != nil {
		return nil, err
	}

	// sterilize trim spaces for some fields.
	sterilize := func(c *export.Request) *export.Request {
		if c != nil {
			space, empty := " ", ""
			c.Repositories = strings.ReplaceAll(c.Repositories, space, empty)
			c.Tags = strings.ReplaceAll(c.Tags, space, empty)
			c.CVEIds = strings.ReplaceAll(c.CVEIds, space, empty)
		}

		return c
	}

	return sterilize(criteria), nil
}

func (sde *ScanDataExport) calculateFileHash(fileName string) (digest.Digest, error) {
	return sde.digestCalculator.Calculate(fileName)
}

func (sde *ScanDataExport) init() {
	if sde.execMgr == nil {
		sde.execMgr = task.NewExecutionManager()
	}

	if sde.scanDataExportDirPath == "" {
		sde.scanDataExportDirPath = export.ScanDataExportDir
	}

	if sde.exportMgr == nil {
		sde.exportMgr = export.NewManager()
	}

	if sde.digestCalculator == nil {
		sde.digestCalculator = &export.SHA256ArtifactDigestCalculator{}
	}

	if sde.filterProcessor == nil {
		sde.filterProcessor = export.NewFilterProcessor()
	}

	if sde.vulnDataSelector == nil {
		sde.vulnDataSelector = export.NewVulnerabilityDataSelector()
	}

	if sde.projectMgr == nil {
		sde.projectMgr = project.New()
	}

	if sde.sysArtifactMgr == nil {
		sde.sysArtifactMgr = systemartifact.Mgr
	}
}

func (sde *ScanDataExport) cleanupCsvFile(ctx job.Context, fileName string, params job.Parameters) {
	logger := ctx.GetLogger()
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		logger.Infof("Export Job Id = %s, CSV Export File = %s does not exist. Nothing to do", params[export.JobID], fileName)
		return
	}
	err := os.Remove(fileName)
	if err != nil {
		logger.Errorf("Export Job Id = %s, CSV Export File = %s could not deleted. Error = %v", params[export.JobID], fileName, err)
		return
	}
}
