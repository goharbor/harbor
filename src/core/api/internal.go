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
	"context"

	"github.com/goharbor/harbor/src/lib/config"

	o "github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
)

// InternalAPI handles request of harbor admin...
type InternalAPI struct {
	BaseController
}

// Prepare validates the URL and parms
func (ia *InternalAPI) Prepare() {
	ia.BaseController.Prepare()
	if !ia.SecurityCtx.IsAuthenticated() {
		ia.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	if !ia.SecurityCtx.IsSysAdmin() {
		ia.SendForbiddenError(errors.New(ia.SecurityCtx.GetUsername()))
		return
	}
}

// RenameAdmin we don't provide flexibility in this API, as this is a workaround.
func (ia *InternalAPI) RenameAdmin() {
	ctx := ia.Ctx.Request.Context()
	if !auth.IsSuperUser(ctx, ia.SecurityCtx.GetUsername()) {
		log.Errorf("User %s is not super user, not allow to rename admin.", ia.SecurityCtx.GetUsername())
		ia.SendForbiddenError(errors.New(ia.SecurityCtx.GetUsername()))
		return
	}
	newName := common.NewHarborAdminName
	if err := user.Ctl.UpdateProfile(ctx, &models.User{
		UserID:   1,
		Username: newName,
	}, "username"); err != nil {
		log.Errorf("Failed to change admin's username, error: %v", err)
		ia.SendInternalServerError(errors.New("failed to rename admin user"))
		return
	}
	log.Debugf("The super user has been renamed to: %s", newName)
	ia.DestroySession()
}

// SyncQuota ...
func (ia *InternalAPI) SyncQuota() {
	if !config.QuotaPerProjectEnable(orm.Context()) {
		ia.SendError(errors.ForbiddenError(nil).WithMessage("quota per project is deactivated"))
		return
	}
	ctx := orm.Context()
	cur := config.ReadOnly(ctx)
	cfgMgr := config.GetCfgManager(ctx)
	if !cur {
		cfgMgr.Set(ctx, common.ReadOnly, true)
		cfgMgr.Save(ctx)
	}
	// For api call, to avoid the timeout, it should be asynchronous
	go func() {
		defer func() {
			ctx := orm.Context()
			cfgMgr.Set(ctx, common.ReadOnly, cur)
			cfgMgr.Save(ctx)
		}()
		log.Info("start to sync quota(API), the system will be set to ReadOnly and back it normal once it done.")
		ctx := orm.NewContext(context.TODO(), o.NewOrm())
		err := quota.RefreshForProjects(ctx)
		if err != nil {
			log.Errorf("fail to sync quota(API), but with error: %v, please try to do it again.", err)
			return
		}
		log.Info("success to sync quota(API).")
	}()
	return
}
