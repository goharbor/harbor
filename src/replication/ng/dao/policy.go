package dao

import (
	"errors"
	"time"

	"github.com/astaxie/beego/orm"
	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/replication/ng/dao/models"
)

// AddRepPolicy insert new policy to DB.
func AddRepPolicy(policy *models.RepPolicy) (int64, error) {
	o := common_dao.GetOrmer()
	now := time.Now()
	policy.CreationTime = now
	policy.UpdateTime = now

	return o.Insert(policy)
}

func filteredRepPolicyQuerySeter(name, namespace string) orm.QuerySeter {
	var qs = common_dao.GetOrmer().QueryTable(new(models.RepPolicy))

	// TODO: just filter polices by name now, and need consider how to  filter namespace.
	qs = qs.Filter("name__icontains", name)

	return qs
}

// GetTotalOfRepPolicies returns the total count of replication policies
func GetTotalOfRepPolicies(name, namespace string) (int64, error) {
	var qs = filteredRepPolicyQuerySeter(name, namespace)
	return qs.Count()
}

// GetPolicies filter policies and pagination.
func GetPolicies(name, namespace string, page, pageSize int64) (policies []*models.RepPolicy, err error) {
	var qs = filteredRepPolicyQuerySeter(name, namespace)

	// Paginate
	if page > 0 && pageSize > 0 {
		qs = qs.Limit(pageSize, (page-1)*pageSize)
	}

	_, err = qs.All(&policies)

	return
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
