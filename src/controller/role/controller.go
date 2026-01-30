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

package role

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/rbac"
	rbac_model "github.com/goharbor/harbor/src/pkg/rbac/model"
	role "github.com/goharbor/harbor/src/pkg/role"
	"github.com/goharbor/harbor/src/pkg/role/model"
)

var (
	// Ctl is a global variable for the default role account controller implementation
	Ctl = NewController()
)

// Controller to handle the requests related with role
type Controller interface {
	// Get ...
	Get(ctx context.Context, id int64, option *Option) (*Role, error)

	// Count returns the total count of roles according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// Create ...
	Create(ctx context.Context, r *Role) (int64, error)

	// Delete ...
	Delete(ctx context.Context, id int64, option ...*Option) error

	// Update ...
	Update(ctx context.Context, r *Role, option *Option) error

	// List ...
	List(ctx context.Context, query *q.Query, option *Option) ([]*Role, error)
}

// controller ...
type controller struct {
	roleMgr role.Manager
	proMgr  project.Manager
	rbacMgr rbac.Manager
}

// NewController ...
func NewController() Controller {
	return &controller{
		roleMgr: role.Mgr,
		proMgr:  pkg.ProjectMgr,
		rbacMgr: rbac.Mgr,
	}
}

// Get ...
func (d *controller) Get(ctx context.Context, id int64, option *Option) (*Role, error) {
	role, err := d.roleMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return d.populate(ctx, role, option)
}

// Count ...
func (d *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return d.roleMgr.Count(ctx, query)
}

// Create ...
func (d *controller) Create(ctx context.Context, r *Role) (int64, error) {

	name := r.Name

	rCreate := &model.Role{
		Name:     name,
		RoleMask: r.RoleMask,
		RoleCode: r.RoleCode,
	}
	roleID, err := d.roleMgr.Create(ctx, rCreate)
	if err != nil {
		return 0, err
	}
	r.ID = roleID
	if err := d.createPermission(ctx, r); err != nil {
		return 0, err
	}
	// fire event
	notification.AddEvent(ctx, &metadata.CreateRoleEventMetadata{
		Ctx:  ctx,
		Role: rCreate,
	})
	return roleID, nil
}

// Delete ...
func (d *controller) Delete(ctx context.Context, id int64, option ...*Option) error {
	rDelete, err := d.roleMgr.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := d.roleMgr.Delete(ctx, id); err != nil {
		return err
	}
	if err := d.rbacMgr.DeletePermissionsByRole(ctx, ROLETYPE, id); err != nil {
		return err
	}
	// fire event
	deleteMetadata := &metadata.DeleteRoleEventMetadata{
		Ctx:  ctx,
		Role: rDelete,
	}
	if len(option) != 0 && option[0].Operator != "" {
		deleteMetadata.Operator = option[0].Operator
	}
	notification.AddEvent(ctx, deleteMetadata)
	return nil
}

// Update ...
func (d *controller) Update(ctx context.Context, r *Role, option *Option) error {
	if r == nil {
		return errors.New("cannot update a nil role").WithCode(errors.BadRequestCode)
	}

	//TODO allow to modify name !
	//	if err := d.roleMgr.Update(ctx, &r.Role, "description", "disabled"); err != nil {
	//		return err
	//	}
	// update the permission
	if option != nil && option.WithPermission {
		if err := d.rbacMgr.DeletePermissionsByRole(ctx, ROLETYPE, r.ID); err != nil && !errors.IsNotFoundErr(err) {
			return err
		}
		if err := d.createPermission(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// List ...
func (d *controller) List(ctx context.Context, query *q.Query, option *Option) ([]*Role, error) {
	role, err := d.roleMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var roles []*Role
	for _, r := range role {
		rb, err := d.populate(ctx, r, option)
		log.Debug("*** role has been populated with " + r.Name + " - " + strconv.Itoa(len(rb.Permissions)))
		if err != nil {
			return nil, err
		}
		roles = append(roles, rb)
	}
	return roles, nil
}

func (d *controller) createPermission(ctx context.Context, r *Role) error {
	if r == nil {
		return nil
	}

	for _, per := range r.Permissions {
		policy := &rbac_model.PermissionPolicy{}
		policy.Scope = "/project/*"

		for _, access := range per.Access {
			policy.Resource = access.Resource.String()
			policy.Action = access.Action.String()
			policy.Effect = access.Effect.String()

			policyID, err := d.rbacMgr.CreateRbacPolicy(ctx, policy)
			if err != nil {
				return err
			}

			_, err = d.rbacMgr.CreatePermission(ctx, &rbac_model.RolePermission{
				RoleType:           ROLETYPE,
				RoleID:             r.ID,
				PermissionPolicyID: policyID,
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *controller) populate(ctx context.Context, r *model.Role, option *Option) (*Role, error) {
	if r == nil {
		return nil, nil
	}
	role := &Role{
		Role: *r,
	}
	role.setLevel()
	role.setEditable()
	// for the v2 role, add prefix to the role name
	if role.Editable {
		role.Name = fmt.Sprintf("%s%s", config.RolePrefix(ctx), r.Name)
	} else {
		role.Name = r.Name
	}
	if option != nil && option.WithPermission {
		if err := d.populatePermissions(ctx, role); err != nil {
			return nil, err
		}
	}
	return role, nil
}

func (d *controller) populatePermissions(ctx context.Context, r *Role) error {
	if r == nil {
		return nil
	}
	rolePermissions, err := d.rbacMgr.GetPermissionsByRole(ctx, ROLETYPE, r.ID)
	log.Debug("*** get permissions by role returned : " + r.Name + " - " + strconv.Itoa((len(rolePermissions))))

	if err != nil {
		log.Errorf("failed to get permissions of role %d: %v", r.ID, err)
		return err
	}
	if len(rolePermissions) == 0 {
		return nil
	}

	// scope: accesses
	accessMap := make(map[string][]*types.Policy)

	// group by scope
	for _, rp := range rolePermissions {
		_, exist := accessMap[rp.Scope]
		if !exist {
			accessMap[rp.Scope] = []*types.Policy{{
				Resource: types.Resource(rp.Resource),
				Action:   types.Action(rp.Action),
				Effect:   types.Effect(rp.Effect),
			}}
		} else {
			accesses := accessMap[rp.Scope]
			accesses = append(accesses, &types.Policy{
				Resource: types.Resource(rp.Resource),
				Action:   types.Action(rp.Action),
				Effect:   types.Effect(rp.Effect),
			})
			accessMap[rp.Scope] = accesses
		}
	}

	var permissions []*Permission
	for scope, accesses := range accessMap {
		p := &Permission{}
		p.Scope = scope
		//TODO MGS this will not show a lot of permissions, as the role permissions are not defined
		p.Kind = LEVELROLE
		p.Namespace = "*"
		p.Access = accesses
		permissions = append(permissions, p)
	}
	r.Permissions = permissions
	return nil
}

// convertScope converts the db scope into role model
// /system    =>  Kind: system  Namespace: /
// /project/* =>  Kind: project Namespace: *
// /project/1 =>  Kind: project Namespace: library
/*
func (d *controller) convertScope(ctx context.Context, scope string) (kind, namespace string, err error) {
	if scope == "" {
		return
	}
	if scope == SCOPEALLPROJECT {
		kind = LEVELPROJECT
		namespace = "*"
	} else {
		kind = LEVELPROJECT
		ns, ok := rbac_project.NamespaceParse(types.Resource(scope))
		if !ok {
			log.Debugf("got no namespace from the resource %s", scope)
			return "", "", errors.Errorf("got no namespace from the resource %s", scope)
		}
		pro, err := d.proMgr.Get(ctx, ns.Identity())
		if err != nil {
			return "", "", err
		}
		namespace = pro.Name
	}
	return
}

// toScope ...
func (d *controller) toScope(ctx context.Context, p *Permission) (string, error) {
	switch p.Kind {
	case LEVELSYSTEM:
		if p.Namespace != "/" {
			return "", errors.New(nil).WithMessage("unknown namespace").WithCode(errors.BadRequestCode)
		}
		return SCOPESYSTEM, nil
	case LEVELPROJECT:
		if p.Namespace == "*" {
			return SCOPEALLPROJECT, nil
		}
		pro, err := d.proMgr.Get(ctx, p.Namespace)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("/project/%d", pro.ProjectID), nil
	}
	return "", errors.New(nil).WithMessage("unknown role kind").WithCode(errors.BadRequestCode)
}
*/
/*
// set the project info if it's a project level role

	func SetProject(ctx context.Context, r *Role) error {
		if r == nil {
			return nil
		}
		if r.Level == LEVELPROJECT {
			pro, err := project.New().Get(ctx, r.Permissions[0].Namespace)
			if err != nil {
				return err
			}
			r.ProjectName = pro.Name
			r.ProjectID = pro.ProjectID
		}
		return nil
	}

	func CreateSec(salt ...string) (string, string, string, error) {
		var secret, pwd string
		options := []retry.Option{
			retry.InitialInterval(time.Millisecond * 500),
			retry.MaxInterval(time.Second * 10),
			retry.Timeout(time.Minute),
			retry.Callback(func(err error, sleep time.Duration) {
				log.Debugf("failed to generate secret for role, retry after %s : %v", sleep, err)
			}),
		}

		if err := retry.Retry(func() error {
			pwd = utils.GenerateRandomString()
			if !IsValidSec(pwd) {
				return errors.New(nil).WithMessage("invalid secret format")
			}
			return nil
		}, options...); err != nil {
			return "", "", "", errors.Wrap(err, "failed to generate an valid random secret for role in one minute, please try again")
		}

		var saltTmp string
		if len(salt) != 0 {
			saltTmp = salt[0]
		} else {
			saltTmp = utils.GenerateRandomString()
		}
		secret = utils.Encrypt(pwd, saltTmp, utils.SHA256)
		return secret, pwd, saltTmp, nil
	}
*/
var (
	hasLower  = regexp.MustCompile(`[a-z]`)
	hasUpper  = regexp.MustCompile(`[A-Z]`)
	hasNumber = regexp.MustCompile(`\d`)
)

/*
func IsValidSec(secret string) bool {
	return len(secret) >= 8 && len(secret) <= 128 && hasLower.MatchString(secret) && hasUpper.MatchString(secret) && hasNumber.MatchString(secret)
}
*/
