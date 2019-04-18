package dao

import (
	"errors"
	"time"

	"github.com/astaxie/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
)

// AddRepPolicy insert new policy to DB.
func AddRepPolicy(policy *models.RepPolicy) (int64, error) {
	o := common_dao.GetOrmer()
	now := time.Now()
	policy.CreationTime = now
	policy.UpdateTime = now

	return o.Insert(policy)
}

// GetPolicies list polices with given query parameters.
func GetPolicies(queries ...*model.PolicyQuery) (int64, []*models.RepPolicy, error) {
	var qs = common_dao.GetOrmer().QueryTable(new(models.RepPolicy))
	var policies []*models.RepPolicy

	if len(queries) == 0 {
		total, err := qs.Count()
		if err != nil {
			return -1, nil, err
		}

		_, err = qs.All(&policies)
		if err != nil {
			return total, nil, err
		}

		return total, policies, nil
	}

	query := queries[0]
	if len(query.Name) != 0 {
		qs = qs.Filter("Name__icontains", query.Name)
	}
	if len(query.Namespace) != 0 {
		// TODO: Namespace filter not implemented yet
	}
	if query.SrcRegistry > 0 {
		qs = qs.Filter("SrcRegistryID__exact", query.SrcRegistry)
	}
	if query.DestRegistry > 0 {
		qs = qs.Filter("DestRegistryID__exact", query.DestRegistry)
	}

	total, err := qs.Count()
	if err != nil {
		return -1, nil, err
	}

	if query.Page > 0 && query.Size > 0 {
		qs = qs.Limit(query.Size, (query.Page-1)*query.Size)
	}
	_, err = qs.All(&policies)
	if err != nil {
		return total, nil, err
	}

	return total, policies, nil
}

// GetRepPolicy return special policy by id.
func GetRepPolicy(id int64) (policy *models.RepPolicy, err error) {
	policy = new(models.RepPolicy)
	err = common_dao.GetOrmer().QueryTable(policy).
		Filter("id", id).One(policy)
	if err == orm.ErrNoRows {
		return nil, nil
	}

	return
}

// GetRepPolicyByName return special policy by name.
func GetRepPolicyByName(name string) (policy *models.RepPolicy, err error) {
	policy = new(models.RepPolicy)
	err = common_dao.GetOrmer().QueryTable(policy).
		Filter("name", name).One(policy)
	if err == orm.ErrNoRows {
		return nil, nil
	}

	return
}

// UpdateRepPolicy update fields by props
func UpdateRepPolicy(policy *models.RepPolicy, props ...string) (err error) {
	var o = common_dao.GetOrmer()

	if policy != nil {
		_, err = o.Update(policy, props...)
	} else {
		err = errors.New("Nil policy")
	}

	return
}

// DeleteRepPolicy will hard delete database item
func DeleteRepPolicy(id int64) error {
	o := common_dao.GetOrmer()

	_, err := o.Delete(&models.RepPolicy{ID: id})
	return err
}
