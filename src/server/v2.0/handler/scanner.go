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
	"github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/scanner"
)

func newScannerAPI() *scannerAPI {
	return &scannerAPI{
		scannerCtl: scanner.DefaultController,
	}
}

type scannerAPI struct {
	BaseAPI
	scannerCtl scanner.Controller
}

func (s *scannerAPI) CreateScanner(ctx context.Context, params operation.CreateScannerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionCreate, rbac.ResourceScanner); err != nil {
		return s.SendError(ctx, err)
	}

	r := &scanner.Registration{IsDefault: false}
	copyToScannerRegistration(r, params.Registration)

	if err := r.Validate(false); err != nil {
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage(err.Error()))
	}

	uuid, err := s.scannerCtl.CreateRegistration(ctx, r)
	if err != nil {
		return s.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%s", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), uuid)
	return operation.NewCreateScannerCreated().WithLocation(location)
}

func (s *scannerAPI) DeleteScanner(ctx context.Context, params operation.DeleteScannerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionDelete, rbac.ResourceScanner); err != nil {
		return s.SendError(ctx, err)
	}

	r, err := s.scannerCtl.GetRegistration(ctx, params.RegistrationID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if r == nil {
		return s.SendError(ctx, errors.NotFoundError(nil).WithMessage("scanner %s not found", params.RegistrationID))
	}

	// Immutable registration is not allowed
	if r.Immutable {
		format := "registration %s is not allowed to delete as it is immutable: scanner API: delete"
		return s.SendError(ctx, errors.ForbiddenError(nil).WithMessage(format, r.Name))
	}

	deleted, err := s.scannerCtl.DeleteRegistration(ctx, r.UUID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewDeleteScannerOK().WithPayload(model.NewScannerRegistration(deleted).ToSwagger(ctx))
}

func (s *scannerAPI) GetScanner(ctx context.Context, params operation.GetScannerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceScanner); err != nil {
		return s.SendError(ctx, err)
	}

	r, err := s.scannerCtl.GetRegistration(ctx, params.RegistrationID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if r == nil {
		return s.SendError(ctx, errors.NotFoundError(nil).WithMessage("scanner %s not found", params.RegistrationID))
	}

	return operation.NewGetScannerOK().WithPayload(model.NewScannerRegistration(r).ToSwagger(ctx))
}

func (s *scannerAPI) GetScannerMetadata(ctx context.Context, params operation.GetScannerMetadataParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceScanner); err != nil {
		return s.SendError(ctx, err)
	}

	meta, err := s.scannerCtl.GetMetadata(ctx, params.RegistrationID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewGetScannerMetadataOK().WithPayload(model.NewScannerMetadata(meta).ToSwagger(ctx))
}

func (s *scannerAPI) ListScanners(ctx context.Context, params operation.ListScannersParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionList, rbac.ResourceScanner); err != nil {
		return s.SendError(ctx, err)
	}

	query, err := s.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return s.SendError(ctx, err)
	}

	// compatible with previous version list scanners API
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

	total, err := s.scannerCtl.GetTotalOfRegistrations(ctx, query)
	if err != nil {
		return s.SendError(ctx, err)
	}

	scanners, err := s.scannerCtl.ListRegistrations(ctx, query)
	if err != nil {
		return s.SendError(ctx, err)
	}

	payload := make([]*models.ScannerRegistration, len(scanners))
	for i, scanner := range scanners {
		payload[i] = model.NewScannerRegistration(scanner).ToSwagger(ctx)
	}

	return operation.NewListScannersOK().
		WithXTotalCount(total).
		WithLink(s.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(payload)
}

func (s *scannerAPI) PingScanner(ctx context.Context, params operation.PingScannerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceScanner); err != nil {
		return s.SendError(ctx, err)
	}

	r := &scanner.Registration{
		Name:             lib.StringValue(params.Settings.Name),
		URL:              lib.StringValue((*string)(params.Settings.URL)),
		Auth:             params.Settings.Auth,
		AccessCredential: params.Settings.AccessCredential,
	}

	if err := r.Validate(false); err != nil {
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage(err.Error()))
	}

	if _, err := s.scannerCtl.Ping(ctx, r); err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewPingScannerOK()
}

func (s *scannerAPI) SetScannerAsDefault(ctx context.Context, params operation.SetScannerAsDefaultParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceScanner); err != nil {
		return s.SendError(ctx, err)
	}

	if params.Payload.IsDefault {
		if err := s.scannerCtl.SetDefaultRegistration(ctx, params.RegistrationID); err != nil {
			return s.SendError(ctx, err)
		}
	}

	return operation.NewSetScannerAsDefaultOK()
}

func (s *scannerAPI) UpdateScanner(ctx context.Context, params operation.UpdateScannerParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionUpdate, rbac.ResourceScanner); err != nil {
		return s.SendError(ctx, err)
	}

	r, err := s.scannerCtl.GetRegistration(ctx, params.RegistrationID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if r == nil {
		return s.SendError(ctx, errors.NotFoundError(nil).WithMessage("scanner %s not found", params.RegistrationID))
	}

	// Immutable registration is not allowed
	if r.Immutable {
		format := "registration %s is not allowed to update as it is immutable: scanner API: update"
		return s.SendError(ctx, errors.ForbiddenError(nil).WithMessage(format, r.Name))
	}

	copyToScannerRegistration(r, params.Registration)

	if err := r.Validate(true); err != nil {
		return s.SendError(ctx, errors.BadRequestError(nil).WithMessage(err.Error()))
	}

	if err := s.scannerCtl.UpdateRegistration(ctx, r); err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewUpdateScannerOK()
}

func copyToScannerRegistration(r *scanner.Registration, req *models.ScannerRegistrationReq) {
	r.Name = lib.StringValue(req.Name)
	r.URL = lib.StringValue((*string)(req.URL))
	r.Description = req.Description
	r.Disabled = lib.BoolValue(req.Disabled)
	r.SkipCertVerify = lib.BoolValue(req.SkipCertVerify)
	r.UseInternalAddr = lib.BoolValue(req.UseInternalAddr)
	r.Auth = req.Auth
	r.AccessCredential = req.AccessCredential
}
