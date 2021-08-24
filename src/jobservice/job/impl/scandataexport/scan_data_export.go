package scandataexport

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/gocarina/gocsv"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/opencontainers/go-digest"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	vulnScanReportView = "vuln_scan_report"
	scanDataExportDir  = "/var/scandata_exports"
	pageSize           = 100
	DigestKey          = "artifact_digest"
	CreateTimestampKey = "create_ts"
)

// ScanDataExport is the struct to implement the scan data export.
// implements the Job interface
type ScanDataExport struct {
	execMgr               task.ExecutionManager
	scanDataExportDirPath string
	exportMgr             export.Manager
	regCli                registry.Client
	digestCalculator      ArtifactDigestCalculator
}

func (sde *ScanDataExport) MaxFails() uint {
	return 1
}

// MaxCurrency of the job. Unlike the WorkerPool concurrency, it controls the limit on the number jobs of that type
// that can be active at one time by within a single redis instance.
// The default value is 0, which means "no limit on job concurrency".
func (sde *ScanDataExport) MaxCurrency() uint {
	return 0
}

// ShouldRetry tells worker if retry the failed job when the fails is
// still less that the number declared by the method 'MaxFails'.
//
// Returns:
//  true for retry and false for none-retry
func (sde *ScanDataExport) ShouldRetry() bool {
	return true
}

// Validate Indicate whether the parameters of job are valid.
// Return:
// error if parameters are not valid. NOTES: If no parameters needed, directly return nil.
func (sde *ScanDataExport) Validate(params job.Parameters) error {
	return nil
}

// Run the business logic here.
// The related arguments will be injected by the workerpool.
//
// ctx Context                   : Job execution context.
// params map[string]interface{} : parameters with key-pair style for the job execution.
//
// Returns:
//  error if failed to run. NOTES: If job is stopped or cancelled, a specified error should be returned
//
func (sde *ScanDataExport) Run(ctx job.Context, params job.Parameters) error {
	logger.Info("Scan data export job started")
	sde.init()
	fileName := fmt.Sprintf("%s/scandata_export_%v.csv", sde.scanDataExportDirPath, params["JobId"])
	err := sde.writeCsvFile(ctx, params, fileName)
	if err != nil {
		logger.Errorf("Error when writing data to CSV: %v", err)
		return err
	}

	hash, err := sde.calculateFileHash(fileName)
	if err != nil {
		logger.Errorf("Error when calculating checksum for generated file: %v", err)
		return err
	}
	logger.Infof("Export Job Id = %v, FileName = %s, Hash = %v", params["JobId"], fileName, hash)

	csvFile, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		logger.Errorf(
			"Export Job Id = %v. Error when moving report file %s to persistent storage: %v", params["JobId"], fileName, err)
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
	err = sde.regCli.PushBlob(repositoryName, hash.String(), stat.Size(), csvFile)
	if err != nil {
		logger.Errorf(
			"Export Job Id = %v. Error when persisting report file %s to persistent storage: %v", params["JobId"], fileName, err)
		return err
	}

	err = sde.execMgr.UpdateExtraAttrs(ctx.SystemContext(), int64(params["JobId"].(float64)), map[string]interface{}{DigestKey: hash.String(), CreateTimestampKey: float64(time.Now().Unix())})
	if err != nil {
		logger.Errorf("Export Job Id = %v. Error when updating execution record : %v", params["JobId"], err)
		return err
	}
	logger.Info("Scan data export job completed")

	return nil
}

func (sde *ScanDataExport) writeCsvFile(ctx job.Context, params job.Parameters, fileName string) error {
	csvFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, os.ModePerm)

	defer csvFile.Close()

	if err != nil {
		logger.Errorf("Failed to create CSV export file %s. Error : %v", fileName, err)
		return err
	}
	logger.Infof("Created CSV export file %s", csvFile.Name())

	var exportParams export.Params

	if criteira, ok := params["Criteria"]; ok {
		logger.Infof("Criteria for export : %v", criteira)
		filterCriteria, err := sde.extractCriteria(params)
		if err != nil {
			return err
		}

		exportParams = export.Params{
			Projects:     filterCriteria.Projects,
			Repositories: filterCriteria.Repositories,
			CVEIds:       filterCriteria.CVEIds,
			Tags:         filterCriteria.Tags,
			Labels:       filterCriteria.Labels,
		}
	} else {
		exportParams = export.Params{
			Projects:     nil,
			Repositories: nil,
			CVEIds:       nil,
			Tags:         nil,
		}
	}

	exportParams.PageNumber = 1
	exportParams.PageSize = pageSize

	for {
		data, err := sde.exportMgr.Fetch(ctx.SystemContext(), exportParams)
		if err != nil {
			logger.Error("Encountered error reading from the report table", err)
			return err
		}
		if len(data) == 0 {
			logger.Infof("No more data to fetch. Exiting...")
			break
		}
		logger.Infof("Export Job Id = %v, Page Number = %d, Page Size = %d Num Records = %d", params["JobId"], exportParams.PageNumber, exportParams.PageSize, len(data))
		// for the first page write the CSV with the headers
		if exportParams.PageNumber == 1 {
			err = gocsv.Marshal(data, csvFile)
		} else {
			err = gocsv.MarshalWithoutHeaders(data, csvFile)
		}
		if err != nil {
			return nil
		}
		exportParams.PageNumber = exportParams.PageNumber + 1
	}
	return nil
}

func (sde *ScanDataExport) extractCriteria(params job.Parameters) (*export.Criteria, error) {
	filterMap, ok := params["Criteria"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("malformed criteria '%v'", params["Criteria"])
	}
	jsonData, err := json.Marshal(filterMap)
	if err != nil {
		return nil, err
	}
	criteria := &export.Criteria{}
	err = criteria.FromJSON(string(jsonData))

	if err != nil {
		return nil, err
	}
	return criteria, nil
}

func (sde *ScanDataExport) calculateFileHash(fileName string) (digest.Digest, error) {
	return sde.digestCalculator.Calculate(fileName)
}

func (sde *ScanDataExport) init() {
	if sde.execMgr == nil {
		sde.execMgr = task.NewExecutionManager()
	}

	if sde.scanDataExportDirPath == "" {
		sde.scanDataExportDirPath = scanDataExportDir
	}

	if sde.exportMgr == nil {
		sde.exportMgr = export.NewManager()
	}

	if sde.regCli == nil {
		sde.regCli = registry.Cli
	}

	if sde.digestCalculator == nil {
		sde.digestCalculator = &SHA256ArtifactDigestCalculator{}
	}
}

// ArtifactDigestCalculator is an interface to be implemented by all file hash calculators
type ArtifactDigestCalculator interface {
	// Calculate returns the hash for a file
	Calculate(fileName string) (digest.Digest, error)
}

type SHA256ArtifactDigestCalculator struct{}

func (calc *SHA256ArtifactDigestCalculator) Calculate(fileName string) (digest.Digest, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return "", err
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return digest.NewDigest(digest.SHA256, hash), nil
}
