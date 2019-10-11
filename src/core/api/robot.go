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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/token"
	"github.com/goharbor/harbor/src/core/config"
)

// RobotAPI ...
type RobotAPI struct {
	BaseController
	project *models.Project
	robot   *models.Robot
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

	if method == http.MethodPut || method == http.MethodDelete {
		id, err := r.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			r.SendBadRequestError(errors.New("invalid robot ID"))
			return
		}

		robot, err := dao.GetRobotByID(id)
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

	var robotReq models.RobotReq
	isValid, err := r.DecodeJSONReqAndValidate(&robotReq)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}

	if err := validateRobotReq(r.project, &robotReq); err != nil {
		r.SendBadRequestError(err)
		return
	}

	// Token duration in minutes
	tokenDuration := time.Duration(config.RobotTokenDuration()) * time.Minute
	expiresAt := time.Now().UTC().Add(tokenDuration).Unix()
	createdName := common.RobotPrefix + robotReq.Name

	// first to add a robot account, and get its id.
	robot := models.Robot{
		Name:        createdName,
		Description: robotReq.Description,
		ProjectID:   r.project.ProjectID,
		ExpiresAt:   expiresAt,
	}
	id, err := dao.AddRobot(&robot)
	if err != nil {
		if err == dao.ErrDupRows {
			r.SendConflictError(errors.New("conflict robot account"))
			return
		}
		r.SendInternalServerError(fmt.Errorf("failed to create robot account: %v", err))
		return
	}

	// generate the token, and return it with response data.
	// token is not stored in the database.
	jwtToken, err := token.New(id, r.project.ProjectID, expiresAt, robotReq.Access)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to valid parameters to generate token for robot account, %v", err))
		err := dao.DeleteRobot(id)
		if err != nil {
			r.SendInternalServerError(fmt.Errorf("failed to delete the robot account: %d, %v", id, err))
		}
		return
	}

	rawTk, err := jwtToken.Raw()
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to sign token for robot account, %v", err))
		err := dao.DeleteRobot(id)
		if err != nil {
			r.SendInternalServerError(fmt.Errorf("failed to delete the robot account: %d, %v", id, err))
		}
		return
	}

	robotRep := models.RobotRep{
		Name:  robot.Name,
		Token: rawTk,
	}

	w := r.Ctx.ResponseWriter
	w.Header().Set("Content-Type", "application/json")

	r.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
	r.Data["json"] = robotRep
	r.ServeJSON()
}

// List list all the robots of a project
func (r *RobotAPI) List() {
	if !r.requireAccess(rbac.ActionList) {
		return
	}

	query := models.RobotQuery{
		ProjectID: r.project.ProjectID,
	}

	count, err := dao.CountRobot(&query)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to list robots on project: %d, %v", r.project.ProjectID, err))
		return
	}
	query.Page, query.Size, err = r.GetPaginationParams()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}

	robots, err := dao.ListRobots(&query)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to get robots %v", err))
		return
	}

	r.SetPaginationHeader(count, query.Page, query.Size)
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

	robot, err := dao.GetRobotByID(id)
	if err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to get robot %d: %v", id, err))
		return
	}
	if robot == nil {
		r.SendNotFoundError(fmt.Errorf("robot %d not found", id))
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

	var robotReq models.RobotReq
	if err := r.DecodeJSONReq(&robotReq); err != nil {
		r.SendBadRequestError(err)
		return
	}

	r.robot.Disabled = robotReq.Disabled

	if err := dao.UpdateRobot(r.robot); err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to update robot %d: %v", r.robot.ID, err))
		return
	}

}

// Delete delete robot by id
func (r *RobotAPI) Delete() {
	if !r.requireAccess(rbac.ActionDelete) {
		return
	}

	if err := dao.DeleteRobot(r.robot.ID); err != nil {
		r.SendInternalServerError(fmt.Errorf("failed to delete robot %d: %v", r.robot.ID, err))
		return
	}
}

func validateRobotReq(p *models.Project, robotReq *models.RobotReq) error {
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
