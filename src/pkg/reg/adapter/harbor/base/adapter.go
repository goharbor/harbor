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

package base

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	common_http_auth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

// New creates an instance of the base adapter
func New(registry *model.Registry) (*Adapter, error) {
	if isLocalHarbor(registry.URL) {
		authorizer := common_http_auth.NewSecretAuthorizer(registry.Credential.AccessSecret)
		httpClient := common_http.NewClient(&http.Client{
			// when it's a local Harbor instance, the code runs inside the same process with
			// core, so insecure transport is ok
			// If using the secure one, as we'll replace the URL with 127.0.0.1 and this will
			// cause error "x509: cannot validate certificate for 127.0.0.1 because it doesn't contain any IP SANs"
			Transport: common_http.GetHTTPTransport(common_http.WithInsecure(true)),
		}, authorizer)
		client, err := NewClient(registry.URL, httpClient)
		if err != nil {
			return nil, err
		}
		return &Adapter{
			Adapter:    native.NewAdapterWithAuthorizer(registry, authorizer),
			Registry:   registry,
			Client:     client,
			httpClient: httpClient,
		}, nil
	}

	var authorizers []modifier.Modifier
	if registry.Credential != nil {
		authorizers = append(authorizers, basic.NewAuthorizer(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret))
	}
	httpClient := common_http.NewClient(&http.Client{
		Transport: common_http.GetHTTPTransport(common_http.WithInsecure(registry.Insecure)),
	}, authorizers...)
	client, err := NewClient(registry.URL, httpClient)
	if err != nil {
		return nil, err
	}
	return &Adapter{
		Adapter:    native.NewAdapter(registry),
		Registry:   registry,
		Client:     client,
		httpClient: httpClient,
	}, nil
}

// Adapter is the base adapter for Harbor
type Adapter struct {
	*native.Adapter
	Registry *model.Registry
	Client   *Client

	httpClient *common_http.Client
}

// GetAPIVersion returns the supported API version of the Harbor instance that the adapter is created for
func (a *Adapter) GetAPIVersion() string {
	return a.Client.APIVersion
}

// Info provides the information of the Harbor registry instance
func (a *Adapter) Info() (*model.RegistryInfo, error) {
	info := &model.RegistryInfo{
		Type: model.RegistryTypeHarbor,
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
		SupportedRepositoryPathComponentType: model.RepositoryPathComponentTypeAtLeastTwo,
		SupportedCopyByChunk:                 true,
	}

	labels, err := a.Client.ListLabels()
	if err != nil {
		return nil, err
	}
	info.SupportedResourceFilters = append(info.SupportedResourceFilters,
		&model.FilterStyle{
			Type:   model.FilterTypeLabel,
			Style:  model.FilterStyleTypeList,
			Values: labels,
		})

	return info, nil
}

// PrepareForPush creates projects
func (a *Adapter) PrepareForPush(resources []*model.Resource) error {
	projects := map[string]*Project{}
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
		projects[projectName] = &Project{
			Name:     projectName,
			Metadata: metadata,
		}
	}

	// Create a list of the project names.
	var ps []string
	for p := range projects {
		// Surround name in 'quotes' to force the server to parse as a string.
		// Handles the case where a project name consists entirely of numbers.
		ps = append(ps, fmt.Sprintf("'%s'", p))
	}
	// query by project names
	q := fmt.Sprintf("name={%s}", strings.Join(ps, " "))
	// get exist projects
	queryProjects, err := a.Client.ListProjectsWithQuery(q, false)
	if err != nil {
		return errors.Wrapf(err, "list projects with query %s", q)
	}

	proxyCacheProjects := make(map[string]bool)
	existProjects := make(map[string]bool)
	for _, p := range queryProjects {
		existProjects[p.Name] = true
		// if project with registry_id, that means this is a proxy cache project.
		if p.RegistryID > 0 {
			proxyCacheProjects[p.Name] = true
		}
	}

	var notExistProjects []*Project
	for _, p := range projects {
		if !existProjects[p.Name] {
			notExistProjects = append(notExistProjects, p)
		}
	}

	for _, project := range notExistProjects {
		if err := a.Client.CreateProject(project.Name, project.Metadata); err != nil {
			if httpErr, ok := err.(*common_http.Error); ok && httpErr.Code == http.StatusConflict {
				log.Debugf("got 409 when trying to create project %s", project.Name)
				continue
			}
			return err
		}
		log.Debugf("project %s created", project.Name)
	}

	// do filter for proxy cache projects.
	for _, res := range resources {
		paths := strings.Split(res.Metadata.Repository.Name, "/")
		projectName := paths[0]
		if proxyCacheProjects[projectName] {
			// set resource skip flag to true if it's a proxy cache project.
			res.Skip = true
		}
	}

	return nil
}

// ListProjects lists projects
func (a *Adapter) ListProjects(filters []*model.Filter) ([]*Project, error) {
	pattern := ""
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			pattern = filter.Value.(string)
			break
		}
	}
	var projects []*Project
	if len(pattern) > 0 {
		substrings := strings.Split(pattern, "/")
		projectPattern := substrings[0]
		names, ok := util.IsSpecificPathComponent(projectPattern)
		if ok {
			for _, name := range names {
				// trim white space in project name
				name = strings.TrimSpace(name)
				project, err := a.Client.GetProject(name)
				if err != nil {
					return nil, err
				}
				if project == nil {
					continue
				}
				projects = append(projects, project)
			}
		}
	}
	if len(projects) > 0 {
		var names []string
		for _, project := range projects {
			names = append(names, project.Name)
		}
		log.Debugf("parsed the projects %v from pattern %s", names, pattern)
		return projects, nil
	}
	return a.Client.ListProjects("")
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

// Project model
type Project struct {
	ID         int64                  `json:"project_id"`
	Name       string                 `json:"name"`
	Metadata   map[string]interface{} `json:"metadata"`
	RegistryID int64                  `json:"registry_id"`
}

func isLocalHarbor(url string) bool {
	return url == os.Getenv("CORE_URL")
}

// check whether the current process is running inside core
func isInCore() bool {
	return len(os.Getenv("EXT_ENDPOINT")) > 0
}
