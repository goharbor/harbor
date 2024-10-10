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

package handler

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security/local"
	robotSc "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	pkg "github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/robot"
)

func newRobotAPI() *robotAPI {
	return &robotAPI{
		robotCtl: robot.Ctl,
	}
}

type robotAPI struct {
	BaseAPI
	robotCtl robot.Controller
}

func (rAPI *robotAPI) CreateRobot(ctx context.Context, params operation.CreateRobotParams) middleware.Responder {
	if err := validateName(params.Robot.Name); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.validate(params.Robot.Duration, params.Robot.Level, params.Robot.Permissions); err != nil {
		return rAPI.SendError(ctx, err)
	}

	sc, err := rAPI.GetSecurityContext(ctx)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	r := &robot.Robot{
		Robot: pkg.Robot{
			Name:        params.Robot.Name,
			Description: params.Robot.Description,
			Duration:    params.Robot.Duration,
			Visible:     true,
		},
		Level:           params.Robot.Level,
		ProjectNameOrID: params.Robot.Permissions[0].Namespace,
	}

	if err := rAPI.requireAccess(ctx, r, rbac.ActionCreate); err != nil {
		return rAPI.SendError(ctx, err)
	}

	var creatorRef int64
	switch s := sc.(type) {
	case *local.SecurityContext:
		creatorRef = int64(s.User().UserID)
	case *robotSc.SecurityContext:
		creatorRef = s.User().ID
	default:
		return rAPI.SendError(ctx, errors.New(nil).WithMessage("invalid security context"))
	}
	r.CreatorType = sc.Name()
	r.CreatorRef = creatorRef

	if err := lib.JSONCopy(&r.Permissions, params.Robot.Permissions); err != nil {
		log.Warningf("failed to call JSONCopy on robot permission when CreateRobot, error: %v", err)
	}

	if err := robot.SetProject(ctx, r); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if _, ok := sc.(*robotSc.SecurityContext); ok {
		creatorRobots, err := rAPI.robotCtl.List(ctx, q.New(q.KeyWords{
			"name":       strings.TrimPrefix(sc.GetUsername(), config.RobotPrefix(ctx)),
			"project_id": r.ProjectID,
		}), &robot.Option{
			WithPermission: true,
		})
		if err != nil {
			return rAPI.SendError(ctx, err)
		}
		if len(creatorRobots) == 0 {
			return rAPI.SendError(ctx, errors.DeniedError(nil))
		}

		if !isValidPermissionScope(params.Robot.Permissions, creatorRobots[0].Permissions) {
			return rAPI.SendError(ctx, errors.New(nil).WithMessagef("permission scope is invalid. It must be equal to or more restrictive than the creator robot's permissions: %s", creatorRobots[0].Name).WithCode(errors.DENIED))
		}
	}

	rid, pwd, err := rAPI.robotCtl.Create(ctx, r)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	created, err := rAPI.robotCtl.Get(ctx, rid, nil)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), created.ID)
	return operation.NewCreateRobotCreated().WithLocation(location).WithPayload(&models.RobotCreated{
		ID:           created.ID,
		Name:         created.Name,
		Secret:       pwd,
		CreationTime: strfmt.DateTime(created.CreationTime),
		ExpiresAt:    created.ExpiresAt,
	})
}

func (rAPI *robotAPI) DeleteRobot(ctx context.Context, params operation.DeleteRobotParams) middleware.Responder {
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	r, err := rAPI.robotCtl.Get(ctx, params.RobotID, nil)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.requireAccess(ctx, r, rbac.ActionDelete); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.robotCtl.Delete(ctx, params.RobotID); err != nil {
		// for the version 1 robot account, has to ignore the no permission error.
		if !r.Editable && errors.IsNotFoundErr(err) {
			return operation.NewDeleteRobotOK()
		}
		return rAPI.SendError(ctx, err)
	}
	return operation.NewDeleteRobotOK()
}

func (rAPI *robotAPI) ListRobot(ctx context.Context, params operation.ListRobotParams) middleware.Responder {
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	query, err := rAPI.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	var projectID int64
	var level string
	// GET /api/v2.0/robots or GET /api/v2.0/robots?level=system to get all of system level robots.
	// GET /api/v2.0/robots?level=project&project_id=1
	if _, ok := query.Keywords["Level"]; ok {
		if !isValidLevel(query.Keywords["Level"].(string)) {
			return rAPI.SendError(ctx, errors.New(nil).WithMessage("bad request error level input").WithCode(errors.BadRequestCode))
		}
		level = query.Keywords["Level"].(string)
		if level == robot.LEVELPROJECT {
			if _, ok := query.Keywords["ProjectID"]; !ok {
				return rAPI.SendError(ctx, errors.BadRequestError(nil).WithMessage("must with project ID when to query project robots"))
			}
			pid, err := strconv.ParseInt(query.Keywords["ProjectID"].(string), 10, 64)
			if err != nil {
				return rAPI.SendError(ctx, errors.BadRequestError(nil).WithMessage("Project ID must be int type."))
			}
			projectID = pid
		}
	} else {
		level = robot.LEVELSYSTEM
		query.Keywords["ProjectID"] = 0
	}
	query.Keywords["Visible"] = true

	r := &robot.Robot{
		ProjectNameOrID: projectID,
		Level:           level,
	}
	if err := rAPI.requireAccess(ctx, r, rbac.ActionList); err != nil {
		return rAPI.SendError(ctx, err)
	}

	total, err := rAPI.robotCtl.Count(ctx, query)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	robots, err := rAPI.robotCtl.List(ctx, query, &robot.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	var results []*models.Robot
	for _, r := range robots {
		results = append(results, model.NewRobot(r).ToSwagger())
	}

	return operation.NewListRobotOK().
		WithXTotalCount(total).
		WithLink(rAPI.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (rAPI *robotAPI) GetRobotByID(ctx context.Context, params operation.GetRobotByIDParams) middleware.Responder {
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	r, err := rAPI.robotCtl.Get(ctx, params.RobotID, &robot.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}
	if err := rAPI.requireAccess(ctx, r, rbac.ActionRead); err != nil {
		return rAPI.SendError(ctx, err)
	}

	return operation.NewGetRobotByIDOK().WithPayload(model.NewRobot(r).ToSwagger())
}

func (rAPI *robotAPI) UpdateRobot(ctx context.Context, params operation.UpdateRobotParams) middleware.Responder {
	var err error
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}
	r, err := rAPI.robotCtl.Get(ctx, params.RobotID, &robot.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	if !r.Editable {
		err = errors.DeniedError(nil).WithMessage("editing of legacy robot is not allowed")
	} else {
		err = rAPI.updateV2Robot(ctx, params, r)
	}
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	return operation.NewUpdateRobotOK()
}

func (rAPI *robotAPI) RefreshSec(ctx context.Context, params operation.RefreshSecParams) middleware.Responder {
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	r, err := rAPI.robotCtl.Get(ctx, params.RobotID, nil)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.requireAccess(ctx, r, rbac.ActionUpdate); err != nil {
		return rAPI.SendError(ctx, err)
	}

	var secret string
	robotSec := &models.RobotSec{}
	if params.RobotSec.Secret != "" {
		if !robot.IsValidSec(params.RobotSec.Secret) {
			return rAPI.SendError(ctx, errors.New("the secret must be 8-128, inclusively, characters long with at least 1 uppercase letter, 1 lowercase letter and 1 number").WithCode(errors.BadRequestCode))
		}
		secret = utils.Encrypt(params.RobotSec.Secret, r.Salt, utils.SHA256)
		robotSec.Secret = ""
	} else {
		sec, pwd, _, err := robot.CreateSec(r.Salt)
		if err != nil {
			return rAPI.SendError(ctx, err)
		}
		secret = sec
		robotSec.Secret = pwd
	}

	r.Secret = secret
	if err := rAPI.robotCtl.Update(ctx, r, nil); err != nil {
		return rAPI.SendError(ctx, err)
	}

	return operation.NewRefreshSecOK().WithPayload(robotSec)
}

func (rAPI *robotAPI) requireAccess(ctx context.Context, r *robot.Robot, action rbac.Action) error {
	if r.Level == robot.LEVELSYSTEM {
		return rAPI.RequireSystemAccess(ctx, action, rbac.ResourceRobot)
	} else if r.Level == robot.LEVELPROJECT {
		var ns interface{}
		if r.ProjectNameOrID != nil {
			ns = r.ProjectNameOrID
		} else if r.ProjectID > 0 {
			ns = r.ProjectID
		} else if r.ProjectName != "" {
			ns = r.ProjectName
		}
		return rAPI.RequireProjectAccess(ctx, ns, action, rbac.ResourceRobot)
	}

	return errors.ForbiddenError(nil)
}

// more validation
func (rAPI *robotAPI) validate(d int64, level string, permissions []*models.RobotPermission) error {
	if !isValidDuration(d) {
		return errors.New(nil).WithMessagef("bad request error duration input: %d, duration must be either -1(Never) or a positive integer", d).WithCode(errors.BadRequestCode)
	}

	if !isValidLevel(level) {
		return errors.New(nil).WithMessagef("bad request error level input: %s", level).WithCode(errors.BadRequestCode)
	}

	if len(permissions) == 0 {
		return errors.New(nil).WithMessage("bad request empty permission").WithCode(errors.BadRequestCode)
	}

	for _, perm := range permissions {
		if len(perm.Access) == 0 {
			return errors.New(nil).WithMessage("bad request empty access").WithCode(errors.BadRequestCode)
		}
	}

	// to create a project robot, the permission must be only one project scope.
	if level == robot.LEVELPROJECT && len(permissions) > 1 {
		return errors.New(nil).WithMessage("bad request permission").WithCode(errors.BadRequestCode)
	}

	provider := rbac.GetPermissionProvider()
	// to validate the access scope
	for _, perm := range permissions {
		if perm.Kind == robot.LEVELSYSTEM {
			polices := provider.GetPermissions(rbac.ScopeSystem)
			for _, acc := range perm.Access {
				if !containsAccess(polices, acc) {
					return errors.New(nil).WithMessagef("bad request permission: %s:%s", acc.Resource, acc.Action).WithCode(errors.BadRequestCode)
				}
			}
		} else if perm.Kind == robot.LEVELPROJECT {
			polices := provider.GetPermissions(rbac.ScopeProject)
			for _, acc := range perm.Access {
				if !containsAccess(polices, acc) {
					return errors.New(nil).WithMessagef("bad request permission: %s:%s", acc.Resource, acc.Action).WithCode(errors.BadRequestCode)
				}
			}
		} else {
			return errors.New(nil).WithMessagef("bad request permission level: %s", perm.Kind).WithCode(errors.BadRequestCode)
		}
	}

	return nil
}

func (rAPI *robotAPI) updateV2Robot(ctx context.Context, params operation.UpdateRobotParams, r *robot.Robot) error {
	if params.Robot.Duration == nil {
		params.Robot.Duration = &r.Duration
	}
	if err := rAPI.validate(*params.Robot.Duration, params.Robot.Level, params.Robot.Permissions); err != nil {
		return err
	}
	if r.Level != robot.LEVELSYSTEM {
		projectID, err := getProjectID(ctx, params.Robot.Permissions[0].Namespace)
		if err != nil {
			return err
		}
		if r.ProjectID != projectID {
			return errors.BadRequestError(nil).WithMessage("cannot update the project id of robot")
		}
	}
	r.ProjectNameOrID = params.Robot.Permissions[0].Namespace
	if err := rAPI.requireAccess(ctx, r, rbac.ActionUpdate); err != nil {
		return err
	}
	if params.Robot.Level != r.Level || params.Robot.Name != r.Name {
		return errors.BadRequestError(nil).WithMessage("cannot update the level or name of robot")
	}

	if r.Duration != *params.Robot.Duration {
		r.Duration = *params.Robot.Duration
		if *params.Robot.Duration == -1 {
			r.ExpiresAt = -1
		} else {
			r.ExpiresAt = r.CreationTime.AddDate(0, 0, int(*params.Robot.Duration)).Unix()
		}
	}

	r.Description = params.Robot.Description
	r.Disabled = params.Robot.Disable
	if len(params.Robot.Permissions) != 0 {
		if err := lib.JSONCopy(&r.Permissions, params.Robot.Permissions); err != nil {
			log.Warningf("failed to call JSONCopy on robot permission when updateV2Robot, error: %v", err)
		}
	}

	if err := rAPI.robotCtl.Update(ctx, r, &robot.Option{
		WithPermission: true,
	}); err != nil {
		return err
	}
	return nil
}

func isValidLevel(l string) bool {
	return l == robot.LEVELSYSTEM || l == robot.LEVELPROJECT
}

func isValidDuration(d int64) bool {
	return d == -1 || (d > 0 && d < math.MaxInt32)
}

// validateName validates the robot name, especially '+' cannot be a valid character
func validateName(name string) error {
	robotNameReg := `^[a-z0-9]+(?:[._-][a-z0-9]+)*$`
	legal := regexp.MustCompile(robotNameReg).MatchString(name)
	if !legal {
		return errors.BadRequestError(nil).WithMessage("robot name is not in lower case or contains illegal characters")
	}
	return nil
}

func containsAccess(policies []*types.Policy, item *models.Access) bool {
	for _, po := range policies {
		if po.Resource.String() == item.Resource && po.Action.String() == item.Action {
			return true
		}
	}
	return false
}

// isValidPermissionScope checks if permission slice A is a subset of permission slice B
func isValidPermissionScope(creating []*models.RobotPermission, creator []*robot.Permission) bool {
	creatorMap := make(map[string]*robot.Permission)
	for _, creatorPerm := range creator {
		key := fmt.Sprintf("%s:%s", creatorPerm.Kind, creatorPerm.Namespace)
		creatorMap[key] = creatorPerm
	}

	hasLessThanOrEqualAccess := func(creating []*models.Access, creator []*types.Policy) bool {
		creatorMap := make(map[string]*types.Policy)
		for _, creatorP := range creator {
			key := fmt.Sprintf("%s:%s:%s", creatorP.Resource, creatorP.Action, creatorP.Effect)
			creatorMap[key] = creatorP
		}
		for _, creatingP := range creating {
			key := fmt.Sprintf("%s:%s:%s", creatingP.Resource, creatingP.Action, creatingP.Effect)
			if _, found := creatorMap[key]; !found {
				return false
			}
		}
		return true
	}

	for _, pCreating := range creating {
		key := fmt.Sprintf("%s:%s", pCreating.Kind, pCreating.Namespace)
		creatingPerm, found := creatorMap[key]
		if !found {
			return false
		}
		if !hasLessThanOrEqualAccess(pCreating.Access, creatingPerm.Access) {
			return false
		}
	}
	return true
}
