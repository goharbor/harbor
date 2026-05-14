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

package aliacr

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
	"github.com/goharbor/harbor/src/pkg/registry/auth/bearer"
)

var rateLimiterTransport http.RoundTripper

const acrQPSLimit = 15

func init() {
	var envAcrQPSLimit, _ = strconv.Atoi(os.Getenv("REG_ADAPTER_ACR_QPS_LIMIT"))
	if envAcrQPSLimit > acrQPSLimit || envAcrQPSLimit < 1 {
		envAcrQPSLimit = acrQPSLimit
	}
	rateLimiterTransport = lib.NewRateLimitedTransport(envAcrQPSLimit, commonhttp.NewTransport())

	if err := adp.RegisterFactory(model.RegistryTypeAliAcr, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeAliAcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeAliAcr)
}

// example:
// https://cr.%s.aliyuncs.com
var regACRServiceURL = regexp.MustCompile(`https://cr\.([\w\-]+)\.aliyuncs\.com`)

func getRegistryURL(url string) (string, error) {
	if url == "" {
		return "", errors.New("empty url")
	}
	rs := regACRServiceURL.FindStringSubmatch(url)
	if rs == nil {
		return url, nil
	}
	return fmt.Sprintf(registryEndpointTpl, rs[1]), nil
}

// example:
// registry.aliyuncs.com:cn-hangzhou:china:cri-xxxxxxxxx
// registry.aliyuncs.com:cn-hangzhou:26842
func parseRegistryService(service string) (*registryServiceInfo, error) {
	parts := strings.Split(service, ":")
	length := len(parts)
	if length < 2 {
		return nil, errors.New("invalid service format: expected 'registry.aliyuncs.com:region:xxxxx'")
	}

	if !strings.EqualFold(parts[0], registryACRService) {
		return nil, errors.New("not a acr service")
	}

	if strings.HasPrefix(parts[length-1], "cri-") {
		return &registryServiceInfo{
			IsACREE:    true,
			RegionID:   parts[1],
			InstanceID: parts[length-1],
		}, nil
	}
	return &registryServiceInfo{
		IsACREE:  false,
		RegionID: parts[1],
	}, nil
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	url, err := getRegistryURL(registry.URL)
	if err != nil {
		return nil, err
	}
	registry.URL = url

	realm, service, err := util.Ping(registry)
	if err != nil {
		return nil, err
	}

	info, err := parseRegistryService(service)
	if err != nil {
		return nil, err
	}

	var acrAPI openapi
	if !info.IsACREE {
		acrAPI, err = newAcrOpenapi(registry.Credential.AccessKey, registry.Credential.AccessSecret, info.RegionID, rateLimiterTransport)
		if err != nil {
			return nil, err
		}
	} else {
		acrAPI, err = newAcreeOpenapi(registry.Credential.AccessKey, registry.Credential.AccessSecret, info.RegionID, info.InstanceID, rateLimiterTransport)
		if err != nil {
			return nil, err
		}
	}
	authorizer := bearer.NewAuthorizer(realm, service, NewAuth(acrAPI), commonhttp.GetHTTPTransport(
		commonhttp.WithInsecure(registry.Insecure),
		commonhttp.WithCACert(registry.CACertificate),
	))
	return &adapter{
		acrAPI:   acrAPI,
		registry: registry,
		Adapter:  native.NewAdapterWithAuthorizer(registry, authorizer),
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

// adapter for to aliyun docker registry
type adapter struct {
	*native.Adapter
	acrAPI   openapi
	registry *model.Registry
}

var _ adp.Adapter = &adapter{}

// Info ...
func (a *adapter) Info() (*model.RegistryInfo, error) {
	info := &model.RegistryInfo{
		Type: model.RegistryTypeAliAcr,
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
	return info, nil
}

func getAdapterInfo() *model.AdapterPattern {
	var endpoints []*model.Endpoint
	// https://help.aliyun.com/document_detail/40654.html?spm=a2c4g.11186623.2.7.58683ae5Q4lo1o
	for _, e := range []string{
		"cn-qingdao",
		"cn-beijing",
		"cn-zhangjiakou",
		"cn-huhehaote",
		"cn-wulanchabu",
		"cn-hangzhou",
		"cn-shanghai",
		"cn-shenzhen",
		"cn-heyuan",
		"cn-guangzhou",
		"cn-chengdu",
		"cn-hongkong",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-southeast-3",
		"ap-southeast-5",
		"ap-south-1",
		"ap-northeast-1",
		"us-west-1",
		"us-east-1",
		"eu-central-1",
		"eu-west-1",
		"me-east-1",
	} {
		endpoints = append(endpoints, &model.Endpoint{
			Key:   e,
			Value: fmt.Sprintf("https://registry.%s.aliyuncs.com", e),
		})
		endpoints = append(endpoints, &model.Endpoint{
			Key:   e + "-vpc",
			Value: fmt.Sprintf("https://registry-vpc.%s.aliyuncs.com", e),
		})
		endpoints = append(endpoints, &model.Endpoint{
			Key:   e + "-internal",
			Value: fmt.Sprintf("https://registry-internal.%s.aliyuncs.com", e),
		})

		endpoints = append(endpoints, &model.Endpoint{
			Key:   e + "-ee-vpc",
			Value: fmt.Sprintf("https://instanceName-registry-vpc.%s.cr.aliyuncs.com", e),
		})

		endpoints = append(endpoints, &model.Endpoint{
			Key:   e + "-ee",
			Value: fmt.Sprintf("https://instanceName-registry.%s.cr.aliyuncs.com", e),
		})
	}
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints:    endpoints,
		},
	}
	return info
}

func (a *adapter) listCandidateNamespaces(namespacePattern string) ([]string, error) {
	var namespaces []string
	if len(namespacePattern) > 0 {
		if nms, ok := util.IsSpecificPathComponent(namespacePattern); ok {
			namespaces = append(namespaces, nms...)
		}
		if len(namespaces) > 0 {
			log.Debugf("parsed the namespaces %v from pattern %s", namespaces, namespacePattern)
			return namespaces, nil
		}
	}

	if a.acrAPI == nil {
		return nil, errors.New("acr api is nil")
	}

	return a.acrAPI.ListNamespace()
}

// FetchArtifacts AliACR not support /v2/_catalog of Registry, we'll list all resources via Aliyun's API
func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	log.Debugf("FetchArtifacts.filters: %#v\n", filters)

	if a.acrAPI == nil {
		return nil, errors.New("acr api is nil")
	}

	var resources []*model.Resource
	// get filter pattern
	var repoPattern string
	var tagsPattern string
	for _, f := range filters {
		if f.Type == model.FilterTypeName {
			repoPattern = f.Value.(string)
		}
	}
	var namespacePattern = strings.Split(repoPattern, "/")[0]

	log.Debugf("\nrepoPattern=%s tagsPattern=%s\n\n", repoPattern, tagsPattern)

	// get namespaces
	namespaces, err := a.listCandidateNamespaces(namespacePattern)
	if err != nil {
		return nil, err
	}
	log.Debugf("got namespaces: %v \n", namespaces)

	// list repos
	var repositories []*repository
	for _, namespace := range namespaces {
		repos, err := a.acrAPI.ListRepository(namespace)
		if err != nil {
			return nil, err
		}

		log.Debugf("\nnamespace: %s \t repositories: %#v\n\n", namespace, repos)

		for _, repo := range repos {
			var ok bool
			var repoName = filepath.Join(repo.Namespace, repo.Name)
			ok, err = util.Match(repoPattern, repoName)
			log.Debugf("\n Repository: %s\t repoPattern: %s\t Match: %v\n", repoName, repoPattern, ok)
			if err != nil {
				return nil, err
			}
			if ok {
				repositories = append(repositories, repo)
			}
		}
	}
	log.Debugf("FetchArtifacts.repositories: %#v\n", repositories)

	var rawResources = make([]*model.Resource, len(repositories))
	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)

	for i, r := range repositories {
		index := i
		repo := r
		runner.AddTask(func() error {
			var tags []string
			tags, err = a.acrAPI.ListRepoTag(repo)
			if err != nil {
				return fmt.Errorf("list tags for repo '%s' error: %v", repo.Name, err)
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

			if len(filterArtifacts) > 0 {
				rawResources[index] = &model.Resource{
					Type:     model.ResourceTypeImage,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: filepath.Join(repo.Namespace, repo.Name),
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
	for _, r := range rawResources {
		if r != nil {
			resources = append(resources, r)
		}
	}

	return resources, nil
}
