// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	api_models "github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/replication/core"
	"github.com/goharbor/harbor/src/replication/event/notification"
	"github.com/goharbor/harbor/src/replication/event/topic"
)

// ReplicationAPI handles API calls for replication
type ReplicationAPI struct {
	BaseController
}

// Prepare does authentication and authorization works
func (r *ReplicationAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsAuthenticated() {
		r.HandleUnauthorized()
		return
	}

	if !r.SecurityCtx.IsSysAdmin() && !r.SecurityCtx.IsSolutionUser() {
		r.HandleForbidden(r.SecurityCtx.GetUsername())
		return
	}
}

// Post trigger a replication according to the specified policy
func (r *ReplicationAPI) Post() {
	replication := &api_models.Replication{}
	r.DecodeJSONReqAndValidate(replication)

	policy, err := core.GlobalController.GetPolicy(replication.PolicyID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get replication policy %d: %v", replication.PolicyID, err))
		return
	}

	if policy.ID == 0 {
		r.HandleNotFound(fmt.Sprintf("replication policy %d not found", replication.PolicyID))
		return
	}

	count, err := dao.GetTotalCountOfRepJobs(&models.RepJobQuery{
		PolicyID:   replication.PolicyID,
		Statuses:   []string{models.JobPending, models.JobRunning},
		Operations: []string{models.RepOpTransfer, models.RepOpDelete},
	})
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to filter jobs of policy %d: %v",
			replication.PolicyID, err))
		return
	}
	if count > 0 {
		r.RenderError(http.StatusPreconditionFailed, "policy has running/pending jobs, new replication can not be triggered")
		return
	}

	if err = startReplication(replication.PolicyID); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to publish replication topic for policy %d: %v", replication.PolicyID, err))
		return
	}
	log.Infof("replication signal for policy %d sent", replication.PolicyID)
}

func startReplication(policyID int64) error {
	return notifier.Publish(topic.StartReplicationTopic,
		notification.StartReplicationNotification{
			PolicyID: policyID,
		})
}
