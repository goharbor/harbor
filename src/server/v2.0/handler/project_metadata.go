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
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/project/metadata"
	"github.com/goharbor/harbor/src/lib/errors"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project_metadata"
	"strconv"
	"strings"
)

func newProjectMetadaAPI() *projectMetadataAPI {
	return &projectMetadataAPI{
		ctl:    metadata.Ctl,
		proCtl: project.Ctl,
	}
}

type projectMetadataAPI struct {
	BaseAPI
	ctl    metadata.Controller
	proCtl project.Controller
}

func (p *projectMetadataAPI) AddProjectMetadatas(ctx context.Context, params operation.AddProjectMetadatasParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := p.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceMetadata); err != nil {
		return p.SendError(ctx, err)
	}
	metadata := params.Metadata
	metadata, err := p.validate(metadata)
	if err != nil {
		return p.SendError(ctx, err)
	}
	project, err := p.proCtl.Get(ctx, projectNameOrID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	if err = p.ctl.Add(ctx, project.ProjectID, metadata); err != nil {
		return p.SendError(ctx, err)
	}
	return operation.NewAddProjectMetadatasOK()
}

func (p *projectMetadataAPI) ListProjectMetadatas(ctx context.Context, params operation.ListProjectMetadatasParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := p.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceMetadata); err != nil {
		return p.SendError(ctx, err)
	}
	project, err := p.proCtl.Get(ctx, projectNameOrID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	metadata, err := p.ctl.Get(ctx, project.ProjectID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	return operation.NewListProjectMetadatasOK().WithPayload(metadata)
}

func (p *projectMetadataAPI) DeleteProjectMetadata(ctx context.Context, params operation.DeleteProjectMetadataParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := p.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete, rbac.ResourceMetadata); err != nil {
		return p.SendError(ctx, err)
	}
	project, err := p.proCtl.Get(ctx, projectNameOrID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	if err = p.ctl.Delete(ctx, project.ProjectID, params.MetaName); err != nil {
		return p.SendError(ctx, err)
	}
	return operation.NewDeleteProjectMetadataOK()
}

func (p *projectMetadataAPI) GetProjectMetadata(ctx context.Context, params operation.GetProjectMetadataParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := p.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceMetadata); err != nil {
		return p.SendError(ctx, err)
	}
	project, err := p.proCtl.Get(ctx, projectNameOrID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	metadata, err := p.ctl.Get(ctx, project.ProjectID, params.MetaName)
	if err != nil {
		return p.SendError(ctx, err)
	}
	return operation.NewGetProjectMetadataOK().WithPayload(metadata)
}

func (p *projectMetadataAPI) UpdateProjectMetadata(ctx context.Context, params operation.UpdateProjectMetadataParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := p.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceMetadata); err != nil {
		return p.SendError(ctx, err)
	}
	metadata := map[string]string{
		params.MetaName: params.Metadata[params.MetaName],
	}
	metadata, err := p.validate(metadata)
	if err != nil {
		return p.SendError(ctx, err)
	}
	project, err := p.proCtl.Get(ctx, projectNameOrID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	if err = p.ctl.Update(ctx, project.ProjectID, metadata); err != nil {
		return p.SendError(ctx, err)
	}
	return operation.NewUpdateProjectMetadataOK()
}

func (p *projectMetadataAPI) validate(metas map[string]string) (map[string]string, error) {
	if len(metas) != 1 {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("only allow one key/value pair")
	}

	key, value := "", ""
	for key, value = range metas {
	}

	switch key {
	case proModels.ProMetaPublic, proModels.ProMetaEnableContentTrust, proModels.ProMetaEnableContentTrustCosign,
		proModels.ProMetaPreventVul, proModels.ProMetaAutoScan:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("invalid value: %s", value)
		}
		metas[key] = strconv.FormatBool(v)
	case proModels.ProMetaSeverity:
		severity := vuln.ParseSeverityVersion3(strings.ToLower(value))
		if severity == vuln.Unknown {
			return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("invalid value: %s", value)
		}
		metas[proModels.ProMetaSeverity] = strings.ToLower(severity.String())
	default:
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("invalid key: %s", key)
	}
	return metas, nil
}
