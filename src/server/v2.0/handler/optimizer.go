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
	"strings"

	"github.com/go-openapi/runtime/middleware"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/optimizer"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	"github.com/goharbor/harbor/src/pkg/scan/rest/auth"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/optimizer"
)

func newOptimizerAPI() *optimizerAPI {
	return &optimizerAPI{
		optimizerCtl: optimizer.DefaultController,
	}
}

type optimizerAPI struct {
	BaseAPI
	optimizerCtl optimizer.Controller
}

func (s *optimizerAPI) CreateOptimizer(ctx context.Context, params operation.CreateOptimizerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceOptimizer); err != nil {
		return s.SendError(ctx, err)
	}

	r := &dao.Registration{IsDefault: false}
	copyToOptimizerRegistration(r, params.Registration)

	if err := r.Validate(false); err != nil {
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage(err.Error()))
	}

	uuid, err := s.optimizerCtl.CreateRegistration(ctx, r)
	if err != nil {
		return s.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%s", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), uuid)
	return operation.NewCreateOptimizerCreated().WithLocation(location)
}

func (s *optimizerAPI) DeleteOptimizer(ctx context.Context, params operation.DeleteOptimizerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionDelete, rbac.ResourceOptimizer); err != nil {
		return s.SendError(ctx, err)
	}

	r, err := s.optimizerCtl.GetRegistration(ctx, params.RegistrationID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if r == nil {
		return s.SendError(ctx, errors.NotFoundError(nil).WithMessagef("optimizer %s not found", params.RegistrationID))
	}

	// Immutable registration is not allowed
	if r.Immutable {
		format := "registration %s is not allowed to delete as it is immutable: optimizer API: delete"
		return s.SendError(ctx, errors.ForbiddenError(nil).WithMessagef(format, r.Name))
	}

	deleted, err := s.optimizerCtl.DeleteRegistration(ctx, r.UUID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewDeleteOptimizerOK().WithPayload(model.NewOptimizerRegistration(deleted).ToSwagger(ctx))
}

func (s *optimizerAPI) GetOptimizer(ctx context.Context, params operation.GetOptimizerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceOptimizer); err != nil {
		return s.SendError(ctx, err)
	}

	r, err := s.optimizerCtl.GetRegistration(ctx, params.RegistrationID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if r == nil {
		return s.SendError(ctx, errors.NotFoundError(nil).WithMessagef("optimizer %s not found", params.RegistrationID))
	}

	return operation.NewGetOptimizerOK().WithPayload(model.NewOptimizerRegistration(r).ToSwagger(ctx))
}

func (s *optimizerAPI) GetOptimizerMetadata(ctx context.Context, params operation.GetOptimizerMetadataParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceOptimizer); err != nil {
		return s.SendError(ctx, err)
	}

	meta, err := s.optimizerCtl.GetMetadata(ctx, params.RegistrationID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewGetOptimizerMetadataOK().WithPayload(model.NewOptimizerMetadata(meta).ToSwagger(ctx))
}

func (s *optimizerAPI) ListOptimizers(ctx context.Context, params operation.ListOptimizersParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceOptimizer); err != nil {
		return s.SendError(ctx, err)
	}

	query, err := s.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return s.SendError(ctx, err)
	}

	values := params.HTTPRequest.URL.Query()
	for _, k := range []string{"name", "description", "url"} {
		if v := values.Get(k); v != "" {
			query.Keywords[k] = &q.FuzzyMatchValue{Value: v}
		}
	}

	for _, k := range []string{"ex_name", "ex_url"} {
		if v := values.Get(k); v != "" {
			query.Keywords[strings.TrimPrefix(k, "ex_")] = v
		}
	}

	total, err := s.optimizerCtl.GetTotalOfRegistrations(ctx, query)
	if err != nil {
		return s.SendError(ctx, err)
	}

	optimizers, err := s.optimizerCtl.ListRegistrations(ctx, query)
	if err != nil {
		return s.SendError(ctx, err)
	}

	payload := make([]*models.OptimizerRegistration, len(optimizers))
	for i, reg := range optimizers {
		payload[i] = model.NewOptimizerRegistration(reg).ToSwagger(ctx)
	}

	return operation.NewListOptimizersOK().
		WithXTotalCount(total).
		WithLink(s.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(payload)
}

func (s *optimizerAPI) PingOptimizer(ctx context.Context, params operation.PingOptimizerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceOptimizer); err != nil {
		return s.SendError(ctx, err)
	}

	r := &dao.Registration{
		Name:             lib.StringValue(params.Settings.Name),
		URL:              lib.StringValue((*string)(params.Settings.URL)),
		Auth:             params.Settings.Auth,
		AccessCredential: params.Settings.AccessCredential,
	}

	if err := r.Validate(false); err != nil {
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage(err.Error()))
	}

	if _, err := s.optimizerCtl.Ping(ctx, r); err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewPingOptimizerOK()
}

func (s *optimizerAPI) SetOptimizerAsDefault(ctx context.Context, params operation.SetOptimizerAsDefaultParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceOptimizer); err != nil {
		return s.SendError(ctx, err)
	}

	if params.Payload.IsDefault {
		if err := s.optimizerCtl.SetDefaultRegistration(ctx, params.RegistrationID); err != nil {
			return s.SendError(ctx, err)
		}
	}

	return operation.NewSetOptimizerAsDefaultOK()
}

func (s *optimizerAPI) UpdateOptimizer(ctx context.Context, params operation.UpdateOptimizerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceOptimizer); err != nil {
		return s.SendError(ctx, err)
	}

	r, err := s.optimizerCtl.GetRegistration(ctx, params.RegistrationID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if r == nil {
		return s.SendError(ctx, errors.NotFoundError(nil).WithMessagef("optimizer %s not found", params.RegistrationID))
	}

	// Immutable registration is not allowed
	if r.Immutable {
		format := "registration %s is not allowed to update as it is immutable: optimizer API: update"
		return s.SendError(ctx, errors.ForbiddenError(nil).WithMessagef(format, r.Name))
	}

	// GET does not return access_credential; an empty value in the update body means
	// "leave unchanged" when the resulting auth type still requires a credential.
	existingAccessCredential := r.AccessCredential
	copyToOptimizerRegistration(r, params.Registration)
	if params.Registration != nil && params.Registration.AccessCredential == "" &&
		optimizerAuthRequiresAccessCredential(r.Auth) {
		r.AccessCredential = existingAccessCredential
	}

	if err := r.Validate(true); err != nil {
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage(err.Error()))
	}

	if err := s.optimizerCtl.UpdateRegistration(ctx, r); err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewUpdateOptimizerOK()
}

func copyToOptimizerRegistration(r *dao.Registration, req *models.OptimizerRegistrationReq) {
	r.Name = lib.StringValue(req.Name)
	r.URL = lib.StringValue((*string)(req.URL))
	r.Description = req.Description
	r.Disabled = lib.BoolValue(req.Disabled)
	r.SkipCertVerify = lib.BoolValue(req.SkipCertVerify)
	r.UseInternalAddr = lib.BoolValue(req.UseInternalAddr)
	r.Auth = req.Auth
	r.AccessCredential = req.AccessCredential
}

// optimizerAuthRequiresAccessCredential matches dao.Registration.Validate: these auth types need a non-empty credential.
func optimizerAuthRequiresAccessCredential(authType string) bool {
	switch authType {
	case auth.Basic, auth.Bearer, auth.APIKey:
		return true
	default:
		return false
	}
}
