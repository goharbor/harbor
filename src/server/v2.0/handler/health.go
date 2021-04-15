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
	"github.com/goharbor/harbor/src/controller/health"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operations "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/health"
)

func newHealthAPI() *healthAPI {
	return &healthAPI{
		ctl: health.Ctl,
	}
}

type healthAPI struct {
	BaseAPI
	ctl health.Controller
}

func (r *healthAPI) GetHealth(ctx context.Context, params operations.GetHealthParams) middleware.Responder {
	status := r.ctl.GetHealth(ctx)
	s := &models.OverallHealthStatus{
		Status: status.Status,
	}
	for _, c := range status.Components {
		s.Components = append(s.Components, &models.ComponentHealthStatus{
			Error:  c.Error,
			Name:   c.Name,
			Status: c.Status,
		})
	}
	return operations.NewGetHealthOK().WithPayload(s)
}
