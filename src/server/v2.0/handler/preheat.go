package handler

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	preheatCtl "github.com/goharbor/harbor/src/controller/p2p/preheat"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/preheat"
)

func newPreheatAPI() *preheatAPI {
	return &preheatAPI{
		preheatCtl: preheatCtl.Ctl,
	}
}

var _ restapi.PreheatAPI = (*preheatAPI)(nil)

type preheatAPI struct {
	BaseAPI
	preheatCtl preheatCtl.Controller
}

func (api *preheatAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

func (api *preheatAPI) CreateInstance(ctx context.Context, params operation.CreateInstanceParams) middleware.Responder {
	var payload *models.InstanceCreatedResp
	return operation.NewCreateInstanceCreated().WithPayload(payload)
}

func (api *preheatAPI) DeleteInstance(ctx context.Context, params operation.DeleteInstanceParams) middleware.Responder {
	var payload *models.InstanceDeletedResp
	return operation.NewDeleteInstanceOK().WithPayload(payload)
}

func (api *preheatAPI) GetInstance(ctx context.Context, params operation.GetInstanceParams) middleware.Responder {
	var payload *models.Instance
	return operation.NewGetInstanceOK().WithPayload(payload)
}

// ListInstances is List p2p instances
func (api *preheatAPI) ListInstances(ctx context.Context, params operation.ListInstancesParams) middleware.Responder {
	var payload []*models.Instance
	return operation.NewListInstancesOK().WithPayload(payload)
}

func (api *preheatAPI) ListProviders(ctx context.Context, params operation.ListProvidersParams) middleware.Responder {

	var providers, err = preheatCtl.Ctl.GetAvailableProviders()
	if err != nil {
		return operation.NewListProvidersInternalServerError()
	}
	var payload = convertProvidersToFrontend(providers)

	return operation.NewListProvidersOK().WithPayload(payload)
}

// UpdateInstance is Update instance
func (api *preheatAPI) UpdateInstance(ctx context.Context, params operation.UpdateInstanceParams) middleware.Responder {
	var payload *models.InstanceUpdateResp
	return operation.NewUpdateInstanceOK().WithPayload(payload)
}

func convertProvidersToFrontend(backend []*provider.Metadata) (frontend []*models.Metadata) {
	frontend = make([]*models.Metadata, 0)
	for _, provider := range backend {
		frontend = append(frontend, &models.Metadata{
			ID:          provider.ID,
			Icon:        provider.Icon,
			Name:        provider.Name,
			Source:      provider.Source,
			Version:     provider.Version,
			Maintainers: provider.Maintainers,
		})
	}
	return
}
