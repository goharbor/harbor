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

package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/immutabletag"
	"github.com/goharbor/harbor/src/pkg/immutabletag/model"
)

// ImmutableTagRuleAPI ...
type ImmutableTagRuleAPI struct {
	BaseController
	ctr       immutabletag.Controller
	projectID int64
	ID        int64
}

// Prepare validates the user and projectID
func (itr *ImmutableTagRuleAPI) Prepare() {
	itr.BaseController.Prepare()
	// Check access permissions
	if !itr.RequireAuthenticated() {
		return
	}

	pid, err := itr.GetInt64FromPath(":pid")
	if err != nil || pid <= 0 {
		text := "invalid project ID: "
		if err != nil {
			text += err.Error()
		} else {
			text += fmt.Sprintf("%d", pid)
		}
		itr.SendError(errors.New(err).WithCode(errors.BadRequestCode))
		return
	}
	itr.projectID = pid
	itr.ctr = immutabletag.ImmuCtr
	ruleID, err := itr.GetInt64FromPath(":id")
	if err == nil || ruleID > 0 {
		itr.ID = ruleID
		itRule, err := itr.ctr.GetImmutableRule(itr.ID)
		if err != nil {
			itr.SendError(err)
			return
		}
		if itRule.ProjectID != itr.projectID {
			err := fmt.Errorf("immutable tag rule %v not found", itr.ID)
			itr.SendError(errors.New(err).WithCode(errors.NotFoundCode))
			return
		}
	}

	if strings.EqualFold(itr.Ctx.Request.Method, "get") {
		if !itr.requireAccess(rbac.ActionList) {
			return
		}
	} else if strings.EqualFold(itr.Ctx.Request.Method, "put") {
		if !itr.requireAccess(rbac.ActionUpdate) {
			return
		}
	} else if strings.EqualFold(itr.Ctx.Request.Method, "post") {
		if !itr.requireAccess(rbac.ActionCreate) {
			return
		}

	} else if strings.EqualFold(itr.Ctx.Request.Method, "delete") {
		if !itr.requireAccess(rbac.ActionDelete) {
			return
		}
	}
}

func (itr *ImmutableTagRuleAPI) requireAccess(action rbac.Action) bool {
	return itr.RequireProjectAccess(itr.projectID, action, rbac.ResourceImmutableTag)
}

// List list all immutable tag rules of current project
func (itr *ImmutableTagRuleAPI) List() {
	rules, err := itr.ctr.ListImmutableRules(itr.projectID)
	if err != nil {
		itr.SendError(err)
		return
	}
	itr.WriteJSONData(rules)
}

// Post create immutable tag rule
func (itr *ImmutableTagRuleAPI) Post() {
	ir := &model.Metadata{}
	isValid, err := itr.DecodeJSONReqAndValidate(ir)
	if !isValid {
		itr.SendError(errors.New(err).WithCode(errors.BadRequestCode))
		return
	}
	ir.ProjectID = itr.projectID
	id, err := itr.ctr.CreateImmutableRule(ir)
	if err != nil {
		itr.SendError(err)
		return
	}
	itr.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Delete delete immutable tag rule
func (itr *ImmutableTagRuleAPI) Delete() {
	if itr.ID <= 0 {
		err := fmt.Errorf("invalid immutable rule id %d", itr.ID)
		itr.SendError(errors.New(err).WithCode(errors.BadRequestCode))
		return
	}
	err := itr.ctr.DeleteImmutableRule(itr.ID)
	if err != nil {
		itr.SendError(err)
		return
	}
}

// Put update an immutable tag rule
func (itr *ImmutableTagRuleAPI) Put() {
	ir := &model.Metadata{}
	if err := itr.DecodeJSONReq(ir); err != nil {
		itr.SendError(errors.New(err).WithCode(errors.BadRequestCode))
		return
	}
	ir.ID = itr.ID
	ir.ProjectID = itr.projectID

	if itr.ID <= 0 {
		err := fmt.Errorf("invalid immutable rule id %d", itr.ID)
		itr.SendError(errors.New(err).WithCode(errors.BadRequestCode))
		return
	}

	if err := itr.ctr.UpdateImmutableRule(itr.projectID, ir); err != nil {
		itr.SendError(err)
		return
	}
}
