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
	"github.com/goharbor/harbor/src/pkg/scan/scanner"
	"github.com/goharbor/harbor/src/pkg/securityhub"
	secHubModel "github.com/goharbor/harbor/src/pkg/securityhub/model"
	"github.com/goharbor/harbor/src/pkg/tag"
)

// Ctl is the global controller for security hub
var Ctl = NewController()

// Options define the option to query summary info
type Options struct {
	WithCVE      bool
	WithArtifact bool
}

// Option define the func to build options
type Option func(*Options)

func newOptions(options ...Option) *Options {
	opts := &Options{}
	for _, f := range options {
		f(opts)
	}
	return opts
}

// WithCVE enable CVE info in summary
func WithCVE(enable bool) Option {
	return func(o *Options) {
		o.WithCVE = enable
	}
}

// WithArtifact enable artifact info in summary
func WithArtifact(enable bool) Option {
	return func(o *Options) {
		o.WithArtifact = enable
	}
}

// Controller controller of security hub
type Controller interface {
	// SecuritySummary returns the security summary of the specified project.
	SecuritySummary(ctx context.Context, projectID int64, options ...Option) (*secHubModel.Summary, error)
	// ListVuls list vulnerabilities by query
	ListVuls(ctx context.Context, scannerUUID string, projectID int64, withTag bool, query *q.Query) ([]*secHubModel.VulnerabilityItem, error)
	// CountVuls get all vulnerability count by query
	CountVuls(ctx context.Context, scannerUUID string, projectID int64, tuneCount bool, query *q.Query) (int64, error)
}

type controller struct {
	scannerMgr scanner.Manager
	secHubMgr  securityhub.Manager
	tagMgr     tag.Manager
}

// NewController ...
func NewController() Controller {
	return &controller{
		scannerMgr: scanner.Mgr,
		secHubMgr:  securityhub.Mgr,
		tagMgr:     tag.Mgr,
	}
}

func (c *controller) SecuritySummary(ctx context.Context, projectID int64, options ...Option) (*secHubModel.Summary, error) {
	opts := newOptions(options...)
	scannerUUID, err := c.scannerMgr.DefaultScannerUUID(ctx)
	if len(scannerUUID) == 0 || err != nil {
		return &secHubModel.Summary{}, nil
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
	if opts.WithCVE {
		sum.DangerousCVEs, err = c.secHubMgr.DangerousCVEs(ctx, scannerUUID, projectID, nil)
		if err != nil {
			return nil, err
		}
	}
	if opts.WithArtifact {
		sum.DangerousArtifacts, err = c.secHubMgr.DangerousArtifacts(ctx, scannerUUID, projectID, nil)
		if err != nil {
			return nil, err
		}
	}
	return sum, nil
}

func (c *controller) scannedArtifactCount(ctx context.Context, projectID int64) (int64, error) {
	scannerUUID, err := c.scannerMgr.DefaultScannerUUID(ctx)
	if err != nil {
		return 0, err
	}
	return c.secHubMgr.ScannedArtifactsCount(ctx, scannerUUID, projectID, nil)
}

func (c *controller) totalArtifactCount(ctx context.Context, projectID int64) (int64, error) {
	return c.secHubMgr.TotalArtifactsCount(ctx, projectID)
}

func (c *controller) ListVuls(ctx context.Context, scannerUUID string, projectID int64, withTag bool, query *q.Query) ([]*secHubModel.VulnerabilityItem, error) {
	vuls, err := c.secHubMgr.ListVuls(ctx, scannerUUID, projectID, query)
	if err != nil {
		return nil, err
	}
	if withTag {
		return c.attachTags(ctx, vuls)
	}
	return vuls, nil
}

func (c *controller) attachTags(ctx context.Context, vuls []*secHubModel.VulnerabilityItem) ([]*secHubModel.VulnerabilityItem, error) {
	// get all artifact_ids
	artifactTagMap := make(map[int64][]string, 0)
	for _, v := range vuls {
		artifactTagMap[v.ArtifactID] = make([]string, 0)
	}

	// get tags in the artifact list
	var artifactIDs []interface{}
	for k := range artifactTagMap {
		artifactIDs = append(artifactIDs, k)
	}
	query := q.New(q.KeyWords{"artifact_id": q.NewOrList(artifactIDs)})
	tags, err := c.tagMgr.List(ctx, query)
	if err != nil {
		return vuls, err
	}
	for _, tag := range tags {
		artifactTagMap[tag.ArtifactID] = append(artifactTagMap[tag.ArtifactID], tag.Name)
	}

	// attach tags, only show 10 tags
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
