package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	pkg "github.com/goharbor/harbor/src/pkg/robot2/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/robot"
	"strconv"
	"strings"
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
	if err := rAPI.validate(params.Robot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.requireAccess(ctx, params.Robot.Level, params.Robot.Permissions[0].Namespace, rbac.ActionCreate); err != nil {
		return rAPI.SendError(ctx, err)
	}

	r := &robot.Robot{
		Robot: pkg.Robot{
			Name:        params.Robot.Name,
			Description: params.Robot.Description,
			ExpiresAt:   params.Robot.ExpiresAt,
		},
		Level: params.Robot.Level,
	}
	lib.JSONCopy(&r.Permissions, params.Robot.Permissions)

	rid, err := rAPI.robotCtl.Create(ctx, r)
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
		Secret:       created.Secret,
		CreationTime: strfmt.DateTime(created.CreationTime),
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

	if err := rAPI.requireAccess(ctx, r.Level, r.ProjectID, rbac.ActionDelete); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.robotCtl.Delete(ctx, params.RobotID); err != nil {
		return rAPI.SendError(ctx, err)
	}
	return operation.NewDeleteRobotOK()
}

func (rAPI *robotAPI) ListRobot(ctx context.Context, params operation.ListRobotParams) middleware.Responder {
	if err := rAPI.RequireAuthenticated(ctx); err != nil {
		return rAPI.SendError(ctx, err)
	}

	query, err := rAPI.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
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

	if err := rAPI.requireAccess(ctx, level, projectID, rbac.ActionList); err != nil {
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

	if err := rAPI.requireAccess(ctx, r.Level, r.ProjectID, rbac.ActionRead); err != nil {
		return rAPI.SendError(ctx, err)
	}

	return operation.NewGetRobotByIDOK().WithPayload(model.NewRobot(r).ToSwagger())
}

func (rAPI *robotAPI) UpdateRobot(ctx context.Context, params operation.UpdateRobotParams) middleware.Responder {
	if err := rAPI.validate(params.Robot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.requireAccess(ctx, params.Robot.Level, params.Robot.Permissions[0].Namespace, rbac.ActionUpdate); err != nil {
		return rAPI.SendError(ctx, err)
	}

	r, err := rAPI.robotCtl.Get(ctx, params.RobotID, &robot.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	if params.Robot.Level != r.Level || params.Robot.Name != r.Name {
		return rAPI.SendError(ctx, errors.BadRequestError(nil).WithMessage("cannot update the level or name of robot"))
	}

	// refresh secret only
	if params.Robot.Secret != r.Secret && params.Robot.Secret != "" {
		key, err := config.SecretKey()
		if err != nil {
			return rAPI.SendError(ctx, err)
		}
		secret, err := utils.ReversibleEncrypt(params.Robot.Secret, key)
		if err != nil {
			return rAPI.SendError(ctx, err)
		}
		r.Secret = secret
		if err := rAPI.robotCtl.Update(ctx, r); err != nil {
			return rAPI.SendError(ctx, err)
		}
	}

	r.Description = params.Robot.Description
	r.ExpiresAt = params.Robot.ExpiresAt
	r.Disabled = params.Robot.Disable
	if len(params.Robot.Permissions) != 0 {
		lib.JSONCopy(&r.Permissions, params.Robot.Permissions)
	}

	if err := rAPI.robotCtl.Update(ctx, r); err != nil {
		return rAPI.SendError(ctx, err)
	}

	return operation.NewUpdateRobotOK()
}

func (rAPI *robotAPI) requireAccess(ctx context.Context, level string, projectIDOrName interface{}, action rbac.Action) error {
	if level == robot.LEVELSYSTEM {
		return rAPI.RequireSysAdmin(ctx)
	} else if level == robot.LEVELPROJECT {
		return rAPI.RequireProjectAccess(ctx, projectIDOrName, action, rbac.ResourceRobot)
	}
	return errors.ForbiddenError(nil)
}

// more validation
func (rAPI *robotAPI) validate(r *models.RobotCreate) error {
	if !isValidLevel(r.Level) {
		return errors.New(nil).WithMessage("bad request error level input").WithCode(errors.BadRequestCode)
	}

	if len(r.Permissions) == 0 {
		return errors.New(nil).WithMessage("bad request empty permission").WithCode(errors.BadRequestCode)
	}

	// to create a project robot, the permission must be only one project scope.
	if r.Level == robot.LEVELPROJECT && len(r.Permissions) > 1 {
		return errors.New(nil).WithMessage("bad request permission").WithCode(errors.BadRequestCode)
	}
	return nil
}

func isValidLevel(l string) bool {
	return l == robot.LEVELSYSTEM || l == robot.LEVELPROJECT
}
