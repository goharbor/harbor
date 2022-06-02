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

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	art "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/repository/model"
)

var (
	// Ctl is a global repository controller instance
	Ctl = NewController()
)

// Controller defines the operations related with repositories
type Controller interface {
	// Ensure the repository specified by the "name" exists, creates it if it doesn't exist.
	// The "name" should contain the namespace part. The "created" will be set as true
	// when the repository is created
	Ensure(ctx context.Context, name string) (created bool, id int64, err error)
	// Count returns the total count of repositories according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List repositories according to the query
	List(ctx context.Context, query *q.Query) (repositories []*model.RepoRecord, err error)
	// Get the repository specified by ID
	Get(ctx context.Context, id int64) (repository *model.RepoRecord, err error)
	// GetByName gets the repository specified by name
	GetByName(ctx context.Context, name string) (repository *model.RepoRecord, err error)
	// Delete the repository specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Update the repository. Specify the properties or all properties will be updated
	Update(ctx context.Context, repository *model.RepoRecord, properties ...string) (err error)
	// AddPullCount increase pull count for the specified repository
	AddPullCount(ctx context.Context, id int64, count uint64) error
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		proMgr:  pkg.ProjectMgr,
		repoMgr: pkg.RepositoryMgr,
		artMgr:  pkg.ArtifactMgr,
		artCtl:  artifact.Ctl,
	}
}

type controller struct {
	proMgr  project.Manager
	repoMgr repository.Manager
	artMgr  art.Manager
	artCtl  artifact.Controller
}

func (c *controller) Ensure(ctx context.Context, name string) (bool, int64, error) {
	// the repository already exists, return directly
	repository, err := c.repoMgr.GetByName(ctx, name)
	if err == nil {
		return false, repository.RepositoryID, nil
	}

	// got other error when get the repository, return the error
	if !errors.IsErr(err, errors.NotFoundCode) {
		return false, 0, err
	}

	// the repository doesn't exist, create it first
	projectName, _ := utils.ParseRepository(name)
	project, err := c.proMgr.Get(ctx, projectName)
	if err != nil {
		return false, 0, err
	}

	var (
		created bool
		id      int64
	)
	// use orm.WithTransaction here to avoid the issue:
	// https://www.postgresql.org/message-id/002e01c04da9%24a8f95c20%2425efe6c1%40lasting.ro
	if err = orm.WithTransaction(func(ctx context.Context) error {
		id, err = c.repoMgr.Create(ctx, &model.RepoRecord{
			ProjectID: project.ProjectID,
			Name:      name,
		})
		if err != nil {
			return err
		}
		created = true
		return nil
	})(orm.SetTransactionOpNameToContext(ctx, "tx-repository-ensure")); err != nil {
		// isn't conflict error, return directly
		if !errors.IsConflictErr(err) {
			return false, 0, err
		}
		// if got conflict error, try to get again
		repository, err = c.repoMgr.GetByName(ctx, name)
		if err != nil {
			return false, 0, err
		}
		id = repository.RepositoryID
	}

	return created, id, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.repoMgr.Count(ctx, query)
}

func (c *controller) List(ctx context.Context, query *q.Query) ([]*model.RepoRecord, error) {
	return c.repoMgr.List(ctx, query)
}

func (c *controller) Get(ctx context.Context, id int64) (*model.RepoRecord, error) {
	return c.repoMgr.Get(ctx, id)
}

func (c *controller) GetByName(ctx context.Context, name string) (*model.RepoRecord, error) {
	return c.repoMgr.GetByName(ctx, name)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	candidates, err := c.artCtl.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": id,
		},
	}, nil)
	if err != nil {
		return err
	}
	for len(candidates) > 0 {
		artifacts := candidates
		candidates = nil
		for _, artifact := range artifacts {
			parents, err := c.artMgr.ListReferences(ctx, &q.Query{
				Keywords: map[string]interface{}{
					"ChildID": artifact.ID,
				},
			})
			if err != nil {
				return err
			}
			if len(parents) > 0 {
				// if the artifact is referenced by others, put it into the artifacts
				// list again, and try to delete it in the next loop
				log.Debugf("the artifact %d is referenced by others, put it into the backlog and try to delete it in the next loop",
					artifact.ID)
				candidates = append(candidates, artifact)
				continue
			}
			if err = c.artCtl.Delete(ctx, artifact.ID); err != nil {
				return err
			}
			log.Debugf("the artifact %d is deleted", artifact.ID)
		}
	}
	return c.repoMgr.Delete(ctx, id)
}

func (c *controller) Update(ctx context.Context, repository *model.RepoRecord, properties ...string) error {
	return c.repoMgr.Update(ctx, repository, properties...)
}

func (c *controller) AddPullCount(ctx context.Context, id int64, count uint64) error {
	return c.repoMgr.AddPullCount(ctx, id, count)
}
