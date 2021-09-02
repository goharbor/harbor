//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package dao

import (
	"context"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	"time"
)

func init() {
	orm.RegisterModel(
		new(model.UserGroup),
	)
}

// DAO the dao for user group
type DAO interface {
	// Add add user group
	Add(ctx context.Context, userGroup model.UserGroup) (int, error)
	// Count query user group count
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Query query user group
	Query(ctx context.Context, query *q.Query) ([]*model.UserGroup, error)
	// Get get user group by id
	Get(ctx context.Context, id int) (*model.UserGroup, error)
	// Delete delete user group by id
	Delete(ctx context.Context, id int) error
	// UpdateName update user group name
	UpdateName(ctx context.Context, id int, groupName string) error
	// ReadOrCreate create a user group or read existing one from db
	ReadOrCreate(ctx context.Context, g *model.UserGroup, keyAttribute string, combinedKeyAttributes ...string) (bool, int64, error)
}

type dao struct {
}

// New create user group DAO
func New() DAO {
	return &dao{}
}

// ErrGroupNameDup ...
var ErrGroupNameDup = errors.ConflictError(nil).WithMessage("duplicated user group name")

// Add - Add User Group
func (d *dao) Add(ctx context.Context, userGroup model.UserGroup) (int, error) {
	query := q.New(q.KeyWords{"GroupName": userGroup.GroupName, "GroupType": common.HTTPGroupType})
	userGroupList, err := d.Query(ctx, query)
	if err != nil {
		return 0, ErrGroupNameDup
	}
	if len(userGroupList) > 0 {
		return 0, ErrGroupNameDup
	}
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	sql := "insert into user_group (group_name, group_type, ldap_group_dn, creation_time, update_time) values (?, ?, ?, ?, ?) RETURNING id"
	var id int
	now := time.Now()

	err = o.Raw(sql, userGroup.GroupName, userGroup.GroupType, utils.TrimLower(userGroup.LdapGroupDN), now, now).QueryRow(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

// Query - Query User Group
func (d *dao) Query(ctx context.Context, query *q.Query) ([]*model.UserGroup, error) {
	query = q.MustClone(query)
	qs, err := orm.QuerySetter(ctx, &model.UserGroup{}, query)
	if err != nil {
		return nil, err
	}
	var usergroups []*model.UserGroup
	if _, err := qs.All(&usergroups); err != nil {
		return nil, err
	}
	return usergroups, nil
}

// Get ...
func (d *dao) Get(ctx context.Context, id int) (*model.UserGroup, error) {
	userGroupList, err := d.Query(ctx, q.New(q.KeyWords{"ID": id}))
	if err != nil {
		return nil, err
	}
	if len(userGroupList) > 0 {
		return userGroupList[0], nil
	}
	return nil, nil
}

// Delete ...
func (d *dao) Delete(ctx context.Context, id int) error {
	userGroup := model.UserGroup{ID: id}
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = o.Delete(&userGroup)
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

// UpdateName ...
func (d *dao) UpdateName(ctx context.Context, id int, groupName string) error {
	log.Debugf("Updating user_group with id:%v, name:%v", id, groupName)
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	sql := "update user_group set group_name = ? where id =  ? "
	_, err = o.Raw(sql, groupName, id).Exec()
	return err
}

// ReadOrCreate read or create user group
func (d *dao) ReadOrCreate(ctx context.Context, g *model.UserGroup, keyAttribute string, combinedKeyAttributes ...string) (bool, int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return false, 0, err
	}
	return o.ReadOrCreate(g, keyAttribute, combinedKeyAttributes...)
}

func (d *dao) onBoardCommonUserGroup(ctx context.Context, g *model.UserGroup, keyAttribute string, combinedKeyAttributes ...string) error {
	g.LdapGroupDN = utils.TrimLower(g.LdapGroupDN)
	created, ID, err := d.ReadOrCreate(ctx, g, keyAttribute, combinedKeyAttributes...)
	if err != nil {
		return err
	}

	if created {
		g.ID = int(ID)
	} else {
		prevGroup, err := d.Get(ctx, int(ID))
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

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	query = q.MustClone(query)
	qs, err := orm.QuerySetterForCount(ctx, &model.UserGroup{}, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}
