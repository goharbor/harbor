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

package usergroup

import (
	"context"
	"errors"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/usergroup/dao"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
)

var (
	// Mgr default user group manager
	Mgr = newManager()
	// ErrDupUserGroup duplicate user group error
	ErrDupUserGroup = errors.New("duplicated user group name found")
)

// Manager interface provide the management functions for user group
type Manager interface {
	// Create create user group
	Create(ctx context.Context, userGroup model.UserGroup) (int, error)
	// List list user group
	List(ctx context.Context, query *q.Query) ([]*model.UserGroup, error)
	// Count get user group count
	Count(ctx context.Context, query *q.Query) (int64, error)
	// Get get user group by id
	Get(ctx context.Context, id int) (*model.UserGroup, error)
	// Populate populate user group from external auth server to Harbor and return the group id
	Populate(ctx context.Context, userGroups []model.UserGroup) ([]int, error)
	// Delete delete user group by id
	Delete(ctx context.Context, id int) error
	// UpdateName update user group's name
	UpdateName(ctx context.Context, id int, groupName string) error
	// Onboard sync the user group from external auth server to Harbor
	Onboard(ctx context.Context, g *model.UserGroup) error
}

type manager struct {
	dao dao.DAO
}

func newManager() Manager {
	return &manager{dao: dao.New()}
}

func (m *manager) Create(ctx context.Context, userGroup model.UserGroup) (int, error) {
	ug, err := m.dao.Query(ctx, q.New(q.KeyWords{"GroupName": userGroup.GroupName, "GroupType": userGroup.GroupType}))
	if err != nil {
		return 0, err
	}
	if len(ug) > 0 {
		return 0, ErrDupUserGroup
	}
	return m.dao.Add(ctx, userGroup)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.UserGroup, error) {
	return m.dao.Query(ctx, query)
}

func (m *manager) Get(ctx context.Context, id int) (*model.UserGroup, error) {
	return m.dao.Get(ctx, id)
}

func (m *manager) Populate(ctx context.Context, userGroups []model.UserGroup) ([]int, error) {
	ugList := make([]int, 0)
	for _, group := range userGroups {
		err := m.Onboard(ctx, &group)
		if err != nil {
			// log the current error and continue
			log.Warningf("failed to onboard user group %+v, error %v, continue with other user groups", group, err)
			continue
		}
		if group.ID > 0 {
			ugList = append(ugList, group.ID)
		}
	}
	return ugList, nil
}

func (m *manager) Delete(ctx context.Context, id int) error {
	return m.dao.Delete(ctx, id)
}

func (m *manager) UpdateName(ctx context.Context, id int, groupName string) error {
	return m.dao.UpdateName(ctx, id, groupName)
}

func (m *manager) Onboard(ctx context.Context, g *model.UserGroup) error {
	if g.GroupType == common.LDAPGroupType {
		return m.onBoardLdapUserGroup(ctx, g)
	}
	return m.onBoardCommonUserGroup(ctx, g, "GroupName", "GroupType")
}

// onBoardLdapUserGroup -- Check if the ldap group name duplicated and onboard the ldap group
func (m *manager) onBoardLdapUserGroup(ctx context.Context, g *model.UserGroup) error {
	g.LdapGroupDN = utils.TrimLower(g.LdapGroupDN)
	// check if any duplicate ldap group name exist
	ug, err := m.dao.Query(ctx, q.New(q.KeyWords{"GroupName": g.GroupName, "GroupType": g.GroupType}))
	if err != nil {
		return err
	}
	if len(ug) > 0 {
		if g.LdapGroupDN == ug[0].LdapGroupDN {
			g.ID = ug[0].ID
			return nil
		}
		// if duplicated with name, fall back to ldap group dn
		if len(g.LdapGroupDN) <= 255 {
			g.GroupName = g.LdapGroupDN
		} else {
			g.GroupName = g.LdapGroupDN[:254]
		}
		log.Warningf("existing duplicate user group with the same name, name the current user group with ldap group DN %v", g.GroupName)
	}
	return m.onBoardCommonUserGroup(ctx, g, "LdapGroupDN", "GroupType")
}

func (m *manager) onBoardCommonUserGroup(ctx context.Context, g *model.UserGroup, keyAttribute string, combinedKeyAttributes ...string) error {
	g.LdapGroupDN = utils.TrimLower(g.LdapGroupDN)
	created, ID, err := m.dao.ReadOrCreate(ctx, g, keyAttribute, combinedKeyAttributes...)
	if err != nil {
		return err
	}
	if created {
		g.ID = int(ID)
	} else {
		prevGroup, err := m.dao.Get(ctx, int(ID))
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

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}
