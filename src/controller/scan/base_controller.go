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
	"reflect"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common/rbac"
	ar "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/robot"
	sc "github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/accessory"
	allowlist "github.com/goharbor/harbor/src/pkg/allowlist/models"
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
	artifactTagKey = "artifact_tag"
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

// launchScanJobParam is a param to launch scan job.
type launchScanJobParam struct {
	ExecutionID  int64
	Registration *scanner.Registration
	Artifact     *ar.Artifact
	Tag          string
	Reports      []*scan.Report
}

// basicController is default implementation of api.Controller interface
type basicController struct {
	// Manage the scan report records
	manager report.Manager
	// Artifact controller
	ar ar.Controller
	// Accessory manager
	acc accessory.Manager
	// Scanner controller
	sc sc.Controller
	// Robot account controller
	rc robot.Controller
	// Tag controller
	tagCtl tag.Controller
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
		// Refer to the default accessory manager
		acc: accessory.Mgr,
		// Refer to the default scanner controller
		sc: sc.DefaultController,
		// Refer to the default robot account controller
		rc: robot.Ctl,
		// Refer to the default tag controller
		tagCtl: tag.Ctl,
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
		reportConverter: postprocessors.Converter,
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
		ok, err := bc.isAccessory(ctx, a)
		if err != nil {
			return err
		}
		if ok {
			return nil
		}

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
		return errors.PreconditionFailedError(nil).WithMessage("scanner %s is deactivated", r.Name)
	}

	artifacts, scannable, err := bc.collectScanningArtifacts(ctx, r, artifact)
	if err != nil {
		return err
	}

	if !scannable {
		return errors.BadRequestError(nil).WithMessage("the configured scanner %s does not support scanning artifact with mime type %s", r.Name, artifact.ManifestMediaType)
	}

	// Parse options
	opts, err := parseOptions(options...)
	if err != nil {
		return errors.Wrap(err, "scan controller: scan")
	}

	var (
		errs                []error
		launchScanJobParams []*launchScanJobParam
	)
	for _, art := range artifacts {
		reports, err := bc.makeReportPlaceholder(ctx, r, art)
		if err != nil {
			if errors.IsConflictErr(err) {
				errs = append(errs, err)
			} else {
				return err
			}
		}

		var tag string
		if art.Digest == artifact.Digest {
			tag = opts.Tag
		}

		if tag == "" {
			latestTag, err := bc.getLatestTagOfArtifact(ctx, art.ID)
			if err != nil {
				return err
			}

			tag = latestTag
		}

		if len(reports) > 0 {
			launchScanJobParams = append(launchScanJobParams, &launchScanJobParam{
				Registration: r,
				Artifact:     art,
				Tag:          tag,
				Reports:      reports,
			})
		}
	}

	// all report placeholder conflicted
	if len(errs) == len(artifacts) {
		return errs[0]
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
	for _, launchScanJobParam := range launchScanJobParams {
		launchScanJobParam.ExecutionID = opts.ExecutionID

		if err := bc.launchScanJob(ctx, launchScanJobParam); err != nil {
			log.G(ctx).Warningf("scan artifact %s@%s failed, error: %v", artifact.RepositoryName, artifact.Digest, err)
			errs = append(errs, err)
		}
	}

	// all scanning of the artifacts failed
	if len(errs) == len(launchScanJobParams) {
		return fmt.Errorf("scan artifact %s@%s failed", artifact.RepositoryName, artifact.Digest)
	}

	return nil
}

// Stop scan job of a given artifact
func (bc *basicController) Stop(ctx context.Context, artifact *ar.Artifact) error {
	if artifact == nil {
		return errors.New("nil artifact to stop scan")
	}
	query := q.New(q.KeyWords{"extra_attrs.artifact.digest": artifact.Digest})
	executions, err := bc.execMgr.List(ctx, query)
	if err != nil {
		return err
	}
	if len(executions) == 0 {
		message := fmt.Sprintf("no scan job for artifact digest=%v", artifact.Digest)
		return errors.BadRequestError(nil).WithMessage(message)
	}
	execution := executions[0]
	return bc.execMgr.Stop(ctx, execution.ID)
}

func (bc *basicController) ScanAll(ctx context.Context, trigger string, async bool) (int64, error) {
	executionID, err := bc.execMgr.Create(ctx, VendorTypeScanAll, 0, trigger)
	if err != nil {
		return 0, err
	}

	if async {
		go func(ctx context.Context) {
			// if async, this is running in another goroutine ensure the execution exists in db
			err := retry.Retry(func() error {
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

		if err := orm.WithTransaction(scan)(orm.SetTransactionOpNameToContext(bc.makeCtx(), "tx-start-scanall")); err != nil {
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
			message = fmt.Sprintf("%s, scanner not found or deactivated for %d of them", message, summary.PreconditionCount)
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

		if err := orm.WithTransaction(create)(orm.SetTransactionOpNameToContext(ctx, "tx-make-report-placeholder")); err != nil {
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
func (bc *basicController) GetSummary(ctx context.Context, artifact *ar.Artifact, mimeTypes []string) (map[string]interface{}, error) {
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
		sum, err := report.GenerateSummary(rp)
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

// DeleteReports ...
func (bc *basicController) DeleteReports(ctx context.Context, digests ...string) error {
	if err := bc.manager.DeleteByDigests(ctx, digests...); err != nil {
		return errors.Wrap(err, "scan controller: delete reports")
	}
	return nil
}

func (bc *basicController) GetVulnerable(ctx context.Context, artifact *ar.Artifact, allowlist allowlist.CVESet) (*Vulnerable, error) {
	if artifact == nil {
		return nil, errors.New("no way to get vulnerable for nil artifact")
	}

	var (
		mimeType string
		reports  []*scan.Report
	)
	for _, m := range []string{v1.MimeTypeNativeReport, v1.MimeTypeGenericVulnerabilityReport} {
		rps, err := bc.GetReport(ctx, artifact, []string{m})
		if err != nil {
			return nil, err
		}

		if len(rps) == 0 {
			continue
		}

		mimeType = m
		reports = rps
		break
	}

	if len(reports) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("report not found")
	}

	scanStatus := reports[0].Status
	for _, report := range reports {
		scanStatus = vuln.MergeScanStatus(scanStatus, report.Status)
	}

	vulnerable := &Vulnerable{
		ScanStatus: scanStatus,
	}

	if !vulnerable.IsScanSuccess() {
		return vulnerable, nil
	}

	raw, err := report.Reports(reports).ResolveData(mimeType)
	if err != nil {
		return nil, err
	}

	if raw == nil {
		return vulnerable, nil
	}

	rp, ok := raw.(*vuln.Report)
	if !ok {
		return nil, errors.Errorf("type mismatch: expect *vuln.Report but got %s", reflect.TypeOf(raw).String())
	}

	if vuls := rp.GetVulnerabilityItemList().Items(); len(vuls) > 0 {
		vulnerable.VulnerabilitiesCount = len(vuls)

		var severity vuln.Severity

		for _, v := range vuls {
			if allowlist.Contains(v.ID) {
				// Append the by passed CVEs specified in the allowlist
				vulnerable.CVEBypassed = append(vulnerable.CVEBypassed, v.ID)

				vulnerable.VulnerabilitiesCount--

				continue
			}

			if severity == "" || v.Severity.Code() > severity.Code() {
				severity = v.Severity
			}
		}

		if severity != "" {
			vulnerable.Severity = &severity
		}
	}

	return vulnerable, nil
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
func (bc *basicController) launchScanJob(ctx context.Context, param *launchScanJobParam) error {
	// don't launch scan job for the artifact which is not supported by the scanner
	if !hasCapability(param.Registration, param.Artifact) {
		return nil
	}

	var ck string
	if param.Registration.UseInternalAddr {
		ck = configCoreInternalAddr
	} else {
		ck = configRegistryEndpoint
	}

	registryAddr, err := bc.config(ck)
	if err != nil {
		return errors.Wrap(err, "scan controller: launch scan job")
	}

	robot, err := bc.makeRobotAccount(ctx, param.Artifact.ProjectID, param.Artifact.RepositoryName, param.Registration)
	if err != nil {
		return errors.Wrap(err, "scan controller: launch scan job")
	}

	// Set job parameters
	scanReq := &v1.ScanRequest{
		Registry: &v1.Registry{
			URL: registryAddr,
		},
		Artifact: &v1.Artifact{
			NamespaceID: param.Artifact.ProjectID,
			Repository:  param.Artifact.RepositoryName,
			Digest:      param.Artifact.Digest,
			Tag:         param.Tag,
			MimeType:    param.Artifact.ManifestMediaType,
		},
	}

	rJSON, err := param.Registration.ToJSON()
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

	mimes := make([]string, len(param.Reports))
	reportUUIDs := make([]string, len(param.Reports))
	for i, report := range param.Reports {
		mimes[i] = report.MimeType
		reportUUIDs[i] = report.UUID
	}

	params := make(map[string]interface{})
	params[sca.JobParamRegistration] = rJSON
	params[sca.JobParameterAuthType] = param.Registration.GetRegistryAuthorizationType()
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
		artifactIDKey:  param.Artifact.ID,
		artifactTagKey: param.Tag,
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

	_, err = bc.taskMgr.Create(ctx, param.ExecutionID, j, extraAttrs)
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

func (bc *basicController) getLatestTagOfArtifact(ctx context.Context, artifactID int64) (string, error) {
	query := q.New(q.KeyWords{"artifact_id": artifactID})
	tags, err := bc.tagCtl.List(ctx, query.First(q.NewSort("push_time", true)), nil)
	if err != nil {
		return "", err
	}

	if len(tags) == 0 {
		return "", nil
	}

	return tags[0].Name, nil
}

func (bc *basicController) isAccessory(ctx context.Context, art *ar.Artifact) (bool, error) {
	ac, err := bc.acc.List(ctx, q.New(q.KeyWords{"ArtifactID": art.Artifact.ID, "digest": art.Artifact.Digest}))
	if err != nil {
		return false, err
	}
	if len(ac) > 0 {
		return true, nil
	}
	return false, nil
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

func getArtifactTag(extraAttrs map[string]interface{}) string {
	var tag string
	if extraAttrs != nil {
		if v, ok := extraAttrs[artifactTagKey]; ok {
			tag, _ = v.(string)
		}
	}

	return tag
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
