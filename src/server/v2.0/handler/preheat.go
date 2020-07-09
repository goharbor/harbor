package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	preheatCtl "github.com/goharbor/harbor/src/controller/p2p/preheat"
	projectCtl "github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	instanceModel "github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/provider"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/preheat"
)

func newPreheatAPI() *preheatAPI {
	return &preheatAPI{
		preheatCtl: preheatCtl.Ctl,
		projectCtl: projectCtl.Ctl,
	}
}

var _ restapi.PreheatAPI = (*preheatAPI)(nil)

// nameRegex is the regex for name validation.
const nameRegex = "^[A-Za-z0-9]+(?:[._-][A-Za-z0-9]+)*$"

type preheatAPI struct {
	BaseAPI
	preheatCtl preheatCtl.Controller
	projectCtl projectCtl.Controller
}

func (api *preheatAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

func (api *preheatAPI) CreateInstance(ctx context.Context, params operation.CreateInstanceParams) middleware.Responder {
	instance, err := convertParamInstanceToModelInstance(params.Instance)
	if err != nil {
		return api.SendError(ctx, err)
	}

	_, err = api.preheatCtl.CreateInstance(ctx, instance)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewCreateInstanceCreated()
}

func (api *preheatAPI) DeleteInstance(ctx context.Context, params operation.DeleteInstanceParams) middleware.Responder {
	instance, err := api.preheatCtl.GetInstanceByName(ctx, params.PreheatInstanceName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	err = api.preheatCtl.DeleteInstance(ctx, instance.ID)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewDeleteInstanceOK()
}

func (api *preheatAPI) GetInstance(ctx context.Context, params operation.GetInstanceParams) middleware.Responder {
	var payload *models.Instance
	instance, err := api.preheatCtl.GetInstanceByName(ctx, params.PreheatInstanceName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	payload, err = convertInstanceToPayload(instance)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewGetInstanceOK().WithPayload(payload)
}

// ListInstances is List p2p instances
func (api *preheatAPI) ListInstances(ctx context.Context, params operation.ListInstancesParams) middleware.Responder {
	var payload []*models.Instance

	query, err := api.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return api.SendError(ctx, err)
	}

	total, err := api.preheatCtl.CountInstance(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	instances, err := api.preheatCtl.ListInstance(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	for _, instance := range instances {
		ins, err := convertInstanceToPayload(instance)
		if err != nil {
			return api.SendError(ctx, err)
		}
		payload = append(payload, ins)
	}
	return operation.NewListInstancesOK().
		WithPayload(payload).WithXTotalCount(total).
		WithLink(api.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
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

// GetPolicy is Get a preheat policy
func (api *preheatAPI) GetPolicy(ctx context.Context, params operation.GetPolicyParams) middleware.Responder {
	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	var payload *models.PreheatPolicy
	policy, err := api.preheatCtl.GetPolicyByName(ctx, project.ProjectID, params.PreheatPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	payload, err = convertPolicyToPayload(policy)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewGetPolicyOK().WithPayload(payload)
}

// CreatePolicy is Create a preheat policy under a project
func (api *preheatAPI) CreatePolicy(ctx context.Context, params operation.CreatePolicyParams) middleware.Responder {
	policy, err := convertParamPolicyToModelPolicy(params.Policy)
	if err != nil {
		return api.SendError(ctx, err)
	}

	_, err = api.preheatCtl.CreatePolicy(ctx, policy)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewCreatePolicyCreated()
}

// UpdatePolicy is Update preheat policy
func (api *preheatAPI) UpdatePolicy(ctx context.Context, params operation.UpdatePolicyParams) middleware.Responder {
	policy, err := convertParamPolicyToModelPolicy(params.Policy)
	if err != nil {
		return api.SendError(ctx, err)
	}

	err = api.preheatCtl.UpdatePolicy(ctx, policy)
	if err != nil {
		return api.SendError(ctx, err)
	}
	return operation.NewUpdatePolicyOK()
}

// DeletePolicy is Delete a preheat policy
func (api *preheatAPI) DeletePolicy(ctx context.Context, params operation.DeletePolicyParams) middleware.Responder {
	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	policy, err := api.preheatCtl.GetPolicyByName(ctx, project.ProjectID, params.PreheatPolicyName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	err = api.preheatCtl.DeletePolicy(ctx, policy.ID)
	if err != nil {
		return api.SendError(ctx, err)
	}

	return operation.NewDeletePolicyOK()
}

// ListPolicies is List preheat policies
func (api *preheatAPI) ListPolicies(ctx context.Context, params operation.ListPoliciesParams) middleware.Responder {
	project, err := api.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return api.SendError(ctx, err)
	}

	query, err := api.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return api.SendError(ctx, err)
	}

	if query != nil {
		query.Keywords["project_id"] = project.ProjectID
	}

	total, err := api.preheatCtl.CountPolicy(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	policies, err := api.preheatCtl.ListPolicies(ctx, query)
	if err != nil {
		return api.SendError(ctx, err)
	}

	var payload []*models.PreheatPolicy
	for _, policy := range policies {
		p, err := convertPolicyToPayload(policy)
		if err != nil {
			return api.SendError(ctx, err)
		}
		payload = append(payload, p)
	}
	return operation.NewListPoliciesOK().WithPayload(payload).WithXTotalCount(total).
		WithLink(api.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String())
}

// convertPolicyToPayload converts model policy to swagger model
func convertPolicyToPayload(policy *policy.Schema) (*models.PreheatPolicy, error) {
	if policy == nil {
		return nil, errors.New("policy can not be nil")
	}

	return &models.PreheatPolicy{
		CreationTime: strfmt.DateTime(policy.CreatedAt),
		Description:  policy.Description,
		Enabled:      policy.Enabled,
		Filters:      policy.FiltersStr,
		ID:           policy.ID,
		Name:         policy.Name,
		ProjectID:    policy.ProjectID,
		ProviderID:   policy.ProviderID,
		Trigger:      policy.TriggerStr,
		UpdateTime:   strfmt.DateTime(policy.UpdatedTime),
	}, nil
}

// convertParamPolicyToPolicy converts params policy to pkg model policy
func convertParamPolicyToModelPolicy(model *models.PreheatPolicy) (*policy.Schema, error) {
	if model == nil {
		return nil, errors.New("policy can not be nil")
	}

	valid, err := regexp.MatchString(nameRegex, model.Name)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("name %s is invalid", model.Name)
	}

	return &policy.Schema{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		ProjectID:   model.ProjectID,
		ProviderID:  model.ProviderID,
		FiltersStr:  model.Filters,
		TriggerStr:  model.Trigger,
		Enabled:     model.Enabled,
		CreatedAt:   time.Time(model.CreationTime),
		UpdatedTime: time.Time(model.UpdateTime),
	}, nil
}

func convertInstanceToPayload(model *instanceModel.Instance) (*models.Instance, error) {
	if model == nil {
		return nil, errors.New("instance can not be nil")
	}

	var authInfo = map[string]string{}
	var err = json.Unmarshal([]byte(model.AuthData), &authInfo)
	if err != nil {
		return nil, err
	}
	return &models.Instance{
		AuthInfo:       authInfo,
		AuthMode:       model.AuthMode,
		Default:        model.Default,
		Description:    model.Description,
		Enabled:        model.Enabled,
		Endpoint:       model.Endpoint,
		ID:             model.ID,
		Insecure:       model.Insecure,
		Name:           model.Name,
		SetupTimestamp: model.SetupTimestamp,
		Status:         "Unknown",
		Vendor:         model.Vendor,
	}, nil
}

func convertParamInstanceToModelInstance(model *models.Instance) (*instanceModel.Instance, error) {
	if model == nil {
		return nil, errors.New("instance can not be nil")
	}

	valid, err := regexp.MatchString(nameRegex, model.Name)
	if err != nil {
		return nil, err
	}

	if !valid {
		return nil, fmt.Errorf("name %s is invalid", model.Name)
	}

	authData, err := json.Marshal(model.AuthInfo)
	if err != nil {
		return nil, err
	}

	return &instanceModel.Instance{
		AuthData:       string(authData),
		AuthMode:       model.AuthMode,
		Default:        model.Default,
		Description:    model.Description,
		Enabled:        model.Enabled,
		Endpoint:       model.Endpoint,
		ID:             model.ID,
		Insecure:       model.Insecure,
		Name:           model.Name,
		SetupTimestamp: model.SetupTimestamp,
		Status:         model.Status,
		Vendor:         model.Vendor,
	}, nil
}
