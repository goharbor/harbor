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
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/registry"
)

func newRegistryAPI() *registryAPI {
	return &registryAPI{
		ctl: registry.Ctl,
	}
}

type registryAPI struct {
	BaseAPI
	ctl registry.Controller
}

func (r *registryAPI) CreateRegistry(ctx context.Context, params operation.CreateRegistryParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceRegistry); err != nil {
		return r.SendError(ctx, err)
	}
	registry := &model.Registry{
		Name:        params.Registry.Name,
		Description: params.Registry.Description,
		Type:        params.Registry.Type,
		URL:         params.Registry.URL,
		Insecure:    params.Registry.Insecure,
	}
	if params.Registry.Credential != nil {
		registry.Credential = &model.Credential{
			Type:         params.Registry.Credential.Type,
			AccessKey:    params.Registry.Credential.AccessKey,
			AccessSecret: params.Registry.Credential.AccessSecret,
		}
	}

	id, err := r.ctl.Create(ctx, registry)
	if err != nil {
		return r.SendError(ctx, err)
	}
	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateRegistryCreated().WithLocation(location)
}

func (r *registryAPI) GetRegistry(ctx context.Context, params operation.GetRegistryParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceRegistry); err != nil {
		return r.SendError(ctx, err)
	}

	registry, err := r.ctl.Get(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewGetRegistryOK().WithPayload(convertRegistry(registry))
}

func (r *registryAPI) ListRegistries(ctx context.Context, params operation.ListRegistriesParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceRegistry); err != nil {
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

	total, err := r.ctl.Count(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	registries, err := r.ctl.List(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	var regs []*models.Registry
	for _, registry := range registries {
		regs = append(regs, convertRegistry(registry))
	}
	return operation.NewListRegistriesOK().WithXTotalCount(total).
		WithLink(r.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(regs)
}

func (r *registryAPI) DeleteRegistry(ctx context.Context, params operation.DeleteRegistryParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionDelete, rbac.ResourceRegistry); err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.ctl.Delete(ctx, params.ID); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewDeleteRegistryOK()
}

func (r *registryAPI) UpdateRegistry(ctx context.Context, params operation.UpdateRegistryParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceRegistry); err != nil {
		return r.SendError(ctx, err)
	}
	registry, err := r.ctl.Get(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, err)
	}
	if params.Registry != nil {
		if params.Registry.Name != nil {
			registry.Name = *params.Registry.Name
		}
		if params.Registry.Description != nil {
			registry.Description = *params.Registry.Description
		}
		if params.Registry.URL != nil {
			registry.URL = *params.Registry.URL
		}
		if params.Registry.Insecure != nil {
			registry.Insecure = *params.Registry.Insecure
		}
		if registry.Credential == nil {
			registry.Credential = &model.Credential{}
		}
		if params.Registry.CredentialType != nil {
			registry.Credential.Type = *params.Registry.CredentialType
		}
		if params.Registry.AccessKey != nil {
			registry.Credential.AccessKey = *params.Registry.AccessKey
		}
		if params.Registry.AccessSecret != nil {
			registry.Credential.AccessSecret = *params.Registry.AccessSecret
		}
	}
	if err := r.ctl.Update(ctx, registry); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewUpdateRegistryOK()
}

func (r *registryAPI) GetRegistryInfo(ctx context.Context, params operation.GetRegistryInfoParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceRegistry); err != nil {
		return r.SendError(ctx, err)
	}

	info, err := r.ctl.GetInfo(ctx, params.ID)
	if err != nil {
		return r.SendError(ctx, err)
	}

	in := &models.RegistryInfo{
		Description: info.Description,
		Type:        string(info.Type),
	}
	for _, filter := range info.SupportedResourceFilters {
		in.SupportedResourceFilters = append(in.SupportedResourceFilters, &models.FilterStyle{
			Style:  filter.Style,
			Type:   string(filter.Type),
			Values: filter.Values,
		})
	}
	for _, trigger := range info.SupportedTriggers {
		in.SupportedTriggers = append(in.SupportedTriggers, string(trigger))
	}
	return operation.NewGetRegistryInfoOK().WithPayload(in)
}

func (r *registryAPI) ListRegistryProviderTypes(ctx context.Context, params operation.ListRegistryProviderTypesParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceReplicationAdapter); err != nil {
		return r.SendError(ctx, err)
	}

	types, err := r.ctl.ListRegistryProviderTypes(ctx)
	if err != nil {
		return r.SendError(ctx, err)
	}

	return operation.NewListRegistryProviderTypesOK().WithPayload(types)
}

func (r *registryAPI) PingRegistry(ctx context.Context, params operation.PingRegistryParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceRegistry); err != nil {
		return r.SendError(ctx, err)
	}

	registry := &model.Registry{}
	var err error
	if params.Registry != nil {
		if params.Registry.ID != nil {
			registry, err = r.ctl.Get(ctx, *params.Registry.ID)
			if err != nil {
				return r.SendError(ctx, err)
			}
		}
		if params.Registry.Type != nil {
			registry.Type = *params.Registry.Type
		}
		if params.Registry.URL != nil {
			url, err := lib.ValidateHTTPURL(*params.Registry.URL)
			if err != nil {
				return r.SendError(ctx, err)
			}
			registry.URL = url
		}
		if params.Registry.Insecure != nil {
			registry.Insecure = *params.Registry.Insecure
		}
		if params.Registry.CredentialType != nil {
			if registry.Credential == nil {
				registry.Credential = &model.Credential{}
			}
			registry.Credential.Type = *params.Registry.CredentialType
		}
		if params.Registry.AccessKey != nil {
			if registry.Credential == nil {
				registry.Credential = &model.Credential{}
			}
			registry.Credential.AccessKey = *params.Registry.AccessKey
		}
		if params.Registry.AccessSecret != nil {
			if registry.Credential == nil {
				registry.Credential = &model.Credential{}
			}
			registry.Credential.AccessSecret = *params.Registry.AccessSecret
		}
	}

	if len(registry.Type) == 0 || len(registry.URL) == 0 {
		return r.SendError(ctx, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("type or url cannot be empty"))
	}

	healthy, err := r.ctl.IsHealthy(ctx, registry)
	if err != nil {
		return r.SendError(ctx, err)
	}

	if !healthy {
		return r.SendError(ctx, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("the registry is unhealthy"))
	}

	return operation.NewPingRegistryOK()
}

func (r *registryAPI) ListRegistryProviderInfos(ctx context.Context, params operation.ListRegistryProviderInfosParams) middleware.Responder {
	if err := r.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceReplicationAdapter); err != nil {
		return r.SendError(ctx, err)
	}

	infos, err := r.ctl.ListRegistryProviderInfos(ctx)
	if err != nil {
		return r.SendError(ctx, err)
	}

	result := map[string]models.RegistryProviderInfo{}
	for key, info := range infos {
		item := models.RegistryProviderInfo{}
		if info.CredentialPattern != nil {
			item.CredentialPattern = &models.RegistryProviderCredentialPattern{
				AccessKeyData:    info.CredentialPattern.AccessKeyData,
				AccessKeyType:    info.CredentialPattern.AccessKeyType,
				AccessSecretData: info.CredentialPattern.AccessSecretData,
				AccessSecretType: info.CredentialPattern.AccessSecretType,
			}
		}
		if info.EndpointPattern != nil {
			item.EndpointPattern = &models.RegistryProviderEndpointPattern{
				EndpointType: info.EndpointPattern.EndpointType,
			}
			for _, endpoint := range info.EndpointPattern.Endpoints {
				item.EndpointPattern.Endpoints = append(item.EndpointPattern.Endpoints, &models.RegistryEndpoint{
					Key:   endpoint.Key,
					Value: endpoint.Value,
				})
			}
		}
		result[key] = item
	}

	return operation.NewListRegistryProviderInfosOK().WithPayload(result)
}
