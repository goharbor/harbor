package azurecr

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/http/modifier"
	common_http_auth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils/log"
	registry_pkg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeAzureAcr, factory); err != nil {
		log.Errorf("Register adapter factory for %s error: %v", model.RegistryTypeAzureAcr, err)
		return
	}
	log.Infof("Factory for adapter %s registered", model.RegistryTypeAzureAcr)
}

func factory(registry *model.Registry) (adp.Adapter, error) {
	client, err := getClient(registry)
	if err != nil {
		return nil, err
	}

	reg, err := native.NewWithClient(registry, client)
	if err != nil {
		return nil, err
	}

	return &adapter{
		registry: registry,
		Native:   reg,
	}, nil
}

type adapter struct {
	*native.Native
	registry *model.Registry
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

// PrepareForPush no preparation needed for Azure container registry
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	return nil
}

// HealthCheck checks health status of a registry
func (a adapter) HealthCheck() (model.HealthStatus, error) {
	err := a.PingGet()
	if err != nil {
		return model.Unhealthy, nil
	}

	return model.Healthy, nil
}

func getClient(registry *model.Registry) (*http.Client, error) {
	if registry.Credential == nil ||
		len(registry.Credential.AccessKey) == 0 || len(registry.Credential.AccessSecret) == 0 {
		return nil, fmt.Errorf("no credential to ping registry %s", registry.URL)
	}

	var cred modifier.Modifier
	if registry.Credential.Type == model.CredentialTypeSecret {
		cred = common_http_auth.NewSecretAuthorizer(registry.Credential.AccessSecret)
	} else {
		cred = auth.NewBasicAuthCredential(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret)
	}

	transport := util.GetHTTPTransport(registry.Insecure)
	modifiers := []modifier.Modifier{
		&auth.UserAgentModifier{
			UserAgent: adp.UserAgentReplication,
		},
		cred,
	}

	client := &http.Client{
		Transport: registry_pkg.NewTransport(transport, modifiers...),
	}

	return client, nil
}
