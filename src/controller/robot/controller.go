package robot

import (
	"context"
	"fmt"
	rbac_common "github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/rbac"
	rbac_model "github.com/goharbor/harbor/src/pkg/rbac/model"
	robot "github.com/goharbor/harbor/src/pkg/robot2"
	"github.com/goharbor/harbor/src/pkg/robot2/model"
	"time"
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
	Create(ctx context.Context, r *Robot) (int64, error)

	// Delete ...
	Delete(ctx context.Context, id int64) error

	// Update ...
	Update(ctx context.Context, r *Robot) error

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
		proMgr:   project.Mgr,
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
func (d *controller) Create(ctx context.Context, r *Robot) (int64, error) {
	if err := d.setProjectID(ctx, r); err != nil {
		return 0, err
	}

	if r.ExpiresAt == 0 {
		tokenDuration := time.Duration(config.RobotTokenDuration()) * time.Minute
		r.ExpiresAt = time.Now().UTC().Add(tokenDuration).Unix()
	}

	key, err := config.SecretKey()
	if err != nil {
		return 0, err
	}
	str := utils.GenerateRandomString()
	secret, err := utils.ReversibleEncrypt(str, key)
	if err != nil {
		return 0, err
	}

	robotID, err := d.robotMgr.Create(ctx, &model.Robot{
		Name:        r.Name,
		Description: r.Description,
		ProjectID:   r.ProjectID,
		ExpiresAt:   r.ExpiresAt,
		Secret:      secret,
	})
	if err != nil {
		return 0, err
	}
	r.ID = robotID
	if err := d.createPermission(ctx, r); err != nil {
		return 0, err
	}
	return robotID, nil
}

// Delete ...
func (d *controller) Delete(ctx context.Context, id int64) error {
	if err := d.robotMgr.Delete(ctx, id); err != nil {
		return err
	}
	if err := d.rbacMgr.DeletePermissionsByRole(ctx, ROBOTTYPE, id); err != nil {
		return err
	}
	return nil
}

// Update ...
func (d *controller) Update(ctx context.Context, r *Robot) error {
	if r == nil {
		return errors.New("cannot update a nil robot").WithCode(errors.BadRequestCode)
	}
	if err := d.robotMgr.Update(ctx, &r.Robot); err != nil {
		return err
	}
	if err := d.setProjectID(ctx, r); err != nil {
		return err
	}
	// update the permission
	if err := d.rbacMgr.DeletePermissionsByRole(ctx, ROBOTTYPE, r.ID); err != nil {
		return err
	}
	if err := d.createPermission(ctx, r); err != nil {
		return err
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
	robot.Name = fmt.Sprintf("%s%s", config.RobotPrefix(), r.Name)
	robot.setLevel()
	if option == nil {
		return robot, nil
	}
	if option.WithPermission {
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
		p := &Permission{}
		kind, namespace, err := d.convertScope(ctx, scope)
		if err != nil {
			log.Errorf("failed to decode scope of robot %d: %v", r.ID, err)
			return err
		}
		p.Scope = scope
		p.Kind = kind
		p.Namespace = namespace
		p.Access = accesses
		permissions = append(permissions, p)
	}
	r.Permissions = permissions
	return nil
}

func (d *controller) setProjectID(ctx context.Context, r *Robot) error {
	if r == nil {
		return nil
	}
	var projectID int64
	switch r.Level {
	case LEVELSYSTEM:
		projectID = 0
	case LEVELPROJECT:
		pro, err := d.proMgr.Get(ctx, r.Permissions[0].Namespace)
		if err != nil {
			return err
		}
		projectID = pro.ProjectID
	default:
		return errors.New(nil).WithMessage("unknown robot account level").WithCode(errors.BadRequestCode)
	}
	r.ProjectID = projectID
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
		ns, ok := rbac_common.ProjectNamespaceParse(types.Resource(scope))
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
