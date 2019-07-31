package aliacr

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"sync"

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
	if err := adp.RegisterFactory(model.RegistryTypeAliAcr, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry)
	}); err != nil {
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

	rawResources := make([]*model.Resource, len(repositories))
	var wg = new(sync.WaitGroup)
	var stopped = make(chan struct{})
	var passportsPool = utils.NewPassportsPool(adp.MaxConcurrency, stopped)

	for i, r := range repositories {
		wg.Add(1)
		go func(index int, repo aliRepo) {
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

			var tags []string
			tags, err = a.getTags(repo, client)
			if err != nil {
				log.Errorf("List tags for repo '%s' error: %v", repo.RepoName, err)
				if !utils.IsChannelClosed(stopped) {
					close(stopped)
				}
				return
			}

			var filterTags []string
			if tagsPattern != "" {
				for _, tag := range tags {
					var ok bool
					ok, err = util.Match(tagsPattern, tag)
					if err != nil {
						log.Errorf("Match tag '%s' error: %v", tag, err)
						if !utils.IsChannelClosed(stopped) {
							close(stopped)
						}
						return
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
			} else {
				rawResources[index] = nil
			}
		}(i, r)
	}
	wg.Wait()

	if utils.IsChannelClosed(stopped) {
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
