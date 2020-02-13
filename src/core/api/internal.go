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
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_quota "github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/pkg/errors"
	"strconv"

	quota "github.com/goharbor/harbor/src/core/api/quota"
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
		if err := ia.ensureQuota(); err != nil {
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

func (ia *InternalAPI) ensureQuota() error {
	projects, err := dao.GetProjects(nil)
	if err != nil {
		return err
	}
	for _, project := range projects {
		pSize, err := dao.CountSizeOfProject(project.ProjectID)
		if err != nil {
			logger.Warningf("error happen on counting size of project:%d , error:%v, just skip it.", project.ProjectID, err)
			continue
		}
		afQuery := &models.ArtifactQuery{
			PID: project.ProjectID,
		}
		afs, err := dao.ListArtifacts(afQuery)
		if err != nil {
			logger.Warningf("error happen on counting number of project:%d , error:%v, just skip it.", project.ProjectID, err)
			continue
		}
		pCount := int64(len(afs))

		// it needs to append the chart count
		if config.WithChartMuseum() {
			count, err := chartController.GetCountOfCharts([]string{project.Name})
			if err != nil {
				err = errors.Wrap(err, fmt.Sprintf("get chart count of project %d failed", project.ProjectID))
				logger.Error(err)
				continue
			}
			pCount = pCount + int64(count)
		}

		quotaMgr, err := common_quota.NewManager("project", strconv.FormatInt(project.ProjectID, 10))
		if err != nil {
			logger.Errorf("Error occurred when to new quota manager %v, just skip it.", err)
			continue
		}
		used := common_quota.ResourceList{
			common_quota.ResourceStorage: pSize,
			common_quota.ResourceCount:   pCount,
		}
		if err := quotaMgr.EnsureQuota(used); err != nil {
			logger.Errorf("cannot ensure quota for the project: %d, err: %v, just skip it.", project.ProjectID, err)
			continue
		}
	}
	return nil
}

// SyncQuota ...
func (ia *InternalAPI) SyncQuota() {
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
		// As the sync function ignores all of duplicate error, it's safe to enable persist DB.
		err := quota.Sync(ia.ProjectMgr, true)
		if err != nil {
			log.Errorf("fail to sync quota(API), but with error: %v, please try to do it again.", err)
			return
		}
		log.Info("success to sync quota(API).")
	}()
	return
}
