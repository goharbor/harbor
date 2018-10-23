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

package replication

import (
	"fmt"
	"net/http"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	reg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/logger"
)

// Replicator call UI's API to start a repliation according to the policy ID
// passed in parameters
type Replicator struct {
	ctx      env.JobContext
	url      string // the URL of UI service
	insecure bool
	policyID int64
	client   *common_http.Client
	logger   logger.Interface
}

// ShouldRetry ...
func (r *Replicator) ShouldRetry() bool {
	return false
}

// MaxFails ...
func (r *Replicator) MaxFails() uint {
	return 0
}

// Validate ....
func (r *Replicator) Validate(params map[string]interface{}) error {
	return nil
}

// Run ...
func (r *Replicator) Run(ctx env.JobContext, params map[string]interface{}) error {
	if err := r.init(ctx, params); err != nil {
		return err
	}
	return r.replicate()
}

func (r *Replicator) init(ctx env.JobContext, params map[string]interface{}) error {
	r.logger = ctx.GetLogger()
	r.ctx = ctx
	if canceled(r.ctx) {
		r.logger.Warning(errCanceled.Error())
		return errCanceled
	}

	r.policyID = (int64)(params["policy_id"].(float64))
	r.url = params["url"].(string)
	r.insecure = params["insecure"].(bool)
	cred := auth.NewSecretAuthorizer(secret())

	r.client = common_http.NewClient(&http.Client{
		Transport: reg.GetHTTPTransport(r.insecure),
	}, cred)

	r.logger.Infof("initialization completed: policy ID: %d, URL: %s, insecure: %v",
		r.policyID, r.url, r.insecure)

	return nil
}

func (r *Replicator) replicate() error {
	if err := r.client.Post(fmt.Sprintf("%s/api/replications", r.url), struct {
		PolicyID int64 `json:"policy_id"`
	}{
		PolicyID: r.policyID,
	}); err != nil {
		r.logger.Errorf("failed to send the replication request to %s: %v", r.url, err)
		return err
	}
	r.logger.Info("the replication request has been sent successfully")
	return nil

}
