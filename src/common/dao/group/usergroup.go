// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package group

import (
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// AddUserGroup - Add User Group
func AddUserGroup(userGroup models.UserGroup) (int, error) {
	o := dao.GetOrmer()
	query := models.UserGroup{GroupType: userGroup.GroupType, GroupProperty: userGroup.GroupProperty}
	result, err := QueryUserGroup(query)
	if err != nil {
		return 0, err
	}
	if len(result) > 0 {
		log.Debugf("Already exist same group type and group property")
		return result[0].ID, nil
	}
	id, err := o.Insert(&userGroup)

	if err != nil {
		return 0, err
	}
	return int(id), err
}

// QueryUserGroup - Query User Group
func QueryUserGroup(query models.UserGroup) ([]*models.UserGroup, error) {
	o := dao.GetOrmer()
	sql := `select id, group_name, group_type, group_property from user_group where 1=1 `
	sqlParam := make([]interface{}, 1)
	groups := []*models.UserGroup{}
	if len(query.GroupName) != 0 {
		sql += ` and group_name like ? `
		sqlParam = append(sqlParam, `%`+dao.Escape(query.GroupName)+`%`)
	}

	if query.GroupType != 0 {
		sql += ` and group_type = ? `
		sqlParam = append(sqlParam, query.GroupType)
	}

	if len(query.GroupProperty) != 0 {
		sql += ` and group_property = ? `
		sqlParam = append(sqlParam, query.GroupProperty)
	}
	_, err := o.Raw(sql, sqlParam).QueryRows(&groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// GetUserGroup ...
func GetUserGroup(id int) (*models.UserGroup, error) {
	userGroup := models.UserGroup{ID: id}
	o := dao.GetOrmer()
	err := o.Read(&userGroup)
	if err != nil {
		return nil, err
	}
	return &userGroup, nil
}

// DeleteUserGroup ...
func DeleteUserGroup(id int) error {
	userGroup := models.UserGroup{ID: id}
	o := dao.GetOrmer()
	_, err := o.Delete(&userGroup)
	return err
}

// UpdateUserGroup ...
func UpdateUserGroup(id int, groupName string) error {
	log.Debugf("Updating user_group with id:%v, name:%v", id, groupName)
	o := dao.GetOrmer()
	sql := "update user_group set group_name = ? where id =  ? "
	_, err := o.Raw(sql, groupName, id).Exec()
	return err
}
