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
	"sync"

	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
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
	roleMgr  role.Manager
	proMgr   project.Manager
	rbacMgr  rbac.Manager
	localMap sync.Map  // L1: process-local pointer cache, zero-alloc reads
	warmOnce sync.Once
}

// NewController ...
func NewController() Controller {
	return &controller{
		roleMgr: role.Mgr,
		proMgr:  pkg.ProjectMgr,
		rbacMgr: rbac.Mgr,
	}
}

// roleCache returns the shared Redis-backed cache (same instance used by the quota
// controller). Falls back to nil if the cache has not been initialized yet.
func (d *controller) roleCache() cache.Cache {
	return cache.Default()
}

func roleCacheKey(id int64) string {
	return fmt.Sprintf("role:%d", id)
}

// warmCache loads all roles with their permissions into L1 (sync.Map) and L2 (Redis).
// Fires once in background on first Get(); also called after every write.
func (d *controller) warmCache(ctx context.Context) {
	roles, err := d.roleMgr.List(ctx, &q.Query{PageSize: -1})
	if err != nil {
		log.Warningf("failed to warm role permission cache: %v", err)
		return
	}
	for _, r := range roles {
		populated, err := d.populate(ctx, r, &Option{WithPermission: true})
		if err != nil || populated == nil {
			continue
		}
		d.localMap.Store(r.ID, populated)                              // L1
		if c := d.roleCache(); c != nil {
			_ = c.Save(ctx, roleCacheKey(r.ID), populated)            // L2
		}
	}
	log.Debugf("role permission cache warmed with %d roles", len(roles))
}

// invalidateRole removes a role from both cache layers so the next Get() re-fetches
// from DB and repopulates. Called after Create/Update/Delete.
func (d *controller) invalidateRole(ctx context.Context, id int64) {
	d.localMap.Delete(id)
	if c := d.roleCache(); c != nil {
		_ = c.Delete(ctx, roleCacheKey(id))
	}
}

// Get ...
func (d *controller) Get(ctx context.Context, id int64, option *Option) (*Role, error) {
	if option != nil && option.WithPermission {
		// Trigger background warm on first call — does not block.
		d.warmOnce.Do(func() { go d.warmCache(orm.Context()) })

		// L1: process-local pointer — O(1), zero alloc, zero serialization.
		if v, ok := d.localMap.Load(id); ok {
			return v.(*Role), nil
		}

		// L2: Redis — shared across nodes, handles multi-node invalidation.
		if c := d.roleCache(); c != nil {
			var cached Role
			if err := c.Fetch(ctx, roleCacheKey(id), &cached); err == nil {
				d.localMap.Store(id, &cached) // promote to L1
				return &cached, nil
			}
		}
	}
	r, err := d.roleMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return d.populate(ctx, r, option)
}

// Count ...
func (d *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return d.roleMgr.Count(ctx, query)
}

// Create ...
func (d *controller) Create(ctx context.Context, r *Role) (int64, error) {

	name := r.Name

	rCreate := &model.Role{
		Name:        name,
		RoleMask:    r.RoleMask,
		RoleCode:    r.RoleCode,
		Description: r.Description,
		CreatedBy:   r.CreatedBy,
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
	d.invalidateRole(ctx, roleID)
	return roleID, nil
}

// Delete ...
func (d *controller) Delete(ctx context.Context, id int64, option ...*Option) error {
	rDelete, err := d.roleMgr.Get(ctx, id)
	if err != nil {
		return err
	}
	if rDelete.IsBuiltin {
		return errors.ForbiddenError(nil).WithMessagef("cannot delete built-in role %d", id)
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
	d.invalidateRole(ctx, id)
	return nil
}

// Update ...
func (d *controller) Update(ctx context.Context, r *Role, option *Option) error {
	if r == nil {
		return errors.New("cannot update a nil role").WithCode(errors.BadRequestCode)
	}
	existing, err := d.roleMgr.Get(ctx, r.ID)
	if err != nil {
		return err
	}
	if existing.IsBuiltin {
		return errors.ForbiddenError(nil).WithMessagef("cannot modify built-in role %d", r.ID)
	}
	// update role record fields
	if err := d.roleMgr.Update(ctx, &model.Role{
		ID:          r.ID,
		Description: r.Description,
	}, "description"); err != nil {
		return err
	}
	// update the permission
	if option != nil && option.WithPermission {
		if err := d.rbacMgr.DeletePermissionsByRole(ctx, ROLETYPE, r.ID); err != nil && !errors.IsNotFoundErr(err) {
			return err
		}
		if err := d.createPermission(ctx, r); err != nil {
			return err
		}
	}
	// fire event
	notification.AddEvent(ctx, &metadata.UpdateRoleEventMetadata{
		Ctx:  ctx,
		Role: existing,
	})
	d.invalidateRole(ctx, r.ID)
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
