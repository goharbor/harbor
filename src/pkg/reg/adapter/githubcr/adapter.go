package githubcr

import (
	"errors"
	"fmt"
	"net/http"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
	"github.com/goharbor/harbor/src/pkg/registry/auth/basic"
)

// !!!! Limits:
// - GHCR not support `/v2/_catalog`, access this API will return 404 status code.
// - NOT support DELETE manifest.

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeGithubCR, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeGithubCR, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeGithubCR)
}

type factory struct{}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r), nil
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return getAdapterPattern()
}

func getAdapterPattern() *model.AdapterPattern {
	return &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeFix,
			Endpoints: []*model.Endpoint{
				{
					Key:   "ghcr.io",
					Value: "https://ghcr.io",
				},
			},
		},
	}
}

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

// adapter for to github container registry
type adapter struct {
	client *common_http.Client
	*native.Adapter
	registry *model.Registry
}

var _ adp.Adapter = &adapter{}

func newAdapter(registry *model.Registry) *adapter {
	var authorizer modifier.Modifier
	if registry.Credential != nil {
		authorizer = basic.NewAuthorizer(
			registry.Credential.AccessKey,
			registry.Credential.AccessSecret)
	}

	var transport = common_http.GetHTTPTransport(common_http.WithInsecure(registry.Insecure))

	return &adapter{
		Adapter:  native.NewAdapter(registry),
		registry: registry,
		client: common_http.NewClient(
			&http.Client{
				Transport: transport,
			},
			authorizer,
		),
	}
}

// Info ...
func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	info = &model.RegistryInfo{
		Type: model.RegistryTypeGithubCR,
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

func (a *adapter) FetchArtifacts(filters []*model.Filter) (resources []*model.Resource, err error) {
	pattern := ""
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			pattern = filter.Value.(string)
			break
		}
	}
	var repositories []string
	// if the pattern of repository name filter is a specific repository name, just returns
	// the parsed repositories and will check the existence later when filtering the tags
	if paths, ok := util.IsSpecificPath(pattern); ok {
		repositories = paths
	} else {
		err = errors.New("only support specific repository name")
		return
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

			artifacts, err := a.listArtifacts(repo, filters)
			if err != nil {
				return fmt.Errorf("failed to list artifacts of repository %s: %v", repo, err)
			}
			if len(artifacts) == 0 {
				return nil
			}
			rawResources[index] = &model.Resource{
				Type:     model.ResourceTypeImage,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repo,
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

	for _, r := range rawResources {
		if r != nil {
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func (a *adapter) listArtifacts(repository string, filters []*model.Filter) ([]*model.Artifact, error) {
	tags, err := a.ListTags(repository)
	if err != nil {
		return nil, err
	}
	var artifacts []*model.Artifact
	for _, tag := range tags {
		artifacts = append(artifacts, &model.Artifact{
			Tags: []string{tag},
		})
	}
	return filter.DoFilterArtifacts(artifacts, filters)
}

func (a *adapter) DeleteManifest(repository, reference string) error {
	return nil
}
