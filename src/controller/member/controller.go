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

package member

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/member"
	"github.com/goharbor/harbor/src/pkg/member/models"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/pkg/usergroup"
)

// Controller defines the operation related to project member
type Controller interface {
	// Get get the project member with ID
	Get(ctx context.Context, projectNameOrID interface{}, memberID int) (*models.Member, error)
	// Create add project member to project
	Create(ctx context.Context, projectNameOrID interface{}, req Request) (int, error)
	// Delete member from project
	Delete(ctx context.Context, projectNameOrID interface{}, memberID int) error
	// List list all project members with condition
	List(ctx context.Context, projectNameOrID interface{}, entityName string, query *q.Query) ([]*models.Member, error)
	// UpdateRole update the project member role
	UpdateRole(ctx context.Context, projectNameOrID interface{}, memberID int, role int) error
	// Count get the total amount of project members
	Count(ctx context.Context, projectNameOrID interface{}, query *q.Query) (int, error)
}

// Request - Project Member Request
type Request struct {
	ProjectID   int64     `json:"project_id"`
	Role        int       `json:"role_id,omitempty"`
	MemberUser  User      `json:"member_user,omitempty"`
	MemberGroup UserGroup `json:"member_group,omitempty"`
}

// User ...
type User struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}

// UserGroup ...
type UserGroup struct {
	ID          int    `json:"id,omitempty"`
	GroupName   string `json:"group_name,omitempty"`
	GroupType   int    `json:"group_type,omitempty"`
	LdapGroupDN string `json:"ldap_group_dn,omitempty"`
}

// ErrDuplicateProjectMember ...
var ErrDuplicateProjectMember = errors.ConflictError(nil).WithMessage("The project member specified already exist")

// ErrInvalidRole ...
var ErrInvalidRole = errors.BadRequestError(nil).WithMessage("Failed to update project member, role is not in 1,2,3")

type controller struct {
	userManager user.Manager
	mgr         member.Manager
	projectMgr  project.Manager
}

// NewController ...
func NewController() Controller {
	return &controller{mgr: member.Mgr, projectMgr: pkg.ProjectMgr, userManager: user.New()}
}

func (c *controller) Count(ctx context.Context, projectNameOrID interface{}, query *q.Query) (int, error) {
	p, err := c.projectMgr.Get(ctx, projectNameOrID)
	if err != nil {
		return 0, err
	}
	return c.mgr.GetTotalOfProjectMembers(ctx, p.ProjectID, query)
}

func (c *controller) UpdateRole(ctx context.Context, projectNameOrID interface{}, memberID int, role int) error {
	p, err := c.projectMgr.Get(ctx, projectNameOrID)
	if err != nil {
		return err
	}
	if p == nil {
		return errors.BadRequestError(nil).WithMessage("project is not found")
	}
	return c.mgr.UpdateRole(ctx, p.ProjectID, memberID, role)
}

func (c *controller) Get(ctx context.Context, projectNameOrID interface{}, memberID int) (*models.Member, error) {
	p, err := c.projectMgr.Get(ctx, projectNameOrID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.BadRequestError(nil).WithMessage("project is not found")
	}
	return c.mgr.Get(ctx, p.ProjectID, memberID)
}

func (c *controller) Create(ctx context.Context, projectNameOrID interface{}, req Request) (int, error) {
	p, err := c.projectMgr.Get(ctx, projectNameOrID)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, errors.BadRequestError(nil).WithMessage("project is not found")
	}
	var member models.Member
	member.ProjectID = p.ProjectID
	member.Role = req.Role
	member.EntityType = common.GroupMember

	if req.MemberUser.UserID > 0 {
		member.EntityID = req.MemberUser.UserID
		member.EntityType = common.UserMember
	} else if req.MemberGroup.ID > 0 {
		member.EntityID = req.MemberGroup.ID
	} else if len(req.MemberUser.Username) > 0 {
		// If username is provided, search userid by username
		var userID int
		member.EntityType = common.UserMember
		u, err := c.userManager.GetByName(ctx, req.MemberUser.Username)
		if err != nil && !errors.IsNotFoundErr(err) {
			return 0, err
		}
		if u != nil {
			userID = u.UserID
		} else {
			userID, err = auth.SearchAndOnBoardUser(ctx, req.MemberUser.Username)
			if err != nil {
				return 0, err
			}
		}
		member.EntityID = userID
	} else if len(req.MemberGroup.LdapGroupDN) > 0 {
		req.MemberGroup.GroupType = common.LDAPGroupType
		// if the ldap group dn already exist
		ugs, err := usergroup.Mgr.List(ctx, q.New(q.KeyWords{"LdapGroupDN": req.MemberGroup.LdapGroupDN, "GroupType": req.MemberGroup.GroupType}))
		if err != nil {
			return 0, err
		}
		if len(ugs) > 0 {
			member.EntityID = ugs[0].ID
			member.EntityType = common.GroupMember
		} else {
			// If groupname provided, use the provided groupname to name this group
			groupID, err := auth.SearchAndOnBoardGroup(ctx, req.MemberGroup.LdapGroupDN, req.MemberGroup.GroupName)
			if err != nil {
				return 0, err
			}
			member.EntityID = groupID
		}

	} else if len(req.MemberGroup.GroupName) > 0 {
		// all group type can be added to project member by name
		ugs, err := usergroup.Mgr.List(ctx, q.New(q.KeyWords{"GroupName": req.MemberGroup.GroupName, "GroupType": req.MemberGroup.GroupType}))
		if err != nil {
			return 0, err
		}
		if len(ugs) == 0 {
			groupID, err := auth.SearchAndOnBoardGroup(ctx, req.MemberGroup.GroupName, "")
			if err != nil {
				return 0, err
			}
			member.EntityID = groupID
		} else {
			member.EntityID = ugs[0].ID
		}

	}
	if member.EntityID <= 0 {
		return 0, fmt.Errorf("can not get valid member entity, request: %+v", req)
	}

	// Check if member already exist in current project
	memberList, err := c.mgr.List(ctx, models.Member{
		ProjectID:  member.ProjectID,
		EntityID:   member.EntityID,
		EntityType: member.EntityType,
	}, nil)
	if err != nil {
		return 0, err
	}
	if len(memberList) > 0 {
		return 0, ErrDuplicateProjectMember
	}

	if !isValidRole(member.Role) {
		// Return invalid role error
		return 0, ErrInvalidRole
	}
	return c.mgr.AddProjectMember(ctx, member)
}

func isValidRole(role int) bool {
	switch role {
	case common.RoleProjectAdmin,
		common.RoleMaintainer,
		common.RoleDeveloper,
		common.RoleGuest,
		common.RoleLimitedGuest:
		return true
	default:
		return false
	}
}

func (c *controller) List(ctx context.Context, projectNameOrID interface{}, entityName string, query *q.Query) ([]*models.Member, error) {
	p, err := c.projectMgr.Get(ctx, projectNameOrID)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.BadRequestError(nil).WithMessage("project is not found")
	}
	pm := models.Member{
		ProjectID:  p.ProjectID,
		Entityname: entityName,
	}
	return c.mgr.List(ctx, pm, query)
}

func (c *controller) Delete(ctx context.Context, projectNameOrID interface{}, memberID int) error {
	p, err := c.projectMgr.Get(ctx, projectNameOrID)
	if err != nil {
		return err
	}
	return c.mgr.Delete(ctx, p.ProjectID, memberID)
}
