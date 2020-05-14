package aliacree

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/registry/auth/bearer"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr_ee"
	"github.com/docker/distribution/registry/client/auth/challenge"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeAliAcrEE, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeAliAcrEE, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeAliAcrEE)
}

// example:
// https://%s-registry.%s.cr.aliyuncs.com
// https://%s-registry-vpc.%s.cr.aliyuncs.com
var regInstanceRegion = regexp.MustCompile("https://([\\w\\-._]+)-(registry|registry-vpc)\\.([\\w\\-]+)\\.cr\\.aliyuncs\\.com")

func getInstanceRegion(url string) (instance, region string, err error) {
	if url == "" {
		return "", "", errors.New("empty url")
	}
	rs := regInstanceRegion.FindStringSubmatch(url)
	if rs == nil || len(rs) != 4 {
		return "", "", errors.New("invalid Aliyun Registry Enterprise edition service url")
	}

	return rs[1], rs[3], nil
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	instanceName, region, err := getInstanceRegion(registry.URL)
	if err != nil {
		return nil, err
	}

	instanceID, err := getInstanceID(region, instanceName, registry.Credential.AccessKey, registry.Credential.AccessSecret)
	if err != nil {
		return nil, err
	}
	realm, service, err := ping(registry)
	if err != nil {
		return nil, err
	}
	credential := NewAuth(region, instanceID, registry.Credential.AccessKey, registry.Credential.AccessSecret)
	authorizer := bearer.NewAuthorizer(realm, service, credential, util.GetHTTPTransport(registry.Insecure))

	return &adapter{
		region:     region,
		instanceID: instanceID,
		registry:   registry,
		domain:     fmt.Sprintf(endpointTpl, region),
		Adapter:    native.NewAdapterWithAuthorizer(registry, authorizer),
		accessKey:  registry.Credential.AccessKey,
		secretKey:  registry.Credential.AccessSecret,
	}, nil
}

func ping(registry *model.Registry) (string, string, error) {
	client := &http.Client{}
	if registry.Insecure {
		client.Transport = commonhttp.GetHTTPTransport(commonhttp.InsecureTransport)
	} else {
		client.Transport = commonhttp.GetHTTPTransport(commonhttp.SecureTransport)
	}

	resp, err := client.Get(registry.URL + "/v2/")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	challenges := challenge.ResponseChallenges(resp)
	for _, challenge := range challenges {
		if challenge.Scheme == "bearer" {
			return challenge.Parameters["realm"], challenge.Parameters["service"], nil
		}
	}
	return "", "", fmt.Errorf("bearer auth scheme isn't supported: %v", challenges)
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

// adapter for to aliyun docker registry enterprise edition
type adapter struct {
	*native.Adapter
	instanceID           string //  instanceID is obligatory
	region               string
	domain               string
	registry             *model.Registry
	accessKey, secretKey string
}

var _ adp.Adapter = &adapter{}

// Info ...
func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	info = &model.RegistryInfo{
		Type: model.RegistryTypeAliAcrEE,
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
				{Key: "cn-hangzhou", Value: "https://{instanceName}-registry.cn-hangzhou.cr.aliyuncs.com"},
				{Key: "cn-shanghai", Value: "https://{instanceName}-registry.cn-shanghai.cr.aliyuncs.com"},
				{Key: "cn-qingdao", Value: "https://{instanceName}-registry.cn-qingdao.cr.aliyuncs.com"},
				{Key: "cn-beijing", Value: "https://{instanceName}-registry.cn-beijing.cr.aliyuncs.com"},
			},
		},
	}
	return info
}

// getInstanceID get ali-acr-ee instanceID via aliyun open-api by given parameters.
func getInstanceID(region, instanceName, ak, sk string) (string, error) {
	client, err := cr_ee.NewClientWithAccessKey(region, ak, sk)
	if err != nil {
		return "", err
	}

	req := cr_ee.CreateListInstanceRequest()
	req.Domain = fmt.Sprintf(endpointTpl, region)
	req.InstanceName = instanceName
	req.InstanceStatus = "RUNNING"

	resp, err := client.ListInstance(req)
	if err != nil {
		return "", err
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("ali response indicates not success, %v", resp.String())
	}

	var instanceID string
	for _, ins := range resp.Instances {
		if ins.InstanceName == instanceName {
			instanceID = ins.InstanceId
			break
		}
	}
	if instanceID == "" {
		return "", fmt.Errorf("cannot retrieve instanceID from aliyun")
	}

	return instanceID, nil
}

// FetchArtifacts AliACR not support /v2/_catalog of Registry, we'll list all resources via Aliyun's API
func (a *adapter) FetchArtifacts(filters []*model.Filter) (resources []*model.Resource, err error) {
	log.Debugf("FetchArtifacts.filters: %#v\n", filters)

	var client *cr_ee.Client
	client, err = cr_ee.NewClientWithAccessKey(a.region, a.registry.Credential.AccessKey, a.registry.Credential.AccessSecret)
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

	var selectedRepos, reposItem []cr_ee.RepositoriesItem
	reposItem, err = a.listAllRepos(client)
	if err != nil {
		return
	}
	if repoPattern != "" {
		for _, repo := range reposItem {
			var ok bool
			ok, err = util.Match(repoPattern, filepath.Join(repo.RepoNamespaceName, repo.RepoName))
			if err != nil {
				return
			}
			if ok {
				selectedRepos = append(selectedRepos, repo)
			}
		}
	} else {
		selectedRepos = append(selectedRepos, reposItem...)
	}

	log.Debugf("FetchArtifacts.selectedRepos: %#v\n", selectedRepos)

	var rawResources = make([]*model.Resource, len(selectedRepos))
	runner := utils.NewLimitedConcurrentRunner(adp.MaxConcurrency)
	defer runner.Cancel()

	for i, r := range selectedRepos {
		index := i
		repo := r
		runner.AddTask(func() error {
			var tags []string
			tags, err = a.getTags(repo, client)
			if err != nil {
				return fmt.Errorf("list tags for repo '%s' error: %v", repo.RepoName, err)
			}

			var filterTags []string
			if tagsPattern != "" {
				for _, tag := range tags {
					var ok bool
					ok, err = util.Match(tagsPattern, tag)
					if err != nil {
						return fmt.Errorf("match tag '%s' error: %v", tag, err)
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
							Name: filepath.Join(repo.RepoNamespaceName, repo.RepoName),
						},
						Vtags:  filterTags,
					},
				}
			}

			return nil
		})
	}
	runner.Wait()

	if runner.IsCancelled() {
		return nil, fmt.Errorf("FetchArtifacts error when collect tags for repos")
	}

	for _, r := range rawResources {
		if r != nil {
			resources = append(resources, r)
		}
	}

	return
}

func (a *adapter) listAllRepos(c *cr_ee.Client) ([]cr_ee.RepositoriesItem, error) {
	req := cr_ee.CreateListRepositoryRequest()
	req.Domain = fmt.Sprintf(endpointTpl, a.region)
	req.InstanceId = a.instanceID
	req.PageNo = "1"
	req.PageSize = "9999" // max repos occupancy is 9999 for ali-acr-ee
	req.RepoStatus = "ALL"

	resp, err := c.ListRepository(req)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ali response indicates not success, %v", resp.String())
	}

	return resp.Repositories, nil
}

func (a *adapter) getTags(repo cr_ee.RepositoriesItem, c *cr_ee.Client) (tags []string, err error) {
	log.Debugf("[ali-acr-ee.getTags]%s: %#v\n", a.domain, repo)
	var tagResponse *cr_ee.ListRepoTagResponse
	for {
		tagResponse, err = a.getRepoTags(repo.RepoId, c)
		if err != nil {
			return
		}
		for _, img := range tagResponse.Images {
			tags = append(tags, img.Tag)
		}

		var totalCount int
		totalCount, err = strconv.Atoi(tagResponse.TotalCount)
		if err != nil {
			return
		}
		if totalCount-(tagResponse.PageNo*tagResponse.PageSize) <= 0 {
			break
		}
	}
	return
}

// HealthCheck check healthy status for ali-acr-ee.
func (a *adapter) HealthCheck() (model.HealthStatus, error) {
	client, err := cr_ee.NewClientWithAccessKey(a.region, a.accessKey, a.secretKey)
	if err != nil {
		return model.Unhealthy, err
	}
	// the purpose of instance details getting is healthy checking
	req := cr_ee.CreateGetInstanceRequest()
	req.Domain = fmt.Sprintf(endpointTpl, a.region)
	req.InstanceId = a.instanceID
	resp, err := client.GetInstance(req)
	if err != nil {
		return model.Unhealthy, err
	}
	if !resp.IsSuccess() {
		return model.Unhealthy, fmt.Errorf("ali-acr-ee response unsuccessful")
	}
	return model.Healthy, nil
}

func (a *adapter) getRepoTags(repoID string, c *cr_ee.Client) (*cr_ee.ListRepoTagResponse, error) {
	req := cr_ee.CreateListRepoTagRequest()
	req.Domain = fmt.Sprintf(endpointTpl, a.region)
	req.InstanceId = a.instanceID
	req.RepoId = repoID
	req.PageNo = "1"
	req.PageSize = "100"
	resp, err := c.ListRepoTag(req)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ali response indicates not success, %v", resp.String())
	}

	return resp, nil
}
