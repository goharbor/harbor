// Copyright Project Harbor Authors
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
	"time"

	"github.com/goharbor/harbor/src/common/utils"

	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/pkg/errors"
)

// ErrGroupNameDup ...
var ErrGroupNameDup = errors.New("duplicated user group name")

// AddUserGroup - Add User Group
func AddUserGroup(userGroup models.UserGroup) (int, error) {
	userGroupList, err := QueryUserGroup(models.UserGroup{GroupName: userGroup.GroupName, GroupType: common.HTTPGroupType})
	if err != nil {
		return 0, ErrGroupNameDup
	}
	if len(userGroupList) > 0 {
		return 0, ErrGroupNameDup
	}
	o := dao.GetOrmer()
	sql := "insert into user_group (group_name, group_type, ldap_group_dn, creation_time, update_time) values (?, ?, ?, ?, ?) RETURNING id"
	var id int
	now := time.Now()

	err = o.Raw(sql, userGroup.GroupName, userGroup.GroupType, utils.TrimLower(userGroup.LdapGroupDN), now, now).QueryRow(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// QueryUserGroup - Query User Group
func QueryUserGroup(query models.UserGroup) ([]*models.UserGroup, error) {
	o := dao.GetOrmer()
	sql := `select id, group_name, group_type, ldap_group_dn from user_group where 1=1 `
	sqlParam := make([]interface{}, 1)
	var groups []*models.UserGroup
	if len(query.GroupName) != 0 {
		sql += ` and group_name = ? `
		sqlParam = append(sqlParam, query.GroupName)
	}

	if query.GroupType != 0 {
		sql += ` and group_type = ? `
		sqlParam = append(sqlParam, query.GroupType)
	}

	if len(query.LdapGroupDN) != 0 {
		sql += ` and ldap_group_dn = ? `
		sqlParam = append(sqlParam, utils.TrimLower(query.LdapGroupDN))
	}
	if query.ID != 0 {
		sql += ` and id = ? `
		sqlParam = append(sqlParam, query.ID)
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
	userGroupList, err := QueryUserGroup(userGroup)
	if err != nil {
		return nil, err
	}
	if len(userGroupList) > 0 {
		return userGroupList[0], nil
	}
	return nil, nil
}

// GetGroupIDByGroupName - Return the group ID by given group name. it is possible less group ID than the given group name if some group doesn't exist.
func GetGroupIDByGroupName(groupName []string, groupType int) ([]int, error) {
	var retGroupID []int
	if len(groupName) == 0 {
		return retGroupID, nil
	}
	sql := fmt.Sprintf("select id from user_group where group_name in ( %s ) and group_type = ? ", dao.ParamPlaceholderForIn(len(groupName)))
	log.Debugf("GetGroupIDByGroupName: statement sql is %v", sql)
	o := dao.GetOrmer()
	cnt, err := o.Raw(sql, groupName, groupType).QueryRows(&retGroupID)
	if err != nil {
		return retGroupID, err
	}
	log.Debugf("Found rows %v", cnt)
	return retGroupID, nil
}

// DeleteUserGroup ...
func DeleteUserGroup(id int) error {
	userGroup := models.UserGroup{ID: id}
	o := dao.GetOrmer()
	_, err := o.Delete(&userGroup)
	if err == nil {
		// Delete all related project members
		sql := `delete from project_member where entity_id = ? and entity_type='g'`
		_, err := o.Raw(sql, id).Exec()
		if err != nil {
			return err
		}
	}
	return err
}

// UpdateUserGroupName ...
func UpdateUserGroupName(id int, groupName string) error {
	log.Debugf("Updating user_group with id:%v, name:%v", id, groupName)
	o := dao.GetOrmer()
	sql := "update user_group set group_name = ? where id =  ? "
	_, err := o.Raw(sql, groupName, id).Exec()
	return err
}

func onBoardCommonUserGroup(g *models.UserGroup, keyAttribute string, combinedKeyAttributes ...string) error {
	g.LdapGroupDN = utils.TrimLower(g.LdapGroupDN)

	o := dao.GetOrmer()
	created, ID, err := o.ReadOrCreate(g, keyAttribute, combinedKeyAttributes...)
	if err != nil {
		return err
	}

	if created {
		g.ID = int(ID)
	} else {
		prevGroup, err := GetUserGroup(int(ID))
		if err != nil {
			return err
		}
		g.ID = prevGroup.ID
		g.GroupName = prevGroup.GroupName
		g.GroupType = prevGroup.GroupType
		g.LdapGroupDN = prevGroup.LdapGroupDN
	}

	return nil
}

// OnBoardUserGroup will check if a usergroup exists in usergroup table, if not insert the usergroup and
// put the id in the pointer of usergroup model, if it does exist, return the usergroup's profile.
func OnBoardUserGroup(g *models.UserGroup) error {
	if g.GroupType == common.LDAPGroupType {
		return onBoardCommonUserGroup(g, "LdapGroupDN", "GroupType")
	}
	return onBoardCommonUserGroup(g, "GroupName", "GroupType")
}
