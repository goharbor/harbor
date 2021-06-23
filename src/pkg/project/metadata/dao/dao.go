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

package dao

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project/metadata/models"
)

// DAO is the data access object interface for project metadata
type DAO interface {
	// Create create metadata instances for the project
	Create(ctx context.Context, projectID int64, name, value string) (int64, error)

	// Delete delete metadata interfaces filtered the query
	Delete(ctx context.Context, query *q.Query) error

	// Update update the value of metadata instance
	Update(ctx context.Context, projectID int64, name, value string) error

	// List returns project metadata instances
	List(ctx context.Context, query *q.Query) ([]*models.ProjectMetadata, error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

// Create create metadata instances for the project
func (d *dao) Create(ctx context.Context, projectID int64, name, value string) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	md := &models.ProjectMetadata{
		ProjectID:    projectID,
		Name:         name,
		Value:        value,
		CreationTime: now,
		UpdateTime:   now,
	}

	id, err := o.Insert(md)
	if err != nil {
		if e := orm.AsConflictError(err, "metadata %s already exists for project %d", name, projectID); e != nil {
			err = e
		}
		return 0, err
	}
	return id, nil
}

// Delete delete metadata interfaces filtered the query
func (d *dao) Delete(ctx context.Context, query *q.Query) error {
	qs, err := orm.QuerySetter(ctx, &models.ProjectMetadata{}, query)
	if err != nil {
		return err
	}

	_, err = qs.Delete()
	return err
}

// Update update the metadata instance
func (d *dao) Update(ctx context.Context, projectID int64, name, value string) error {
	qs, err := orm.QuerySetter(ctx, &models.ProjectMetadata{}, nil)
	if err != nil {
		return err
	}

	qs = qs.Filter("project_id", projectID).Filter("name", name)

	_, err = qs.Update(orm.Params{"value": value})
	return err
}

// List returns project metadata instances
func (d *dao) List(ctx context.Context, query *q.Query) ([]*models.ProjectMetadata, error) {
	qs, err := orm.QuerySetter(ctx, &models.ProjectMetadata{}, query)
	if err != nil {
		return nil, err
	}

	mds := []*models.ProjectMetadata{}
	if _, err := qs.All(&mds); err != nil {
		return nil, err
	}

	return mds, nil
}
