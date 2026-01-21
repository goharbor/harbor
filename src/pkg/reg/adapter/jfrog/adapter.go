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

package jfrog

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

func init() {
	err := adp.RegisterFactory(model.RegistryTypeJfrogArtifactory, new(factory))
	if err != nil {
		log.Errorf("failed to register factory for jfrog artifactory: %v", err)
		return
	}
	log.Infof("the factory of jfrog artifactory adapter was registered")

	// if the defaultSupportedTypes doesn't include the types, set it by env JFROG_SUPPORTED_REPOSITORY_TYPES.
	// get supported repository types from environment variable
	types := os.Getenv("JFROG_SUPPORTED_REPOSITORY_TYPES")
	if types != "" {
		supportedTypes = strings.Split(types, ",")
	}
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
	_                     adp.Adapter          = (*adapter)(nil)
	_                     adp.ArtifactRegistry = (*adapter)(nil)
	defaultSupportedTypes                      = []string{"docker", "oci", "helm", "helm-oci"}
	supportedTypes                             = defaultSupportedTypes
)

// Adapter is for images replications between harbor and jfrog artifactory image repository
type adapter struct {
	*native.Adapter
	registry *model.Registry
	client   *client
}

var _ adp.Adapter = (*adapter)(nil)

// Info gets info about jfrog artifactory adapter
func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	info = &model.RegistryInfo{
		Type: model.RegistryTypeJfrogArtifactory,
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
	}
	return
}

func newAdapter(registry *model.Registry) (adp.Adapter, error) {
	return &adapter{
		Adapter:  native.NewAdapter(registry),
		registry: registry,
		client:   newClient(registry),
	}, nil
}

// PrepareForPush creates local docker repository in jfrog artifactory
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	var namespaces []string
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
			return errors.New("the name of namespace cannot be null")
		}
		path := strings.Split(resource.Metadata.Repository.Name, "/")
		if len(path) > 0 {
			namespaces = append(namespaces, path[0])
		}
	}

	repositories, err := a.listAllRepositories()
	if err != nil {
		return err
	}
	existedRepositories := make(map[string]struct{})
	for _, repo := range repositories {
		existedRepositories[repo.Key] = struct{}{}
	}

	for _, namespace := range namespaces {
		if _, ok := existedRepositories[namespace]; ok {
			log.Debugf("Namespace %s already existed in remote, skip create it", namespace)
		} else {
			err := a.client.createDockerRepository(namespace)
			if err != nil {
				log.Errorf("Create Namespace %s error: %v", namespace, err)
				return err
			}
		}
	}

	return nil
}

// FetchArtifacts fetches artifacts from jfrog
func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	repositories, err := a.listRepositories(filters)
	if err != nil {
		return nil, err
	}
	if len(repositories) == 0 {
		return nil, nil
	}

	var rawResources = make([]*model.Resource, len(repositories))
	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)

	for i, r := range repositories {
		index := i
		repo := r
		runner.AddTask(func() error {
			artifacts, err := a.listArtifacts(repo.Name, filters)
			if err != nil {
				return fmt.Errorf("failed to list artifacts of repository %s: %v", repo.Name, err)
			}
			if len(artifacts) == 0 {
				return nil
			}
			rawResources[index] = &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repo.Name,
					},
					Artifacts: artifacts,
				},
			}

			return nil
		})
	}
	if err = runner.Wait(); err != nil {
		return nil, fmt.Errorf("failed to fetch artifacts: %v", err)
	}
	var resources []*model.Resource
	for _, r := range rawResources {
		if r != nil {
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func (a *adapter) listAllRepositories() ([]*repository, error) {
	return a.client.getAllRepositories()
}

// listRepositories lists repositories from jfrog
func (a *adapter) listRepositories(filters []*model.Filter) ([]*model.Repository, error) {
	pattern := ""
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			pattern = filter.Value.(string)
			break
		}
	}
	var totalRepos []string
	// if the pattern of repository name filter is a specific repository name, just returns
	// the parsed repositories and will check the existence later when filtering the tags
	if paths, ok := util.IsSpecificPath(pattern); ok {
		totalRepos = paths
	} else {
		// search repositories from catalog API
		repositories, err := a.listAllRepositories()
		if err != nil {
			return nil, err
		}
		for _, r := range repositories {
			url := buildRepositoryURL(a.client.url, r.PackageType, r.Key)
			regClient := registry.NewClientWithAuthorizer(url, basic.NewAuthorizer(a.client.username, a.client.password), a.client.insecure, a.registry.CACertificate)
			repos, err := regClient.Catalog()
			if err != nil {
				// log the error and continue to the next repository
				log.Debugf("List catalog for repository %s error: %v", r.Key, err)
				continue
			}

			for _, repo := range repos {
				totalRepos = append(totalRepos, fmt.Sprintf("%s/%s", r.Key, repo))
			}
		}
	}

	var result []*model.Repository
	for _, repository := range totalRepos {
		result = append(result, &model.Repository{
			Name: repository,
		})
	}
	return filter.DoFilterRepositories(result, filters)
}

func buildRepositoryURL(baseURL, packageType, repoKey string) string {
	return fmt.Sprintf("%s/artifactory/api/%s/%s", baseURL, strings.ToLower(packageType), repoKey)
}

// listArtifacts lists one repository tags
func (a *adapter) listArtifacts(repository string, filters []*model.Filter) ([]*model.Artifact, error) {
	// split docker registry name and repo name
	key, repoName := "", ""
	s := strings.Split(repository, "/")
	if len(s) > 1 {
		key = s[0]
		repoName = strings.Join(s[1:], "/")
	}

	// because repository can be other types like oci, helm, try to list tags with all supported types
	var tags []string
	for _, t := range supportedTypes {
		url := buildRepositoryURL(a.client.url, t, key)
		regClient := registry.NewClientWithAuthorizer(url, basic.NewAuthorizer(a.client.username, a.client.password), a.client.insecure, a.registry.CACertificate)
		var err error
		tags, err = regClient.ListTags(repoName)
		if err != nil {
			log.Debugf("List tags for repository %s with type %s error: %v", repository, t, err)
			continue
		}
		break
	}

	var artifacts []*model.Artifact
	for _, tag := range tags {
		artifacts = append(artifacts, &model.Artifact{
			Tags: []string{tag},
		})
	}
	return filter.DoFilterArtifacts(artifacts, filters)
}

// PushBlob can not use naive PushBlob due to MonolithicUpload, Jfrog now just support push by chunk
// related issue: https://www.jfrog.com/jira/browse/RTFACT-19344
func (a *adapter) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	location, err := a.preparePushBlob(repository)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/v2/%s/blobs/uploads/%s", a.registry.URL, repository, location)
	req, err := http.NewRequest(http.MethodPatch, url, blob)
	if err != nil {
		return err
	}
	rangeSize := fmt.Sprintf("%d", size)
	req.Header.Set("Content-Length", rangeSize)
	req.Header.Set("Content-Range", fmt.Sprintf("0-%s", rangeSize))
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := a.client.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return a.ackPushBlob(repository, digest, location, rangeSize)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}

func (a *adapter) preparePushBlob(repository string) (string, error) {
	url := fmt.Sprintf("%s/v2/%s/blobs/uploads/", a.registry.URL, repository)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Length", "0")
	resp, err := a.client.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return resp.Header.Get("Docker-Upload-Uuid"), nil
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}

	return "", err
}

func (a *adapter) ackPushBlob(repository, digest, location, _ string) error {
	url := fmt.Sprintf("%s/v2/%s/blobs/uploads/%s?digest=%s", a.registry.URL, repository, location, digest)
	req, err := http.NewRequest(http.MethodPut, url, nil)
	if err != nil {
		return err
	}

	resp, err := a.client.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = &common_http.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}

	return err
}
