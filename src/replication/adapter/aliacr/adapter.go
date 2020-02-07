package aliacr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
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
// for aliacr Enterprise edition
// https://(%s)registry.%s.cr.aliyuncs.com
var regRegion = regexp.MustCompile("https://.*(registry|cr)\\.([\\w\\-]+)\\..*aliyuncs\\.com")

func getRegion(url string) (region string, err error) {
	if url == "" {
		return "", errors.New("empty url")
	}
	rs := regRegion.FindStringSubmatch(url)
	if rs == nil {
		return "", errors.New("Invalid Rgistry|CR service url")
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

	credential := NewAuth(region, registry.Credential.AccessKey, registry.Credential.AccessSecret)
	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: util.GetHTTPTransport(registry.Insecure),
	}, credential)
	nativeRegistry, err := native.NewAdapterWithCustomizedAuthorizer(registry, authorizer)
	if err != nil {
		return nil, err
	}

	return &adapter{
		region:   region,
		registry: registry,
		domain:   fmt.Sprintf(endpointTpl, region),
		Adapter:  nativeRegistry,
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
	}
	return
}

func getAdapterInfo() *model.AdapterPattern {
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints: []*model.Endpoint{
				{Key: "cn-hangzhou", Value: "https://registry.cn-hangzhou.aliyuncs.com"},
				{Key: "cn-shanghai", Value: "https://registry.cn-shanghai.aliyuncs.com"},
				{Key: "cn-qingdao", Value: "https://registry.cn-qingdao.aliyuncs.com"},
				{Key: "cn-beijing", Value: "https://registry.cn-beijing.aliyuncs.com"},
				{Key: "cn-zhangjiakou", Value: "https://registry.cn-zhangjiakou.aliyuncs.com"},
				{Key: "cn-huhehaote", Value: "https://registry.cn-huhehaote.aliyuncs.com"},
				{Key: "cn-shenzhen", Value: "https://registry.cn-shenzhen.aliyuncs.com"},
				{Key: "cn-chengdu", Value: "https://registry.cn-chengdu.aliyuncs.com"},
				{Key: "cn-hongkong", Value: "https://registry.cn-hongkong.aliyuncs.com"},
				{Key: "ap-southeast-1", Value: "https://registry.ap-southeast-1.aliyuncs.com"},
				{Key: "ap-southeast-2", Value: "https://registry.ap-southeast-2.aliyuncs.com"},
				{Key: "ap-southeast-3", Value: "https://registry.ap-southeast-3.aliyuncs.com"},
				{Key: "ap-southeast-5", Value: "https://registry.ap-southeast-5.aliyuncs.com"},
				{Key: "ap-northeast-1", Value: "https://registry.ap-northeast-1.aliyuncs.com"},
				{Key: "ap-south-1", Value: "https://registry.ap-south-1.aliyuncs.com"},
				{Key: "eu-central-1", Value: "https://registry.eu-central-1.aliyuncs.com"},
				{Key: "eu-west-1", Value: "https://registry.eu-west-1.aliyuncs.com"},
				{Key: "us-west-1", Value: "https://registry.us-west-1.aliyuncs.com"},
				{Key: "us-east-1", Value: "https://registry.us-east-1.aliyuncs.com"},
				{Key: "me-east-1", Value: "https://registry.me-east-1.aliyuncs.com"},
			},
		},
	}
	return info
}

// FetchImages AliACR not support /v2/_catalog of Registry, we'll list all resources via Aliyun's API
func (a *adapter) FetchImages(filters []*model.Filter) (resources []*model.Resource, err error) {
	log.Debugf("FetchImages.filters: %#v\n", filters)

	var client *cr.Client
	client, err = cr.NewClientWithAccessKey(a.region, a.registry.Credential.AccessKey, a.registry.Credential.AccessSecret)
	if err != nil {
		return
	}

	// get filter pattern
	var repoPattern string
	var tagsPattern string
	for _, filter := range filters {
		if filter.Type == model.FilterTypeName {
			repoPattern = filter.Value.(string)
		}
		if filter.Type == model.FilterTypeTag {
			tagsPattern = filter.Value.(string)
		}
	}

	// list repos
	var repositories []aliRepo
	for {
		var repoListResp *aliRepoResp
		repoListResp, err = a.listRepo(a.region, client)
		if err != nil {
			return
		}
		if repoPattern != "" {
			for _, repo := range repoListResp.Data.Repos {

				var ok bool
				ok, err = util.Match(repoPattern, filepath.Join(repo.RepoNamespace, repo.RepoName))
				if err != nil {
					return
				}
				if ok {
					repositories = append(repositories, repo)
				}
			}
		} else {
			repositories = append(repositories, repoListResp.Data.Repos...)
		}

		if repoListResp.Data.Total-(repoListResp.Data.Page*repoListResp.Data.PageSize) <= 0 {
			break
		}
	}
	log.Debugf("FetchImages.repositories: %#v\n", repositories)

	var rawResources = make([]*model.Resource, len(repositories))
	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)
	defer runner.Cancel()

	for i, r := range repositories {
		index := i
		repo := r
		runner.AddTask(func() error {
			var tags []string
			tags, err = a.getTags(repo, client)
			if err != nil {
				return fmt.Errorf("List tags for repo '%s' error: %v", repo.RepoName, err)
			}

			var filterTags []string
			if tagsPattern != "" {
				for _, tag := range tags {
					var ok bool
					ok, err = util.Match(tagsPattern, tag)
					if err != nil {
						return fmt.Errorf("Match tag '%s' error: %v", tag, err)
					}
					if ok {
						filterTags = append(filterTags, tag)
					}
				}
			} else {
				filterTags = tags
			}

			if len(filterTags) > 0 {
				rawResources[index] = &model.Resource{
					Type:     model.ResourceTypeImage,
					Registry: a.registry,
					Metadata: &model.ResourceMetadata{
						Repository: &model.Repository{
							Name: filepath.Join(repo.RepoNamespace, repo.RepoName),
						},
						Vtags:  filterTags,
						Labels: []string{},
					},
				}
			}

			return nil
		})
	}
	runner.Wait()

	if runner.IsCancelled() {
		return nil, fmt.Errorf("FetchImages error when collect tags for repos")
	}

	for _, r := range rawResources {
		if r != nil {
			resources = append(resources, r)
		}
	}

	return
}

func (a *adapter) listRepo(region string, c *cr.Client) (resp *aliRepoResp, err error) {
	var reposReq = cr.CreateGetRepoListRequest()
	var reposResp = cr.CreateGetRepoListResponse()
	reposReq.SetDomain(a.domain)
	reposResp, err = c.GetRepoList(reposReq)
	if err != nil {
		return
	}
	resp = &aliRepoResp{}
	json.Unmarshal(reposResp.GetHttpContentBytes(), resp)

	return
}

func (a *adapter) getTags(repo aliRepo, c *cr.Client) (tags []string, err error) {
	log.Debugf("[ali-acr.getTags]%s: %#v\n", a.domain, repo)
	var tagsReq = cr.CreateGetRepoTagsRequest()
	var tagsResp = cr.CreateGetRepoTagsResponse()
	tagsReq.SetDomain(a.domain)
	tagsReq.RepoNamespace = repo.RepoNamespace
	tagsReq.RepoName = repo.RepoName
	for {
		fmt.Printf("[GetRepoTags.req] %#v\n", tagsReq)
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
	}

	return
}
