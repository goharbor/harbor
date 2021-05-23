package dao

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

// DAO defines the interface to access the robot data model
type DAO interface {
	// Create ...
	Create(ctx context.Context, r *model.Robot) (int64, error)

	// Update ...
	Update(ctx context.Context, r *model.Robot, props ...string) error

	// Get ...
	Get(ctx context.Context, id int64) (*model.Robot, error)

	// Count returns the total count of robots according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// List ...
	List(ctx context.Context, query *q.Query) ([]*model.Robot, error)

	// Delete ...
	Delete(ctx context.Context, id int64) error

	// DeleteByProjectID ...
	DeleteByProjectID(ctx context.Context, projectID int64) error
}

// New creates a default implementation for Dao
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Create(ctx context.Context, r *model.Robot) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	r.CreationTime = time.Now()
	id, err := ormer.Insert(r)
	if err != nil {
		return 0, orm.WrapConflictError(err, "robot account %d:%s already exists", r.ProjectID, r.Name)
	}
	return id, err
}

func (d *dao) Update(ctx context.Context, r *model.Robot, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(r, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("robot %d not found", r.ID)
	}
	return nil
}

func (d *dao) Get(ctx context.Context, id int64) (*model.Robot, error) {
	r := &model.Robot{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := ormer.Read(r); err != nil {
		return nil, orm.WrapNotFoundError(err, "robot %d not found", id)
	}
	return r, nil
}

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	qs, err := orm.QuerySetterForCount(ctx, &model.Robot{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}

func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.Robot{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("robot account %d not found", id)
	}
	return nil
}

func (d *dao) List(ctx context.Context, query *q.Query) ([]*model.Robot, error) {
	robots := []*model.Robot{}

	qs, err := orm.QuerySetter(ctx, &model.Robot{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&robots); err != nil {
		return nil, err
	}
	return robots, nil
}

func (d *dao) DeleteByProjectID(ctx context.Context, projectID int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	_, err = ormer.Raw("DELETE FROM robot WHERE project_id = ?", projectID).Exec()

	return err
}
