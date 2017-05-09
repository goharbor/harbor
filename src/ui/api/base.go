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
	"github.com/vmware/harbor/src/ui/projectmanager"
)

// BaseController ...
type BaseController struct {
	api.BaseAPI
	// SecurityCxt is the security context used to authN &authZ
	SecurityCxt security.Context
	// ProManager is the project manager which abstracts the operations
	// related to projects
	ProManager projectmanager.ProjectManager
}

// Prepare inits security context and project manager from beego
// context
func (b *BaseController) Prepare() {
	ok := false
	ctx := b.Ctx.Input.GetData(filter.HarborSecurityContext)
	b.SecurityCxt, ok = ctx.(security.Context)
	if !ok {
		log.Error("failed to get security context")
		b.CustomAbort(http.StatusInternalServerError, "")
	}

	pm := b.Ctx.Input.GetData(filter.HarborProjectManager)
	b.ProManager, ok = pm.(projectmanager.ProjectManager)
	if !ok {
		log.Error("failed to get project manager")
		b.CustomAbort(http.StatusInternalServerError, "")
	}
}
