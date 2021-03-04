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

package scan

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common/rbac"
	ar "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/robot"
	sc "github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	sca "github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/postprocessors"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/google/uuid"
)

// DefaultController is a default singleton scan API controller.
var DefaultController = NewController()

// const definitions
const (
	VendorTypeScanAll = "SCAN_ALL"

	configRegistryEndpoint = "registryEndpoint"
	configCoreInternalAddr = "coreInternalAddr"

	artfiactKey     = "artifact"
	registrationKey = "registration"

	artifactIDKey  = "artifact_id"
	reportUUIDsKey = "report_uuids"
	robotIDKey     = "robot_id"
)

func init() {
	// keep only the latest created 5 scan all execution records
	task.SetExecutionSweeperCount(VendorTypeScanAll, 5)
}

// uuidGenerator is a func template which is for generating UUID.
type uuidGenerator func() (string, error)

// configGetter is a func template which is used to wrap the config management
// utility methods.
type configGetter func(cfg string) (string, error)

// basicController is default implementation of api.Controller interface
type basicController struct {
	// Manage the scan report records
	manager report.Manager
	// Artifact controller
	ar ar.Controller
	// Scanner controller
	sc sc.Controller
	// Robot account controller
	rc robot.Controller
	// UUID generator
	uuid uuidGenerator
	// Configuration getter func
	config configGetter

	cloneCtx func(context.Context) context.Context
	makeCtx  func() context.Context

	execMgr task.ExecutionManager
	taskMgr task.Manager
	// Converter for V1 report to V2 report
	reportConverter postprocessors.NativeScanReportConverter
}

// NewController news a scan API controller
func NewController() Controller {
	return &basicController{
		// New report manager
		manager: report.NewManager(),
		// Refer to the default artifact controller
		ar: ar.Ctl,
		// Refer to the default scanner controller
		sc: sc.DefaultController,
		// Refer to the default robot account controller
		rc: robot.Ctl,
		// Generate UUID with uuid lib
		uuid: func() (string, error) {
			aUUID, err := uuid.NewUUID()
			if err != nil {
				return "", err
			}

			return aUUID.String(), nil
		},
		// Get the required configuration options
		config: func(cfg string) (string, error) {
			switch cfg {
			case configRegistryEndpoint:
				return config.ExtEndpoint()
			case configCoreInternalAddr:
				return config.InternalCoreURL(), nil
			default:
				return "", errors.Errorf("configuration option %s not defined", cfg)
			}
		},

		cloneCtx: orm.Clone,
		makeCtx:  orm.Context,

		execMgr: task.ExecMgr,
		taskMgr: task.Mgr,
		// Get the scan V1 to V2 report converters
		reportConverter: postprocessors.NewNativeToRelationalSchemaConverter(),
	}
}

// Collect artifacts itself or its children (exclude child which is image index and not supported by the scanner) when the artifact is scannable.
// Report placeholders will be created to track when scan the artifact.
// The reports of these artifacts will make together when get the reports of the artifact.
// There are two scenarios when artifact is scannable:
// 1. The scanner has capability for the artifact directly, eg the artifact is docker image.
// 2. The artifact is image index and the scanner has capability for any artifact which is referenced by the artifact.
func (bc *basicController) collectScanningArtifacts(ctx context.Context, r *scanner.Registration, artifact *ar.Artifact) ([]*ar.Artifact, bool, error) {
	var (
		scannable bool
		artifacts []*ar.Artifact
	)

	walkFn := func(a *ar.Artifact) error {
		supported := hasCapability(r, a)

		if !supported && a.IsImageIndex() {
			// image index not supported by the scanner, so continue to walk its children
			return nil
		}

		artifacts = append(artifacts, a)

		if supported {
			scannable = true
			return ar.ErrSkip // this artifact supported by the scanner, skip to walk its children
		}

		return nil
	}

	if err := bc.ar.Walk(ctx, artifact, walkFn, nil); err != nil {
		return nil, false, err
	}

	return artifacts, scannable, nil
}

// Scan ...
func (bc *basicController) Scan(ctx context.Context, artifact *ar.Artifact, options ...Option) error {
	if artifact == nil {
		return errors.New("nil artifact to scan")
	}

	r, err := bc.sc.GetRegistrationByProject(ctx, artifact.ProjectID)
	if err != nil {
		return errors.Wrap(err, "scan controller: scan")
	}

	// In case it does not exist
	if r == nil {
		return errors.PreconditionFailedError(nil).WithMessage("no available scanner for project: %d", artifact.ProjectID)
	}

	// Check if it is disabled
	if r.Disabled {
		return errors.PreconditionFailedError(nil).WithMessage("scanner %s is disabled", r.Name)
	}

	artifacts, scannable, err := bc.collectScanningArtifacts(ctx, r, artifact)
	if err != nil {
		return err
	}

	if !scannable {
		return errors.BadRequestError(nil).WithMessage("the configured scanner %s does not support scanning artifact with mime type %s", r.Name, artifact.ManifestMediaType)
	}

	type Param struct {
		Artifact *ar.Artifact
		Reports  []*scan.Report
	}

	params := []*Param{}

	var errs []error
	for _, art := range artifacts {
		reports, err := bc.makeReportPlaceholder(ctx, r, art)
		if err != nil {
			if errors.IsConflictErr(err) {
				errs = append(errs, err)
			} else {
				return err
			}
		}

		if len(reports) > 0 {
			params = append(params, &Param{Artifact: art, Reports: reports})
		}
	}

	// all report placeholder conflicted
	if len(errs) == len(artifacts) {
		return errs[0]
	}

	// Parse options
	opts, err := parseOptions(options...)
	if err != nil {
		return errors.Wrap(err, "scan controller: scan")
	}

	if opts.ExecutionID == 0 {
		extraAttrs := map[string]interface{}{
			artfiactKey: map[string]interface{}{
				"id":              artifact.ID,
				"project_id":      artifact.ProjectID,
				"repository_name": artifact.RepositoryName,
				"digest":          artifact.Digest,
			},
			registrationKey: map[string]interface{}{
				"id":   r.ID,
				"name": r.Name,
			},
		}
		executionID, err := bc.execMgr.Create(ctx, job.ImageScanJob, r.ID, task.ExecutionTriggerManual, extraAttrs)
		if err != nil {
			return err
		}

		opts.ExecutionID = executionID
	}

	errs = errs[:0]
	for _, param := range params {
		if err := bc.launchScanJob(ctx, opts.ExecutionID, param.Artifact, r, param.Reports); err != nil {
			log.G(ctx).Warningf("scan artifact %s@%s failed, error: %v", artifact.RepositoryName, artifact.Digest, err)
			errs = append(errs, err)
		}
	}

	// all scanning of the artifacts failed
	if len(errs) == len(params) {
		return fmt.Errorf("scan artifact %s@%s failed", artifact.RepositoryName, artifact.Digest)
	}

	return nil
}

func (bc *basicController) ScanAll(ctx context.Context, trigger string, async bool) (int64, error) {
	executionID, err := bc.execMgr.Create(ctx, VendorTypeScanAll, 0, trigger)
	if err != nil {
		return 0, err
	}

	if async {
		go func(ctx context.Context) {
			// if async, this is running in another goroutine ensure the execution exists in db
			err := lib.RetryUntil(func() error {
				_, err := bc.execMgr.Get(ctx, executionID)
				return err
			})
			if err != nil {
				log.Errorf("failed to get the execution %d for the scan all", executionID)
				return
			}

			bc.startScanAll(ctx, executionID)
		}(bc.makeCtx())
	} else {
		if err := bc.startScanAll(ctx, executionID); err != nil {
			return 0, err
		}
	}

	return executionID, nil
}

func (bc *basicController) startScanAll(ctx context.Context, executionID int64) error {
	batchSize := 50

	summary := struct {
		TotalCount        int `json:"total_count"`
		SubmitCount       int `json:"submit_count"`
		ConflictCount     int `json:"conflict_count"`
		PreconditionCount int `json:"precondition_count"`
		UnsupportCount    int `json:"unsupport_count"`
		UnknowCount       int `json:"unknow_count"`
	}{}

	for artifact := range ar.Iterator(ctx, batchSize, nil, nil) {
		summary.TotalCount++

		scan := func(ctx context.Context) error {
			return bc.Scan(ctx, artifact, WithExecutionID(executionID))
		}

		if err := orm.WithTransaction(scan)(ctx); err != nil {
			// Just logged
			log.Errorf("failed to scan artifact %s, error %v", artifact, err)

			switch errors.ErrCode(err) {
			case errors.ConflictCode:
				// a previous scan process is ongoing for the artifact
				summary.ConflictCount++
			case errors.PreconditionCode:
				// scanner not found or it's disabled
				summary.PreconditionCount++
			case errors.BadRequestCode:
				// artifact is unsupport
				summary.UnsupportCount++
			default:
				summary.UnknowCount++
			}
		} else {
			summary.SubmitCount++
		}
	}

	extraAttrs := map[string]interface{}{"summary": summary}
	if err := bc.execMgr.UpdateExtraAttrs(ctx, executionID, extraAttrs); err != nil {
		log.Errorf("failed to set the summary info for the scan all execution, error: %v", err)
		return err
	}

	if summary.SubmitCount > 0 { // at least one artifact submitted to the job service
		return nil
	}

	// not artifact found
	if summary.TotalCount == 0 {
		if err := bc.execMgr.MarkDone(ctx, executionID, "no artifact found"); err != nil {
			log.Errorf("failed to mark the execution %d to be done, error: %v", executionID, err)
			return err
		}
	} else if summary.PreconditionCount+summary.UnknowCount == 0 { // not scan job submitted and no failed
		message := fmt.Sprintf("%d artifact(s) found", summary.TotalCount)

		if summary.UnsupportCount > 0 {
			message = fmt.Sprintf("%s, %d artifact(s) not scannable", message, summary.UnsupportCount)
		}

		if summary.ConflictCount > 0 {
			message = fmt.Sprintf("%s, %d artifact(s) have a previous ongoing scan process", message, summary.ConflictCount)
		}

		message = fmt.Sprintf("%s, but no scan job submitted to the job service", message)

		if err := bc.execMgr.MarkDone(ctx, executionID, message); err != nil {
			log.Errorf("failed to mark the execution %d to be done, error: %v", executionID, err)
			return err
		}
	} else { // not scan job submitted and failed
		message := fmt.Sprintf("%d artifact(s) found", summary.TotalCount)

		if summary.PreconditionCount > 0 {
			message = fmt.Sprintf("%s, scanner not found or disabled for %d of them", message, summary.PreconditionCount)
		}

		if summary.UnknowCount > 0 {
			message = fmt.Sprintf("%s, internal error happened for %d of them", message, summary.UnknowCount)
		}

		message = fmt.Sprintf("%s, but no scan job submitted to the job service", message)
		if err := bc.execMgr.MarkError(ctx, executionID, message); err != nil {
			log.Errorf("failed to mark the execution %d to be error, error: %v", executionID, err)
			return err
		}
	}

	return nil
}

func (bc *basicController) makeReportPlaceholder(ctx context.Context, r *scanner.Registration, art *ar.Artifact) ([]*scan.Report, error) {
	mimeTypes := r.GetProducesMimeTypes(art.ManifestMediaType)

	oldReports, err := bc.manager.GetBy(bc.cloneCtx(ctx), art.Digest, r.UUID, mimeTypes)
	if err != nil {
		return nil, err
	}

	if err := bc.assembleReports(ctx, oldReports...); err != nil {
		return nil, err
	}

	if len(oldReports) > 0 {
		for _, oldReport := range oldReports {
			if !job.Status(oldReport.Status).Final() {
				return nil, errors.ConflictError(nil).WithMessage("a previous scan process is %s", oldReport.Status)
			}
		}

		for _, oldReport := range oldReports {
			if err := bc.manager.Delete(ctx, oldReport.UUID); err != nil {
				return nil, err
			}
		}
	}

	var reports []*scan.Report

	for _, pm := range r.GetProducesMimeTypes(art.ManifestMediaType) {
		report := &scan.Report{
			Digest:           art.Digest,
			RegistrationUUID: r.UUID,
			MimeType:         pm,
		}

		create := func(ctx context.Context) error {
			reportUUID, err := bc.manager.Create(ctx, report)
			if err != nil {
				return err
			}
			report.UUID = reportUUID

			return nil
		}

		if err := orm.WithTransaction(create)(ctx); err != nil {
			return nil, err
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// GetReport ...
func (bc *basicController) GetReport(ctx context.Context, artifact *ar.Artifact, mimeTypes []string) ([]*scan.Report, error) {
	if artifact == nil {
		return nil, errors.New("no way to get report for nil artifact")
	}

	mimes := make([]string, 0)
	mimes = append(mimes, mimeTypes...)
	if len(mimes) == 0 {
		// Retrieve native  and the new generic format as default
		mimes = append(mimes, v1.MimeTypeNativeReport, v1.MimeTypeGenericVulnerabilityReport)
	}

	// Get current scanner settings
	r, err := bc.sc.GetRegistrationByProject(ctx, artifact.ProjectID)
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: get report")
	}

	if r == nil {
		return nil, errors.NotFoundError(nil).WithMessage("no scanner registration configured for project: %d", artifact.ProjectID)
	}

	artifacts, scannable, err := bc.collectScanningArtifacts(ctx, r, artifact)
	if err != nil {
		return nil, err
	}

	if !scannable {
		return nil, errors.NotFoundError(nil).WithMessage("report not found for %s@%s", artifact.RepositoryName, artifact.Digest)
	}

	groupReports := make([][]*scan.Report, len(artifacts))

	var wg sync.WaitGroup
	for i, a := range artifacts {
		wg.Add(1)

		go func(i int, a *ar.Artifact) {
			defer wg.Done()

			reports, err := bc.manager.GetBy(bc.cloneCtx(ctx), a.Digest, r.UUID, mimes)
			if err != nil {
				log.Warningf("get reports of %s@%s failed, error: %v", a.RepositoryName, a.Digest, err)
				return
			}

			groupReports[i] = reports
		}(i, a)
	}
	wg.Wait()

	var reports []*scan.Report
	for _, group := range groupReports {
		if len(group) != 0 {
			reports = append(reports, group...)
		} else {
			// NOTE: If the artifact is OCI image, this happened when the artifact is not scanned,
			// but its children artifacts may scanned so return empty report
			return nil, nil
		}
	}

	if len(reports) == 0 {
		return nil, nil
	}

	if err := bc.assembleReports(ctx, reports...); err != nil {
		return nil, err
	}

	return reports, nil
}

// GetSummary ...
func (bc *basicController) GetSummary(ctx context.Context, artifact *ar.Artifact, mimeTypes []string, options ...report.Option) (map[string]interface{}, error) {
	if artifact == nil {
		return nil, errors.New("no way to get report summaries for nil artifact")
	}

	// Get reports first
	rps, err := bc.GetReport(ctx, artifact, mimeTypes)
	if err != nil {
		return nil, err
	}

	summaries := make(map[string]interface{}, len(rps))
	for _, rp := range rps {
		sum, err := report.GenerateSummary(rp, options...)
		if err != nil {
			return nil, err
		}

		if s, ok := summaries[rp.MimeType]; ok {
			r, err := report.MergeSummary(rp.MimeType, s, sum)
			if err != nil {
				return nil, err
			}

			summaries[rp.MimeType] = r
		} else {
			summaries[rp.MimeType] = sum
		}
	}

	return summaries, nil
}

// GetScanLog ...
func (bc *basicController) GetScanLog(ctx context.Context, uuid string) ([]byte, error) {
	if len(uuid) == 0 {
		return nil, errors.New("empty uuid to get scan log")
	}

	reportUUIDs := vuln.ParseReportIDs(uuid)
	tasks, err := bc.listScanTasks(ctx, reportUUIDs)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	reportUUIDToTasks := map[string]*task.Task{}
	for _, task := range tasks {
		for _, reportUUID := range getReportUUIDs(task.ExtraAttrs) {
			reportUUIDToTasks[reportUUID] = task
		}
	}

	errs := map[string]error{}
	logs := make(map[string][]byte, len(tasks))

	var (
		mu sync.Mutex
		wg sync.WaitGroup
	)
	for _, reportUUID := range reportUUIDs {
		wg.Add(1)

		go func(reportUUID string) {
			defer wg.Done()

			task, ok := reportUUIDToTasks[reportUUID]
			if !ok {
				return
			}

			log, err := bc.taskMgr.GetLog(ctx, task.ID)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errs[reportUUID] = err
			} else {
				logs[reportUUID] = log
			}
		}(reportUUID)
	}
	wg.Wait()

	if len(reportUUIDs) == 1 {
		return logs[reportUUIDs[0]], errs[reportUUIDs[0]]
	}

	if len(errs) == len(reportUUIDs) {
		for _, err := range errs {
			return nil, err
		}
	}

	var b bytes.Buffer

	multiLogs := len(logs) > 1
	for _, reportUUID := range reportUUIDs {
		log, ok := logs[reportUUID]
		if !ok || len(log) == 0 {
			continue
		}

		if multiLogs {
			if b.Len() > 0 {
				b.WriteString("\n\n\n\n")
			}
			b.WriteString(fmt.Sprintf("---------- Logs of report %s ----------\n", reportUUID))
		}

		b.Write(log)
	}

	return b.Bytes(), nil
}

func (bc *basicController) UpdateReport(ctx context.Context, report *sca.CheckInReport) error {
	rpl, err := bc.manager.GetBy(ctx, report.Digest, report.RegistrationUUID, []string{report.MimeType})
	if err != nil {
		return errors.Wrap(err, "scan controller: handle job hook")
	}

	logger := log.G(ctx)

	if len(rpl) == 0 {
		fields := log.Fields{
			"report_digest":     report.Digest,
			"registration_uuid": report.RegistrationUUID,
			"mime_type":         report.MimeType,
		}
		logger.WithFields(fields).Warningf("no report found to update data")

		return errors.NotFoundError(nil).WithMessage("no report found to update data")
	}

	logger.Debugf("Converting report ID %s to  the new V2 schema", rpl[0].UUID)

	_, reportData, err := bc.reportConverter.ToRelationalSchema(ctx, rpl[0].UUID, rpl[0].RegistrationUUID, rpl[0].Digest, report.RawReport)
	if err != nil {
		return errors.Wrapf(err, "Failed to convert vulnerability data to new schema for report UUID : %s", rpl[0].UUID)
	}
	// update the original report with the new summarized report with all vulnerability data removed.
	// this is required since the top level layers relay on the vuln.Report struct that
	// contains additional metadata within the report which if stored in the new columns within the scan_report table
	// would be redundant
	if err := bc.manager.UpdateReportData(ctx, rpl[0].UUID, reportData); err != nil {
		return errors.Wrap(err, "scan controller: handle job hook")
	}

	logger.Debugf("Converted report ID %s to the new V2 schema", rpl[0].UUID)

	return nil
}

// DeleteReports ...
func (bc *basicController) DeleteReports(ctx context.Context, digests ...string) error {
	if err := bc.manager.DeleteByDigests(ctx, digests...); err != nil {
		return errors.Wrap(err, "scan controller: delete reports")
	}
	return nil
}

// makeRobotAccount creates a robot account based on the arguments for scanning.
func (bc *basicController) makeRobotAccount(ctx context.Context, projectID int64, repository string, registration *scanner.Registration) (*robot.Robot, error) {
	// Use uuid as name to avoid duplicated entries.
	UUID, err := bc.uuid()
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}

	projectName := strings.Split(repository, "/")[0]

	robotReq := &robot.Robot{
		Robot: model.Robot{
			Name:        fmt.Sprintf("%s-%s", registration.Name, UUID),
			Description: "for scan",
			ProjectID:   projectID,
		},
		Level: robot.LEVELPROJECT,
		Permissions: []*robot.Permission{
			{
				Kind:      "project",
				Namespace: projectName,
				Access: []*types.Policy{
					{
						Resource: rbac.ResourceRepository,
						Action:   rbac.ActionPull,
					},
					{
						Resource: rbac.ResourceRepository,
						Action:   rbac.ActionScannerPull,
					},
				},
			},
		},
	}

	rb, pwd, err := bc.rc.Create(ctx, robotReq)
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}

	r, err := bc.rc.Get(ctx, rb, &robot.Option{WithPermission: false})
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}
	r.Secret = pwd
	return r, nil
}

// launchScanJob launches a job to run scan
func (bc *basicController) launchScanJob(ctx context.Context, executionID int64, artifact *ar.Artifact, registration *scanner.Registration, reports []*scan.Report) error {
	// don't launch scan job for the artifact which is not supported by the scanner
	if !hasCapability(registration, artifact) {
		return nil
	}

	var ck string
	if registration.UseInternalAddr {
		ck = configCoreInternalAddr
	} else {
		ck = configRegistryEndpoint
	}

	registryAddr, err := bc.config(ck)
	if err != nil {
		return errors.Wrap(err, "scan controller: launch scan job")
	}

	robot, err := bc.makeRobotAccount(ctx, artifact.ProjectID, artifact.RepositoryName, registration)
	if err != nil {
		return errors.Wrap(err, "scan controller: launch scan job")
	}

	// Set job parameters
	scanReq := &v1.ScanRequest{
		Registry: &v1.Registry{
			URL: registryAddr,
		},
		Artifact: &v1.Artifact{
			NamespaceID: artifact.ProjectID,
			Repository:  artifact.RepositoryName,
			Digest:      artifact.Digest,
			MimeType:    artifact.ManifestMediaType,
		},
	}

	rJSON, err := registration.ToJSON()
	if err != nil {
		return errors.Wrap(err, "scan controller: launch scan job")
	}

	sJSON, err := scanReq.ToJSON()
	if err != nil {
		return errors.Wrap(err, "launch scan job")
	}

	robotJSON, err := robot.ToJSON()
	if err != nil {
		return errors.Wrap(err, "launch scan job")
	}

	mimes := make([]string, len(reports))
	reportUUIDs := make([]string, len(reports))
	for i, report := range reports {
		mimes[i] = report.MimeType
		reportUUIDs[i] = report.UUID
	}

	params := make(map[string]interface{})
	params[sca.JobParamRegistration] = rJSON
	params[sca.JobParameterAuthType] = registration.GetRegistryAuthorizationType()
	params[sca.JobParameterRequest] = sJSON
	params[sca.JobParameterMimes] = mimes
	params[sca.JobParameterRobot] = robotJSON

	j := &task.Job{
		Name: job.ImageScanJob,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
	}

	// keep the report uuids in array so that when ?| operator support by the FilterRaw method of beego's orm
	// we can list the tasks of the scan reports by one SQL
	extraAttrs := map[string]interface{}{
		artifactIDKey:  artifact.ID,
		robotIDKey:     robot.ID,
		reportUUIDsKey: reportUUIDs,
	}

	// NOTE: due to the limitation of the beego's orm, the List method of the task manager not support ?! operator for the jsonb field,
	// we cann't list the tasks for scan reports of uuid1, uuid2 by SQL `SELECT * FROM task WHERE (extra_attrs->'report_uuids')::jsonb ?| array['uuid1', 'uuid2']`
	// or by `SELECT * FROM task WHERE id IN (SELECT id FROM task WHERE (extra_attrs->'report_uuids')::jsonb ?| array['uuid1', 'uuid2'])`
	// so save {"report:uuid1": "1", "report:uuid2": "2"} in the extra_attrs of the task, and then list it with
	// SQL `SELECT * FROM task WHERE extra_attrs->>'report:uuid1' = '1'` in loop
	for _, reportUUID := range reportUUIDs {
		extraAttrs["report:"+reportUUID] = "1"
	}

	_, err = bc.taskMgr.Create(ctx, executionID, j, extraAttrs)
	return err
}

// listScanTasks returns the tasks of the reports
func (bc *basicController) listScanTasks(ctx context.Context, reportUUIDs []string) ([]*task.Task, error) {
	if len(reportUUIDs) == 0 {
		return nil, nil
	}

	tasks := make([]*task.Task, len(reportUUIDs))
	errs := make([]error, len(reportUUIDs))

	var wg sync.WaitGroup
	for i, reportUUID := range reportUUIDs {
		wg.Add(1)

		go func(ix int, reportUUID string) {
			defer wg.Done()

			task, err := bc.getScanTask(bc.cloneCtx(ctx), reportUUID)
			if err == nil {
				tasks[ix] = task
			} else if !errors.IsNotFoundErr(err) {
				errs[ix] = err
			} else {
				log.G(ctx).Warningf("task for the scan report %s not found", reportUUID)
			}
		}(i, reportUUID)
	}
	wg.Wait()

	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	var results []*task.Task
	for _, task := range tasks {
		if task != nil {
			results = append(results, task)
		}
	}

	return results, nil
}

func (bc *basicController) getScanTask(ctx context.Context, reportUUID string) (*task.Task, error) {
	query := q.New(q.KeyWords{"extra_attrs." + "report:" + reportUUID: "1"})
	tasks, err := bc.taskMgr.List(bc.cloneCtx(ctx), query)
	if err != nil {
		return nil, err
	}
	if len(tasks) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("task for report %s not found", reportUUID)
	}

	return tasks[0], nil
}

func (bc *basicController) assembleReports(ctx context.Context, reports ...*scan.Report) error {
	reportUUIDs := make([]string, len(reports))
	for i, report := range reports {
		reportUUIDs[i] = report.UUID
	}

	tasks, err := bc.listScanTasks(ctx, reportUUIDs)
	if err != nil {
		return err
	}

	reportUUIDToTasks := map[string]*task.Task{}
	for _, task := range tasks {
		for _, reportUUID := range getReportUUIDs(task.ExtraAttrs) {
			reportUUIDToTasks[reportUUID] = task
		}
	}

	for _, report := range reports {
		if task, ok := reportUUIDToTasks[report.UUID]; ok {
			report.Status = task.Status
			report.StartTime = task.StartTime
			report.EndTime = task.EndTime
		} else {
			report.Status = job.ErrorStatus.String()
		}
		completeReport, err := bc.reportConverter.FromRelationalSchema(ctx, report.UUID, report.Digest, report.Report)
		if err != nil {
			return err
		}
		report.Report = completeReport
	}

	return nil
}

func getArtifactID(extraAttrs map[string]interface{}) int64 {
	var artifactID float64
	if extraAttrs != nil {
		if v, ok := extraAttrs[artifactIDKey]; ok {
			artifactID, _ = v.(float64) // int64 Unmarshal to float64
		}
	}

	return int64(artifactID)
}

func getReportUUIDs(extraAttrs map[string]interface{}) []string {
	var reportUUIDs []string

	if extraAttrs != nil {
		value, ok := extraAttrs[reportUUIDsKey]
		if ok {
			arr, _ := value.([]interface{})
			for _, el := range arr {
				if s, ok := el.(string); ok {
					reportUUIDs = append(reportUUIDs, s)
				}
			}
		}
	}

	return reportUUIDs
}

func getRobotID(extraAttrs map[string]interface{}) int64 {
	var trackID float64
	if extraAttrs != nil {
		if v, ok := extraAttrs[robotIDKey]; ok {
			trackID, _ = v.(float64) // int64 Unmarshal to float64
		}
	}

	return int64(trackID)
}

func parseOptions(options ...Option) (*Options, error) {
	ops := &Options{}
	for _, op := range options {
		if err := op(ops); err != nil {
			return nil, err
		}
	}

	return ops, nil
}
