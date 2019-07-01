package quayio

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

type adapter struct {
	*native.Adapter
	registry *model.Registry
	client   *common_http.Client
}

func init() {
	err := adp.RegisterFactory(model.RegistryTypeQuayio, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry)
	}, nil)
	if err != nil {
		log.Errorf("failed to register factory for Quay.io: %v", err)
		return
	}
	log.Infof("the factory of Quay.io adapter was registered")
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	modifiers := []modifier.Modifier{
		&auth.UserAgentModifier{
			UserAgent: adp.UserAgentReplication,
		},
	}

	var authorizer modifier.Modifier
	if registry.Credential != nil && len(registry.Credential.AccessKey) != 0 {
		authorizer = auth.NewAPIKeyAuthorizer("Authorization", fmt.Sprintf("Bearer %s", registry.Credential.AccessKey), auth.APIKeyInHeader)
	}

	if authorizer != nil {
		modifiers = append(modifiers, authorizer)
	}
	nativeRegistryAdapter, err := native.NewAdapterWithCustomizedAuthorizer(registry, authorizer)
	if err != nil {
		return nil, err
	}

	return &adapter{
		Adapter:  nativeRegistryAdapter,
		registry: registry,
		client: common_http.NewClient(
			&http.Client{
				Transport: util.GetHTTPTransport(registry.Insecure),
			},
			modifiers...,
		),
	}, nil
}

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeQuayio,
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

// HealthCheck checks health status of a registry
func (a *adapter) HealthCheck() (model.HealthStatus, error) {
	err := a.PingSimple()
	if err != nil {
		return model.Unhealthy, nil
	}

	return model.Healthy, nil
}

// PrepareForPush does the prepare work that needed for pushing/uploading the resource
// eg: create the namespace or repository
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	namespaces := []string{}
	for _, resource := range resources {
		if resource == nil {
			return errors.New("the resource cannot be null")
		}
		if resource.Metadata == nil {
			return errors.New("the metadata of resource cannot be null")
		}
		if resource.Metadata.Repository == nil {
			return errors.New("the namespace of resource cannot be null")
		}
		if len(resource.Metadata.Repository.Name) == 0 {
			return errors.New("the name of the namespace cannot be null")
		}
		paths := strings.Split(resource.Metadata.Repository.Name, "/")
		namespace := paths[0]
		namespaces = append(namespaces, namespace)
	}

	for _, namespace := range namespaces {
		err := a.createNamespace(&model.Namespace{
			Name: namespace,
		})
		if err != nil {
			return fmt.Errorf("create namespace '%s' in Quay.io error: %v", namespace, err)
		}
		log.Debugf("namespace %s created", namespace)
	}
	return nil
}

// createNamespace creates a new namespace in Quay.io
func (a *adapter) createNamespace(namespace *model.Namespace) error {
	ns, err := a.getNamespace(namespace.Name)
	if err != nil {
		return fmt.Errorf("check existence of namespace '%s' error: %v", namespace.Name, err)
	}

	// If the namespace already exist, return succeeded directly.
	if ns != nil {
		log.Infof("Namespace %s already exist in Quay.io, skip it.", namespace.Name)
		return nil
	}

	org := &orgCreate{
		Name:  namespace.Name,
		Email: namespace.GetStringMetadata("email", namespace.Name),
	}
	b, err := json.Marshal(org)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, buildOrgURL(""), bytes.NewReader(b))
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Errorf("create namespace error: %d -- %s", resp.StatusCode, string(body))
	return fmt.Errorf("%d -- %s", resp.StatusCode, body)
}

// getNamespace get namespace from Quay.io, if the namespace not found, two nil would be returned.
func (a *adapter) getNamespace(namespace string) (*model.Namespace, error) {
	req, err := http.NewRequest(http.MethodGet, buildOrgURL(namespace), nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode/100 != 2 {
		log.Errorf("get namespace error: %d -- %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("%d -- %s", resp.StatusCode, body)
	}

	return &model.Namespace{
		Name: namespace,
	}, nil
}
