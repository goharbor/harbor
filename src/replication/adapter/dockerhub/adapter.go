package dockerhub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeDockerHub, factory); err != nil {
		log.Errorf("Register adapter factory for %s error: %v", model.RegistryTypeDockerHub, err)
		return
	}
	log.Infof("Factory for adapter %s registered", model.RegistryTypeDockerHub)
}

func factory(registry *model.Registry) (adp.Adapter, error) {
	client, err := NewClient(registry)
	if err != nil {
		return nil, err
	}

	dockerRegistryAdapter, err := native.NewAdapter(&model.Registry{
		URL:        registryURL,
		Credential: registry.Credential,
		Insecure:   registry.Insecure,
	})
	if err != nil {
		return nil, err
	}

	return &adapter{
		client:   client,
		registry: registry,
		Adapter:  dockerRegistryAdapter,
	}, nil
}

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
	resp, err := a.client.Do(http.MethodGet, listNamespacePath, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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

	resp, err := a.client.Do(http.MethodPost, createNamespacePath, bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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
	resp, err := a.client.Do(http.MethodGet, getNamespacePath(namespace), nil)
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

// FetchImages fetches images
func (a *adapter) FetchImages(filters []*model.Filter) ([]*model.Resource, error) {
	var repos []Repo
	nameFilter, err := a.getStringFilterValue(model.FilterTypeName, filters)
	if err != nil {
		return nil, err
	}
	tagFilter, err := a.getStringFilterValue(model.FilterTypeTag, filters)
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
	var wg = new(sync.WaitGroup)
	var stopped = make(chan struct{})
	var passportsPool = utils.NewPassportsPool(adp.MaxConcurrency, stopped)

	for i, r := range repos {
		wg.Add(1)
		go func(index int, repo Repo) {
			defer func() {
				wg.Done()
			}()

			// Return false means no passport acquired, and no valid passport will be dispatched any more.
			// For example, some crucial errors happened and all tasks should be cancelled.
			if ok := passportsPool.Apply(); !ok {
				return
			}
			defer func() {
				passportsPool.Revoke()
			}()

			name := fmt.Sprintf("%s/%s", repo.Namespace, repo.Name)
			log.Infof("Routine started to collect tags for repo: %s", name)

			// If name filter set, skip repos that don't match the filter pattern.
			if len(nameFilter) != 0 {
				m, err := util.Match(nameFilter, name)
				if err != nil {
					if !utils.IsChannelClosed(stopped) {
						close(stopped)
					}
					log.Errorf("match repo name '%s' against pattern '%s' error: %v", name, nameFilter, err)
					return
				}
				if !m {
					return
				}
			}

			var tags []string
			page := 1
			pageSize := 100
			for {
				pageTags, err := a.getTags(repo.Namespace, repo.Name, page, pageSize)
				if err != nil {
					if !utils.IsChannelClosed(stopped) {
						close(stopped)
					}
					log.Errorf("get tags for repo '%s/%s' from DockerHub error: %v", repo.Namespace, repo.Name, err)
					return
				}
				for _, t := range pageTags.Tags {
					// If tag filter set, skip tags that don't match the filter pattern.
					if len(tagFilter) != 0 {
						m, err := util.Match(tagFilter, t.Name)
						if err != nil {
							if !utils.IsChannelClosed(stopped) {
								close(stopped)
							}
							log.Errorf("match tag name '%s' against pattern '%s' error: %v", t.Name, tagFilter, err)
							return
						}

						if !m {
							continue
						}
					}
					tags = append(tags, t.Name)
				}

				if len(pageTags.Next) == 0 {
					break
				}
				page++
			}

			if len(tags) == 0 {
				rawResources[index] = nil
			} else {
				rawResources[index] = &model.Resource{
					Type:     model.ResourceTypeImage,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: name,
						},
						Vtags: tags,
					},
				}
			}
		}(i, r)
	}
	wg.Wait()

	if utils.IsChannelClosed(stopped) {
		return nil, fmt.Errorf("FetchImages error when collect tags for repos")
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

	resp, err := a.client.Do(http.MethodDelete, deleteTagPath(parts[0], parts[1], reference), nil)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
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
	resp, err := a.client.Do(http.MethodGet, listReposPath(namespace, name, page, pageSize), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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
	resp, err := a.client.Do(http.MethodGet, listTagsPath(namespace, repo, page, pageSize), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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

// getFilter gets specific type filter value from filters list.
func (a *adapter) getStringFilterValue(filterType model.FilterType, filters []*model.Filter) (string, error) {
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
