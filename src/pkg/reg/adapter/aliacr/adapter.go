package aliacr

import (
	"encoding/json"
	"errors"
	"fmt"

	"path/filepath"
	"regexp"
	"strings"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
	"github.com/goharbor/harbor/src/pkg/registry/auth/bearer"

	commonhttp "github.com/goharbor/harbor/src/common/http"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeAliAcr, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeAliAcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeAliAcr)
}

// example:
// https://registry.%s.aliyuncs.com
// https://cr.%s.aliyuncs.com
var regRegion = regexp.MustCompile("https://(registry|cr)\\.([\\w\\-]+)\\.aliyuncs\\.com")

func getRegion(url string) (region string, err error) {
	if url == "" {
		return "", errors.New("empty url")
	}
	rs := regRegion.FindStringSubmatch(url)
	if rs == nil {
		return "", errors.New("invalid Rgistry|CR service url")
	}
	// fmt.Println(rs)
	return rs[2], nil
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	region, err := getRegion(registry.URL)
	if err != nil {
		return nil, err
	}
	// fix url (allow user input cr service url)
	registry.URL = fmt.Sprintf(registryEndpointTpl, region)
	realm, service, err := util.Ping(registry)
	if err != nil {
		return nil, err
	}
	credential := NewAuth(region, registry.Credential.AccessKey, registry.Credential.AccessSecret)
	authorizer := bearer.NewAuthorizer(realm, service, credential, commonhttp.GetHTTPTransport(commonhttp.WithInsecure(registry.Insecure)))
	return &adapter{
		region:   region,
		registry: registry,
		domain:   fmt.Sprintf(endpointTpl, region),
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
	region   string
	domain   string
	registry *model.Registry
}

var _ adp.Adapter = &adapter{}

// Info ...
func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	info = &model.RegistryInfo{
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
	return
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
	}
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints:    endpoints,
		},
	}
	return info
}

func (a *adapter) listNamespaces(c *cr.Client) (namespaces []string, err error) {
	// list namespaces
	var nsReq = cr.CreateGetNamespaceListRequest()
	var nsResp = cr.CreateGetNamespaceListResponse()
	nsReq.SetDomain(a.domain)
	nsResp, err = c.GetNamespaceList(nsReq)
	if err != nil {
		return
	}
	var resp = &aliACRNamespaceResp{}
	err = json.Unmarshal(nsResp.GetHttpContentBytes(), resp)
	if err != nil {
		return
	}
	for _, ns := range resp.Data.Namespaces {
		namespaces = append(namespaces, ns.Namespace)
	}

	log.Debugf("FetchArtifacts.listNamespaces: %#v\n", namespaces)

	return
}

func (a *adapter) listCandidateNamespaces(c *cr.Client, namespacePattern string) (namespaces []string, err error) {
	if len(namespacePattern) > 0 {
		if nms, ok := util.IsSpecificPathComponent(namespacePattern); ok {
			namespaces = append(namespaces, nms...)
		}
		if len(namespaces) > 0 {
			log.Debugf("parsed the namespaces %v from pattern %s", namespaces, namespacePattern)
			return namespaces, nil
		}
	}

	return a.listNamespaces(c)
}

// FetchArtifacts AliACR not support /v2/_catalog of Registry, we'll list all resources via Aliyun's API
func (a *adapter) FetchArtifacts(filters []*model.Filter) (resources []*model.Resource, err error) {
	log.Debugf("FetchArtifacts.filters: %#v\n", filters)

	var client *cr.Client
	client, err = cr.NewClientWithAccessKey(a.region, a.registry.Credential.AccessKey, a.registry.Credential.AccessSecret)
	if err != nil {
		return
	}

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
	var namespaces []string
	namespaces, err = a.listCandidateNamespaces(client, namespacePattern)
	if err != nil {
		return
	}
	log.Debugf("got namespaces: %v \n", namespaces)

	// list repos
	var repositories []aliRepo
	for _, namespace := range namespaces {
		var repos []aliRepo
		repos, err = a.listReposByNamespace(a.region, namespace, client)
		if err != nil {
			return
		}

		log.Debugf("\nnamespace: %s \t repositories: %#v\n\n", namespace, repos)

		for _, repo := range repos {
			var ok bool
			var repoName = filepath.Join(repo.RepoNamespace, repo.RepoName)
			ok, err = util.Match(repoPattern, repoName)
			log.Debugf("\n Repository: %s\t repoPattern: %s\t Match: %v\n", repoName, repoPattern, ok)
			if err != nil {
				return
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
			tags, err = a.getTags(repo, client)
			if err != nil {
				return fmt.Errorf("list tags for repo '%s' error: %v", repo.RepoName, err)
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
							Name: filepath.Join(repo.RepoNamespace, repo.RepoName),
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

	return
}

func (a *adapter) listReposByNamespace(region string, namespace string, c *cr.Client) (repos []aliRepo, err error) {
	var reposReq = cr.CreateGetRepoListByNamespaceRequest()
	var reposResp = cr.CreateGetRepoListByNamespaceResponse()
	reposReq.SetDomain(a.domain)
	reposReq.RepoNamespace = namespace
	var page = 1
	for {
		reposReq.Page = requests.NewInteger(page)
		reposResp, err = c.GetRepoListByNamespace(reposReq)
		if err != nil {
			return
		}
		var resp = &aliReposResp{}
		err = json.Unmarshal(reposResp.GetHttpContentBytes(), resp)
		if err != nil {
			return
		}
		repos = append(repos, resp.Data.Repos...)

		if resp.Data.Total-(resp.Data.Page*resp.Data.PageSize) <= 0 {
			break
		}
		page++
	}
	return
}

func (a *adapter) getTags(repo aliRepo, c *cr.Client) (tags []string, err error) {
	log.Debugf("[ali-acr.getTags]%s: %#v\n", a.domain, repo)
	var tagsReq = cr.CreateGetRepoTagsRequest()
	var tagsResp = cr.CreateGetRepoTagsResponse()
	tagsReq.SetDomain(a.domain)
	tagsReq.RepoNamespace = repo.RepoNamespace
	tagsReq.RepoName = repo.RepoName
	var page = 1
	for {
		tagsReq.Page = requests.NewInteger(page)
		tagsResp, err = c.GetRepoTags(tagsReq)
		if err != nil {
			return
		}

		var resp = &aliTagResp{}
		json.Unmarshal(tagsResp.GetHttpContentBytes(), resp)
		for _, tag := range resp.Data.Tags {
			tags = append(tags, tag.Tag)
		}

		if resp.Data.Total-(resp.Data.Page*resp.Data.PageSize) <= 0 {
			break
		}
		page++
	}

	return
}
