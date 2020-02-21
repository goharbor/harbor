package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
)

// ListInstanceQuery defines the query params of the instance record.
type ListInstanceQuery struct {
	Page     uint
	PageSize uint
	Keyword  string
}

// AddInstance adds a new distribution instance.
func AddInstance(instance *models.Instance) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(instance)
}

// GetInstance gets instance from db by id.
func GetInstance(id int64) (*models.Instance, error) {
	o := dao.GetOrmer()
	di := models.Instance{ID: id}
	err := o.Read(&di, "ID")
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &di, err
}

// UpdateInstance updates distribution instance.
func UpdateInstance(instance *models.Instance, props ...string) error {
	o := dao.GetOrmer()
	_, err := o.Update(instance, props...)
	return err
}

// DeleteInstance deletes one distribution instance by id.
func DeleteInstance(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.Instance{ID: id})
	return err
}

// ListInstances lists instances by query parmas.
func ListInstances(query *ListInstanceQuery) ([]*models.Instance, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(&models.Instance{})

	if query != nil {
		offset := (query.Page - 1) * query.PageSize
		qs = qs.Offset(offset).Limit(query.PageSize)
		// keyword match
		if len(query.Keyword) > 0 {
			qs = qs.Filter("name__contains", query.Keyword)
		}
	}

	var ins []*models.Instance
	_, err := qs.All(&ins)
	return ins, err
}
