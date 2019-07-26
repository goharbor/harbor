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
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/pkg/errors"
)

// QuotaAPI handles request to /api/quotas/
type QuotaAPI struct {
	BaseController
	quota *models.Quota
}

// Prepare validates the URL and the user
func (qa *QuotaAPI) Prepare() {
	qa.BaseController.Prepare()

	if !qa.SecurityCtx.IsAuthenticated() {
		qa.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	if !qa.SecurityCtx.IsSysAdmin() {
		qa.SendForbiddenError(errors.New(qa.SecurityCtx.GetUsername()))
		return
	}

	if len(qa.GetStringFromPath(":id")) != 0 {
		id, err := qa.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			text := "invalid quota ID: "
			if err != nil {
				text += err.Error()
			} else {
				text += fmt.Sprintf("%d", id)
			}
			qa.SendBadRequestError(errors.New(text))
			return
		}

		quota, err := dao.GetQuota(id)
		if err != nil {
			qa.SendInternalServerError(fmt.Errorf("failed to get quota %d, error: %v", id, err))
			return
		}

		if quota == nil {
			qa.SendNotFoundError(fmt.Errorf("quota %d not found", id))
			return
		}

		qa.quota = quota
	}
}

// Get returns quota by id
func (qa *QuotaAPI) Get() {
	query := &models.QuotaQuery{
		ID: qa.quota.ID,
	}

	quotas, err := dao.ListQuotas(query)
	if err != nil {
		qa.SendInternalServerError(fmt.Errorf("failed to get quota %d, error: %v", qa.quota.ID, err))
		return
	}

	if len(quotas) == 0 {
		qa.SendNotFoundError(fmt.Errorf("quota %d not found", qa.quota.ID))
		return
	}

	qa.Data["json"] = quotas[0]
	qa.ServeJSON()
}

// Put update the quota
func (qa *QuotaAPI) Put() {
	var req *models.QuotaUpdateRequest
	if err := qa.DecodeJSONReq(&req); err != nil {
		qa.SendBadRequestError(err)
		return
	}

	if err := quota.Validate(qa.quota.Reference, req.Hard); err != nil {
		qa.SendBadRequestError(err)
		return
	}

	mgr, err := quota.NewManager(qa.quota.Reference, qa.quota.ReferenceID)
	if err != nil {
		qa.SendInternalServerError(fmt.Errorf("failed to create quota manager, error: %v", err))
		return
	}

	if err := mgr.UpdateQuota(req.Hard); err != nil {
		qa.SendInternalServerError(fmt.Errorf("failed to update hard limits of the quota, error: %v", err))
		return
	}
}

// List returns quotas by query
func (qa *QuotaAPI) List() {
	page, size, err := qa.GetPaginationParams()
	if err != nil {
		qa.SendBadRequestError(err)
		return
	}

	query := &models.QuotaQuery{
		Reference:   qa.GetString("reference"),
		ReferenceID: qa.GetString("reference_id"),
		Pagination: models.Pagination{
			Page: page,
			Size: size,
		},
		Sorting: models.Sorting{
			Sort: qa.GetString("sort"),
		},
	}

	total, err := dao.GetTotalOfQuotas(query)
	if err != nil {
		qa.SendInternalServerError(fmt.Errorf("failed to query database for total of quotas, error: %v", err))
		return
	}

	quotas, err := dao.ListQuotas(query)
	if err != nil {
		qa.SendInternalServerError(fmt.Errorf("failed to query database for quotas, error: %v", err))
		return
	}

	qa.SetPaginationHeader(total, page, size)
	qa.Data["json"] = quotas
	qa.ServeJSON()
}
