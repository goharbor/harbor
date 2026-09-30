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

package pushcount

import (
	"context"

	"github.com/goharbor/harbor/src/lib/orm"
)

// Mgr is the global push count manager
var Mgr = New()

// Manager records how many artifacts have been pushed into each project.
// The count is kept independently of the artifact table so that it stays
// monotonic when artifacts or repositories are deleted.
type Manager interface {
	// Add increases the push count of the specified project
	Add(ctx context.Context, projectID int64, count uint64) error
	// Get returns the push count of the specified project, 0 if nothing was pushed yet
	Get(ctx context.Context, projectID int64) (int64, error)
}

// New returns a default implementation of Manager
func New() Manager {
	return &manager{}
}

type manager struct{}

func (m *manager) Add(ctx context.Context, projectID int64, count uint64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	sql := `INSERT INTO project_push_count (project_id, push_count, update_time) VALUES (?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT (project_id) DO UPDATE SET push_count = project_push_count.push_count + EXCLUDED.push_count, update_time = EXCLUDED.update_time`
	_, err = ormer.Raw(sql, projectID, count).Exec()
	return err
}

func (m *manager) Get(ctx context.Context, projectID int64) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	var count int64
	err = ormer.Raw(`SELECT COALESCE(SUM(push_count), 0) FROM project_push_count WHERE project_id = ?`, projectID).QueryRow(&count)
	return count, err
}
