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

package repository

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
)

var (
	// Ctl is a global repository controller instance
	Ctl = NewController(repository.Mgr)
)

// Controller defines the operations related with repositories
type Controller interface {
	// Ensure the repository specified by the "name" exists under the project,
	// creates it if it doesn't exist. The "name" should contain the namespace part.
	// The "created" will be set as true when the repository is created
	Ensure(ctx context.Context, projectID int64, name string) (created bool, id int64, err error)
	// Get the repository specified by ID
	Get(ctx context.Context, id int64) (repository *models.RepoRecord, err error)
}

// NewController creates an instance of the default repository controller
func NewController(repoMgr repository.Manager) Controller {
	return &controller{
		repoMgr: repoMgr,
	}
}

type controller struct {
	repoMgr repository.Manager
}

func (c *controller) Ensure(ctx context.Context, projectID int64, name string) (bool, int64, error) {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"name": name,
		},
	}
	_, repositories, err := c.repoMgr.List(ctx, query)
	if err != nil {
		return false, 0, err
	}
	// the repository already exists, return directly
	if len(repositories) > 0 {
		return false, repositories[0].RepositoryID, nil
	}

	// the repository doesn't exist, create it first
	id, err := c.repoMgr.Create(ctx, &models.RepoRecord{
		ProjectID: projectID,
		Name:      name,
	})
	if err != nil {
		// if got conflict error, try to get again
		if ierror.IsConflictErr(err) {
			_, repositories, err = c.repoMgr.List(ctx, query)
			if err != nil {
				return false, 0, err
			}
			if len(repositories) > 0 {
				return false, repositories[0].RepositoryID, nil
			}
		}
		return false, 0, err
	}
	return true, id, nil
}

func (c *controller) Get(ctx context.Context, id int64) (*models.RepoRecord, error) {
	return c.repoMgr.Get(ctx, id)
}
