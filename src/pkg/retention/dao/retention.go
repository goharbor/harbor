package dao

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
)

func CreatePolicy(p *models.RetentionPolicy) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(p)
}

func UpdatePolicy(p *models.RetentionPolicy) error {
	o := dao.GetOrmer()
	_, err := o.Update(p)
	return err
}

func DeletePolicy(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.RetentionPolicy{
		ID: id,
	})
	return err
}

func GetPolicy(id int64) (*models.RetentionPolicy, error) {
	o := dao.GetOrmer()
	p := &models.RetentionPolicy{
		ID: id,
	}
	if err := o.Read(p); err != nil {
		return nil, err
	} else {
		return p, nil
	}
}
