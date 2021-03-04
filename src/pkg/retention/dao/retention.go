package dao

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
)

// CreatePolicy Create Policy
func CreatePolicy(p *models.RetentionPolicy) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(p)
}

// UpdatePolicy Update Policy
func UpdatePolicy(p *models.RetentionPolicy, cols ...string) error {
	o := dao.GetOrmer()
	_, err := o.Update(p, cols...)
	return err
}

// DeletePolicy Update Policy
func DeletePolicy(id int64) error {
	o := dao.GetOrmer()
	p := &models.RetentionPolicy{
		ID: id,
	}
	_, err := o.Delete(p)
	return err
}

// GetPolicy Get Policy
func GetPolicy(id int64) (*models.RetentionPolicy, error) {
	o := dao.GetOrmer()
	p := &models.RetentionPolicy{
		ID: id,
	}
	if err := o.Read(p); err != nil {
		return nil, err
	}
	return p, nil
}
