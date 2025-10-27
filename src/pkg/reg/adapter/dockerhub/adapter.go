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

package dockerhub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeDockerHub, new(factory)); err != nil {
		log.Errorf("Register adapter factory for %s error: %v", model.RegistryTypeDockerHub, err)
		return
	}
	log.Infof("Factory for adapter %s registered", model.RegistryTypeDockerHub)
}

func newAdapter(registry *model.Registry) (adp.Adapter, error) {
	client, err := NewClient(registry)
	if err != nil {
		return nil, err
	}

	return &adapter{
		client:   client,
		registry: registry,
		Adapter: native.NewAdapter(&model.Registry{
			URL:        registryURL,
			Credential: registry.Credential,
			Insecure:   registry.Insecure,
		}),
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

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

type adapter struct {
	*native.Adapter
	registry *model.Registry
	client   *Client
}

// Ensure '*adapter' implements interface 'Adapter'.
var _ adp.Adapter = (*adapter)(nil)

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeDockerHub,
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
		SupportedRepositoryPathComponentType: model.RepositoryPathComponentTypeOnlyTwo,
	}, nil
}

func getAdapterInfo() *model.AdapterPattern {
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeFix,
			Endpoints: []*model.Endpoint{
				{
					Key:   "hub.docker.com",
					Value: "https://hub.docker.com",
				},
			},
		},
	}
	return info
}

// Rate-limit aware wrapper function for client.Do()
// - Avoids being hit by limit by pausing requests when less than 'lowMark' requests remaining.
// - Pauses for given time when limit is hit.
// - Allows 2 more attempts before giving up.
// Reason: Observed (02/2024) penalty for hitting the limit is 120s, normal reset is 60s,
// so it is better to not hit the wall.
func (a *adapter) limitAwareDo(method string, path string, body io.Reader) (*http.Response, error) {
	const lowMark = 8
	var attemptsLeft = 3
	for attemptsLeft > 0 {
		clientResp, clientErr := a.client.Do(method, path, body)
		if clientErr != nil {
			return clientResp, clientErr
		}
		if clientResp.StatusCode != http.StatusTooManyRequests {
			reqsLeft, err := strconv.ParseInt(clientResp.Header.Get("x-ratelimit-remaining"), 10, 64)
			if err != nil {
				return clientResp, clientErr
			}
			if reqsLeft < lowMark {
				resetTSC, err := strconv.ParseInt(clientResp.Header.Get("x-ratelimit-reset"), 10, 64)
				if err == nil {
					dur := time.Until(time.Unix(resetTSC, 0))
					log.Infof("Rate-limit exhaustion eminent, sleeping for %.1f seconds", dur.Seconds())
					time.Sleep(dur)
					log.Info("Sleep finished, resuming operation")
				}
			}
			return clientResp, clientErr
		}
		var dur = time.Duration(0)
		seconds, err := strconv.ParseInt(clientResp.Header.Get("retry-after"), 10, 64)
		if err != nil {
			expireTime, err := http.ParseTime(clientResp.Header.Get("retry-after"))
			if err != nil {
				return nil, errors.New("blocked by dockerhub rate-limit and missing retry-after header")
			}
			dur = time.Until(expireTime)
		} else {
			dur = time.Duration(seconds) * time.Second
		}
		log.Infof("Rate-limit exhausted, sleeping for %.1f seconds", dur.Seconds())
		time.Sleep(dur)
		log.Info("Sleep finished, resuming operation")
		attemptsLeft--
	}
	return nil, errors.New("unable to get past dockerhub rate-limit")
}

// PrepareForPush does the prepare work that needed for pushing/uploading the resource
// eg: create the namespace or repository
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	namespaces := map[string]struct{}{}
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
		namespaces[namespace] = struct{}{}
	}

	for namespace := range namespaces {
		err := a.CreateNamespace(&model.Namespace{
			Name: namespace,
		})
		if err != nil {
			return fmt.Errorf("create namespace '%s' in DockerHub error: %v", namespace, err)
		}
		log.Debugf("namespace %s created", namespace)
	}
	return nil
}

func (a *adapter) listNamespaces() ([]string, error) {
	resp, err := a.limitAwareDo(http.MethodGet, listNamespacePath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		log.Errorf("list namespace error: %s", string(body))
		return nil, fmt.Errorf("%s", string(body))
	}

	namespaces := NamespacesResp{}
	err = json.Unmarshal(body, &namespaces)
	if err != nil {
		return nil, err
	}
	log.Debugf("got namespaces %v by calling the listing namespaces API", namespaces)
	return namespaces.Namespaces, nil
}

// CreateNamespace creates a new namespace in DockerHub
func (a *adapter) CreateNamespace(namespace *model.Namespace) error {
	ns, err := a.getNamespace(namespace.Name)
	if err != nil {
		return fmt.Errorf("check existence of namespace '%s' error: %v", namespace.Name, err)
	}

	// If the namespace already exist, return succeeded directly.
	if ns != nil {
		log.Infof("Namespace %s already exist in DockerHub, skip it.", namespace.Name)
		return nil
	}

	req := &NewOrgReq{
		Name:     namespace.Name,
		FullName: namespace.GetStringMetadata(metadataKeyFullName, namespace.Name),
		Company:  namespace.GetStringMetadata(metadataKeyCompany, namespace.Name),
	}
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := a.limitAwareDo(http.MethodPost, createNamespacePath, bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		log.Errorf("create namespace error: %d -- %s", resp.StatusCode, string(body))
		return fmt.Errorf("%d -- %s", resp.StatusCode, body)
	}

	return nil
}

// getNamespace get namespace from DockerHub, if the namespace not found, two nil would be returned.
func (a *adapter) getNamespace(namespace string) (*model.Namespace, error) {
	resp, err := a.limitAwareDo(http.MethodGet, getNamespacePath(namespace), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
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

// FetchArtifacts fetches images
func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	var repos []Repo
	nameFilter, err := a.getStringFilterValue(model.FilterTypeName, filters)
	if err != nil {
		return nil, err
	}

	namespaces, err := a.listCandidateNamespaces(nameFilter)
	if err != nil {
		return nil, err
	}
	log.Debugf("got %d namespaces", len(namespaces))
	for _, ns := range namespaces {
		page := 1
		pageSize := 100
		n := 0
		for {
			pageRepos, err := a.getRepos(ns, "", page, pageSize)
			if err != nil {
				return nil, fmt.Errorf("get repos for namespace '%s' from DockerHub error: %v", ns, err)
			}
			repos = append(repos, pageRepos.Repos...)

			n += len(pageRepos.Repos)
			if len(pageRepos.Next) == 0 {
				break
			}

			page++
		}
		log.Debugf("got %d repositories for namespace %s", n, ns)
	}

	var rawResources = make([]*model.Resource, len(repos))
	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)
	for i, r := range repos {
		index := i
		repo := r
		runner.AddTask(func() error {
			name := fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)
			log.Debugf("Routine started to collect tags for repo: %s", name)

			// If name filter set, skip repos that don't match the filter pattern.
			if len(nameFilter) != 0 {
				m, err := util.Match(nameFilter, name)
				if err != nil {
					return fmt.Errorf("match repo name '%s' against pattern '%s' error: %v", name, nameFilter, err)
				}
				if !m {
					return nil
				}
			}

			var tags []string
			page := 1
			pageSize := 100
			for {
				pageTags, err := a.getTags(repo.Namespace, repo.Name, page, pageSize)
				if err != nil {
					return fmt.Errorf("get tags for repo '%s/%s' from DockerHub error: %v", repo.Namespace, repo.Name, err)
				}
				for _, t := range pageTags.Tags {
					tags = append(tags, t.Name)
				}

				if len(pageTags.Next) == 0 {
					break
				}
				page++
			}

			var artifacts []*model.Artifact
			for _, tag := range tags {
				artifacts = append(artifacts, &model.Artifact{
					Tags: []string{tag},
				})
			}
			filterArtifacts, err := filter.DoFilterArtifacts(artifacts, filters)
			if err != nil {
				return err
			}

			if len(tags) > 0 {
				rawResources[index] = &model.Resource{
					Type:     model.ResourceTypeImage,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: name,
						},
						Artifacts: filterArtifacts,
					},
				}
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

func (a *adapter) listCandidateNamespaces(pattern string) ([]string, error) {
	namespaces := []string{}
	if len(pattern) > 0 {
		substrings := strings.Split(pattern, "/")
		namespacePattern := substrings[0]
		if nms, ok := util.IsSpecificPathComponent(namespacePattern); ok {
			namespaces = append(namespaces, nms...)
		}
	}
	if len(namespaces) > 0 {
		log.Debugf("parsed the namespaces %v from pattern %s", namespaces, pattern)
		return namespaces, nil
	}
	return a.listNamespaces()
}

// DeleteManifest ...
// Note: DockerHub only supports delete by tag
func (a *adapter) DeleteManifest(repository, reference string) error {
	parts := strings.Split(repository, "/")
	if len(parts) != 2 {
		return fmt.Errorf("dockerhub only support repo in format <namespace>/<name>, but got: %s", repository)
	}

	resp, err := a.limitAwareDo(http.MethodDelete, deleteTagPath(parts[0], parts[1], reference), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		log.Errorf("Delete tag error: %d -- %s", resp.StatusCode, string(body))
		return fmt.Errorf("%d -- %s", resp.StatusCode, string(body))
	}

	return nil
}

// getRepos gets a page of repos from DockerHub
func (a *adapter) getRepos(namespace, name string, page, pageSize int) (*ReposResp, error) {
	resp, err := a.limitAwareDo(http.MethodGet, listReposPath(namespace, name, page, pageSize), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		log.Errorf("list repos error: %d -- %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("%d -- %s", resp.StatusCode, string(body))
	}

	repos := &ReposResp{}
	err = json.Unmarshal(body, repos)
	if err != nil {
		return nil, fmt.Errorf("unmarshal repos list %s error: %v", string(body), err)
	}

	return repos, nil
}

// getTags gets a page of tags for a repo from DockerHub
func (a *adapter) getTags(namespace, repo string, page, pageSize int) (*TagsResp, error) {
	resp, err := a.limitAwareDo(http.MethodGet, listTagsPath(namespace, repo, page, pageSize), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		log.Errorf("list tags error: %d -- %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("%d -- %s", resp.StatusCode, body)
	}

	tags := &TagsResp{}
	err = json.Unmarshal(body, tags)
	if err != nil {
		return nil, fmt.Errorf("unmarshal tags list %s error: %v", string(body), err)
	}

	return tags, nil
}

// getStringFilterValue gets specific type filter value from filters list.
func (a *adapter) getStringFilterValue(filterType string, filters []*model.Filter) (string, error) {
	for _, f := range filters {
		if f.Type == filterType {
			v, ok := f.Value.(string)
			if !ok {
				msg := fmt.Sprintf("expect filter value to be string, but got: %v", f.Value)
				log.Error(msg)
				return "", errors.New(msg)
			}
			return v, nil
		}
	}
	return "", nil
}
