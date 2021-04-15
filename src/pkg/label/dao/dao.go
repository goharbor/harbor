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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/label/model"
	"time"
)

// DAO is the data access object interface for label
type DAO interface {
	// Get the specified label
	Get(ctx context.Context, id int64) (label *model.Label, err error)
	// Create the label
	Create(ctx context.Context, label *model.Label) (id int64, err error)
	// Count returns the total count of Labels according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// Update the label
	Update(ctx context.Context, label *model.Label) error
	// Delete the label
	Delete(ctx context.Context, id int64) (err error)
	// List ...
	List(ctx context.Context, query *q.Query) ([]*model.Label, error)

	// List labels that added to the artifact specified by the ID
	ListByArtifact(ctx context.Context, artifactID int64) (labels []*model.Label, err error)
	// Create label reference
	CreateReference(ctx context.Context, reference *model.Reference) (id int64, err error)
	// Delete the label reference specified by ID
	DeleteReference(ctx context.Context, id int64) (err error)
	// Delete label references specified by query
	DeleteReferences(ctx context.Context, query *q.Query) (n int64, err error)
}

// New creates an instance of the default DAO
func New() DAO {
	return &defaultDAO{}
}

type defaultDAO struct{}

func (d *defaultDAO) Get(ctx context.Context, id int64) (*model.Label, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	label := &model.Label{
		ID: id,
	}
	if err = ormer.Read(label); err != nil {
		if e := orm.AsNotFoundError(err, "label %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return label, nil
}

func (d *defaultDAO) Create(ctx context.Context, label *model.Label) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(label)
	if err != nil {
		if e := orm.AsConflictError(err, "label %s already exists", label.Name); e != nil {
			err = e
		}
	}
	return id, err
}

func (d *defaultDAO) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.Label{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (d *defaultDAO) Update(ctx context.Context, label *model.Label) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	label.UpdateTime = time.Now()
	n, err := ormer.Update(label)
	if n == 0 {
		if e := orm.AsConflictError(err, "label %s already exists", label.Name); e != nil {
			err = e
		}
		return err
	}
	return err
}

func (d *defaultDAO) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.Label{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("label %d not found", id)
	}
	return nil
}

func (d *defaultDAO) List(ctx context.Context, query *q.Query) ([]*model.Label, error) {
	robots := []*model.Label{}
	qs, err := orm.QuerySetter(ctx, &model.Label{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&robots); err != nil {
		return nil, err
	}
	return robots, nil
}

func (d *defaultDAO) ListByArtifact(ctx context.Context, artifactID int64) ([]*model.Label, error) {
	sql := `select label.* from harbor_label label 
				join label_reference ref on label.id = ref.label_id 
				where ref.artifact_id = ?`
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	labels := []*model.Label{}
	if _, err = ormer.Raw(sql, artifactID).QueryRows(&labels); err != nil {
		return nil, err
	}
	return labels, nil
}
func (d *defaultDAO) CreateReference(ctx context.Context, ref *model.Reference) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(ref)
	if err != nil {
		if e := orm.AsConflictError(err, "label %d is already added to the artifact %d",
			ref.LabelID, ref.ArtifactID); e != nil {
			err = e
		} else if e := orm.AsForeignKeyError(err, "the reference tries to refer a non existing label %d or artifact %d",
			ref.LabelID, ref.ArtifactID); e != nil {
			err = errors.New(e).WithCode(errors.NotFoundCode).WithMessage(e.Message)
		}
	}
	return id, err
}

func (d *defaultDAO) DeleteReference(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.Reference{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("label reference %d not found", id)
	}
	return nil
}

func (d *defaultDAO) DeleteReferences(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetter(ctx, &model.Reference{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Delete()
}
