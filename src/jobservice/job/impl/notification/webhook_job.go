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
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
)

// WebhookJob implements the job interface, which send notification by http or https.
type WebhookJob struct {
	client *http.Client
	logger logger.Interface
	ctx    job.Context
}

// MaxFails returns that how many times this job can fail, get this value from ctx.
func (wj *WebhookJob) MaxFails() (result uint) {
	// Default max fails count is 3
	result = 3
	if maxFails, exist := os.LookupEnv(maxFails); exist {
		mf, err := strconv.ParseUint(maxFails, 10, 32)
		if err != nil {
			logger.Warningf("Fetch webhook job maxFails error: %s", err.Error())
			return result
		}
		result = uint(mf)
	}
	return result
}

// MaxCurrency is implementation of same method in Interface.
func (wj *WebhookJob) MaxCurrency() uint {
	return 0
}

// ShouldRetry ...
func (wj *WebhookJob) ShouldRetry() bool {
	return true
}

// Validate implements the interface in job/Interface
func (wj *WebhookJob) Validate(_ job.Parameters) error {
	return nil
}

// Run implements the interface in job/Interface
func (wj *WebhookJob) Run(ctx job.Context, params job.Parameters) error {
	if err := wj.init(ctx, params); err != nil {
		return err
	}

	wj.logger.Info("start to run webhook job")

	if err := wj.execute(ctx, params); err != nil {
		wj.logger.Errorf("exit webhook job, error: %s", err)
		return err
	}

	wj.logger.Info("success to run webhook job")
	return nil
}

// init webhook job
func (wj *WebhookJob) init(ctx job.Context, params map[string]interface{}) error {
	wj.logger = ctx.GetLogger()
	wj.ctx = ctx

	// default use secure transport
	wj.client = httpHelper.clients[secure]
	if v, ok := params["skip_cert_verify"]; ok {
		if skipCertVerify, ok := v.(bool); ok && skipCertVerify {
			// if skip cert verify is true, it means not verify remote cert, use insecure client
			wj.client = httpHelper.clients[insecure]
		}
	}
	return nil
}

// execute webhook job
func (wj *WebhookJob) execute(_ job.Context, params map[string]interface{}) error {
	payload := params["payload"].(string)
	address := params["address"].(string)

	req, err := http.NewRequest(http.MethodPost, address, bytes.NewReader([]byte(payload)))
	if err != nil {
		return errors.Wrap(err, "error to generate request")
	}

	if h, ok := params["header"].(string); ok {
		header := make(http.Header)
		if err = json.Unmarshal([]byte(h), &header); err != nil {
			return errors.Wrap(err, "error to unmarshal header")
		}
		req.Header = header
	}

	wj.logger.Infof("send request to remote endpoint, body: %s", payload)

	resp, err := wj.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "error to send request")
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			wj.logger.Errorf("error to read response body, error: %s", err)
		}

		return errors.Errorf("abnormal response code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
