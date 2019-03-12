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
	"fmt"
	"net/http"
	"strconv"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/utils/log"
	registry_pkg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

func init() {
	// TODO add more information to the info
	info := &adp.Info{
		Type:                   model.RegistryTypeHarbor,
		SupportedResourceTypes: []model.ResourceType{model.ResourceTypeRepository},
	}
	// TODO passing coreServiceURL and tokenServiceURL
	coreServiceURL := "http://core:8080"
	tokenServiceURL := ""
	if err := adp.RegisterFactory(info, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry, coreServiceURL, tokenServiceURL), nil
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

// The registry URL and core service URL are different when the adapter
// is created for a local Harbor. If the "coreServicrURL" is null, the
// registry URL will be used as the coreServiceURL instead
func newAdapter(registry *model.Registry, coreServiceURL string,
	tokenServiceURL string) *adapter {
	transport := registry_pkg.GetHTTPTransport(registry.Insecure)
	modifiers := []modifier.Modifier{
		&auth.UserAgentModifier{
			UserAgent: adp.UserAgentReplicator,
		},
	}
	if registry.Credential != nil {
		authorizer := auth.NewBasicAuthCredential(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret)
		modifiers = append(modifiers, authorizer)
	}

	url := registry.URL
	if len(coreServiceURL) > 0 {
		url = coreServiceURL
	}

	return &adapter{
		registry:       registry,
		coreServiceURL: url,
		client: common_http.NewClient(
			&http.Client{
				Transport: transport,
			}, modifiers...),
		DefaultImageRegistry: adp.NewDefaultImageRegistry(registry, tokenServiceURL),
	}
}

// TODO implement the function
func (a *adapter) ListNamespaces(*model.NamespaceQuery) ([]*model.Namespace, error) {
	return nil, nil
}
func (a *adapter) CreateNamespace(namespace *model.Namespace) error {
	project := &struct {
		Name     string                 `json:"project_name"`
		Metadata map[string]interface{} `json:"metadata"`
	}{
		Name: namespace.Name,
	}
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

	err := a.client.Post(a.coreServiceURL+"/api/projects", project)
	if httpErr, ok := err.(*common_http.Error); ok && httpErr.Code == http.StatusConflict {
		log.Debugf("got 409 when trying to create project %s", namespace.Name)
		return nil
	}
	return err
}
func (a *adapter) GetNamespace(namespace string) (*model.Namespace, error) {
	project, err := a.getProject(namespace)
	if err != nil {
		return nil, err
	}
	return &model.Namespace{
		Name:     namespace,
		Metadata: project.Metadata,
	}, nil
}

// TODO implement filter
func (a *adapter) FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	resources := []*model.Resource{}
	for _, namespace := range namespaces {
		project, err := a.getProject(namespace)
		if err != nil {
			return nil, err
		}
		repositories := []*repository{}
		url := fmt.Sprintf("%s/api/repositories?project_id=%d", a.coreServiceURL, project.ID)
		if err = a.client.Get(url, &repositories); err != nil {
			return nil, err
		}

		for _, repository := range repositories {
			url := fmt.Sprintf("%s/api/repositories/%s/tags", a.coreServiceURL, repository.Name)
			tags := []*tag{}
			if err = a.client.Get(url, &tags); err != nil {
				return nil, err
			}
			vtags := []string{}
			for _, tag := range tags {
				vtags = append(vtags, tag.Name)
			}
			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeRepository,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Namespace: namespace,
					Name:      repository.Name,
					Vtags:     vtags,
				},
			})
		}
	}

	return resources, nil
}

type project struct {
	ID       int64                  `json:"project_id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

type repository struct {
	Name string `json:"name"`
}

type tag struct {
	Name string `json:"name"`
}

func (a *adapter) getProject(name string) (*project, error) {
	// TODO need an API to exact match project by name
	projects := []*project{}
	url := fmt.Sprintf("%s/api/projects?name=%s&page=1&page_size=1000", a.coreServiceURL, name)
	if err := a.client.Get(url, &projects); err != nil {
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
