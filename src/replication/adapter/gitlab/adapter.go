package gitlab

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
	"net/http"
	"strings"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeGitLab, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeGitLab, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeGitLab)
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

type adapter struct {
	*native.Adapter
	registry        *model.Registry
	url             string
	username        string
	token           string
	clientGitlabAPI *Client
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	var credential auth.Credential
	if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
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
		registry:        registry,
		url:             registry.URL,
		clientGitlabAPI: NewClient(registry),
		Adapter:         dockerRegistryAdapter,
	}, nil
}

func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeGitLab,
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

// FetchImages fetches images
func (a *adapter) FetchImages(filters []*model.Filter) ([]*model.Resource, error) {
	var resources []*model.Resource
	var projects []*Project
	var err error
	pattern := ""
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			pattern = filter.Value.(string)
			break
		}
	}

	if len(pattern) > 0 {
		substrings := strings.Split(pattern, "/")
		projectPattern := substrings[1]
		names, ok := util.IsSpecificPathComponent(projectPattern)
		if ok {
			for _, name := range names {
				var projectsByName, err = a.clientGitlabAPI.getProjectsByName(name)
				if err != nil {
					return nil, err
				}
				if projectsByName == nil {
					continue
				}
				projects = append(projects, projectsByName...)
			}
		}
	}
	if len(projects) == 0 {
		projects, err = a.clientGitlabAPI.getProjects()
		if err != nil {
			return nil, err
		}
	}
	var pathPatterns []string

	if paths, ok := util.IsSpecificPath(pattern); ok {
		pathPatterns = paths
	}

	for _, project := range projects {
		if !existPatterns(project.FullPath, pathPatterns) {
			continue
		}
		repositories, err := a.clientGitlabAPI.getRepositories(project.ID)
		if err != nil {
			return nil, err
		}
		if len(repositories) == 0 {
			continue
		}
		for _, repository := range repositories {
			if !existPatterns(repository.Path, pathPatterns) {
				continue
			}
			vTags, err := a.clientGitlabAPI.getTags(project.ID, repository.ID)
			if err != nil {
				return nil, err
			}
			if len(vTags) == 0 {
				continue
			}
			tags := []string{}
			for _, vTag := range vTags {
				if !existPatterns(vTag.Path, pathPatterns) {
					continue
				}
				tags = append(tags, vTag.Name)
			}
			info := make(map[string]interface{})
			info["location"] = repository.Location
			info["path"] = repository.Path

			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name:     strings.ToLower(repository.Path),
						Metadata: info,
					},
					Vtags: tags,
				},
			})
		}
	}
	return resources, nil
}

func existPatterns(path string, patterns []string) bool {
	correct := false
	if len(patterns) > 0 {
		for _, pathPattern := range patterns {
			if strings.HasPrefix(strings.ToLower(path), strings.ToLower(pathPattern)) {
				correct = true
				break
			}
		}
	} else {
		correct = true
	}
	return correct
}
