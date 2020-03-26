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

	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/core/config"
	ierror "github.com/goharbor/harbor/src/lib/error"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/pkg/errors"
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
	if !dao.IsSuperUser(ia.SecurityCtx.GetUsername()) {
		log.Errorf("User %s is not super user, not allow to rename admin.", ia.SecurityCtx.GetUsername())
		ia.SendForbiddenError(errors.New(ia.SecurityCtx.GetUsername()))
		return
	}
	newName := common.NewHarborAdminName
	if err := dao.ChangeUserProfile(models.User{
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

// QuotaSwitcher ...
type QuotaSwitcher struct {
	Enabled bool
}

// SwitchQuota ...
func (ia *InternalAPI) SwitchQuota() {
	var req QuotaSwitcher
	if err := ia.DecodeJSONReq(&req); err != nil {
		ia.SendBadRequestError(err)
		return
	}
	cur := config.ReadOnly()
	// quota per project from disable to enable, it needs to update the quota usage bases on the DB records.
	if !config.QuotaPerProjectEnable() && req.Enabled {
		if !cur {
			config.GetCfgManager().Set(common.ReadOnly, true)
			config.GetCfgManager().Save()
		}

		ctx := orm.NewContext(ia.Ctx.Request.Context(), o.NewOrm())
		if err := quota.RefreshForProjects(ctx); err != nil {
			ia.SendInternalServerError(err)
			return
		}
	}
	defer func() {
		config.GetCfgManager().Set(common.ReadOnly, cur)
		config.GetCfgManager().Set(common.QuotaPerProjectEnable, req.Enabled)
		config.GetCfgManager().Save()
	}()
	return
}

// SyncQuota ...
func (ia *InternalAPI) SyncQuota() {
	if !config.QuotaPerProjectEnable() {
		ia.SendError(ierror.ForbiddenError(nil).WithMessage("quota per project is disabled"))
		return
	}

	cur := config.ReadOnly()
	cfgMgr := config.GetCfgManager()
	if !cur {
		cfgMgr.Set(common.ReadOnly, true)
		cfgMgr.Save()
	}
	// For api call, to avoid the timeout, it should be asynchronous
	go func() {
		defer func() {
			cfgMgr.Set(common.ReadOnly, cur)
			cfgMgr.Save()
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
