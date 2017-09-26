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
	"net/http"

	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/security"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/filter"
	"github.com/vmware/harbor/src/ui/promgr"
)

// BaseController ...
type BaseController struct {
	api.BaseAPI
	// SecurityCtx is the security context used to authN &authZ
	SecurityCtx security.Context
	// ProjectMgr is the project manager which abstracts the operations
	// related to projects
	ProjectMgr promgr.ProjectManager
}

const (
	//ReplicationJobType ...
	ReplicationJobType = "replication"
	//ScanJobType ...
	ScanJobType = "scan"
)

// Prepare inits security context and project manager from request
// context
func (b *BaseController) Prepare() {
	ctx, err := filter.GetSecurityContext(b.Ctx.Request)
	if err != nil {
		log.Errorf("failed to get security context: %v", err)
		b.CustomAbort(http.StatusInternalServerError, "")
	}
	b.SecurityCtx = ctx

	pm, err := filter.GetProjectManager(b.Ctx.Request)
	if err != nil {
		log.Errorf("failed to get project manager: %v", err)
		b.CustomAbort(http.StatusInternalServerError, "")
	}
	b.ProjectMgr = pm
}
