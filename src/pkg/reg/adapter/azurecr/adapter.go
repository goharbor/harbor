package azurecr

import (
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeAzureAcr, new(factory)); err != nil {
		log.Errorf("Register adapter factory for %s error: %v", model.RegistryTypeAzureAcr, err)
		return
	}
	log.Infof("Factory for adapter %s registered", model.RegistryTypeAzureAcr)
}

func newAdapter(registry *model.Registry) (adp.Adapter, error) {
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
	return nil
}

type adapter struct {
	*native.Adapter
}

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeAzureAcr,
		SupportedResourceTypes: []string{
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
		SupportedTriggers: []string{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}, nil
}
