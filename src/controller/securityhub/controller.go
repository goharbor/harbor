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

package securityhub

import (
	"context"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/scanner"
	"github.com/goharbor/harbor/src/pkg/securityhub"
	secHubModel "github.com/goharbor/harbor/src/pkg/securityhub/model"
	"github.com/goharbor/harbor/src/pkg/tag"
)

// Ctl is the global controller for security hub
var Ctl = NewController()

// Controller controller of security hub
type Controller interface {
	// SecuritySummary returns the security summary of the specified project.
	SecuritySummary(ctx context.Context, projectID int64, withCVE bool, withArtifact bool) (*secHubModel.Summary, error)
	// ListVuls list vulnerabilities by query
	ListVuls(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*secHubModel.VulnerabilityItem, error)
	// CountVuls get all vulnerability count by query
	CountVuls(ctx context.Context, scannerUUID string, projectID int64, tuneCount bool, query *q.Query) (int64, error)
	// DefaultScannerUUID get default scanner UUID
	DefaultScannerUUID(ctx context.Context) (string, error)
}

type controller struct {
	artifactMgr artifact.Manager
	scannerMgr  scanner.Manager
	secHubMgr   securityhub.Manager
	tagMgr      tag.Manager
}

// NewController ...
func NewController() Controller {
	return &controller{
		artifactMgr: artifact.NewManager(),
		scannerMgr:  scanner.New(),
		secHubMgr:   securityhub.Mgr,
		tagMgr:      tag.Mgr,
	}
}

func (c *controller) SecuritySummary(ctx context.Context, projectID int64, withCVE bool, withArtifact bool) (*secHubModel.Summary, error) {
	scannerUUID, err := c.DefaultScannerUUID(ctx)
	if err != nil {
		return nil, err
	}
	sum, err := c.secHubMgr.Summary(ctx, scannerUUID, projectID, nil)
	if err != nil {
		return nil, err
	}
	sum.TotalArtifactCnt, err = c.totalArtifactCount(ctx, projectID)
	if err != nil {
		return nil, err
	}
	sum.ScannedCnt, err = c.secHubMgr.ScannedArtifactsCount(ctx, scannerUUID, projectID, nil)
	if err != nil {
		return nil, err
	}
	if withCVE {
		sum.DangerousCVEs, err = c.secHubMgr.DangerousCVEs(ctx, scannerUUID, projectID, nil)
		if err != nil {
			return nil, err
		}
	}
	if withArtifact {
		sum.DangerousArtifacts, err = c.secHubMgr.DangerousArtifacts(ctx, scannerUUID, projectID, nil)
		if err != nil {
			return nil, err
		}
	}
	return sum, nil
}

func (c *controller) totalArtifactCount(ctx context.Context, projectID int64) (int64, error) {
	if projectID == 0 {
		return c.artifactMgr.Count(ctx, nil)
	}
	return c.artifactMgr.Count(ctx, q.New(q.KeyWords{"project_id": projectID}))
}

// DefaultScannerUUID returns the default scanner uuid.
func (c *controller) DefaultScannerUUID(ctx context.Context) (string, error) {
	reg, err := c.scannerMgr.GetDefault(ctx)
	if err != nil {
		return "", err
	}
	return reg.UUID, nil
}

func (c *controller) ListVuls(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*secHubModel.VulnerabilityItem, error) {
	vuls, err := c.secHubMgr.ListVuls(ctx, scannerUUID, projectID, query)
	if err != nil {
		return nil, err
	}
	resultList, err := c.attachTags(ctx, vuls)
	if err != nil {
		return nil, err
	}
	return resultList, nil
}

func (c *controller) attachTags(ctx context.Context, vuls []*secHubModel.VulnerabilityItem) ([]*secHubModel.VulnerabilityItem, error) {
	var artifactIds []interface{}
	artifactTagMap := make(map[int64][]string, 0)
	for _, v := range vuls {
		artifactTagMap[v.ArtifactID] = make([]string, 0)
	}
	for k := range artifactTagMap {
		artifactIds = append(artifactIds, k)
	}
	query := q.New(q.KeyWords{"artifact_id": q.NewOrList(artifactIds)})
	tags, err := c.tagMgr.List(ctx, query)
	if err != nil {
		return vuls, err
	}
	for _, tag := range tags {
		artifactTagMap[tag.ArtifactID] = append(artifactTagMap[tag.ArtifactID], tag.Name)
	}
	for _, v := range vuls {
		if len(artifactTagMap[v.ArtifactID]) > 10 {
			v.Tags = artifactTagMap[v.ArtifactID][:10]
			continue
		}
		v.Tags = artifactTagMap[v.ArtifactID]
	}
	return vuls, nil
}

func (c *controller) CountVuls(ctx context.Context, scannerUUID string, projectID int64, tuneCount bool, query *q.Query) (int64, error) {
	return c.secHubMgr.TotalVuls(ctx, scannerUUID, projectID, tuneCount, query)
}
