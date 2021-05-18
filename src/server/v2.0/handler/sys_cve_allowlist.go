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

package handler

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/allowlist"
	"github.com/goharbor/harbor/src/pkg/allowlist/models"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"

	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/system_cve_allowlist"
)

type systemCVEAllowListAPI struct {
	BaseAPI
	mgr allowlist.Manager
}

func newSystemCVEAllowListAPI() *systemCVEAllowListAPI {
	return &systemCVEAllowListAPI{
		mgr: allowlist.NewDefaultManager(),
	}
}

func (s systemCVEAllowListAPI) PutSystemCVEAllowlist(ctx context.Context, params system_cve_allowlist.PutSystemCVEAllowlistParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceConfiguration); err != nil {
		return s.SendError(ctx, err)
	}
	l := models.CVEAllowlist{}
	l.ExpiresAt = params.Allowlist.ExpiresAt
	for _, it := range params.Allowlist.Items {
		l.Items = append(l.Items, models.CVEAllowlistItem{CVEID: it.CVEID})
	}
	if err := s.mgr.SetSys(ctx, l); err != nil {
		return s.SendError(ctx, err)
	}
	return system_cve_allowlist.NewPutSystemCVEAllowlistOK()
}

func (s systemCVEAllowListAPI) GetSystemCVEAllowlist(ctx context.Context, params system_cve_allowlist.GetSystemCVEAllowlistParams) middleware.Responder {
	if err := s.RequireAuthenticated(ctx); err != nil {
		return s.SendError(ctx, err)
	}
	l, err := s.mgr.GetSys(ctx)
	if err != nil {
		return s.SendError(ctx, err)
	}
	return system_cve_allowlist.NewGetSystemCVEAllowlistOK().WithPayload(model.NewCVEAllowlist(l).ToSwagger())
}
