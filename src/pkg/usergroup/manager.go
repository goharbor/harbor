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
	List(ctx context.Context, query model.UserGroup) ([]*model.UserGroup, error)
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
	query := model.UserGroup{
		GroupName: userGroup.GroupName,
		GroupType: userGroup.GroupType,
	}
	ug, err := m.dao.Query(ctx, query)
	if err != nil {
		return 0, err
	}
	if len(ug) > 0 {
		return 0, ErrDupUserGroup
	}
	return m.dao.Add(ctx, userGroup)
}

func (m *manager) List(ctx context.Context, query model.UserGroup) ([]*model.UserGroup, error) {
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
			return ugList, err
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
		return m.onBoardCommonUserGroup(ctx, g, "LdapGroupDN", "GroupType")
	}
	return m.onBoardCommonUserGroup(ctx, g, "GroupName", "GroupType")
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
