package dtr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func init() {
	err := adp.RegisterFactory(model.RegistryTypeDTR, new(factory))
	if err != nil {
		log.Errorf("failed to register factory for dtr: %v", err)
		return
	}
	log.Infof("the factory of dtr adapter was registered")
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r), nil
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

var (
	_ adp.Adapter = (*adapter)(nil)
)

type adapter struct {
	*native.Adapter
	registry     *model.Registry
	url          string
	username     string
	token        string
	clientDTRAPI *Client
}

func newAdapter(registry *model.Registry) *adapter {
	return &adapter{
		registry:     registry,
		url:          registry.URL,
		clientDTRAPI: NewClient(registry),
		Adapter:      native.NewAdapter(registry),
	}
}

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeAzureAcr,
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
	}, nil
}

// FetchArtifacts ...
func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	var resources []*model.Resource

	repositories, err := a.clientDTRAPI.getRepositories()
	if err != nil {
		log.Error("Failed to lookup repositories from DTR")
		return nil, err
	}
	if len(repositories) == 0 {
		return nil, nil
	}
	log.Debugf("%d of repositories pre filter", len(repositories))
	repositories, err = filter.DoFilterRepositories(repositories, filters)
	if err != nil {
		return nil, err
	}
	log.Debugf("%d of repositories post filter", len(repositories))

	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)

	for _, r := range repositories {
		repo := r
		runner.AddTask(func() error {
			artifacts, err := a.listArtifacts(repo.Name, filters)
			if err != nil {
				return fmt.Errorf("failed to list artifacts of repository %s: %v", repo.Name, err)
			}
			log.Debugf("%s has %d artifacts", repo.Name, len(artifacts))

			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repo.Name,
					},
					Artifacts: artifacts,
				},
			})
			return nil
		})
	}

	if err = runner.Wait(); err != nil {
		return nil, err
	}

	return resources, nil
}

// PrepareForPush creates docker repository in DTR
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	var dtrNamespaces []Account
	var repos []string
	namespaces := make(map[string]string)
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
			namespaces[path[0]] = path[0]
		}
		if len(resource.Metadata.Repository.Name) > 0 {
			repos = append(repos, resource.Metadata.Repository.Name)
		}
	}

	dtrNamespaces, err := a.clientDTRAPI.getNamespaces()
	if err != nil {
		log.Errorf("Failed to lookup namespaces from DTR: %v", err)
		return err
	}

	existingNamespaces := make(map[string]struct{})
	for _, namespace := range dtrNamespaces {
		existingNamespaces[namespace.Name] = struct{}{}
	}

	for namespace := range namespaces {
		if _, ok := existingNamespaces[namespace]; ok {
			log.Debugf("Namespace %s already existed in remote, skip create it", namespace)
		} else {
			err = a.clientDTRAPI.createNamespace(namespace)
			if err != nil {
				log.Errorf("Create Namespace %s error: %v", namespace, err)
				return err
			}
		}
	}

	repositories, err := a.clientDTRAPI.getRepositories()
	if err != nil {
		log.Errorf("Failed to lookup repositories from DTR: %v", err)
		return err
	}

	existingRepositories := make(map[string]struct{})
	for _, repo := range repositories {
		existingRepositories[repo.Name] = struct{}{}
	}

	for _, repo := range repos {
		if _, ok := existingRepositories[repo]; ok {
			log.Debugf("Repo %s already existed in remote, skip create it", repo)
		} else {
			err = a.clientDTRAPI.createRepository(repo)
			if err != nil {
				log.Errorf("Create Repository %s error: %v", repo, err)
				return err
			}
		}
	}

	return nil
}

func (a *adapter) listArtifacts(repository string, filters []*model.Filter) ([]*model.Artifact, error) {
	tags, err := a.clientDTRAPI.getTags(repository)
	if err != nil {
		return nil, fmt.Errorf("list tags for repo '%s' error: %v", repository, err)
	}
	var artifacts []*model.Artifact
	for _, tag := range tags {
		artifacts = append(artifacts, &model.Artifact{
			Tags: []string{tag},
		})
	}
	return filter.DoFilterArtifacts(artifacts, filters)
}
