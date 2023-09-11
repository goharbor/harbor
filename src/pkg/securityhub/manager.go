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
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/securityhub/dao"
	"github.com/goharbor/harbor/src/pkg/securityhub/model"
)

var (
	// Mgr is the global security manager
	Mgr = NewManager()
)

// Manager is used to manage the security manager.
type Manager interface {
	// Summary returns the summary of the scan cve reports.
	Summary(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (*model.Summary, error)
	// DangerousArtifacts returns the most dangerous artifact for the given scanner.
	DangerousArtifacts(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*model.DangerousArtifact, error)
	// TotalArtifactsCount return the count of artifacts.
	TotalArtifactsCount(ctx context.Context, projectID int64) (int64, error)
	// ScannedArtifactsCount return the count of scanned artifacts.
	ScannedArtifactsCount(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (int64, error)
	// DangerousCVEs returns the most dangerous CVEs for the given scanner.
	DangerousCVEs(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*scan.VulnerabilityRecord, error)
	// TotalVuls return the count of vulnerabilities
	TotalVuls(ctx context.Context, scannerUUID string, projectID int64, tuneCount bool, query *q.Query) (int64, error)
	// ListVuls returns vulnerabilities list
	ListVuls(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*model.VulnerabilityItem, error)
}

// NewManager news security manager.
func NewManager() Manager {
	return &securityManager{
		dao: dao.New(),
	}
}

// securityManager is a default implementation of security manager.
type securityManager struct {
	dao dao.SecurityHubDao
}

func (s *securityManager) TotalArtifactsCount(ctx context.Context, projectID int64) (int64, error) {
	return s.dao.TotalArtifactsCount(ctx, projectID)
}

func (s *securityManager) Summary(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (*model.Summary, error) {
	return s.dao.Summary(ctx, scannerUUID, projectID, query)
}

func (s *securityManager) DangerousArtifacts(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*model.DangerousArtifact, error) {
	return s.dao.DangerousArtifacts(ctx, scannerUUID, projectID, query)
}

func (s *securityManager) ScannedArtifactsCount(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (int64, error) {
	return s.dao.ScannedArtifactsCount(ctx, scannerUUID, projectID, query)
}

func (s *securityManager) DangerousCVEs(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*scan.VulnerabilityRecord, error) {
	return s.dao.DangerousCVEs(ctx, scannerUUID, projectID, query)
}

func (s *securityManager) TotalVuls(ctx context.Context, scannerUUID string, projectID int64, tuneCount bool, query *q.Query) (int64, error) {
	return s.dao.CountVulnerabilities(ctx, scannerUUID, projectID, tuneCount, query)
}

func (s *securityManager) ListVuls(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*model.VulnerabilityItem, error) {
	return s.dao.ListVulnerabilities(ctx, scannerUUID, projectID, query)
}
