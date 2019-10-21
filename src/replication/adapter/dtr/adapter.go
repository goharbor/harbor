package dtr

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeDTR, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry)
	}); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeDTR, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeDTR)
}

type adapter struct {
	*native.Adapter
	registry     *model.Registry
	url          string
	username     string
	token        string
	clientDTRAPI *Client
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	log.Debugf("Entering DTR newAdapter")

	var credential auth.Credential
	if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
		log.Debugf("Setting up DTR for basic auth")
		credential = auth.NewBasicAuthCredential(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret)
	}
	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: util.GetHTTPTransport(registry.Insecure),
	}, credential)

	dockerRegistryAdapter, err := native.NewAdapterWithCustomizedAuthorizer(&model.Registry{
		Name:       registry.Name,
		URL:        registry.URL,
		Credential: registry.Credential,
		Insecure:   registry.Insecure,
	}, authorizer)
	if err != nil {
		return nil, err
	}

	return &adapter{
		registry:     registry,
		url:          registry.URL,
		clientDTRAPI: NewClient(registry),
		Adapter:      dockerRegistryAdapter,
	}, nil
}

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeAzureAcr,
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

// FetchImages ...
func (a *adapter) FetchImages(filters []*model.Filter) ([]*model.Resource, error) {
	repositories, err := a.clientDTRAPI.getRepositories()
	if err != nil {
		log.Error("Failed to lookup repositories from DTR")
		return nil, err
	}
	if len(repositories) == 0 {
		return nil, nil
	}
	log.Debugf("Time to filter")
	for _, filter := range filters {
		if err = filter.DoFilter(&repositories); err != nil {
			return nil, err
		}
	}

	var rawResources = make([]*model.Resource, len(repositories))
	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)
	defer runner.Cancel()

	for i, r := range repositories {
		index := i
		repo := r
		runner.AddTask(func() error {
			vTags, err := a.clientDTRAPI.getVTags(repo.Name)
			if err != nil {
				return fmt.Errorf("List tags for repo '%s' error: %v", repo.Name, err)
			}
			if len(vTags) == 0 {
				return nil
			}
			for _, filter := range filters {
				if err = filter.DoFilter(&vTags); err != nil {
					return fmt.Errorf("Filter tags %v error: %v", vTags, err)
				}
			}
			if len(vTags) == 0 {
				return nil
			}
			tags := []string{}
			for _, vTag := range vTags {
				tags = append(tags, vTag.Name)
			}
			rawResources[index] = &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repo.Name,
					},
					Vtags: tags,
				},
			}

			return nil
		})
	}
	runner.Wait()

	if runner.IsCancelled() {
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
		log.Error("Failed to lookup namespaces from DTR: %v", err)
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
		log.Error("Failed to lookup repositories from DTR: %v", err)
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
