package handler

import (
	"context"
	"fmt"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"regexp"
	"strings"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	pkg_robot "github.com/goharbor/harbor/src/pkg/robot"
	pkg "github.com/goharbor/harbor/src/pkg/robot/model"
	handler_model "github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/robotv1"
)

func newRobotV1API() *robotV1API {
	return &robotV1API{
		robotCtl:   robot.Ctl,
		robotMgr:   pkg_robot.Mgr,
		projectCtr: project.Ctl,
	}
}

type robotV1API struct {
	BaseAPI
	robotCtl   robot.Controller
	robotMgr   pkg_robot.Manager
	projectCtr project.Controller
}

func (rAPI *robotV1API) CreateRobotV1(ctx context.Context, params operation.CreateRobotV1Params) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := rAPI.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceRobot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	if err := rAPI.validate(ctx, params); err != nil {
		return rAPI.SendError(ctx, err)
	}

	r := &robot.Robot{
		Robot: pkg.Robot{
			Name:        params.Robot.Name,
			Description: params.Robot.Description,
			ExpiresAt:   params.Robot.ExpiresAt,
			Visible:     true,
		},
		Level: robot.LEVELPROJECT,
	}

	projectName, ok := projectNameOrID.(string)
	if !ok {
		p, err := rAPI.projectCtr.Get(ctx, projectNameOrID, project.Metadata(false))
		if err != nil {
			log.Errorf("failed to get project %s: %v", projectNameOrID, err)
			return rAPI.SendError(ctx, err)
		}
		projectName = p.Name
	}

	permission := &robot.Permission{
		Kind:      "project",
		Namespace: projectName,
	}

	var policies []*types.Policy
	for _, acc := range params.Robot.Access {
		policy := &types.Policy{
			Action: types.Action(acc.Action),
			Effect: types.Effect(acc.Effect),
		}
		res, err := getRawResource(acc.Resource)
		if err != nil {
			return rAPI.SendError(ctx, err)
		}
		policy.Resource = types.Resource(res)
		policies = append(policies, policy)
	}
	permission.Access = policies
	r.Permissions = append(r.Permissions, permission)

	rid, pwd, err := rAPI.robotCtl.Create(ctx, r)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	created, err := rAPI.robotCtl.Get(ctx, rid, nil)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), created.ID)
	return operation.NewCreateRobotV1Created().WithLocation(location).WithPayload(&models.RobotCreated{
		ID:           created.ID,
		Name:         created.Name,
		Secret:       pwd,
		CreationTime: strfmt.DateTime(created.CreationTime),
	})
}

func (rAPI *robotV1API) DeleteRobotV1(ctx context.Context, params operation.DeleteRobotV1Params) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := rAPI.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete, rbac.ResourceRobot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	pro, err := rAPI.projectCtr.Get(ctx, projectNameOrID)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}
	r, err := rAPI.robotCtl.List(ctx, q.New(q.KeyWords{"ProjectID": pro.ProjectID, "ID": params.RobotID}), &robot.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}
	if len(r) == 0 {
		return rAPI.SendError(ctx, errors.NotFoundError(fmt.Errorf("cannot find robot with project id: %d and id: %d", pro.ProjectID, params.RobotID)))
	}

	// ignore the not permissions error.
	if err := rAPI.robotCtl.Delete(ctx, params.RobotID); err != nil && !errors.IsNotFoundErr(err) {
		return rAPI.SendError(ctx, err)
	}
	return operation.NewDeleteRobotV1OK()
}

func (rAPI *robotV1API) ListRobotV1(ctx context.Context, params operation.ListRobotV1Params) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := rAPI.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceRobot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	query, err := rAPI.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}
	query.Keywords["Visible"] = true

	pro, err := rAPI.projectCtr.Get(ctx, projectNameOrID)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	query.Keywords["ProjectID"] = pro.ProjectID

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
		results = append(results, handler_model.NewRobot(r).ToSwagger())
	}

	return operation.NewListRobotV1OK().
		WithXTotalCount(total).
		WithLink(rAPI.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (rAPI *robotV1API) GetRobotByIDV1(ctx context.Context, params operation.GetRobotByIDV1Params) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := rAPI.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceRobot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	pro, err := rAPI.projectCtr.Get(ctx, projectNameOrID)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}

	r, err := rAPI.robotCtl.List(ctx, q.New(q.KeyWords{"ProjectID": pro.ProjectID, "ID": params.RobotID}), &robot.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}
	if len(r) == 0 {
		return rAPI.SendError(ctx, errors.NotFoundError(fmt.Errorf("cannot find robot with project id: %d and id: %d", pro.ProjectID, params.RobotID)))
	}

	return operation.NewGetRobotByIDV1OK().WithPayload(handler_model.NewRobot(r[0]).ToSwagger())
}

func (rAPI *robotV1API) UpdateRobotV1(ctx context.Context, params operation.UpdateRobotV1Params) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := rAPI.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceRobot); err != nil {
		return rAPI.SendError(ctx, err)
	}

	pro, err := rAPI.projectCtr.Get(ctx, projectNameOrID)
	if err != nil {
		return rAPI.SendError(ctx, err)
	}
	r, err := rAPI.robotCtl.List(ctx, q.New(q.KeyWords{"ProjectID": pro.ProjectID, "ID": params.RobotID}), &robot.Option{
		WithPermission: true,
	})
	if err != nil {
		return rAPI.SendError(ctx, err)
	}
	if len(r) == 0 {
		return rAPI.SendError(ctx, errors.NotFoundError(fmt.Errorf("cannot find robot with project id: %d and id: %d", pro.ProjectID, params.RobotID)))
	}
	robot := r[0]

	// for v1 API, only update the disable.
	robot.Disabled = params.Robot.Disable
	if err := rAPI.robotCtl.Update(ctx, robot, nil); err != nil {
		return rAPI.SendError(ctx, err)
	}

	return operation.NewUpdateRobotV1OK()
}

func (rAPI *robotV1API) validate(ctx context.Context, params operation.CreateRobotV1Params) error {
	if params.Robot == nil {
		return errors.New(nil).WithMessage("bad request no robot").WithCode(errors.BadRequestCode)
	}
	if len(params.Robot.Access) == 0 {
		return errors.New(nil).WithMessage("bad request no access").WithCode(errors.BadRequestCode)
	}

	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	pro, err := rAPI.projectCtr.Get(ctx, projectNameOrID)
	if err != nil {
		return err
	}

	policies := rbac_project.GetPoliciesOfProject(pro.ProjectID)

	mp := map[string]bool{}
	for _, policy := range policies {
		mp[policy.String()] = true
	}

	for _, policy := range params.Robot.Access {
		p := &types.Policy{}
		lib.JSONCopy(p, policy)
		if !mp[p.String()] {
			return errors.New(nil).WithMessage("%s action of %s resource not exist in project %s", policy.Action, policy.Resource, projectNameOrID).WithCode(errors.BadRequestCode)
		}
	}

	return nil
}

// /project/1/repository => repository
func getRawResource(resource string) (string, error) {
	resourceReg := regexp.MustCompile("^/project/[0-9]+/(?P<repository>[a-z-]+)$")
	matches := resourceReg.FindStringSubmatch(resource)
	if len(matches) <= 1 {
		return "", errors.New(nil).WithMessage("bad resource %s", resource).WithCode(errors.BadRequestCode)
	}
	return matches[1], nil
}
