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

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/ldap"
	"github.com/goharbor/harbor/src/pkg/usergroup"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
)

var (
	// Ctl Global instance of the UserGroup controller
	Ctl = newController()
)

// Controller manages the user group
type Controller interface {
	// Delete delete user group
	Delete(ctx context.Context, id int) error
	// Update update the user group name
	Update(ctx context.Context, id int, groupName string) error
	// Create create user group
	Create(ctx context.Context, group model.UserGroup) (int, error)
	// Get get user group by id
	Get(ctx context.Context, id int) (*model.UserGroup, error)
	// Ensure if this user group doesn't exist in Harbor, create it in Harbor, if exist, do nothing
	Ensure(ctx context.Context, group *model.UserGroup) error
	// Populate populate user group and get the user group's id
	Populate(ctx context.Context, userGroups []model.UserGroup) ([]int, error)
	// List list user groups
	List(ctx context.Context, q *q.Query) ([]*model.UserGroup, error)
	// Count user group count
	Count(ctx context.Context, q *q.Query) (int64, error)
}

type controller struct {
	mgr usergroup.Manager
}

func newController() Controller {
	return &controller{mgr: usergroup.Mgr}
}

func (c *controller) List(ctx context.Context, query *q.Query) ([]*model.UserGroup, error) {
	return c.mgr.List(ctx, query)
}

func (c *controller) Populate(ctx context.Context, userGroups []model.UserGroup) ([]int, error) {
	return c.mgr.Populate(ctx, userGroups)
}

func (c *controller) Ensure(ctx context.Context, group *model.UserGroup) error {
	return c.mgr.Onboard(ctx, group)
}

func (c *controller) Delete(ctx context.Context, id int) error {
	return c.mgr.Delete(ctx, id)
}

func (c *controller) Update(ctx context.Context, id int, groupName string) error {
	ug, err := c.mgr.List(ctx, q.New(q.KeyWords{"ID": id}))
	if err != nil {
		return err
	}
	if len(ug) == 0 {
		return errors.NotFoundError(nil).WithMessage("the user group with id %v is not found", id)
	}
	return c.mgr.UpdateName(ctx, id, groupName)
}

func (c *controller) Create(ctx context.Context, group model.UserGroup) (int, error) {
	if group.GroupType == common.LDAPGroupType {
		ldapGroup, err := auth.SearchGroup(ctx, group.LdapGroupDN)
		if err == ldap.ErrNotFound || ldapGroup == nil {
			return 0, errors.BadRequestError(nil).WithMessage("LDAP Group DN is not found: DN:%v", group.LdapGroupDN)
		}
		if err == ldap.ErrDNSyntax {
			return 0, errors.BadRequestError(nil).WithMessage("invalid DN syntax. DN: %v", group.LdapGroupDN)
		}
		if err != nil {
			return 0, err
		}

	}
	id, err := c.mgr.Create(ctx, group)
	if err != nil && err == usergroup.ErrDupUserGroup {
		return 0, errors.ConflictError(nil).
			WithMessage("duplicate user group, group name:%v, group type: %v, ldap group DN: %v",
				group.GroupName, group.GroupType, group.LdapGroupDN)
	}

	return id, err
}

func (c *controller) Get(ctx context.Context, id int) (*model.UserGroup, error) {
	return c.mgr.Get(ctx, id)
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.mgr.Count(ctx, query)
}
