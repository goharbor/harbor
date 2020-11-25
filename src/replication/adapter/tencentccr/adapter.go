package tencentccr

import (
"github.com/goharbor/harbor/src/lib/log"
adp "github.com/goharbor/harbor/src/replication/adapter"
"github.com/goharbor/harbor/src/replication/adapter/native"
"github.com/goharbor/harbor/src/replication/model"
)

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeTencentCcr, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeTencentCcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeTencentCcr)

}

type adapter struct {
	*native.Adapter
	registry *model.Registry
	region   string
	// forceEndpoint *string
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	log.Infof("ccr debug by richie newAdapter:%s", registry.URL)
	return &adapter{
		Adapter: native.NewAdapter(registry),
	}, nil

}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return getAdapterInfo()
}

// getAdapterInfo ...
func getAdapterInfo() *model.AdapterPattern {
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints: []*model.Endpoint{
				{
					Key:   "Tecent-CCR",
					Value: "https://ccr.ccs.tencentyun.com",
				},
			},
		},
	}
	return info
}

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeTencentCcr,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeImage,
		},
		SupportedResourceFilters: []*model.FilterStyle{
			{
				Type:  model.FilterTypeName,
				Style: model.FilterStyleTypeText,
			},
			{
				Type:  model.FilterTypeTag,
				Style: model.FilterStyleTypeText,
			},
		},
		SupportedTriggers: []model.TriggerType{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}, nil
}
