// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/robot"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/pkg/errors"
	"net/http"
	"strconv"
)

// RobotAPI ...
type RobotAPI struct {
	BaseController
	project *models.Project
	ctr     robot.Controller
	robot   *model.Robot
}

// Prepare ...
func (r *RobotAPI) Prepare() {

	r.BaseController.Prepare()
	method := r.Ctx.Request.Method

	if !r.SecurityCtx.IsAuthenticated() {
		r.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	pid, err := r.GetInt64FromPath(":pid")
	if err != nil || pid <= 0 {
		var errMsg string
		if err != nil {
			errMsg = "failed to get project ID " + err.Error()
		} else {
			errMsg = "invalid project ID: " + fmt.Sprintf("%d", pid)
		}
		r.SendBadRequestError(errors.New(errMsg))
		return
	}
	project, err := r.ProjectMgr.Get(pid)
	if err != nil {
		r.ParseAndHandleError(fmt.Sprintf("failed to get project %d", pid), err)
		return
	}
	if project == nil {
		r.SendNotFoundError(fmt.Errorf("project %d not found", pid))
		return
	}
	r.project = project
	r.ctr = robot.RobotCtr

	if method == http.MethodPut || method == http.MethodDelete {
		id, err := r.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			r.SendBadRequestError(errors.New("invalid robot ID"))
			return
		}
		robot, err := r.ctr.GetRobotAccount(id)
		if err != nil {
			r.SendInternalServerError(fmt.Errorf("failed to get robot %d: %v", id, err))
			return
		}

		if robot == nil {
			r.SendNotFoundError(fmt.Errorf("robot %d not found", id))
			return
		}

		r.robot = robot
	}
}

func (r *RobotAPI) requireAccess(action rbac.Action) bool {
	return r.RequireProjectAccess(r.project.ProjectID, action, rbac.ResourceRobot)
}

// Post ...
func (r *RobotAPI) Post() {
	if !r.requireAccess(rbac.ActionCreate) {
		return
	}

	var robotReq model.RobotCreate
	isValid, err := r.DecodeJSONReqAndValidate(&robotReq)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}
	robotReq.Visible = true
	robotReq.ByPassPolicyCheck = false
	robotReq.ProjectID = r.project.ProjectID

	if err := validateRobotReq(r.project, &robotReq); err != nil {
		r.SendBadRequestError(err)
		return
	}

	robot, err := r.ctr.CreateRobotAccount(&robotReq)
	if err != nil {
		if err == dao.ErrDupRows {
			r.SendConflictError(errors.New("conflict robot account"))
			return
		}
		r.SendInternalServerError(errors.Wrap(err, "robot API: post"))
		return
	}

	w := r.Ctx.ResponseWriter
	w.Header().Set("Content-Type", "application/json")

	robotRep := model.RobotRep{
		Name:  robot.Name,
		Token: robot.Token,
	}

	r.Redirect(http.StatusCreated, strconv.FormatInt(robot.ID, 10))
	r.Data["json"] = robotRep
	r.ServeJSON()
}

// List list all the robots of a project
func (r *RobotAPI) List() {
	if !r.requireAccess(rbac.ActionList) {
		return
	}

	keywords := make(map[string]interface{})
	keywords["ProjectID"] = r.project.ProjectID
	keywords["Visible"] = true
	query := &q.Query{
		Keywords: keywords,
	}
	robots, err := r.ctr.ListRobotAccount(query)
	if err != nil {
		r.SendInternalServerError(errors.Wrap(err, "robot API: list"))
		return
	}
	count := len(robots)
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}

	r.SetPaginationHeader(int64(count), page, size)
	r.Data["json"] = robots
	r.ServeJSON()
}

// Get get robot by id
func (r *RobotAPI) Get() {
	if !r.requireAccess(rbac.ActionRead) {
		return
	}

	id, err := r.GetInt64FromPath(":id")
	if err != nil || id <= 0 {
		r.SendBadRequestError(fmt.Errorf("invalid robot ID: %s", r.GetStringFromPath(":id")))
		return
	}

	robot, err := r.ctr.GetRobotAccount(id)
	if err != nil {
		r.SendInternalServerError(errors.Wrap(err, "robot API: get robot"))
		return
	}
	if robot == nil {
		r.SendNotFoundError(fmt.Errorf("robot API: robot %d not found", id))
		return
	}
	if !robot.Visible {
		r.SendForbiddenError(fmt.Errorf("robot API: robot %d is invisible", id))
		return
	}

	r.Data["json"] = robot
	r.ServeJSON()
}

// Put disable or enable a robot account
func (r *RobotAPI) Put() {
	if !r.requireAccess(rbac.ActionUpdate) {
		return
	}

	var robotReq model.RobotCreate
	if err := r.DecodeJSONReq(&robotReq); err != nil {
		r.SendBadRequestError(err)
		return
	}

	r.robot.Disabled = robotReq.Disabled

	if err := r.ctr.UpdateRobotAccount(r.robot); err != nil {
		r.SendInternalServerError(errors.Wrap(err, "robot API: update"))
		return
	}

}

// Delete delete robot by id
func (r *RobotAPI) Delete() {
	if !r.requireAccess(rbac.ActionDelete) {
		return
	}

	if err := r.ctr.DeleteRobotAccount(r.robot.ID); err != nil {
		r.SendInternalServerError(errors.Wrap(err, "robot API: delete"))
		return
	}
}

func validateRobotReq(p *models.Project, robotReq *model.RobotCreate) error {
	if len(robotReq.Access) == 0 {
		return errors.New("access required")
	}

	namespace, _ := rbac.Resource(fmt.Sprintf("/project/%d", p.ProjectID)).GetNamespace()
	policies := project.GetAllPolicies(namespace)

	mp := map[string]bool{}
	for _, policy := range policies {
		mp[policy.String()] = true
	}

	for _, policy := range robotReq.Access {
		if !mp[policy.String()] {
			return fmt.Errorf("%s action of %s resource not exist in project %s", policy.Action, policy.Resource, p.Name)
		}
	}

	return nil
}
