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
	"fmt"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/acceleration"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/acceleration/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/acceleration"
)

func newAccelerationAPI() *accelerationAPI {
	return &accelerationAPI{
		mgr: acceleration.Mgr,
	}
}

type accelerationAPI struct {
	BaseAPI
	mgr acceleration.Manager
}

func (r *accelerationAPI) CreateAccelerationService(ctx context.Context, params operation.CreateAccelerationServiceParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceAcceleration); err != nil {
		return r.SendError(ctx, err)
	}
	accel := &model.AccelerationService{
		Name:        params.Acceleration.Name,
		Description: params.Acceleration.Description,
		Type:        params.Acceleration.Type,
		URL:         params.Acceleration.URL,
		Insecure:    params.Acceleration.Insecure,
	}
	if params.Acceleration.Credential != nil {
		accel.Credential = &model.Credential{
			Type:         params.Acceleration.Credential.Type,
			AccessKey:    params.Acceleration.Credential.AccessKey,
			AccessSecret: params.Acceleration.Credential.AccessSecret,
		}
	}

	id, err := r.mgr.Create(ctx, accel)
	if err != nil {
		return r.SendError(ctx, err)
	}
	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateAccelerationServiceCreated().WithLocation(location)
}

func (r *accelerationAPI) GetAccelerationService(ctx context.Context, params operation.GetAccelerationServiceParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceAcceleration); err != nil {
		return r.SendError(ctx, err)
	}

	accel, err := r.mgr.Get(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewGetAccelerationServiceOK().WithPayload(convertAcceleration(accel))
}

func (r *accelerationAPI) ListAccelerationServices(ctx context.Context, params operation.ListAccelerationServicesParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceAcceleration); err != nil {
		return r.SendError(ctx, err)
	}

	query, err := r.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return r.SendError(ctx, err)
	}
	// keep backward compatibility for the "name" query
	if params.Name != nil {
		query.Keywords["Name"] = q.NewFuzzyMatchValue(*params.Name)
	}

	total, err := r.mgr.Count(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	accs, err := r.mgr.List(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	var accels []*models.Acceleration
	for _, acc := range accs {
		accels = append(accels, convertAcceleration(acc))
	}
	return operation.NewListAccelerationServicesOK().WithXTotalCount(total).
		WithLink(r.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(accels)
}

func (r *accelerationAPI) DeleteAccelerationService(ctx context.Context, params operation.DeleteAccelerationServiceParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionDelete, rbac.ResourceAcceleration); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.mgr.Delete(ctx, params.ID); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewDeleteAccelerationServiceOK()
}

func (r *accelerationAPI) UpdateAccelerationService(ctx context.Context, params operation.UpdateAccelerationServiceParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceAcceleration); err != nil {
		return r.SendError(ctx, err)
	}
	accel, err := r.mgr.Get(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	if params.Acceleration != nil {
		if params.Acceleration.Name != nil {
			accel.Name = *params.Acceleration.Name
		}
		if params.Acceleration.Description != nil {
			accel.Description = *params.Acceleration.Description
		}
		if params.Acceleration.URL != nil {
			accel.URL = *params.Acceleration.URL
		}
		if params.Acceleration.Insecure != nil {
			accel.Insecure = *params.Acceleration.Insecure
		}
		if accel.Credential == nil {
			accel.Credential = &model.Credential{}
		}
		if params.Acceleration.CredentialType != nil {
			accel.Credential.Type = *params.Acceleration.CredentialType
		}
		if params.Acceleration.AccessKey != nil {
			accel.Credential.AccessKey = *params.Acceleration.AccessKey
		}
		if params.Acceleration.AccessSecret != nil {
			accel.Credential.AccessSecret = *params.Acceleration.AccessSecret
		}
	}
	if err := r.mgr.Update(ctx, accel); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewUpdateAccelerationServiceOK()
}

func (r *accelerationAPI) PingAccelerationService(ctx context.Context, params operation.PingAccelerationServiceParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceAcceleration); err != nil {
		return r.SendError(ctx, err)
	}

	return operation.NewPingAccelerationServiceOK()
}

func convertAcceleration(registry *model.AccelerationService) *models.Acceleration {
	r := &models.Acceleration{
		CreationTime: strfmt.DateTime(registry.CreationTime),
		Description:  registry.Description,
		ID:           registry.ID,
		Insecure:     registry.Insecure,
		Name:         registry.Name,
		Status:       registry.Status,
		Type:         string(registry.Type),
		UpdateTime:   strfmt.DateTime(registry.UpdateTime),
		URL:          registry.URL,
	}
	if registry.Credential != nil {
		credential := &models.AccelerationCredential{
			AccessKey: registry.Credential.AccessKey,
			Type:      string(registry.Credential.Type),
		}
		if len(registry.Credential.AccessSecret) > 0 {
			credential.AccessSecret = "*****"
		}
		r.Credential = credential
	}
	return r
}
