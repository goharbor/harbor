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
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	cj "github.com/goharbor/harbor/src/common/job"
	jm "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/rbac"
	ar "github.com/goharbor/harbor/src/controller/artifact"
	sc "github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	sca "github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/all"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/google/uuid"
)

// DefaultController is a default singleton scan API controller.
var DefaultController = NewController()

const (
	configRegistryEndpoint = "registryEndpoint"
	configCoreInternalAddr = "coreInternalAddr"
)

// uuidGenerator is a func template which is for generating UUID.
type uuidGenerator func() (string, error)

// configGetter is a func template which is used to wrap the config management
// utility methods.
type configGetter func(cfg string) (string, error)

// jcGetter is a func template which is used to get the job service client.
type jcGetter func() cj.Client

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
	// Job service client
	jc jcGetter
	// UUID generator
	uuid uuidGenerator
	// Configuration getter func
	config configGetter
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
		rc: robot.RobotCtr,
		// Refer to the default job service client
		jc: func() cj.Client {
			return cj.GlobalClient
		},
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

	r, err := bc.sc.GetRegistrationByProject(artifact.ProjectID)
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
		return errors.Errorf("the configured scanner %s does not support scanning artifact with mime type %s", r.Name, artifact.ManifestMediaType)
	}

	type Param struct {
		Artifact      *ar.Artifact
		TrackID       string
		ProducesMimes []string
	}

	params := []*Param{}

	var errs []error
	for _, art := range artifacts {
		trackID, producesMimes, err := bc.makeReportPlaceholder(ctx, r, art, options...)
		if err != nil {
			if errors.IsConflictErr(err) {
				errs = append(errs, err)
			} else {
				return err
			}
		}

		if len(producesMimes) > 0 {
			params = append(params, &Param{Artifact: art, TrackID: trackID, ProducesMimes: producesMimes})
		}
	}

	// all report placeholder conflicted
	if len(errs) == len(artifacts) {
		return errs[0]
	}

	errs = errs[:0]
	for _, param := range params {
		if err := bc.scanArtifact(ctx, r, param.Artifact, param.TrackID, param.ProducesMimes); err != nil {
			log.Warningf("scan artifact %s@%s failed, error: %v", artifact.RepositoryName, artifact.Digest, err)
			errs = append(errs, err)
		}
	}

	// all scanning of the artifacts failed
	if len(errs) == len(params) {
		return fmt.Errorf("scan artifact %s@%s failed", artifact.RepositoryName, artifact.Digest)
	}

	return nil
}

func (bc *basicController) makeReportPlaceholder(ctx context.Context, r *scanner.Registration, art *ar.Artifact, options ...Option) (string, []string, error) {
	trackID, err := bc.uuid()
	if err != nil {
		return "", nil, errors.Wrap(err, "scan controller: scan")
	}

	// Parse options
	ops, err := parseOptions(options...)
	if err != nil {
		return "", nil, errors.Wrap(err, "scan controller: scan")
	}

	create := func(ctx context.Context, digest, registrationUUID, mimeType, trackID string, status job.Status) error {
		reportPlaceholder := &scan.Report{
			Digest:           digest,
			RegistrationUUID: registrationUUID,
			Status:           status.String(),
			StatusCode:       status.Code(),
			TrackID:          trackID,
			MimeType:         mimeType,
		}
		// Set requester if it is specified
		if len(ops.Requester) > 0 {
			reportPlaceholder.Requester = ops.Requester
		} else {
			// Use the trackID as the requester
			reportPlaceholder.Requester = trackID
		}

		_, e := bc.manager.Create(reportPlaceholder)
		return e
	}

	if hasCapability(r, art) {
		var producesMimes []string

		for _, pm := range r.GetProducesMimeTypes(art.ManifestMediaType) {
			if err = create(ctx, art.Digest, r.UUID, pm, trackID, job.PendingStatus); err != nil {
				return "", nil, err
			}

			producesMimes = append(producesMimes, pm)
		}

		if len(producesMimes) > 0 {
			return trackID, producesMimes, nil
		}
	}

	err = create(ctx, art.Digest, r.UUID, v1.MimeTypeNativeReport, trackID, job.ErrorStatus)
	return "", nil, err
}

func (bc *basicController) scanArtifact(ctx context.Context, r *scanner.Registration, artifact *ar.Artifact, trackID string, producesMimes []string) error {
	jobID, err := bc.launchScanJob(trackID, artifact, r, producesMimes)
	if err != nil {
		// Update the status to the concrete error
		// Change status code to normal error code
		if e := bc.manager.UpdateStatus(trackID, err.Error(), 0); e != nil {
			err = errors.Wrap(e, err.Error())
		}

		return errors.Wrap(err, "scan controller: scan")
	}

	// Insert the generated job ID now
	// It will not block the whole process. If any errors happened, just logged.
	if err := bc.manager.UpdateScanJobID(trackID, jobID); err != nil {
		logger.Error(errors.Wrap(err, "scan controller: scan"))
	}

	return nil
}

// GetReport ...
func (bc *basicController) GetReport(ctx context.Context, artifact *ar.Artifact, mimeTypes []string) ([]*scan.Report, error) {
	if artifact == nil {
		return nil, errors.New("no way to get report for nil artifact")
	}

	mimes := make([]string, 0)
	mimes = append(mimes, mimeTypes...)
	if len(mimes) == 0 {
		// Retrieve native as default
		mimes = append(mimes, v1.MimeTypeNativeReport)
	}

	// Get current scanner settings
	r, err := bc.sc.GetRegistrationByProject(artifact.ProjectID)
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

			reports, err := bc.manager.GetBy(a.Digest, r.UUID, mimes)
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
			// NOTE: If the artifact is OCI image, this happened when the artifact is not scanned.
			// If the artifact is OCI image index, this happened when the artifact is not scanned,
			// but its children artifacts may scanned so return empty report
			return nil, nil
		}
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
func (bc *basicController) GetScanLog(uuid string) ([]byte, error) {
	if len(uuid) == 0 {
		return nil, errors.New("empty uuid to get scan log")
	}

	// Get by uuid
	sr, err := bc.manager.Get(uuid)
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: get scan log")
	}

	if sr == nil {
		// Not found
		return nil, nil
	}

	// Not job error
	if sr.StatusCode == job.ErrorStatus.Code() {
		jst := job.Status(sr.Status)
		if jst.Code() == -1 {
			return []byte(sr.Status), nil
		}
	}

	// Job log
	return bc.jc().GetJobLog(sr.JobID)
}

// HandleJobHooks ...
func (bc *basicController) HandleJobHooks(trackID string, change *job.StatusChange) error {
	if len(trackID) == 0 {
		return errors.New("empty track ID")
	}

	if change == nil {
		return errors.New("nil change object")
	}

	// Clear robot account
	// Only when the job is successfully done!
	if change.Status == job.SuccessStatus.String() {
		if v, ok := change.Metadata.Parameters[sca.JobParameterRobotID]; ok {
			if rid, y := v.(float64); y {
				if err := robot.RobotCtr.DeleteRobotAccount(int64(rid)); err != nil {
					// Should not block the main flow, just logged
					log.Error(errors.Wrap(err, "scan controller: handle job hook"))
				} else {
					log.Debugf("Robot account with id %d for the scan %s is removed", int64(rid), trackID)
				}
			}
		}
	}

	// Check in data
	if len(change.CheckIn) > 0 {
		checkInReport := &sca.CheckInReport{}
		if err := checkInReport.FromJSON(change.CheckIn); err != nil {
			return errors.Wrap(err, "scan controller: handle job hook")
		}

		rpl, err := bc.manager.GetBy(
			checkInReport.Digest,
			checkInReport.RegistrationUUID,
			[]string{checkInReport.MimeType})
		if err != nil {
			return errors.Wrap(err, "scan controller: handle job hook")
		}

		if len(rpl) == 0 {
			return errors.New("no report found to update data")
		}

		if err := bc.manager.UpdateReportData(
			rpl[0].UUID,
			checkInReport.RawReport,
			change.Metadata.Revision); err != nil {
			return errors.Wrap(err, "scan controller: handle job hook")
		}

		return nil
	}

	return bc.manager.UpdateStatus(trackID, change.Status, change.Metadata.Revision)
}

// DeleteReports ...
func (bc *basicController) DeleteReports(digests ...string) error {
	if err := bc.manager.DeleteByDigests(digests...); err != nil {
		return errors.Wrap(err, "scan controller: delete reports")
	}

	return nil
}

// GetStats ...
func (bc *basicController) GetStats(requester string) (*all.Stats, error) {
	sts, err := bc.manager.GetStats(requester)
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: delete reports")
	}

	return sts, nil
}

// makeRobotAccount creates a robot account based on the arguments for scanning.
func (bc *basicController) makeRobotAccount(projectID int64, repository string, registration *scanner.Registration) (*model.Robot, error) {
	// Use uuid as name to avoid duplicated entries.
	UUID, err := bc.uuid()
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}

	resource := rbac.NewProjectNamespace(projectID).Resource(rbac.ResourceRepository)
	robotReq := &model.RobotCreate{
		Name:        fmt.Sprintf("%s-%s", registration.Name, UUID),
		Description: "for scan",
		ProjectID:   projectID,
		Access: []*types.Policy{
			{Resource: resource, Action: rbac.ActionPull},
			{Resource: resource, Action: rbac.ActionScannerPull},
		},
	}

	rb, err := bc.rc.CreateRobotAccount(robotReq)
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}

	return rb, nil
}

// launchScanJob launches a job to run scan
func (bc *basicController) launchScanJob(trackID string, artifact *ar.Artifact, registration *scanner.Registration, mimes []string) (jobID string, err error) {
	var ck string
	if registration.UseInternalAddr {
		ck = configCoreInternalAddr
	} else {
		ck = configRegistryEndpoint
	}

	registryAddr, err := bc.config(ck)
	if err != nil {
		return "", errors.Wrap(err, "scan controller: launch scan job")
	}

	robot, err := bc.makeRobotAccount(artifact.ProjectID, artifact.RepositoryName, registration)
	if err != nil {
		return "", errors.Wrap(err, "scan controller: launch scan job")
	}

	basic := fmt.Sprintf("%s:%s", robot.Name, robot.Token)
	authorization := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(basic)))

	// Set job parameters
	scanReq := &v1.ScanRequest{
		Registry: &v1.Registry{
			URL:           registryAddr,
			Authorization: authorization,
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
		return "", errors.Wrap(err, "scan controller: launch scan job")
	}

	sJSON, err := scanReq.ToJSON()
	if err != nil {
		return "", errors.Wrap(err, "launch scan job")
	}

	params := make(map[string]interface{})
	params[sca.JobParamRegistration] = rJSON
	params[sca.JobParameterRequest] = sJSON
	params[sca.JobParameterMimes] = mimes
	params[sca.JobParameterRobotID] = robot.ID

	// Launch job
	callbackURL, err := bc.config(configCoreInternalAddr)
	if err != nil {
		return "", errors.Wrap(err, "launch scan job")
	}
	hookURL := fmt.Sprintf("%s/service/notifications/jobs/scan/%s", callbackURL, trackID)

	j := &jm.JobData{
		Name: job.ImageScanJob,
		Metadata: &jm.JobMetadata{
			JobKind: job.KindGeneric,
		},
		Parameters: params,
		StatusHook: hookURL,
	}

	return bc.jc().SubmitJob(j)
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
