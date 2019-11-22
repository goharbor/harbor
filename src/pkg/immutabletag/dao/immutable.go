package dao

import (
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/immutabletag/dao/model"
)

// ImmutableRuleDao defines the interface to access the ImmutableRule data model
type ImmutableRuleDao interface {
	CreateImmutableRule(ir *model.ImmutableRule) (int64, error)
	UpdateImmutableRule(projectID int64, ir *model.ImmutableRule) (int64, error)
	ToggleImmutableRule(id int64, status bool) (int64, error)
	GetImmutableRule(id int64) (*model.ImmutableRule, error)
	QueryImmutableRuleByProjectID(projectID int64) ([]model.ImmutableRule, error)
	QueryEnabledImmutableRuleByProjectID(projectID int64) ([]model.ImmutableRule, error)
	DeleteImmutableRule(id int64) (int64, error)
}

// New creates a default implementation for ImmutableRuleDao
func New() ImmutableRuleDao {
	return &immutableRuleDao{}
}

type immutableRuleDao struct{}

// CreateImmutableRule creates the Immutable Rule
func (i *immutableRuleDao) CreateImmutableRule(ir *model.ImmutableRule) (int64, error) {
	ir.Disabled = false
	o := dao.GetOrmer()
	return o.Insert(ir)
}

// UpdateImmutableRule update the immutable rules
func (i *immutableRuleDao) UpdateImmutableRule(projectID int64, ir *model.ImmutableRule) (int64, error) {
	ir.ProjectID = projectID
	o := dao.GetOrmer()
	return o.Update(ir, "TagFilter")
}

// ToggleImmutableRule enable/disable immutable rules
func (i *immutableRuleDao) ToggleImmutableRule(id int64, status bool) (int64, error) {
	o := dao.GetOrmer()
	ir := &model.ImmutableRule{ID: id, Disabled: status}
	return o.Update(ir, "Disabled")
}

// GetImmutableRule get immutable rule
func (i *immutableRuleDao) GetImmutableRule(id int64) (*model.ImmutableRule, error) {
	o := dao.GetOrmer()
	ir := &model.ImmutableRule{ID: id}
	err := o.Read(ir)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return ir, nil
}

// QueryImmutableRuleByProjectID get all immutable rule by project
func (i *immutableRuleDao) QueryImmutableRuleByProjectID(projectID int64) ([]model.ImmutableRule, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(&model.ImmutableRule{}).Filter("ProjectID", projectID)
	r := make([]model.ImmutableRule, 0)
	_, err := qs.All(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to get immutable tag rule by projectID %d, error: %v", projectID, err)
	}
	return r, nil
}

// QueryEnabledImmutableRuleByProjectID get all enabled immutable rule by project
func (i *immutableRuleDao) QueryEnabledImmutableRuleByProjectID(projectID int64) ([]model.ImmutableRule, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(&model.ImmutableRule{}).Filter("ProjectID", projectID).Filter("Disabled", false)
	var r []model.ImmutableRule
	_, err := qs.All(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled immutable tag rule for by projectID %d, error: %v", projectID, err)
	}
	return r, nil
}

// DeleteImmutableRule delete the immutable rule
func (i *immutableRuleDao) DeleteImmutableRule(id int64) (int64, error) {
	o := dao.GetOrmer()
	ir := &model.ImmutableRule{ID: id}
	return o.Delete(ir)
}
