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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/distribution"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
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

func (s *scanAPI) Prepare(ctx context.Context, _ string, params any) middleware.Responder {
	if err := unescapePathParams(params, "RepositoryName"); err != nil {
		s.SendError(ctx, err)
	}

	return nil
}

func (s *scanAPI) StopScanArtifact(ctx context.Context, params operation.StopScanArtifactParams) middleware.Responder {
	scanType := v1.ScanTypeVulnerability
	if params.ScanType != nil && validScanType(params.ScanType.ScanType) {
		scanType = params.ScanType.ScanType
	}
	res := rbac.ResourceScan
	if scanType == v1.ScanTypeSbom {
		res = rbac.ResourceSBOM
	}
	if err := s.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionStop, res); err != nil {
		return s.SendError(ctx, err)
	}

	// get the artifact
	curArtifact, err := s.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if err := s.scanCtl.Stop(ctx, curArtifact, params.ScanType.ScanType); err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewStopScanArtifactAccepted()
}

func (s *scanAPI) ScanArtifact(ctx context.Context, params operation.ScanArtifactParams) middleware.Responder {
	scanType := v1.ScanTypeVulnerability
	options := []scan.Option{}
	if !distribution.IsDigest(params.Reference) {
		options = append(options, scan.WithTag(params.Reference))
	}
	if params.ScanType != nil && validScanType(params.ScanType.ScanType) {
		scanType = params.ScanType.ScanType
		options = append(options, scan.WithScanType(scanType))
	}
	res := rbac.ResourceScan
	if scanType == v1.ScanTypeSbom {
		res = rbac.ResourceSBOM
	}
	if err := s.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, res); err != nil {
		return s.SendError(ctx, err)
	}
	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	artifact, err := s.artCtl.GetByReference(ctx, repository, params.Reference, nil)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if err := s.scanCtl.Scan(ctx, artifact, options...); err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewScanArtifactAccepted()
}

func (s *scanAPI) GetReportLog(ctx context.Context, params operation.GetReportLogParams) middleware.Responder {
	if err := s.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceScan); err != nil {
		return s.SendError(ctx, err)
	}
	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	a, err := s.artCtl.GetByReference(ctx, repository, params.Reference, nil)
	if err != nil {
		return s.SendError(ctx, err)
	}

	bytes, err := s.scanCtl.GetScanLog(ctx, a, params.ReportID)
	if err != nil {
		return s.SendError(ctx, err)
	}

	if bytes == nil {
		// Not found
		return s.SendError(ctx, errors.NotFoundError(nil).WithMessagef("report with uuid %s does not exist", params.ReportID))
	}

	return operation.NewGetReportLogOK().WithPayload(string(bytes))
}

func validScanType(scanType string) bool {
	return scanType == "sbom" || scanType == "vulnerability"
}
