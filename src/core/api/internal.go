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
	"net/http"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// InternalAPI handles request of harbor admin...
type InternalAPI struct {
	BaseController
}

// Prepare validates the URL and parms
func (ia *InternalAPI) Prepare() {
	ia.BaseController.Prepare()
	if !ia.SecurityCtx.IsAuthenticated() {
		ia.HandleUnauthorized()
		return
	}
	if !ia.SecurityCtx.IsSysAdmin() {
		ia.HandleForbidden(ia.SecurityCtx.GetUsername())
		return
	}
}

// SyncRegistry ...
func (ia *InternalAPI) SyncRegistry() {
	err := SyncRegistry(ia.ProjectMgr)
	if err != nil {
		ia.HandleInternalServerError(err.Error())
		return
	}
}

// RenameAdmin we don't provide flexibility in this API, as this is a workaround.
func (ia *InternalAPI) RenameAdmin() {
	if !dao.IsSuperUser(ia.SecurityCtx.GetUsername()) {
		log.Errorf("User %s is not super user, not allow to rename admin.", ia.SecurityCtx.GetUsername())
		ia.CustomAbort(http.StatusForbidden, "")
	}
	newName := common.NewHarborAdminName
	if err := dao.ChangeUserProfile(models.User{
		UserID:   1,
		Username: newName,
	}, "username"); err != nil {
		log.Errorf("Failed to change admin's username, error: %v", err)
		ia.CustomAbort(http.StatusInternalServerError, "Failed to rename admin user.")
	}
	log.Debugf("The super user has been renamed to: %s", newName)
	ia.DestroySession()
}
