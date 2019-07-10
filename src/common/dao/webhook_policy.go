package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
)

// GetWebhookPolicy return webhook policy by id
func GetWebhookPolicy(id int64) (*models.WebhookPolicy, error) {
	policy := new(models.WebhookPolicy)
	o := orm.NewOrm()
	err := o.QueryTable(policy).Filter("id", id).One(policy)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return policy, err
}

// GetWebhookPolicyByName return webhook policy by name
func GetWebhookPolicyByName(name string, projectID int64) (*models.WebhookPolicy, error) {
	policy := new(models.WebhookPolicy)
	o := GetOrmer()
	err := o.QueryTable(policy).Filter("name", name).Filter("projectID", projectID).One(policy)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return policy, err
}

// GetWebhookPolicies returns all webhook policy in project
func GetWebhookPolicies(projectID int64) (int64, []*models.WebhookPolicy, error) {
	var policies []*models.WebhookPolicy
	qs := GetOrmer().QueryTable(new(models.WebhookPolicy)).Filter("ProjectID", projectID)

	total, err := qs.All(&policies)
	if err != nil {
		return -1, nil, err
	}
	return total, policies, nil

}

// AddWebhookPolicy insert new webhook policy to DB
func AddWebhookPolicy(policy *models.WebhookPolicy) (int64, error) {
	o := GetOrmer()
	return o.Insert(policy)
}

// UpdateWebhookPolicy update t specified webhook policy
func UpdateWebhookPolicy(policy *models.WebhookPolicy) error {
	o := GetOrmer()
	_, err := o.Update(policy)
	return err
}

// DeleteWebhookPolicy delete webhook policy by id
func DeleteWebhookPolicy(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.WebhookPolicy{ID: id})
	return err
}
