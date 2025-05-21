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

package notification

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
)

// SlackJob implements the job interface, which send notification to slack by slack incoming webhooks.
type SlackJob struct {
	client *http.Client
	logger logger.Interface
}

// MaxFails returns that how many times this job can fail.
func (sj *SlackJob) MaxFails() (result uint) {
	// Default max fails count is 3
	result = 3
	if maxFails, exist := os.LookupEnv(maxFails); exist {
		mf, err := strconv.ParseUint(maxFails, 10, 32)
		if err != nil {
			logger.Warningf("Fetch slack job maxFails error: %s", err.Error())
			return result
		}
		result = uint(mf)
	}
	return result
}

// MaxCurrency is implementation of same method in Interface.
func (sj *SlackJob) MaxCurrency() uint {
	return 1
}

// ShouldRetry ...
func (sj *SlackJob) ShouldRetry() bool {
	return true
}

// Validate implements the interface in job/Interface
func (sj *SlackJob) Validate(params job.Parameters) error {
	if params == nil {
		// Params are required
		return errors.New("missing parameter of slack job")
	}

	payload, ok := params["payload"]
	if !ok {
		return errors.Errorf("missing job parameter 'payload'")
	}
	_, ok = payload.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'payload', expecting string but got %s", reflect.TypeOf(payload).String())
	}

	address, ok := params["address"]
	if !ok {
		return errors.Errorf("missing job parameter 'address'")
	}
	_, ok = address.(string)
	if !ok {
		return errors.Errorf("malformed job parameter 'address', expecting string but got %s", reflect.TypeOf(address).String())
	}
	return nil
}

// Run implements the interface in job/Interface
func (sj *SlackJob) Run(ctx job.Context, params job.Parameters) error {
	if err := sj.init(ctx, params); err != nil {
		return err
	}

	sj.logger.Info("start to run slack job")

	err := sj.execute(params)
	if err != nil {
		sj.logger.Errorf("exit slack job, error: %s", err)
	} else {
		sj.logger.Info("success to run slack job")
	}
	// Wait a second for slack rate limit, refer to https://api.slack.com/docs/rate-limits
	time.Sleep(time.Second)
	return err
}

// init slack job
func (sj *SlackJob) init(ctx job.Context, params map[string]any) error {
	sj.logger = ctx.GetLogger()

	// default use secure transport
	sj.client = httpHelper.clients[secure]
	if v, ok := params["skip_cert_verify"]; ok {
		if skipCertVerify, ok := v.(bool); ok && skipCertVerify {
			// if skip cert verify is true, it means not verify remote cert, use insecure client
			sj.client = httpHelper.clients[insecure]
		}
	}
	return nil
}

// execute slack job
func (sj *SlackJob) execute(params map[string]any) error {
	payload := params["payload"].(string)
	address := params["address"].(string)

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader([]byte(payload)))
	if err != nil {
		return errors.Wrap(err, "error to generate request")
	}
	req.Header.Set("Content-Type", "application/json")

	sj.logger.Infof("send request to remote endpoint, body: %s", payload)

	resp, err := sj.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error to send request")
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			sj.logger.Errorf("error to read response body, error: %s", err)
		}

		return errors.Errorf("abnormal response code: %d, body: %s", resp.StatusCode, string(body))
	}
	return nil
}
