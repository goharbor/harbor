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
	"encoding/base64"
	"fmt"

	sc "github.com/goharbor/harbor/src/api/scanner"
	cj "github.com/goharbor/harbor/src/common/job"
	jm "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/robot"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	sca "github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/all"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/errs"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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

// Scan ...
func (bc *basicController) Scan(artifact *v1.Artifact, options ...Option) error {
	if artifact == nil {
		return errors.New("nil artifact to scan")
	}

	// Parse options
	ops, err := parseOptions(options...)
	if err != nil {
		return errors.Wrap(err, "scan controller: scan")
	}

	r, err := bc.sc.GetRegistrationByProject(artifact.NamespaceID)
	if err != nil {
		return errors.Wrap(err, "scan controller: scan")
	}

	// In case it does not exist
	if r == nil {
		return errs.WithCode(errs.PreconditionFailed, errs.Errorf("no available scanner for project: %d", artifact.NamespaceID))
	}

	// Check if it is disabled
	if r.Disabled {
		return errs.WithCode(errs.PreconditionFailed, errs.Errorf("scanner %s is disabled", r.Name))
	}

	// Check the health of the registration by ping.
	// The metadata of the scanner adapter is also returned.
	meta, err := bc.sc.Ping(r)
	if err != nil {
		return errors.Wrap(err, "scan controller: scan")
	}

	// Generate a UUID as track ID which groups the report records generated
	// by the specified registration for the digest with given mime type.
	trackID, err := bc.uuid()
	if err != nil {
		return errors.Wrap(err, "scan controller: scan")
	}

	producesMimes := make([]string, 0)
	matched := false
	statusConflict := false
	for _, ca := range meta.Capabilities {
		for _, cm := range ca.ConsumesMimeTypes {
			if cm == artifact.MimeType {
				matched = true
				break
			}
		}

		if matched {
			for _, pm := range ca.ProducesMimeTypes {
				// Create report placeholder first
				reportPlaceholder := &scan.Report{
					Digest:           artifact.Digest,
					RegistrationUUID: r.UUID,
					Status:           job.PendingStatus.String(),
					StatusCode:       job.PendingStatus.Code(),
					TrackID:          trackID,
					MimeType:         pm,
				}
				// Set requester if it is specified
				if len(ops.Requester) > 0 {
					reportPlaceholder.Requester = ops.Requester
				} else {
					// Use the trackID as the requester
					reportPlaceholder.Requester = trackID
				}

				_, e := bc.manager.Create(reportPlaceholder)
				if e != nil {
					// Check if it is a status conflict error with common error format.
					// Common error returned if and only if status conflicts.
					if !statusConflict {
						statusConflict = errs.AsError(e, errs.Conflict)
					}

					// Recorded by error wrap and logged at the same time.
					if err == nil {
						err = e
					} else {
						err = errors.Wrap(e, err.Error())
					}

					logger.Error(errors.Wrap(e, "scan controller: scan"))
					continue
				}

				producesMimes = append(producesMimes, pm)
			}

			break
		}
	}

	// Scanner does not support scanning the given artifact.
	if !matched {
		return errors.Errorf("the configured scanner %s does not support scanning artifact with mime type %s", r.Name, artifact.MimeType)
	}

	// If all the record are created failed.
	if len(producesMimes) == 0 {
		// Return the last error
		if statusConflict {
			return errs.WithCode(errs.Conflict, errs.Wrap(err, "scan controller: scan"))
		}

		return errors.Wrap(err, "scan controller: scan")
	}

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
func (bc *basicController) GetReport(artifact *v1.Artifact, mimeTypes []string) ([]*scan.Report, error) {
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
	r, err := bc.sc.GetRegistrationByProject(artifact.NamespaceID)
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: get report")
	}

	if r == nil {
		return nil, ierror.NotFoundError(nil).WithMessage("no scanner registration configured for project: %d", artifact.NamespaceID)
	}

	return bc.manager.GetBy(artifact.Digest, r.UUID, mimes)
}

// GetSummary ...
func (bc *basicController) GetSummary(artifact *v1.Artifact, mimeTypes []string, options ...report.Option) (map[string]interface{}, error) {
	if artifact == nil {
		return nil, errors.New("no way to get report summaries for nil artifact")
	}

	// Get reports first
	rps, err := bc.GetReport(artifact, mimeTypes)
	if err != nil {
		return nil, err
	}

	summaries := make(map[string]interface{}, len(rps))
	for _, rp := range rps {
		sum, err := report.GenerateSummary(rp, options...)
		if err != nil {
			return nil, err
		}

		summaries[rp.MimeType] = sum
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
func (bc *basicController) makeRobotAccount(projectID int64, repository string) (*model.Robot, error) {
	// Use uuid as name to avoid duplicated entries.
	UUID, err := bc.uuid()
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}

	resource := rbac.NewProjectNamespace(projectID).Resource(rbac.ResourceRepository)
	robotReq := &model.RobotCreate{
		Name:        UUID,
		Description: "for scan",
		ProjectID:   projectID,
		Access:      []*rbac.Policy{{Resource: resource, Action: rbac.ActionScannerPull}},
	}

	rb, err := bc.rc.CreateRobotAccount(robotReq)
	if err != nil {
		return nil, errors.Wrap(err, "scan controller: make robot account")
	}

	return rb, nil
}

// launchScanJob launches a job to run scan
func (bc *basicController) launchScanJob(trackID string, artifact *v1.Artifact, registration *scanner.Registration, mimes []string) (jobID string, err error) {
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

	robot, err := bc.makeRobotAccount(artifact.NamespaceID, artifact.Repository)
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
		Artifact: artifact,
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
