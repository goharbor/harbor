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

	beego_orm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
)

func init() {
	beego_orm.RegisterModel(&tag.Tag{})
}

// DAO is the data access object for tag
type DAO interface {
	// Count returns the total count of tags according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List tags according to the query
	List(ctx context.Context, query *q.Query) (tags []*tag.Tag, err error)
	// Get the tag specified by ID
	Get(ctx context.Context, id int64) (tag *tag.Tag, err error)
	// Create the tag
	Create(ctx context.Context, tag *tag.Tag) (id int64, err error)
	// Update the tag. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, tag *tag.Tag, props ...string) (err error)
	// Delete the tag specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// DeleteOfArtifact deletes all tags attached to the artifact
	DeleteOfArtifact(ctx context.Context, artifactID int64) (err error)
}

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &tag.Tag{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}
func (d *dao) List(ctx context.Context, query *q.Query) ([]*tag.Tag, error) {
	tags := []*tag.Tag{}
	qs, err := orm.QuerySetter(ctx, &tag.Tag{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&tags); err != nil {
		return nil, err
	}
	return tags, nil
}
func (d *dao) Get(ctx context.Context, id int64) (*tag.Tag, error) {
	tag := &tag.Tag{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(tag); err != nil {
		if e := orm.AsNotFoundError(err, "tag %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return tag, nil
}
func (d *dao) Create(ctx context.Context, tag *tag.Tag) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(tag)
	if err != nil {
		if e := orm.AsConflictError(err, "tag %s already exists under the repository %d",
			tag.Name, tag.RepositoryID); e != nil {
			err = e
		} else if e := orm.AsForeignKeyError(err, "the tag %s tries to attach to a non existing artifact %d",
			tag.Name, tag.ArtifactID); e != nil {
			err = e
		}
	}
	return id, err
}
func (d *dao) Update(ctx context.Context, tag *tag.Tag, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(tag, props...)
	if err != nil {
		if e := orm.AsForeignKeyError(err, "the tag %d tries to attach to a non existing artifact %d",
			tag.ID, tag.ArtifactID); e != nil {
			err = e
		}
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("tag %d not found", tag.ID)
	}
	return nil
}
func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&tag.Tag{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessagef("tag %d not found", id)
	}
	return nil
}

func (d *dao) DeleteOfArtifact(ctx context.Context, artifactID int64) error {
	qs, err := orm.QuerySetter(ctx, &tag.Tag{}, &q.Query{
		Keywords: map[string]any{
			"ArtifactID": artifactID,
		},
	})
	if err != nil {
		return err
	}
	_, err = qs.Delete()
	return err
}
