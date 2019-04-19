package dockerhub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeDockerHub, func(registry *model.Registry) (adp.Adapter, error) {
		registry.URL = baseURL
		client, err := NewClient(registry)
		if err != nil {
			return nil, err
		}
		reg, err := adp.NewDefaultImageRegistry(&model.Registry{
			Name:       registry.Name,
			URL:        registryURL,
			Credential: registry.Credential,
			Insecure:   registry.Insecure,
		})
		if err != nil {
			return nil, err
		}

		return &adapter{
			client:               client,
			registry:             registry,
			DefaultImageRegistry: reg,
		}, nil
	}); err != nil {
		log.Errorf("Register adapter factory for %s error: %v", model.RegistryTypeDockerHub, err)
		return
	}
	log.Infof("Factory for adapter %s registered", model.RegistryTypeDockerHub)
}

type adapter struct {
	*adp.DefaultImageRegistry
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

// ListNamespaces lists namespaces from DockerHub with the provided query conditions.
func (a *adapter) ListNamespaces(query *model.NamespaceQuery) ([]*model.Namespace, error) {
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
	var result []*model.Namespace
	for _, ns := range namespaces.Namespaces {
		// If query set, skip the namespace that doesn't match the query.
		if query != nil && len(query.Name) > 0 && strings.Index(ns, query.Name) != -1 {
			continue
		}

		result = append(result, &model.Namespace{
			Name: ns,
		})
	}
	return result, nil
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

	namespaces, err := a.ListNamespaces(nil)
	if err != nil {
		return nil, err
	}
	for _, ns := range namespaces {
		page := 1
		pageSize := 100
		for {
			pageRepos, err := a.getRepos(ns.Name, "", page, pageSize)
			if err != nil {
				return nil, fmt.Errorf("get repos for namespace '%s' from DockerHub error: %v", ns, err)
			}
			repos = append(repos, pageRepos.Repos...)

			if len(pageRepos.Next) == 0 {
				break
			}
			page++
		}
	}

	log.Infof("%d repos found for namespaces: %v", len(repos), namespaces)
	var resources []*model.Resource
	// TODO(ChenDe): Get tags for repos in parallel
	for _, repo := range repos {
		// If name filter set, skip repos that don't match the filter pattern.
		if len(nameFilter) != 0 {
			m, err := util.Match(nameFilter, repo.Name)
			if err != nil {
				return nil, fmt.Errorf("match repo name '%s' against pattern '%s' error: %v", repo.Name, nameFilter, err)
			}

			if !m {
				continue
			}
		}

		var tags []string
		page := 1
		pageSize := 100
		for {
			pageTags, err := a.getTags(repo.Namespace, repo.Name, page, pageSize)
			if err != nil {
				return nil, fmt.Errorf("get tags for repo '%s/%s' from DockerHub error: %v", repo.Namespace, repo.Name, err)
			}
			for _, t := range pageTags.Tags {
				// If tag filter set, skip tags that don't match the filter pattern.
				if len(tagFilter) != 0 {
					m, err := util.Match(tagFilter, t.Name)
					if err != nil {
						return nil, fmt.Errorf("match tag name '%s' against pattern '%s' error: %v", t.Name, tagFilter, err)
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

		// If the repo has no tags, skip it
		if len(tags) == 0 {
			continue
		}

		resources = append(resources, &model.Resource{
			Type:     model.ResourceTypeImage,
			Registry: a.registry,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: fmt.Sprintf("%s/%s", repo.Namespace, repo.Name),
				},
				Vtags: tags,
			},
		})
	}

	return resources, nil
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
