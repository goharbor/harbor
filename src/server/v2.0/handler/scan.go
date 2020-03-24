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

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/scan"
	ierror "github.com/goharbor/harbor/src/lib/error"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/scan"
)

func newScanAPI() *scanAPI {
	return &scanAPI{
		artCtl:  artifact.Ctl,
		scanCtl: scan.DefaultController,
	}
}

type scanAPI struct {
	BaseAPI
	artCtl  artifact.Controller
	scanCtl scan.Controller
}

func (s *scanAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	if err := unescapePathParams(params, "RepositoryName"); err != nil {
		s.SendError(ctx, err)
	}

	return nil
}

func (s *scanAPI) ScanArtifact(ctx context.Context, params operation.ScanArtifactParams) middleware.Responder {
	if err := s.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceScan); err != nil {
		return s.SendError(ctx, err)
	}

	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	artifact, err := s.artCtl.GetByReference(ctx, repository, params.Reference, nil)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if err := s.scanCtl.Scan(ctx, artifact); err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewScanArtifactAccepted()
}

func (s *scanAPI) GetReportLog(ctx context.Context, params operation.GetReportLogParams) middleware.Responder {
	if err := s.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceScan); err != nil {
		return s.SendError(ctx, err)
	}

	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	_, err := s.artCtl.GetByReference(ctx, repository, params.Reference, nil)
	if err != nil {
		return s.SendError(ctx, err)
	}

	bytes, err := s.scanCtl.GetScanLog(params.ReportID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if bytes == nil {
		// Not found
		return s.SendError(ctx, ierror.NotFoundError(nil).WithMessage("report with uuid %s does not exist", params.ReportID))
	}

	return operation.NewGetReportLogOK().WithPayload(string(bytes))
}
