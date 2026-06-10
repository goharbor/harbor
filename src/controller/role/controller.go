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
	"sync"
	"sync/atomic"
	"time"

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

const (
	// roleVersionCacheKey is the Redis key holding a global "version token" that
	// changes on every in-app role write. Nodes poll it (throttled) to detect
	// changes made on other nodes and invalidate their L1 cache accordingly.
	roleVersionCacheKey = "role:version"
	// roleCacheL1TTL / roleCacheL2TTL bound how long a stale entry can survive an
	// out-of-band change (e.g. a direct SQL UPDATE to role_permission that bypasses
	// the controller and therefore does not bump the version token).
	roleCacheL1TTL = 30 * time.Minute
	roleCacheL2TTL = 30 * time.Minute
	// roleVersionCheckInterval throttles how often a node re-reads the version
	// token from Redis, bounding cross-node propagation of in-app changes while
	// keeping the hot Get() path at L1's zero-network-hop cost.
	roleVersionCheckInterval = time.Second
)

// roleEntry is the L1 (process-local) cache value. It carries the local
// generation it was cached under (invalidated when the global version token
// changes) and an absolute expiry (the out-of-band TTL backstop).
type roleEntry struct {
	role      *Role
	gen       int64
	expiresAt time.Time
}

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
	localMap sync.Map // L1: id -> *roleEntry, zero-alloc reads
	warmOnce sync.Once

	// Throttle state for the cross-node version check (lock-free).
	versionCheckedAt atomic.Int64           // unixnano of the last version fetch
	lastVersion      atomic.Pointer[string] // last-seen global version token
	localGen         atomic.Int64           // bumped whenever the token changes
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

// currentGen returns this node's local cache generation, refreshing it from the
// global version token at most once per roleVersionCheckInterval. A change in the
// token (written by an in-app role write on any node) bumps the local generation,
// which invalidates every L1 entry cached under the previous generation. The
// throttle keeps the hot Get() path at L1's zero-network-hop cost.
func (d *controller) currentGen(ctx context.Context) int64 {
	now := time.Now().UnixNano()
	last := d.versionCheckedAt.Load()
	if now-last > int64(roleVersionCheckInterval) && d.versionCheckedAt.CompareAndSwap(last, now) {
		if c := d.roleCache(); c != nil {
			var token string
			if err := c.Fetch(ctx, roleVersionCacheKey, &token); err == nil {
				if p := d.lastVersion.Load(); p == nil || *p != token {
					d.localGen.Add(1)
					d.lastVersion.Store(&token)
				}
			}
		}
	}
	return d.localGen.Load()
}

// bumpVersion writes a fresh global version token so other nodes invalidate their
// L1 caches on their next throttled check, and advances this node's own generation
// immediately (so the writer does not wait out its own throttle window).
func (d *controller) bumpVersion(ctx context.Context) {
	token := strconv.FormatInt(time.Now().UnixNano(), 10)
	if c := d.roleCache(); c != nil {
		_ = c.Save(ctx, roleVersionCacheKey, token) // no TTL: the version key persists
	}
	d.lastVersion.Store(&token)
	d.localGen.Add(1)
}

// storeL1 caches a role in the process-local L1, stamped with the generation it
// was cached under and an absolute expiry (the out-of-band TTL backstop).
func (d *controller) storeL1(id int64, r *Role, gen int64) {
	d.localMap.Store(id, &roleEntry{role: r, gen: gen, expiresAt: time.Now().Add(roleCacheL1TTL)})
}

// warmCache loads all roles with their permissions into L1 and L2.
// Fires once in background on first Get().
func (d *controller) warmCache(ctx context.Context) {
	roles, err := d.roleMgr.List(ctx, &q.Query{PageSize: -1})
	if err != nil {
		log.Warningf("failed to warm role permission cache: %v", err)
		return
	}
	gen := d.currentGen(ctx)
	for _, r := range roles {
		populated, err := d.populate(ctx, r, &Option{WithPermission: true})
		if err != nil || populated == nil {
			continue
		}
		if c := d.roleCache(); c != nil {
			_ = c.Save(ctx, roleCacheKey(r.ID), populated, roleCacheL2TTL) // L2
		}
		d.storeL1(r.ID, populated, gen) // L1
	}
	log.Debugf("role permission cache warmed with %d roles", len(roles))
}

// invalidateRole is called after an in-app Create/Update/Delete (which have
// already committed to the DB). Order matters:
//  1. bump the global version token so other nodes drop their L1 entries on their
//     next throttled check (<= roleVersionCheckInterval);
//  2. delete the changed role's L2 key so the next reader re-reads the DB for the
//     role whose data actually changed (other roles' L2 entries survive and
//     re-promote cheaply once their L1 is gen-invalidated);
//  3. drop this node's local L1 entry.
//
// Residual race (bounded, documented): a reader that read the DB *before* this
// write commits can repopulate the role's L2 key *after* step 2, re-caching stale
// data. That window is bounded by the L2 TTL (roleCacheL2TTL) and cleared by the
// next write. Closing it fully would require versioned L2 payloads or pub/sub.
func (d *controller) invalidateRole(ctx context.Context, id int64) {
	d.bumpVersion(ctx)
	if c := d.roleCache(); c != nil {
		_ = c.Delete(ctx, roleCacheKey(id))
	}
	d.localMap.Delete(id)
}

// Get returns a role (optionally with permissions). When permissions are
// requested it is served from a two-level cache. Read order: refresh the
// cross-node version (throttled) -> L1 -> L2 -> DB.
func (d *controller) Get(ctx context.Context, id int64, option *Option) (*Role, error) {
	if option != nil && option.WithPermission {
		// Trigger background warm on first call — does not block.
		d.warmOnce.Do(func() { go d.warmCache(orm.Context()) })

		// Refresh the cross-node version first so L1 validity is judged against
		// the latest token (throttled to one Redis read per interval at most).
		gen := d.currentGen(ctx)

		// L1: process-local — valid only if cached under the current generation
		// and not past its TTL.
		if v, ok := d.localMap.Load(id); ok {
			e := v.(*roleEntry)
			if e.gen == gen && time.Now().Before(e.expiresAt) {
				return e.role, nil
			}
		}

		// L2: Redis (honours its own TTL — expired keys are already gone).
		if c := d.roleCache(); c != nil {
			var cached Role
			if err := c.Fetch(ctx, roleCacheKey(id), &cached); err == nil {
				d.storeL1(id, &cached, gen) // promote to L1
				return &cached, nil
			}
		}

		// DB: source of truth. Populate both layers.
		r, err := d.roleMgr.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		populated, err := d.populate(ctx, r, option)
		if err != nil {
			return nil, err
		}
		if populated != nil {
			if c := d.roleCache(); c != nil {
				_ = c.Save(ctx, roleCacheKey(id), populated, roleCacheL2TTL)
			}
			d.storeL1(id, populated, gen)
		}
		return populated, nil
	}

	// Without permissions there is nothing cacheable — read straight through.
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
