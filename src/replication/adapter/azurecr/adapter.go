package azurecr

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeAzureAcr, factory, nil); err != nil {
		log.Errorf("Register adapter factory for %s error: %v", model.RegistryTypeAzureAcr, err)
		return
	}
	log.Infof("Factory for adapter %s registered", model.RegistryTypeAzureAcr)
}

func factory(registry *model.Registry) (adp.Adapter, error) {
	dockerRegistryAdapter, err := native.NewAdapter(registry)
	if err != nil {
		return nil, err
	}
	return &adapter{
		Adapter: dockerRegistryAdapter,
	}, nil
}

type adapter struct {
	*native.Adapter
}

// Ensure '*adapter' implements interface 'Adapter'.
var _ adp.Adapter = (*adapter)(nil)

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeAzureAcr,
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
