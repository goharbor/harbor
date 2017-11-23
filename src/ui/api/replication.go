// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

	"github.com/vmware/harbor/src/replication/core"
	"github.com/vmware/harbor/src/ui/api/models"
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

	if !r.SecurityCtx.IsSysAdmin() {
		r.HandleForbidden(r.SecurityCtx.GetUsername())
		return
	}
}

// Post trigger a replication according to the specified policy
func (r *ReplicationAPI) Post() {
	replication := &models.Replication{}
	r.DecodeJSONReqAndValidate(replication)

	policy, err := core.DefaultController.GetPolicy(replication.PolicyID)
	if err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to get replication policy %d: %v", replication.PolicyID, err))
		return
	}

	if policy.ID == 0 {
		r.HandleNotFound(fmt.Sprintf("replication policy %d not found", replication.PolicyID))
		return
	}

	if err = core.DefaultController.Replicate(replication.PolicyID); err != nil {
		r.HandleInternalServerError(fmt.Sprintf("failed to trigger the replication policy %d: %v", replication.PolicyID, err))
		return
	}
}
