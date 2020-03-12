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

package handler

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	api_preheat "github.com/goharbor/harbor/src/api/preheat"
	dao_models "github.com/goharbor/harbor/src/pkg/p2p/preheat/dao/models"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/preheat"
)

func newPreheatAPI() *preheatAPI {
	return &preheatAPI{
		preheatCtl: api_preheat.DefaultController,
	}
}

type preheatAPI struct {
	BaseAPI
	preheatCtl api_preheat.Controller
}

// CreateInstance is Create p2p instances
func (p *preheatAPI) CreateInstance(ctx context.Context, params preheat.CreateInstanceParams) middleware.Responder {
	id, err := api_preheat.DefaultController.CreateInstance(&dao_models.Metadata{
		ID:             params.Instance.ID,
		Name:           params.Instance.Name,
		Description:    params.Instance.Description,
		Provider:       params.Instance.Provider,
		Endpoint:       params.Instance.Endpoint,
		AuthMode:       params.Instance.AuthMode,
		AuthData:       params.Instance.AuthData,
		Status:         params.Instance.Status,
		Enabled:        params.Instance.Enabled,
		SetupTimestamp: params.Instance.SetupTimestamp,
		Extensions:     params.Instance.Extensions,
	})
	if err != nil {
		return p.SendError(ctx, err)
	}

	return preheat.NewCreateInstanceCreated().WithPayload(&models.InstanceCreatedResp{
		ID: id,
	})
}

// DeleteInstance is Delete instance
func (p *preheatAPI) DeleteInstance(ctx context.Context, params preheat.DeleteInstanceParams) middleware.Responder {
	if err := api_preheat.DefaultController.DeleteInstance(params.InstanceID); err != nil {
		return p.SendError(ctx, err)
	}

	return preheat.NewDeleteInstanceOK().WithPayload(&models.InstanceDeletedResp{
		Removed: params.InstanceID,
	})
}

// GetInstance is Get instance
func (p *preheatAPI) GetInstance(ctx context.Context, params preheat.GetInstanceParams) middleware.Responder {
	instance, err := api_preheat.DefaultController.GetInstance(params.InstanceID)
	if err != nil {
		return p.SendError(ctx, err)
	}

	return preheat.NewGetInstanceOK().WithPayload(&models.Instance{
		AuthData:       instance.AuthData,
		AuthMode:       instance.AuthMode,
		Description:    instance.Description,
		Enabled:        instance.Enabled,
		Endpoint:       instance.Endpoint,
		Extensions:     instance.Extensions,
		ID:             instance.ID,
		Name:           instance.Name,
		Provider:       instance.Provider,
		SetupTimestamp: instance.SetupTimestamp,
		Status:         instance.Status,
	})
}

// ListInstances is List p2p instances
func (p *preheatAPI) ListInstances(ctx context.Context, params preheat.ListInstancesParams) middleware.Responder {
	queryParams := &dao_models.QueryParam{}
	if params.PageSize != nil {
		queryParams.PageSize = uint(*params.PageSize)
	} else {
		queryParams.PageSize = 10
	}
	if params.Page != nil {
		queryParams.Page = uint(*params.Page)
	} else {
		queryParams.Page = 1
	}
	if params.Q != nil {
		queryParams.Keyword = *params.Q
	}

	data, err := api_preheat.DefaultController.ListInstances(queryParams)
	if err != nil {
		return p.SendError(ctx, err)
	}

	var instances []*models.Instance
	for _, i := range data {
		instances = append(instances, &models.Instance{
			AuthData:       i.AuthData,
			AuthMode:       i.AuthMode,
			Description:    i.Description,
			Enabled:        i.Enabled,
			Endpoint:       i.Endpoint,
			Extensions:     i.Extensions,
			ID:             i.ID,
			Name:           i.Name,
			Provider:       i.Provider,
			SetupTimestamp: i.SetupTimestamp,
			Status:         i.Status,
		})
	}

	return preheat.NewListInstancesOK().WithPayload(instances)
}

// ListPreheatHistories is List preheats history
func (p *preheatAPI) ListPreheatHistories(ctx context.Context, params preheat.ListPreheatHistoriesParams) middleware.Responder {
	queryParams := &dao_models.QueryParam{}
	if params.PageSize != nil {
		queryParams.PageSize = uint(*params.PageSize)
	} else {
		queryParams.PageSize = 10
	}
	if params.Page != nil {
		queryParams.Page = uint(*params.Page)
	} else {
		queryParams.Page = 1
	}
	if params.Q != nil {
		queryParams.Keyword = *params.Q
	}

	data, err := api_preheat.DefaultController.LoadHistoryRecords(queryParams)
	if err != nil {
		return p.SendError(ctx, err)
	}

	var histories []*models.PreheatHistory
	for _, i := range data {
		histories = append(histories, &models.PreheatHistory{
			FinishTime: i.FinishTime,
			Image:      i.Image,
			Instance:   i.Instance,
			Provider:   i.Provider,
			StartTime:  i.StartTime,
			Status:     i.Status,
			TaskID:     i.TaskID,
		})
	}

	return preheat.NewListPreheatHistoriesOK().WithPayload(histories)
}

// ListProviders is List available p2p providers.
func (p *preheatAPI) ListProviders(ctx context.Context, params preheat.ListProvidersParams) middleware.Responder {
	data, err := api_preheat.DefaultController.GetAvailableProviders()
	if err != nil {
		return p.SendError(ctx, err)
	}

	var providers []*models.Provider
	for _, i := range data {
		providers = append(providers, &models.Provider{
			AuthMode:    i.AuthMode,
			Icon:        i.Icon,
			ID:          i.ID,
			Maintainers: i.Maintainers,
			Name:        i.Name,
			Source:      i.Source,
			Version:     i.Version,
		})
	}

	return preheat.NewListProvidersOK().WithPayload(providers)
}

// PreheatImages is Start to preheat images
func (p *preheatAPI) PreheatImages(ctx context.Context, params preheat.PreheatImagesParams) middleware.Responder {
	preheatingImages, ok := params.PreheatReq["images"]
	if !ok {
		return preheat.NewPreheatImagesBadRequest().WithPayload([]*models.Error{
			{
				Message: "missing images",
			},
		})
	}

	imageList, ok := preheatingImages.([]interface{})
	if !ok {
		return preheat.NewPreheatImagesBadRequest().WithPayload([]*models.Error{
			{
				Message: "'images' should be an array",
			},
		})
	}

	if len(imageList) == 0 {
		return preheat.NewPreheatImagesBadRequest().WithPayload([]*models.Error{
			{
				Message: "no images submitted",
			},
		})
	}

	var imageRepos []dao_models.ImageRepository
	for _, img := range imageList {
		imageRepos = append(imageRepos, dao_models.ImageRepository(img.(string)))
	}
	result, err := api_preheat.DefaultController.PreheatImages(imageRepos...)
	if err != nil {
		return p.SendError(ctx, err)
	}

	preheatings := make(map[string][]models.PreheatingStatus)
	for k, v := range result {
		var statuses []models.PreheatingStatus
		for _, s := range *v {
			statuses = append(statuses, models.PreheatingStatus{
				Error:      s.Error,
				FinishTime: s.FinishTime,
				StartTime:  s.StartTime,
				Status:     s.Status,
				TaskID:     s.TaskID,
			})
		}
		preheatings[k] = statuses
	}

	return preheat.NewPreheatImagesOK().WithPayload(preheatings)
}

// UpdateInstance is Update instance
func (p *preheatAPI) UpdateInstance(ctx context.Context, params preheat.UpdateInstanceParams) middleware.Responder {
	if err := api_preheat.DefaultController.UpdateInstance(params.InstanceID, params.PropertySet); err != nil {
		return p.SendError(ctx, err)
	}

	return preheat.NewUpdateInstanceOK().WithPayload(&models.InstanceUpdateResp{
		Updated: params.InstanceID,
	})
}
