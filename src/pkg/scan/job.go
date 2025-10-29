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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

const (
	// JobParamRegistration ...
	JobParamRegistration = "registration"
	// JobParameterRequest ...
	JobParameterRequest = "scanRequest"
	// JobParameterMimes ...
	JobParameterMimes = "mimeTypes"
	// JobParameterAuthType ...
	JobParameterAuthType = "authType"
	// JobParameterRobot ...
	JobParameterRobot = "robotAccount"

	checkTimeout       = 30 * time.Minute
	firstCheckInterval = 2 * time.Second

	authorizationBearer = "Bearer"
	authorizationBasic  = "Basic"

	service = "harbor-registry"
)

// CheckInReport defines model for checking in the scan report with specified mime.
type CheckInReport struct {
	Digest           string `json:"digest"`
	RegistrationUUID string `json:"registration_uuid"`
	MimeType         string `json:"mime_type"`
	RawReport        string `json:"raw_report"`
}

// FromJSON parse json to CheckInReport
func (cir *CheckInReport) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty JSON data")
	}

	return json.Unmarshal([]byte(jsonData), cir)
}

// ToJSON marshal CheckInReport to JSON
func (cir *CheckInReport) ToJSON() (string, error) {
	jsonData, err := json.Marshal(cir)
	if err != nil {
		return "", errors.Wrap(err, "To JSON: CheckInReport")
	}

	return string(jsonData), nil
}

// Job for running scan in the job service with async way
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
		// Params are required
		return errors.New("missing parameter of scan job")
	}

	if _, err := extractRegistration(params); err != nil {
		return errors.Wrap(err, "job validate")
	}

	if _, err := ExtractScanReq(params); err != nil {
		return errors.Wrap(err, "job validate")
	}

	if _, err := extractMimeTypes(params); err != nil {
		return errors.Wrap(err, "job validate")
	}

	if _, err := extractRobotAccount(params); err != nil {
		return errors.Wrap(err, "job validate")
	}

	authType, err := extractAuthType(params)
	if err != nil {
		return errors.Wrap(err, "job validate")
	}

	if authType != authorizationBearer && authType != authorizationBasic {
		return errors.Wrapf(err, "job validate: not support auth type %s", authType)
	}

	return nil
}

// Run the job
func (j *Job) Run(ctx job.Context, params job.Parameters) error {
	// Get logger
	myLogger := ctx.GetLogger()
	startTime := time.Now()

	// shouldStop checks if the job should be stopped
	shouldStop := func() bool {
		if cmd, ok := ctx.OPCommand(); ok && cmd == job.StopCommand {
			myLogger.Info("scan job being stopped")
			return true
		}

		return false
	}

	// Ignore errors as they have been validated already
	r, _ := extractRegistration(params)
	req, _ := ExtractScanReq(params)
	mimeTypes, _ := extractMimeTypes(params)
	scanType := v1.ScanTypeVulnerability
	if len(req.RequestType) > 0 {
		scanType = req.RequestType[0].Type
	}
	handler := GetScanHandler(scanType)

	// Print related infos to log
	printJSONParameter(JobParamRegistration, removeRegistrationAuthInfo(r), myLogger)
	printJSONParameter(JobParameterRequest, removeScanAuthInfo(req), myLogger)
	myLogger.Infof("Report mime types: %v\n", mimeTypes)

	if shouldStop() {
		return nil
	}

	// Submit scan request to the scanner adapter
	client, err := r.Client(v1.DefaultClientPool)
	if err != nil {
		return logAndWrapError(myLogger, err, "scan job: get client")
	}

	// Ignore the namespace ID here
	req.Artifact.NamespaceID = 0

	robotAccount, _ := extractRobotAccount(params)

	var authorization string
	var tokenURL string

	authType, _ := extractAuthType(params)
	if authType == authorizationBearer {
		tokenURL, err = getInternalTokenServiceEndpoint(ctx)
		if err != nil {
			return errors.Wrap(err, "scan job: get token service endpoint")
		}
		authorization, err = makeBearerAuthorization(robotAccount, tokenURL, req.Artifact.Repository)
	} else {
		authorization, err = makeBasicAuthorization(robotAccount)
	}
	if err != nil {
		_ = logAndWrapError(myLogger, err, "scan job: make authorization")
	}

	if shouldStop() {
		return nil
	}

	req.Registry.Authorization = authorization
	resp, err := client.SubmitScan(req)
	if err != nil {
		return logAndWrapError(myLogger, err, "scan job: submit scan request")
	}

	// For collecting errors
	errs := make([]error, len(mimeTypes))
	rawReports := make([]string, len(mimeTypes))

	// Concurrently retrieving report by different mime types
	wg := &sync.WaitGroup{}
	wg.Add(len(mimeTypes))

	for i, mimeType := range mimeTypes {
		go func(i int, m string) {
			defer wg.Done()

			// Log info
			myLogger.Infof("Get report for mime type: %s", m)

			// Loop check if the report is ready
			tm := time.NewTimer(firstCheckInterval)
			defer tm.Stop()

			for {
				select {
				case t := <-tm.C:
					if shouldStop() {
						return
					}

					myLogger.Debugf("check scan report for mime %s at %s", m, t.Format("2006/01/02 15:04:05"))

					reportURLParameter, err := handler.URLParameter(req)
					if err != nil {
						errs[i] = errors.Wrap(err, "scan job: get report url")
						return
					}
					rawReport, err := fetchScanReportFromScanner(client, resp.ID, m, reportURLParameter)
					if err != nil {
						// Not ready yet
						if notReadyErr, ok := err.(*v1.ReportNotReadyError); ok {
							// Reset to the new check interval
							tm.Reset(time.Duration(notReadyErr.RetryAfter) * time.Second)
							myLogger.Infof("Report with mime type %s is not ready yet, retry after %d seconds", m, notReadyErr.RetryAfter)
							continue
						}
						errs[i] = errors.Wrap(err, fmt.Sprintf("scan job: fetch scan report, mimetype %v", m))
						return
					}
					rawReports[i] = rawReport
					return
				case <-ctx.SystemContext().Done():
					// Terminated by system
					return
				case <-time.After(checkTimeout):
					errs[i] = errors.New("check scan report timeout")
					return
				}
			}
		}(i, mimeType)
	}

	// Wait for all the retrieving routines are completed
	wg.Wait()

	if shouldStop() {
		return nil
	}

	// Merge errors
	for _, e := range errs {
		if e != nil {
			if err != nil {
				err = errors.Wrap(e, err.Error())
			} else {
				err = e
			}
		}
	}

	// Log error to the job log
	if err != nil {
		myLogger.Error(err)
		return err
	}

	for i, mimeType := range mimeTypes {
		rp, err := handler.GetPlaceHolder(ctx.SystemContext(), req.Artifact.Repository, req.Artifact.Digest, r.UUID, mimeType)
		if err != nil {
			return err
		}
		myLogger.Debugf("Converting report ID %s to the new V2 schema", rp.UUID)

		reportData, err := handler.PostScan(ctx, req, rp, rawReports[i], startTime, robotAccount)
		if err != nil {
			myLogger.Errorf("handler failed at PostScan, report %s, error %v", rp.UUID, err)
			return err
		}

		// update the original report with the new summarized report with all vulnerability data removed.
		// this is required since the top level layers relay on the vuln.Report struct that
		// contains additional metadata within the report which if stored in the new columns within the scan_report table
		// would be redundant
		if err := handler.Update(ctx.SystemContext(), rp.UUID, reportData); err != nil {
			myLogger.Errorf("Failed to update report data for report %s, error %v", rp.UUID, err)
			return err
		}
		myLogger.Debugf("Converted report ID %s to the new V2 schema", rp.UUID)
	}

	return nil
}

func fetchScanReportFromScanner(client v1.Client, requestID string, mimType string, urlParameter string) (rawReport string, err error) {
	rawReport, err = client.GetScanReport(requestID, mimType, urlParameter)
	if err != nil {
		return "", err
	}
	// Make sure the data is aligned with the v1 spec.
	if _, err = report.ResolveData(mimType, []byte(rawReport)); err != nil {
		return "", err
	}
	return rawReport, nil
}

// ExtractScanReq extracts the scan request from the job parameters.
func ExtractScanReq(params job.Parameters) (*v1.ScanRequest, error) {
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

	req := &v1.ScanRequest{}
	if err := req.FromJSON(jsonData); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	reqType := v1.ScanTypeVulnerability
	// attach the request with ProducesMimeTypes and Parameters
	if len(req.RequestType) > 0 {
		// current only support requestType with one element for each request
		if len(req.RequestType[0].Type) > 0 {
			reqType = req.RequestType[0].Type
		}
		handler := GetScanHandler(reqType)
		if handler == nil {
			return nil, errors.Errorf("failed to get scan handler, request type %v", reqType)
		}
		req.RequestType[0].ProducesMimeTypes = handler.RequestProducesMineTypes()
		req.RequestType[0].Parameters = handler.RequestParameters()
	}
	return req, nil
}

func logAndWrapError(logger logger.Interface, err error, message string) error {
	e := errors.Wrap(err, message)
	logger.Error(e)

	return e
}

func printJSONParameter(parameter string, v string, logger logger.Interface) {
	logger.Debugf("%s:\n", parameter)
	printPrettyJSON([]byte(v), logger)
}

func printPrettyJSON(in []byte, logger logger.Interface) {
	var out bytes.Buffer
	if err := json.Indent(&out, in, "", "  "); err != nil {
		logger.Errorf("Print pretty JSON error: %s", err)
		return
	}

	logger.Infof("%s\n", out.String())
}

func removeScanAuthInfo(sr *v1.ScanRequest) string {
	req := &v1.ScanRequest{
		Artifact: sr.Artifact,
		Registry: &v1.Registry{
			URL:           sr.Registry.URL,
			Authorization: "[HIDDEN]",
		},
		RequestType: sr.RequestType,
	}

	str, err := req.ToJSON()
	if err != nil {
		logger.Error(errors.Wrap(err, "scan job: remove auth for scan request"))
	}

	return str
}

func removeRegistrationAuthInfo(sr *scanner.Registration) string {
	req := &scanner.Registration{
		ID:               sr.ID,
		UUID:             sr.UUID,
		Name:             sr.Name,
		Description:      sr.Description,
		URL:              sr.URL,
		Disabled:         sr.Disabled,
		IsDefault:        sr.IsDefault,
		Health:           sr.Health,
		Auth:             sr.Auth,
		AccessCredential: "[HIDDEN]",
		SkipCertVerify:   sr.SkipCertVerify,
		UseInternalAddr:  sr.UseInternalAddr,
		Immutable:        sr.Immutable,
		Adapter:          sr.Adapter,
		Vendor:           sr.Vendor,
		Version:          sr.Version,
		Metadata:         sr.Metadata,
		CreateTime:       sr.CreateTime,
		UpdateTime:       sr.UpdateTime,
	}

	str, err := req.ToJSON()
	if err != nil {
		logger.Error(errors.Wrap(err, "scan job: remove auth for registration"))
	}

	return str
}

func extractRegistration(params job.Parameters) (*scanner.Registration, error) {
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

	r := &scanner.Registration{}
	if err := r.FromJSON(jsonData); err != nil {
		return nil, err
	}

	if err := r.Validate(true); err != nil {
		return nil, err
	}

	return r, nil
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

func extractMimeTypes(params job.Parameters) ([]string, error) {
	v, ok := params[JobParameterMimes]
	if !ok {
		return nil, errors.Errorf("missing job parameter '%s'", JobParameterMimes)
	}

	l, ok := v.([]any)
	if !ok {
		return nil, errors.Errorf(
			"malformed job parameter '%s', expecting []any but got %s",
			JobParameterMimes,
			reflect.TypeOf(v).String(),
		)
	}

	mimes := make([]string, 0)
	for _, v := range l {
		mime, ok := v.(string)
		if !ok {
			return nil, errors.Errorf("expect string but got %s", reflect.TypeOf(v).String())
		}

		mimes = append(mimes, mime)
	}

	return mimes, nil
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

// makeBasicAuthorization creates authorization from a robot account based on the arguments for scanning.
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
