package notification

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
)

// GetNotificationPolicy return notification policy by id
func GetNotificationPolicy(id int64) (*models.NotificationPolicy, error) {
	policy := new(models.NotificationPolicy)
	o := dao.GetOrmer()
	err := o.QueryTable(policy).Filter("id", id).One(policy)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return policy, err
}

// GetNotificationPolicyByName return notification policy by name
func GetNotificationPolicyByName(name string, projectID int64) (*models.NotificationPolicy, error) {
	policy := new(models.NotificationPolicy)
	o := dao.GetOrmer()
	err := o.QueryTable(policy).Filter("name", name).Filter("projectID", projectID).One(policy)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return policy, err
}

// GetNotificationPolicies returns all notification policy in project
func GetNotificationPolicies(projectID int64) ([]*models.NotificationPolicy, error) {
	var policies []*models.NotificationPolicy
	qs := dao.GetOrmer().QueryTable(new(models.NotificationPolicy)).Filter("ProjectID", projectID)

	_, err := qs.All(&policies)
	if err != nil {
		return nil, err
	}
	return policies, nil

}

// AddNotificationPolicy insert new notification policy to DB
func AddNotificationPolicy(policy *models.NotificationPolicy) (int64, error) {
	if policy == nil {
		return 0, errors.New("nil policy")
	}
	o := dao.GetOrmer()
	return o.Insert(policy)
}

// UpdateNotificationPolicy update t specified notification policy
func UpdateNotificationPolicy(policy *models.NotificationPolicy) error {
	if policy == nil {
		return errors.New("nil policy")
	}
	o := dao.GetOrmer()
	_, err := o.Update(policy)
	return err
}

// DeleteNotificationPolicy delete notification policy by id
func DeleteNotificationPolicy(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.NotificationPolicy{ID: id})
	return err
}
