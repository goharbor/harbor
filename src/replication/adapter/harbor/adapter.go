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

package harbor

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	common_http_auth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeHarbor, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry)
	}); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeHarbor, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeHarbor)
}

type adapter struct {
	*adp.DefaultImageRegistry
	registry       *model.Registry
	coreServiceURL string
	client         *common_http.Client
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	transport := util.GetHTTPTransport(registry.Insecure)
	modifiers := []modifier.Modifier{
		&auth.UserAgentModifier{
			UserAgent: adp.UserAgentReplication,
		},
	}
	if registry.Credential != nil {
		var authorizer modifier.Modifier
		if registry.Credential.Type == model.CredentialTypeSecret {
			authorizer = common_http_auth.NewSecretAuthorizer(registry.Credential.AccessSecret)
		} else {
			authorizer = auth.NewBasicAuthCredential(
				registry.Credential.AccessKey,
				registry.Credential.AccessSecret)
		}
		modifiers = append(modifiers, authorizer)
	}

	// The registry URL and core service URL are different when the adapter
	// is created for a local Harbor. If the "registry.CoreURL" is null, the
	// registry URL will be used as the coreServiceURL instead
	url := registry.URL
	if len(registry.CoreURL) > 0 {
		url = registry.CoreURL
	}

	reg, err := adp.NewDefaultImageRegistry(registry)
	if err != nil {
		return nil, err
	}
	return &adapter{
		registry:       registry,
		coreServiceURL: url,
		client: common_http.NewClient(
			&http.Client{
				Transport: transport,
			}, modifiers...),
		DefaultImageRegistry: reg,
	}, nil
}

func (a *adapter) Info() (*model.RegistryInfo, error) {
	info := &model.RegistryInfo{
		Type:             model.RegistryTypeHarbor,
		SupportNamespace: true,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeRepository,
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
			// TODO add support for label filter
			// {
			//	 Type:  model.FilterTypeLabel,
			//	 Style: model.FilterStyleTypeText,
			// },
		},
		SupportedTriggers: []model.TriggerType{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
			model.TriggerTypeEventBased,
		},
	}

	sys := &struct {
		ChartRegistryEnabled bool `json:"with_chartmuseum"`
	}{}
	if err := a.client.Get(a.coreServiceURL+"/api/systeminfo", sys); err != nil {
		return nil, err
	}
	if sys.ChartRegistryEnabled {
		info.SupportedResourceTypes = append(info.SupportedResourceTypes, model.ResourceTypeChart)
	}
	return info, nil
}

func (a *adapter) ListNamespaces(query *model.NamespaceQuery) ([]*model.Namespace, error) {
	var namespaces []*model.Namespace
	name := ""
	if query != nil {
		name = query.Name
	}
	projects, err := a.getProjects(name)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		namespaces = append(namespaces, &model.Namespace{
			Name:     project.Name,
			Metadata: project.Metadata,
		})
	}

	return namespaces, nil
}

func (a *adapter) ConvertResourceMetadata(metadata *model.ResourceMetadata, namespace *model.Namespace) (*model.ResourceMetadata, error) {
	if metadata == nil {
		return nil, errors.New("the metadata cannot be null")
	}
	name := metadata.GetResourceName()
	strs := strings.SplitN(name, "/", 2)
	if len(strs) < 2 {
		return nil, fmt.Errorf("unsupported resource name %s, at least contains one '/'", name)
	}
	meta := &model.ResourceMetadata{
		Vtags:  metadata.Vtags,
		Labels: metadata.Labels,
	}
	meta.Namespace = &model.Namespace{
		Name: strs[0],
	}
	if metadata.Namespace != nil {
		meta.Namespace.Metadata = metadata.Namespace.Metadata
	}
	meta.Repository = &model.Repository{
		Name: strs[1],
	}
	if metadata.Repository != nil {
		meta.Repository.Metadata = metadata.Repository.Metadata
	}
	// replace the namespace if it is specified
	if namespace == nil || len(namespace.Name) == 0 {
		return meta, nil
	}
	if strings.Contains(namespace.Name, "/") {
		return nil, fmt.Errorf("the namespace %s cannot contain '/'", namespace.Name)
	}
	meta.Namespace.Name = namespace.Name
	if namespace.Metadata != nil {
		meta.Namespace.Metadata = namespace.Metadata
	}
	return meta, nil
}
func (a *adapter) PrepareForPush(resource *model.Resource) error {
	if resource == nil {
		return errors.New("the resource cannot be null")
	}
	if resource.Metadata == nil {
		return errors.New("the metadata of resource cannot be null")
	}
	if resource.Metadata.Namespace == nil {
		return errors.New("the namespace of resource cannot be null")
	}
	if len(resource.Metadata.Namespace.Name) == 0 {
		return errors.New("the name of the namespace cannot be null")
	}
	project := &struct {
		Name     string                 `json:"project_name"`
		Metadata map[string]interface{} `json:"metadata"`
	}{
		Name:     resource.Metadata.Namespace.Name,
		Metadata: resource.Metadata.Namespace.Metadata,
	}

	// TODO
	/*
		// handle the public of the project
		if meta, exist := namespace.Metadata["public"]; exist {
			public := true
			// if one of them is "private", the set the public as false
			for _, value := range meta.(map[string]interface{}) {
				b, err := strconv.ParseBool(value.(string))
				if err != nil {
					return err
				}
				if !b {
					public = false
					break
				}
			}
			project.Metadata = map[string]interface{}{
				"public": public,
			}
		}
	*/

	err := a.client.Post(a.coreServiceURL+"/api/projects", project)
	if httpErr, ok := err.(*common_http.Error); ok && httpErr.Code == http.StatusConflict {
		log.Debugf("got 409 when trying to create project %s", resource.Metadata.Namespace.Name)
		return nil
	}
	return err
}

type project struct {
	ID       int64                  `json:"project_id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (a *adapter) getProjects(name string) ([]*project, error) {
	projects := []*project{}
	url := fmt.Sprintf("%s/api/projects?name=%s&page=1&page_size=1000", a.coreServiceURL, name)
	if err := a.client.Get(url, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (a *adapter) getProject(name string) (*project, error) {
	// TODO need an API to exact match project by name
	projects, err := a.getProjects(name)
	if err != nil {
		return nil, err
	}

	for _, pro := range projects {
		if pro.Name == name {
			p := &project{
				ID:   pro.ID,
				Name: name,
			}
			if pro.Metadata != nil {
				metadata := map[string]interface{}{}
				for key, value := range pro.Metadata {
					metadata[key] = value
				}
				p.Metadata = metadata
			}
			return p, nil
		}
	}
	return nil, fmt.Errorf("project %s not found", name)
}
