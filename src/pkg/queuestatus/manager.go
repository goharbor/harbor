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

package queuestatus

import (
	"context"

	"github.com/goharbor/harbor/src/pkg/queuestatus/dao"
	"github.com/goharbor/harbor/src/pkg/queuestatus/model"
)

var (
	// Mgr default user group manager
	Mgr = newManager()
)

// Manager the manager for queue status
type Manager interface {
	// List list queue status
	List(ctx context.Context) ([]*model.JobQueueStatus, error)
	// AllJobTypeStatus get all job type status
	AllJobTypeStatus(ctx context.Context) (map[string]bool, error)
	// Get get queue status by JobType
	Get(ctx context.Context, jobType string) (*model.JobQueueStatus, error)
	// UpdateStatus update queue status
	UpdateStatus(ctx context.Context, jobType string, paused bool) error
	// CreateOrUpdate create a queue status or update it if it already exists
	CreateOrUpdate(ctx context.Context, status *model.JobQueueStatus) (int64, error)
}

type manager struct {
	dao dao.DAO
}

func newManager() Manager {
	return &manager{dao: dao.New()}
}

func (m *manager) List(ctx context.Context) ([]*model.JobQueueStatus, error) {
	return m.dao.Query(ctx, nil)
}

func (m *manager) Get(ctx context.Context, jobType string) (*model.JobQueueStatus, error) {
	return m.dao.GetByJobType(ctx, jobType)
}

func (m *manager) UpdateStatus(ctx context.Context, jobType string, paused bool) error {
	return m.dao.UpdateStatus(ctx, jobType, paused)
}

func (m manager) CreateOrUpdate(ctx context.Context, status *model.JobQueueStatus) (int64, error) {
	return m.dao.InsertOrUpdate(ctx, status)
}

func (m *manager) AllJobTypeStatus(ctx context.Context) (map[string]bool, error) {
	statuses, err := m.List(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool)
	for _, status := range statuses {
		result[status.JobType] = status.Paused
	}
	return result, nil
}
