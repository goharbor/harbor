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
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common/api"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	common_http_auth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"

	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeHarbor, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeHarbor, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeHarbor)
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

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
	_ adp.ChartRegistry    = (*adapter)(nil)
)

type adapter struct {
	*native.Adapter
	registry *model.Registry
	url      string
	client   *common_http.Client
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	var transport *http.Transport
	if registry.URL == config.GetCoreURL() {
		transport = common_http.GetHTTPTransport(common_http.SecureTransport)
	} else {
		transport = util.GetHTTPTransport(registry.Insecure)
	}
	// local Harbor instance
	if registry.Credential != nil && registry.Credential.Type == model.CredentialTypeSecret {
		authorizer := common_http_auth.NewSecretAuthorizer(registry.Credential.AccessSecret)
		return &adapter{
			registry: registry,
			url:      registry.URL,
			client: common_http.NewClient(
				&http.Client{
					Transport: transport,
				}, authorizer),
			Adapter: native.NewAdapterWithAuthorizer(registry, authorizer),
		}, nil
	}

	var authorizers []modifier.Modifier
	if registry.Credential != nil {
		authorizers = append(authorizers, basic.NewAuthorizer(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret))
	}
	return &adapter{
		registry: registry,
		url:      registry.URL,
		client: common_http.NewClient(
			&http.Client{
				Transport: transport,
			}, authorizers...),
		Adapter: native.NewAdapter(registry),
	}, nil
}

func (a *adapter) Info() (*model.RegistryInfo, error) {
	info := &model.RegistryInfo{
		Type: model.RegistryTypeHarbor,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeArtifact,
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
	}

	sys := &struct {
		ChartRegistryEnabled bool `json:"with_chartmuseum"`
	}{}
	if err := a.client.Get(fmt.Sprintf("%s/api/%s/systeminfo", a.getURL(), api.APIVersion), sys); err != nil {
		return nil, err
	}
	if sys.ChartRegistryEnabled {
		info.SupportedResourceTypes = append(info.SupportedResourceTypes, model.ResourceTypeChart)
	}
	labels := []*struct {
		Name string `json:"name"`
	}{}
	// label isn't supported in some previous version of Harbor
	if err := a.client.Get(fmt.Sprintf("%s/api/%s/labels?scope=g", a.getURL(), api.APIVersion), &labels); err != nil {
		if e, ok := err.(*common_http.Error); !ok || e.Code != http.StatusNotFound {
			return nil, err
		}
	} else {
		ls := []string{}
		for _, label := range labels {
			ls = append(ls, label.Name)
		}
		labelFilter := &model.FilterStyle{
			Type:   model.FilterTypeLabel,
			Style:  model.FilterStyleTypeList,
			Values: ls,
		}
		info.SupportedResourceFilters = append(info.SupportedResourceFilters, labelFilter)
	}
	return info, nil
}

func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	projects := map[string]*project{}
	for _, resource := range resources {
		if resource == nil {
			return errors.New("the resource cannot be null")
		}
		if resource.Metadata == nil {
			return errors.New("the metadata of resource cannot be null")
		}
		if resource.Metadata.Repository == nil {
			return errors.New("the repository of resource cannot be null")
		}
		if len(resource.Metadata.Repository.Name) == 0 {
			return errors.New("the name of the repository cannot be null")
		}

		paths := strings.Split(resource.Metadata.Repository.Name, "/")
		projectName := paths[0]
		// handle the public properties
		metadata := abstractPublicMetadata(resource.Metadata.Repository.Metadata)
		pro, exist := projects[projectName]
		if exist {
			metadata = mergeMetadata(pro.Metadata, metadata)
		}
		projects[projectName] = &project{
			Name:     projectName,
			Metadata: metadata,
		}
	}
	for _, project := range projects {
		pro := struct {
			Name     string                 `json:"project_name"`
			Metadata map[string]interface{} `json:"metadata"`
		}{
			Name:     project.Name,
			Metadata: project.Metadata,
		}
		err := a.client.Post(fmt.Sprintf("%s/api/%s/projects", a.getURL(), api.APIVersion), pro)
		if err != nil {
			if httpErr, ok := err.(*common_http.Error); ok && httpErr.Code == http.StatusConflict {
				log.Debugf("got 409 when trying to create project %s", project.Name)
				continue
			}
			return err
		}
		log.Debugf("project %s created", project.Name)
	}
	return nil
}

func abstractPublicMetadata(metadata map[string]interface{}) map[string]interface{} {
	if metadata == nil {
		return nil
	}
	public, exist := metadata["public"]
	if !exist {
		return nil
	}
	return map[string]interface{}{
		"public": public,
	}
}

// currently, mergeMetadata only handles the public metadata
func mergeMetadata(metadata1, metadata2 map[string]interface{}) map[string]interface{} {
	public := parsePublic(metadata1) && parsePublic(metadata2)
	return map[string]interface{}{
		"public": strconv.FormatBool(public),
	}
}

func parsePublic(metadata map[string]interface{}) bool {
	if metadata == nil {
		return false
	}
	pub, exist := metadata["public"]
	if !exist {
		return false
	}
	public, ok := pub.(bool)
	if ok {
		return public
	}
	pubstr, ok := pub.(string)
	if ok {
		public, err := strconv.ParseBool(pubstr)
		if err != nil {
			log.Errorf("failed to parse %s to bool: %v", pubstr, err)
			return false
		}
		return public
	}
	return false
}

type project struct {
	ID       int64                  `json:"project_id"`
	Name     string                 `json:"name"`
	Metadata map[string]interface{} `json:"metadata"`
}

func (a *adapter) getProjects(name string) ([]*project, error) {
	projects := []*project{}
	url := fmt.Sprintf("%s/api/%s/projects?name=%s&page=1&page_size=500", a.getURL(), api.APIVersion, name)
	if err := a.client.GetAndIteratePagination(url, &projects); err != nil {
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
	return nil, nil
}

// when the adapter is created for local Harbor, returns the "http://127.0.0.1:8080"
// as URL to avoid issue https://github.com/goharbor/harbor-helm/issues/222
// when harbor is deployed on Kubernetes
func (a *adapter) getURL() string {
	if a.registry.Type == model.RegistryTypeHarbor && a.registry.Name == "Local" {
		if common_http.InternalTLSEnabled() {
			return "https://core:8443"
		}
		return "http://127.0.0.1:8080"
	}
	return a.url
}
