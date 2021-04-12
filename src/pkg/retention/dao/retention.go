package dao

import (
	"context"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
)

// CreatePolicy Create Policy
func CreatePolicy(ctx context.Context, p *models.RetentionPolicy) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	return o.Insert(p)
}

// UpdatePolicy Update Policy
func UpdatePolicy(ctx context.Context, p *models.RetentionPolicy, cols ...string) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = o.Update(p, cols...)
	return err
}

// DeletePolicy Update Policy
func DeletePolicy(ctx context.Context, id int64) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	p := &models.RetentionPolicy{
		ID: id,
	}
	_, err = o.Delete(p)
	return err
}

// GetPolicy Get Policy
func GetPolicy(ctx context.Context, id int64) (*models.RetentionPolicy, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	p := &models.RetentionPolicy{
		ID: id,
	}
	if err := o.Read(p); err != nil {
		return nil, err
	}
	return p, nil
}
