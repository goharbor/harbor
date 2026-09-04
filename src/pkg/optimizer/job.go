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

package optimizer

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/goharbor/harbor/src/common"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	dockerfileoptdao "github.com/goharbor/harbor/src/pkg/dockerfileoptimization/dao"
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	v1 "github.com/goharbor/harbor/src/pkg/optimizer/rest/v1"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

const (
	// JobParamRegistration is the key of the optimizer registration job parameter
	JobParamRegistration = "registration"
	// JobParameterRequest is the key of the optimize request job parameter
	JobParameterRequest = "optimizeRequest"
	// JobParameterAuthType is the key of the registry auth type job parameter
	JobParameterAuthType = "authType"
	// JobParameterRobot is the key of the robot account job parameter
	JobParameterRobot = "robotAccount"

	// checkTimeout is the overall timeout for polling the adapter for a report.
	// The LLM call inside the adapter has its own 90s timeout, so 10 minutes is
	// generous without holding a job worker for the scan-style 30 minutes.
	checkTimeout       = 10 * time.Minute
	firstCheckInterval = 2 * time.Second

	authorizationBearer = "Bearer"
	authorizationBasic  = "Basic"

	service = "harbor-registry"
)

// Job runs the artifact optimization in the job service asynchronously: it submits
// the optimize request to the registered optimizer adapter, polls for the report and
// persists the outcome into the dockerfile_optimization table.
type Job struct{}

// MaxFails for defining the number of retries
func (j *Job) MaxFails() uint {
	return 1
}

// MaxCurrency is implementation of same method in Interface.
func (j *Job) MaxCurrency() uint {
	return 0
}

// ShouldRetry indicates if the job should be retried
func (j *Job) ShouldRetry() bool {
	return false
}

// Validate the parameters of this job
func (j *Job) Validate(params job.Parameters) error {
	if params == nil {
		return errors.New("missing parameters of optimize job")
	}

	if _, err := extractRegistration(params); err != nil {
		return errors.Wrap(err, "optimize job validate")
	}

	if _, err := extractOptimizeReq(params); err != nil {
		return errors.Wrap(err, "optimize job validate")
	}

	if _, err := extractRobotAccount(params); err != nil {
		return errors.Wrap(err, "optimize job validate")
	}

	authType, err := extractAuthType(params)
	if err != nil {
		return errors.Wrap(err, "optimize job validate")
	}

	if authType != authorizationBearer && authType != authorizationBasic {
		return errors.Errorf("optimize job validate: not support auth type %s", authType)
	}

	return nil
}

// Run the job
func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	myLogger := ctx.GetLogger()

	shouldStop := func() bool {
		if cmd, ok := ctx.OPCommand(); ok && cmd == job.StopCommand {
			myLogger.Info("optimize job being stopped")
			return true
		}
		return false
	}

	// Ignore errors as they have been validated already
	r, _ := extractRegistration(params)
	req, _ := extractOptimizeReq(params)
	robotAccount, _ := extractRobotAccount(params)

	optDAO := dockerfileoptdao.New()
	repo := req.Artifact.Repository
	digest := req.Artifact.Digest

	// failWith marks the DB row as errored and returns the wrapped error so the
	// task status reflects the failure.
	failWith := func(err error, message string) error {
		e := errors.Wrap(err, message)
		myLogger.Error(e)
		if derr := optDAO.UpdateStatus(ctx.SystemContext(), repo, digest, dockerfileoptdao.StatusError, e.Error()); derr != nil {
			myLogger.Errorf("failed to persist error status: %v", derr)
		}
		return e
	}

	if shouldStop() {
		return nil
	}

	if err := optDAO.UpdateStatus(ctx.SystemContext(), repo, digest, dockerfileoptdao.StatusRunning, ""); err != nil {
		myLogger.Errorf("failed to mark optimization running: %v", err)
	}

	client, err := r.Client(v1.DefaultClientPool)
	if err != nil {
		return failWith(err, "optimize job: get client")
	}

	var authorization string

	authType, _ := extractAuthType(params)
	if authType == authorizationBearer {
		tokenURL, err := getInternalTokenServiceEndpoint(ctx)
		if err != nil {
			return failWith(err, "optimize job: get token service endpoint")
		}
		authorization, err = makeBearerAuthorization(robotAccount, tokenURL, req.Artifact.Repository)
		if err != nil {
			return failWith(err, "optimize job: make bearer authorization")
		}
	} else {
		authorization, err = makeBasicAuthorization(robotAccount)
		if err != nil {
			return failWith(err, "optimize job: make basic authorization")
		}
	}

	if shouldStop() {
		return nil
	}

	req.Registry.Authorization = authorization
	resp, err := client.SubmitOptimize(req)
	if err != nil {
		return failWith(err, "optimize job: submit optimize request")
	}

	myLogger.Infof("Optimize request submitted, tracking ID: %s", resp.ID)

	// Poll the adapter until the report is ready or the timeout hits.
	rawReport, err := j.pollReport(ctx, client, resp.ID, shouldStop)
	if err != nil {
		return failWith(err, "optimize job: fetch optimization report")
	}
	if rawReport == "" {
		// stopped or terminated by system
		return nil
	}

	report := &v1.OptimizationReport{}
	if err := report.FromJSON(rawReport); err != nil {
		return failWith(err, "optimize job: parse optimization report")
	}

	switch report.Status {
	case v1.ReportStatusSuccess:
		rec, err := optDAO.GetByArtifact(ctx.SystemContext(), repo, digest)
		if err != nil {
			return failWith(err, "optimize job: load pending record")
		}
		rec.Dockerfile = report.Dockerfile
		rec.OptimizedDockerfile = report.OptimizedDockerfile
		rec.AttestationManifestDigest = report.AttestationManifestDigest
		rec.StatementDigest = report.StatementDigest
		rec.Generated = report.Generated
		rec.Status = dockerfileoptdao.StatusSuccess
		rec.Error = ""
		if err := optDAO.Upsert(ctx.SystemContext(), rec); err != nil {
			return failWith(err, "optimize job: persist optimization result")
		}
		myLogger.Info("Optimization succeeded and result persisted")
		return nil
	case v1.ReportStatusFailed:
		msg := "optimization failed"
		if report.Error != nil {
			msg = fmt.Sprintf("%s: %s", report.Error.Code, report.Error.Message)
		}
		if err := optDAO.UpdateStatus(ctx.SystemContext(), repo, digest, dockerfileoptdao.StatusError, msg); err != nil {
			myLogger.Errorf("failed to persist error status: %v", err)
		}
		return errors.New(msg)
	default:
		return failWith(errors.Errorf("unexpected report status %q", report.Status), "optimize job: parse optimization report")
	}
}

// pollReport loops calling GetOptimizationReport until the report is ready. Returns
// the raw report, or "" if the job was stopped/terminated externally.
func (j *Job) pollReport(ctx job.Context, client v1.Client, requestID string, shouldStop func() bool) (string, error) {
	tm := time.NewTimer(firstCheckInterval)
	defer tm.Stop()

	timeout := time.After(checkTimeout)

	for {
		select {
		case <-tm.C:
			if shouldStop() {
				return "", nil
			}

			rawReport, err := client.GetOptimizationReport(requestID)
			if err != nil {
				if notReadyErr, ok := err.(*v1.ReportNotReadyError); ok {
					tm.Reset(time.Duration(notReadyErr.RetryAfter) * time.Second)
					ctx.GetLogger().Infof("Report is not ready yet, retry after %d seconds", notReadyErr.RetryAfter)
					continue
				}
				return "", err
			}
			return rawReport, nil
		case <-ctx.SystemContext().Done():
			// Terminated by system
			return "", nil
		case <-timeout:
			return "", errors.New("check optimization report timeout")
		}
	}
}

func extractRegistration(params job.Parameters) (*dao.Registration, error) {
	v, ok := params[JobParamRegistration]
	if !ok {
		return nil, errors.Errorf("missing job parameter '%s'", JobParamRegistration)
	}

	jsonData, ok := v.(string)
	if !ok {
		return nil, errors.Errorf(
			"malformed job parameter '%s', expecting string but got %s",
			JobParamRegistration,
			reflect.TypeOf(v).String(),
		)
	}

	r := &dao.Registration{}
	if err := r.FromJSON(jsonData); err != nil {
		return nil, err
	}

	if err := r.Validate(true); err != nil {
		return nil, err
	}

	return r, nil
}

func extractOptimizeReq(params job.Parameters) (*v1.OptimizeRequest, error) {
	v, ok := params[JobParameterRequest]
	if !ok {
		return nil, errors.Errorf("missing job parameter '%s'", JobParameterRequest)
	}

	jsonData, ok := v.(string)
	if !ok {
		return nil, errors.Errorf(
			"malformed job parameter '%s', expecting string but got %s",
			JobParameterRequest,
			reflect.TypeOf(v).String(),
		)
	}

	req := &v1.OptimizeRequest{}
	if err := req.FromJSON(jsonData); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func extractRobotAccount(params job.Parameters) (*model.Robot, error) {
	v, ok := params[JobParameterRobot]
	if !ok {
		return nil, errors.Errorf("missing job parameter '%s'", JobParameterRobot)
	}

	jsonData, ok := v.(string)
	if !ok {
		return nil, errors.Errorf(
			"malformed job parameter '%s', expecting string but got %s",
			JobParameterRobot,
			reflect.TypeOf(v).String(),
		)
	}
	r := &model.Robot{}

	if err := r.FromJSON(jsonData); err != nil {
		return nil, err
	}

	return r, nil
}

func extractAuthType(params job.Parameters) (string, error) {
	v, ok := params[JobParameterAuthType]
	if !ok {
		return "", errors.Errorf("missing job parameter '%s'", JobParameterAuthType)
	}

	authType, ok := v.(string)
	if !ok {
		return "", errors.Errorf(
			"malformed job parameter '%s', expecting string but got %s",
			JobParameterAuthType,
			reflect.TypeOf(v).String(),
		)
	}

	return authType, nil
}

func getInternalTokenServiceEndpoint(ctx job.Context) (string, error) {
	cfgMgr, ok := config.FromContext(ctx.SystemContext())
	if !ok {
		return "", errors.Errorf("failed to get config manager")
	}

	return cfgMgr.Get(ctx.SystemContext(), common.CoreURL).GetString() + "/service/token", nil
}

// makeBasicAuthorization creates authorization from a robot account.
func makeBasicAuthorization(robotAccount *model.Robot) (string, error) {
	basic := fmt.Sprintf("%s:%s", robotAccount.Name, robotAccount.Secret)
	encoded := base64.StdEncoding.EncodeToString([]byte(basic))

	return fmt.Sprintf("Basic %s", encoded), nil
}

// makeBearerAuthorization creates bearer token from a robot account
func makeBearerAuthorization(robotAccount *model.Robot, tokenURL string, repository string) (string, error) {
	u, err := url.Parse(tokenURL)
	if err != nil {
		return "", err
	}

	query := u.Query()
	query.Add("service", service)
	query.Add("scope", fmt.Sprintf("repository:%s:pull", repository))
	u.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}

	auth, _ := makeBasicAuthorization(robotAccount)
	req.Header.Set("Authorization", auth)

	client := &http.Client{
		Transport: commonhttp.GetHTTPTransport(commonhttp.WithInsecure(true)),
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get bearer token failed, %s", string(data))
	}

	token := &models.Token{}
	if err = json.Unmarshal(data, token); err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", token.GetToken()), nil
}
