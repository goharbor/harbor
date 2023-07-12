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
	"github.com/goharbor/harbor/src/server/v2.0/models"
	securityModel "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/securityhub"

	"github.com/goharbor/harbor/src/controller/securityhub"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	secHubModel "github.com/goharbor/harbor/src/pkg/securityhub/model"
)

func newSecurityAPI() *securityAPI {
	return &securityAPI{
		controller: securityhub.Ctl,
	}
}

type securityAPI struct {
	BaseAPI
	controller securityhub.Controller
}

func (s *securityAPI) GetSecuritySummary(ctx context.Context,
	params securityModel.GetSecuritySummaryParams) middleware.Responder {
	if err := s.RequireSystemAccess(ctx, rbac.ActionRead, rbac.ResourceSecurityHub); err != nil {
		return s.SendError(ctx, err)
	}
	summary, err := s.controller.SecuritySummary(ctx, 0, securityhub.WithCVE(*params.WithDangerousCVE), securityhub.WithArtifact(*params.WithDangerousArtifact))
	if err != nil {
		return s.SendError(ctx, err)
	}
	sum := toSecuritySummaryModel(summary)
	return securityModel.NewGetSecuritySummaryOK().WithPayload(sum)
}

func toSecuritySummaryModel(summary *secHubModel.Summary) *models.SecuritySummary {
	return &models.SecuritySummary{
		CriticalCnt:        summary.CriticalCnt,
		HighCnt:            summary.HighCnt,
		MediumCnt:          summary.MediumCnt,
		LowCnt:             summary.LowCnt,
		NoneCnt:            summary.NoneCnt,
		UnknownCnt:         summary.UnknownCnt,
		FixableCnt:         summary.FixableCnt,
		TotalVuls:          summary.CriticalCnt + summary.HighCnt + summary.MediumCnt + summary.LowCnt + summary.NoneCnt + summary.UnknownCnt,
		TotalArtifact:      summary.TotalArtifactCnt,
		ScannedCnt:         summary.ScannedCnt,
		DangerousCves:      toDangerousCves(summary.DangerousCVEs),
		DangerousArtifacts: toDangerousArtifacts(summary.DangerousArtifacts),
	}
}
func toDangerousArtifacts(artifacts []*secHubModel.DangerousArtifact) []*models.DangerousArtifact {
	var result []*models.DangerousArtifact
	for _, artifact := range artifacts {
		result = append(result, &models.DangerousArtifact{
			ProjectID:      artifact.Project,
			RepositoryName: artifact.Repository,
			Digest:         artifact.Digest,
			CriticalCnt:    artifact.CriticalCnt,
			HighCnt:        artifact.HighCnt,
			MediumCnt:      artifact.MediumCnt,
		})
	}
	return result
}

func toDangerousCves(cves []*scan.VulnerabilityRecord) []*models.DangerousCVE {
	var result []*models.DangerousCVE
	for _, vul := range cves {
		result = append(result, &models.DangerousCVE{
			CVEID:       vul.CVEID,
			Package:     vul.Package,
			Version:     vul.PackageVersion,
			Severity:    vul.Severity,
			CvssScoreV3: *vul.CVE3Score,
		})
	}
	return result
}
