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
	"os"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	reg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/replication/model"
)

// Scheduler is a job running in Jobservice which can be used as
// a scheduler when submitting it as a scheduled job. It receives
// a URL and data, and post the data to the URL when it is running
type Scheduler struct {
	ctx job.Context
}

// ShouldRetry ...
func (s *Scheduler) ShouldRetry() bool {
	return false
}

// MaxFails ...
func (s *Scheduler) MaxFails() uint {
	return 0
}

// Validate ....
func (s *Scheduler) Validate(params job.Parameters) error {
	return nil
}

// Run ...
func (s *Scheduler) Run(ctx job.Context, params job.Parameters) error {
	cmd, exist := ctx.OPCommand()
	if exist && cmd == job.StopCommand {
		return nil
	}
	logger := ctx.GetLogger()

	url := params["url"].(string)
	url = fmt.Sprintf("%s/api/replication/executions?trigger=%s", url, model.TriggerTypeScheduled)
	policyID := (int64)(params["policy_id"].(float64))
	cred := auth.NewSecretAuthorizer(os.Getenv("JOBSERVICE_SECRET"))
	client := common_http.NewClient(&http.Client{
		Transport: reg.GetHTTPTransport(true),
	}, cred)
	if err := client.Post(url, struct {
		PolicyID int64 `json:"policy_id"`
	}{
		PolicyID: policyID,
	}); err != nil {
		logger.Errorf("failed to run the schedule job: %v", err)
		return err
	}
	logger.Info("the schedule job finished")
	return nil
}
