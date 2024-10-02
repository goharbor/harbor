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

package joblog

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/pkg/joblog/dao"
	"github.com/goharbor/harbor/src/pkg/joblog/models"
)

// Mgr is the global job log manager instance
var Mgr = New()

// Manager is used for job log management
type Manager interface {
	// Get the job log specified by ID
	Get(ctx context.Context, uuid string) (jobLog *models.JobLog, err error)
	// Create the job log
	Create(ctx context.Context, jobLog *models.JobLog) (id int64, err error)
	// DeleteBefore the job log specified by time
	DeleteBefore(ctx context.Context, t time.Time) (id int64, err error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{
		dao: dao.New(),
	}
}

type manager struct {
	dao dao.DAO
}

// Get ...
func (m *manager) Get(ctx context.Context, uuid string) (jobLog *models.JobLog, err error) {
	return m.dao.Get(ctx, uuid)
}

// Create ...
func (m *manager) Create(ctx context.Context, jobLog *models.JobLog) (id int64, err error) {
	return m.dao.Create(ctx, jobLog)
}

// DeleteBefore ...
func (m *manager) DeleteBefore(ctx context.Context, t time.Time) (id int64, err error) {
	return m.dao.DeleteBefore(ctx, t)
}
