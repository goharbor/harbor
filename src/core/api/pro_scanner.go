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
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/pkg/scan/api/scanner"
	"github.com/pkg/errors"
)

// ProjectScannerAPI provides rest API for managing the project level scanner(s).
type ProjectScannerAPI struct {
	// The base controller to provide common utilities
	BaseController
	// Scanner controller for operating scanner registrations.
	c scanner.Controller
	// ID of the project
	pid int64
}

// Prepare sth. for the subsequent actions
func (sa *ProjectScannerAPI) Prepare() {
	// Call super prepare method
	sa.BaseController.Prepare()

	// Check access permissions
	if !sa.SecurityCtx.IsAuthenticated() {
		sa.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	// Get ID of the project
	pid, err := sa.GetInt64FromPath(":pid")
	if err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: get project scanners"))
		return
	}
	sa.pid = pid

	sa.c = scanner.DefaultController
}

// GetProjectScanner gets the project level scanner
func (sa *ProjectScannerAPI) GetProjectScanner() {
	// Check access permissions
	resource := rbac.NewProjectNamespace(sa.pid).Resource(rbac.ResourceConfiguration)
	if !sa.SecurityCtx.Can(rbac.ActionRead, resource) {
		sa.SendForbiddenError(errors.New(sa.SecurityCtx.GetUsername()))
		return
	}
	r, err := sa.c.GetRegistrationByProject(sa.pid)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: get project scanners"))
		return
	}

	if r != nil {
		sa.Data["json"] = r
	} else {
		sa.Data["json"] = make(map[string]interface{})
	}

	sa.ServeJSON()
}

// SetProjectScanner sets the project level scanner
func (sa *ProjectScannerAPI) SetProjectScanner() {
	// Check access permissions
	resource := rbac.NewProjectNamespace(sa.pid).Resource(rbac.ResourceConfiguration)
	if !sa.SecurityCtx.Can(rbac.ActionUpdate, resource) {
		sa.SendForbiddenError(errors.New(sa.SecurityCtx.GetUsername()))
		return
	}

	body := make(map[string]string)
	if err := sa.DecodeJSONReq(&body); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: set project scanners"))
		return
	}

	uuid, ok := body["uuid"]
	if !ok || len(uuid) == 0 {
		sa.SendBadRequestError(errors.New("missing scanner uuid when setting project scanner"))
		return
	}

	if err := sa.c.SetRegistrationByProject(sa.pid, uuid); err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: set project scanners"))
		return
	}
}
