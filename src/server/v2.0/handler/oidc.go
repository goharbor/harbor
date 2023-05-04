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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	oidcpkg "github.com/goharbor/harbor/src/pkg/oidc"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/oidc"
)

type oidcAPI struct {
	BaseAPI
}

func newOIDCAPI() *oidcAPI {
	return &oidcAPI{}
}

func (o oidcAPI) PingOIDC(ctx context.Context, params oidc.PingOIDCParams) middleware.Responder {
	if err := o.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceConfiguration); err != nil {
		return o.SendError(ctx, err)
	}
	err := oidcpkg.TestEndpoint(oidcpkg.Conn{
		URL:        params.Endpoint.URL,
		VerifyCert: params.Endpoint.VerifyCert,
	})

	if err != nil {
		log.Errorf("Failed to verify connection: %+v, err: %v", params.Endpoint, err)
		return o.SendError(ctx, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("failed to verify connection"))
	}
	return oidc.NewPingOIDCOK()
}
