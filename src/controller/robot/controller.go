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

package robot

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/rbac"
	rbac_model "github.com/goharbor/harbor/src/pkg/rbac/model"
	robot "github.com/goharbor/harbor/src/pkg/robot"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

var (
	// Ctl is a global variable for the default robot account controller implementation
	Ctl = NewController()
)

// Controller to handle the requests related with robot account
type Controller interface {
	// Get ...
	Get(ctx context.Context, id int64, option *Option) (*Robot, error)

	// Count returns the total count of robots according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// Create ...
	Create(ctx context.Context, r *Robot) (int64, string, error)

	// Delete ...
	Delete(ctx context.Context, id int64, option ...*Option) error

	// Update ...
	Update(ctx context.Context, r *Robot, option *Option) error

	// List ...
	List(ctx context.Context, query *q.Query, option *Option) ([]*Robot, error)
}

// controller ...
type controller struct {
	robotMgr robot.Manager
	proMgr   project.Manager
	rbacMgr  rbac.Manager
}

// NewController ...
func NewController() Controller {
	return &controller{
		robotMgr: robot.Mgr,
		proMgr:   pkg.ProjectMgr,
		rbacMgr:  rbac.Mgr,
	}
}

// Get ...
func (d *controller) Get(ctx context.Context, id int64, option *Option) (*Robot, error) {
	robot, err := d.robotMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return d.populate(ctx, robot, option)
}

// Count ...
func (d *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return d.robotMgr.Count(ctx, query)
}

// Create ...
func (d *controller) Create(ctx context.Context, r *Robot) (int64, string, error) {
	var expiresAt int64
	if r.Duration == -1 {
		expiresAt = -1
	} else {
		durationStr := strconv.FormatInt(r.Duration, 10)
		duration, err := strconv.Atoi(durationStr)
		if err != nil {
			return 0, "", err
		}
		expiresAt = time.Now().AddDate(0, 0, duration).Unix()
	}

	secret, pwd, salt, err := CreateSec()
	if err != nil {
		return 0, "", err
	}

	name := r.Name
	// for the project level robot, set the name pattern as projectname+robotname, and + is a illegal character.
	if r.Level == LEVELPROJECT {
		name = fmt.Sprintf("%s+%s", r.ProjectName, r.Name)
	}

	rCreate := &model.Robot{
		Name:        name,
		Description: r.Description,
		ProjectID:   r.ProjectID,
		ExpiresAt:   expiresAt,
		Secret:      secret,
		Duration:    r.Duration,
		Salt:        salt,
		Visible:     r.Visible,
		CreatorRef:  r.CreatorRef,
		CreatorType: r.CreatorType,
	}
	robotID, err := d.robotMgr.Create(ctx, rCreate)
	if err != nil {
		return 0, "", err
	}
	r.ID = robotID
	if err := d.createPermission(ctx, r); err != nil {
		return 0, "", err
	}
	// fire event
	notification.AddEvent(ctx, &metadata.CreateRobotEventMetadata{
		Ctx:   ctx,
		Robot: rCreate,
	})
	return robotID, pwd, nil
}

// Delete ...
func (d *controller) Delete(ctx context.Context, id int64, option ...*Option) error {
	rDelete, err := d.robotMgr.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := d.robotMgr.Delete(ctx, id); err != nil {
		return err
	}
	if err := d.rbacMgr.DeletePermissionsByRole(ctx, ROBOTTYPE, id); err != nil {
		return err
	}
	// fire event
	deleteMetadata := &metadata.DeleteRobotEventMetadata{
		Ctx:   ctx,
		Robot: rDelete,
	}
	if len(option) != 0 && option[0].Operator != "" {
		deleteMetadata.Operator = option[0].Operator
	}
	notification.AddEvent(ctx, deleteMetadata)
	return nil
}

// Update ...
func (d *controller) Update(ctx context.Context, r *Robot, option *Option) error {
	if r == nil {
		return errors.New("cannot update a nil robot").WithCode(errors.BadRequestCode)
	}
	if err := d.robotMgr.Update(ctx, &r.Robot, "secret", "description", "disabled", "duration", "expiresat"); err != nil {
		return err
	}
	// update the permission
	if option != nil && option.WithPermission {
		if err := d.rbacMgr.DeletePermissionsByRole(ctx, ROBOTTYPE, r.ID); err != nil && !errors.IsNotFoundErr(err) {
			return err
		}
		if err := d.createPermission(ctx, r); err != nil {
			return err
		}
	}
	return nil
}

// List ...
func (d *controller) List(ctx context.Context, query *q.Query, option *Option) ([]*Robot, error) {
	robots, err := d.robotMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var robotAccounts []*Robot
	for _, r := range robots {
		rb, err := d.populate(ctx, r, option)
		if err != nil {
			return nil, err
		}
		robotAccounts = append(robotAccounts, rb)
	}
	return robotAccounts, nil
}

func (d *controller) createPermission(ctx context.Context, r *Robot) error {
	if r == nil {
		return nil
	}

	for _, per := range r.Permissions {
		policy := &rbac_model.PermissionPolicy{}
		scope, err := d.toScope(ctx, per)
		if err != nil {
			return err
		}
		policy.Scope = scope

		for _, access := range per.Access {
			policy.Resource = access.Resource.String()
			policy.Action = access.Action.String()
			policy.Effect = access.Effect.String()

			policyID, err := d.rbacMgr.CreateRbacPolicy(ctx, policy)
			if err != nil {
				return err
			}

			_, err = d.rbacMgr.CreatePermission(ctx, &rbac_model.RolePermission{
				RoleType:           ROBOTTYPE,
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

func (d *controller) populate(ctx context.Context, r *model.Robot, option *Option) (*Robot, error) {
	if r == nil {
		return nil, nil
	}
	robot := &Robot{
		Robot: *r,
	}
	robot.setLevel()
	robot.setEditable()
	// for the v2 robots, add prefix to the robot name
	// for the v1 legacy robots, keep the robot name
	if robot.Editable {
		robot.Name = fmt.Sprintf("%s%s", config.RobotPrefix(ctx), r.Name)
	} else {
		robot.Name = r.Name
	}
	if option != nil && option.WithPermission {
		if err := d.populatePermissions(ctx, robot); err != nil {
			return nil, err
		}
	}
	return robot, nil
}

func (d *controller) populatePermissions(ctx context.Context, r *Robot) error {
	if r == nil {
		return nil
	}
	rolePermissions, err := d.rbacMgr.GetPermissionsByRole(ctx, ROBOTTYPE, r.ID)
	if err != nil {
		log.Errorf("failed to get permissions of robot %d: %v", r.ID, err)
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
		kind, namespace, err := d.convertScope(ctx, scope)
		if err != nil {
			// if the project is removed, do not list it in the robot account permission.
			if errors.IsNotFoundErr(err) {
				continue
			}
			log.Errorf("failed to decode scope of robot %d: %v", r.ID, err)
			return err
		}
		p := &Permission{}
		p.Scope = scope
		p.Kind = kind
		p.Namespace = namespace
		p.Access = accesses
		permissions = append(permissions, p)
	}
	r.Permissions = permissions
	return nil
}

// convertScope converts the db scope into robot model
// /system    =>  Kind: system  Namespace: /
// /project/* =>  Kind: project Namespace: *
// /project/1 =>  Kind: project Namespace: library
func (d *controller) convertScope(ctx context.Context, scope string) (kind, namespace string, err error) {
	if scope == "" {
		return
	}
	if scope == SCOPESYSTEM {
		kind = LEVELSYSTEM
		namespace = "/"
	} else if scope == SCOPEALLPROJECT {
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
	return "", errors.New(nil).WithMessage("unknown robot kind").WithCode(errors.BadRequestCode)
}

// set the project info if it's a project level robot
func SetProject(ctx context.Context, r *Robot) error {
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
			log.Debugf("failed to generate secret for robot, retry after %s : %v", sleep, err)
		}),
	}

	if err := retry.Retry(func() error {
		pwd = utils.GenerateRandomString()
		if !IsValidSec(pwd) {
			return errors.New(nil).WithMessage("invalid secret format")
		}
		return nil
	}, options...); err != nil {
		return "", "", "", errors.Wrap(err, "failed to generate an valid random secret for robot in one minute, please try again")
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

var (
	hasLower  = regexp.MustCompile(`[a-z]`)
	hasUpper  = regexp.MustCompile(`[A-Z]`)
	hasNumber = regexp.MustCompile(`\d`)
)

func IsValidSec(secret string) bool {
	return len(secret) >= 8 && len(secret) <= 128 && hasLower.MatchString(secret) && hasUpper.MatchString(secret) && hasNumber.MatchString(secret)
}
