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
	"strconv"
	"strings"

	"github.com/go-openapi/runtime/middleware"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/project/metadata"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/pattern"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project_metadata"
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
	project, err := p.proCtl.Get(ctx, projectNameOrID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	metadata := params.Metadata
	metadata, err = p.validate(ctx, project.ProjectID, metadata)
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
	if params.MetaName == proModels.ProMetaProxyCacheFilterKind {
		existing, err := p.ctl.Get(ctx, project.ProjectID, proModels.ProMetaProxyCacheFilterPattern)
		if err != nil {
			return p.SendError(ctx, err)
		}
		if patVal, ok := existing[proModels.ProMetaProxyCacheFilterPattern]; ok && patVal != "" {
			return p.SendError(ctx, errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessagef("cannot delete %s when %s is not empty, please clear the pattern first",
					proModels.ProMetaProxyCacheFilterKind, proModels.ProMetaProxyCacheFilterPattern))
		}
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
	project, err := p.proCtl.Get(ctx, projectNameOrID)
	if err != nil {
		return p.SendError(ctx, err)
	}
	metadata := map[string]string{
		params.MetaName: params.Metadata[params.MetaName],
	}
	metadata, err = p.validate(ctx, project.ProjectID, metadata)
	if err != nil {
		return p.SendError(ctx, err)
	}
	if err = p.ctl.Update(ctx, project.ProjectID, metadata); err != nil {
		return p.SendError(ctx, err)
	}
	return operation.NewUpdateProjectMetadataOK()
}

func (p *projectMetadataAPI) validate(ctx context.Context, projectID int64, metas map[string]string) (map[string]string, error) {
	if len(metas) != 1 {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("only allow one key/value pair")
	}

	key, value := "", ""
	for key, value = range metas { // nolint:revive
	}

	switch key {
	case proModels.ProMetaPublic, proModels.ProMetaEnableContentTrust, proModels.ProMetaEnableContentTrustCosign,
		proModels.ProMetaAutoSBOMGen, proModels.ProMetaPreventVul, proModels.ProMetaAutoScan, proModels.ProMetaReuseSysCVEAllowlist,
		proModels.ProMetaProxyCacheLocalOnNotFound, proModels.ProMetaProxyReferrerAPI:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("invalid value: %s", value)
		}
		metas[key] = strconv.FormatBool(v)
	case proModels.ProMetaSeverity:
		severity := vuln.ParseSeverityVersion3(strings.ToLower(value))
		if severity == vuln.Unknown {
			return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("invalid value: %s", value)
		}
		metas[proModels.ProMetaSeverity] = strings.ToLower(severity.String())
	case proModels.ProMetaProxySpeed:
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("invalid value: %s", value)
		}
		metas[proModels.ProMetaProxySpeed] = strconv.FormatInt(v, 10)
	case proModels.ProMetaMaxUpstreamConn:
		v, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return nil, errors.New(nil).
				WithCode(errors.BadRequestCode).
				WithMessagef("invalid value: %s", value)
		}

		if v <= 0 && v != -1 {
			return nil, errors.New(nil).
				WithCode(errors.BadRequestCode).
				WithMessagef("invalid value for %s: %d", key, v)
		}
		metas[proModels.ProMetaMaxUpstreamConn] = strconv.FormatInt(v, 10)
	case proModels.ProMetaProxyCacheFilterPattern:
		if err := p.requireProxyProject(ctx, projectID); err != nil {
			return nil, err
		}
		// The API accepts one key per request, so the effective kind is
		// whatever is already stored; empty means the doublestar default.
		var kind string
		existing, err := p.ctl.Get(ctx, projectID, proModels.ProMetaProxyCacheFilterKind)
		if err != nil {
			return nil, err
		}
		if k, ok := existing[proModels.ProMetaProxyCacheFilterKind]; ok {
			kind = k
		}
		if err := pattern.ValidateRepositoryFilter(value, kind); err != nil {
			return nil, errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessagef("invalid proxy_cache_filter_pattern: %q, err: %v", value, err)
		}
		metas[proModels.ProMetaProxyCacheFilterPattern] = value
	case proModels.ProMetaProxyCacheFilterKind:
		if err := p.requireProxyProject(ctx, projectID); err != nil {
			return nil, err
		}
		if err := pattern.ValidateKind(value); err != nil {
			return nil, errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessagef("invalid proxy_cache_filter_kind: %v", err)
		}
		// The new kind must keep the stored pattern (the only possible source,
		// since the API accepts one key per request) compilable.
		existing, err := p.ctl.Get(ctx, projectID, proModels.ProMetaProxyCacheFilterPattern)
		if err != nil {
			return nil, err
		}
		if pat, ok := existing[proModels.ProMetaProxyCacheFilterPattern]; ok && pat != "" {
			if err := pattern.ValidateRepositoryFilter(pat, value); err != nil {
				return nil, errors.New(nil).WithCode(errors.BadRequestCode).
					WithMessagef("existing proxy_cache_filter_pattern %q is invalid for new kind %q: %v", pat, value, err)
			}
		}
		metas[proModels.ProMetaProxyCacheFilterKind] = value
	default:
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("invalid key: %s", key)
	}
	return metas, nil
}

// requireProxyProject rejects proxy-cache-filter metadata changes on projects
// that are not proxy cache projects.
func (p *projectMetadataAPI) requireProxyProject(ctx context.Context, projectID int64) error {
	pro, err := p.proCtl.Get(ctx, projectID)
	if err != nil {
		return err
	}
	if !pro.IsProxy() {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("can not update the normal project with proxy cache filter metadata")
	}
	return nil
}
