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

package tencentcr

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
	"github.com/goharbor/harbor/src/pkg/registry/auth/bearer"
)

var (
	errInvalidTcrEndpoint = errors.New("[tencent-tcr.newAdapter] Invalid TCR instance endpoint")
	rateLimiterTransport  http.RoundTripper
)

func init() {
	var envTcrQPSLimit, _ = strconv.Atoi(os.Getenv("REG_ADAPTER_TCR_QPS_LIMIT"))
	if envTcrQPSLimit > tcrQPSLimit || envTcrQPSLimit < 1 {
		envTcrQPSLimit = tcrQPSLimit
	}
	rateLimiterTransport = lib.NewRateLimitedTransport(envTcrQPSLimit, commonhttp.NewTransport())

	if err := adp.RegisterFactory(model.RegistryTypeTencentTcr, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeTencentTcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeTencentTcr)
}

type factory struct{}

/**
	* Implement Factory Interface
**/
var _ adp.Factory = &factory{}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return getAdapterInfo()
}

func getAdapterInfo() *model.AdapterPattern {
	return &model.AdapterPattern{}
}

type adapter struct {
	*native.Adapter
	registryID *string
	regionName *string
	tcrClient  *tcr.Client
	pageSize   *int64
	client     *commonhttp.Client
	registry   *model.Registry
}

/**
	* Implement Adapter Interface
**/
var _ adp.Adapter = &adapter{}

func newAdapter(registry *model.Registry) (a *adapter, err error) {
	if !isSecretID(registry.Credential.AccessKey) {
		err = errors.New("[tencent-tcr.newAdapter] Please use SecretId/SecretKey, NOT docker login Username/Password")
		log.Debugf("[tencent-tcr.newAdapter] error=%v", err)
		return
	}

	// Query TCR instance info via endpoint.
	var registryURL *url.URL
	registryURL, _ = url.Parse(registry.URL)

	// only validate registryURL.Host in non-UT scenario
	if os.Getenv("UTTEST") != "true" {
		if !strings.Contains(registryURL.Host, ".tencentcloudcr.com") {
			log.Errorf("[tencent-tcr.newAdapter] errInvalidTcrEndpoint=%v", err)
			return nil, errInvalidTcrEndpoint
		}
	}

	realm, service, err := util.Ping(registry)
	log.Debugf("[tencent-tcr.newAdapter] realm=%s, service=%s error=%v", realm, service, err)
	if err != nil {
		log.Errorf("[tencent-tcr.newAdapter] ping failed. error=%v", err)
		return
	}

	// Create TCR API client
	var tcrCredential = common.NewCredential(registry.Credential.AccessKey, registry.Credential.AccessSecret)
	var cfp = profile.NewClientProfile()
	var client *tcr.Client
	// temp client used to get TCR instance info
	client, err = tcr.NewClient(tcrCredential, regions.Guangzhou, cfp)
	if err != nil {
		return
	}

	var req = tcr.NewDescribeInstancesRequest()
	req.AllRegion = common.BoolPtr(true)
	req.Filters = []*tcr.Filter{
		{
			Name:   common.StringPtr("RegistryName"),
			Values: []*string{common.StringPtr(strings.ReplaceAll(registryURL.Host, ".tencentcloudcr.com", ""))},
		},
	}
	var resp *tcr.DescribeInstancesResponse
	resp, err = client.DescribeInstances(req)
	if err != nil {
		log.Errorf("DescribeInstances error=%s", err.Error())
		return
	}
	if *resp.Response.TotalCount == 0 {
		err = fmt.Errorf("[tencent-tcr.newAdapter] Can not get TCR instance info. RequestId=%s", *resp.Response.RequestId)
		return
	}
	var instanceInfo = resp.Response.Registries[0]
	log.Debugf("[tencent-tcr.InstanceInfo] registry.URL=%s, host=%s, PublicDomain=%s, RegionName=%s, RegistryId=%s",
		registry.URL, registryURL.Host, *instanceInfo.PublicDomain, *instanceInfo.RegionName, *instanceInfo.RegistryId)

	// rebuild TCR SDK client
	client = &tcr.Client{}
	client.Init(*instanceInfo.RegionName).
		WithCredential(tcrCredential).
		WithProfile(cfp).
		WithHttpTransport(rateLimiterTransport)

	var credential = NewAuth(instanceInfo.RegistryId, client)
	var transport = commonhttp.GetHTTPTransport(commonhttp.WithInsecure(registry.Insecure))
	var authorizer = bearer.NewAuthorizer(realm, service, credential, transport)

	return &adapter{
		registry:   registry,
		registryID: instanceInfo.RegistryId,
		regionName: instanceInfo.RegionName,
		tcrClient:  client,
		pageSize:   common.Int64Ptr(20),
		client: commonhttp.NewClient(
			&http.Client{
				Transport: transport,
			},
			credential,
		),
		Adapter: native.NewAdapterWithAuthorizer(registry, authorizer),
	}, nil
}

func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	info = &model.RegistryInfo{
		Type: model.RegistryTypeTencentTcr,
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

func (a *adapter) PrepareForPush(resources []*model.Resource) (err error) {
	log.Debugf("[tencent-tcr.PrepareForPush]")
	for _, resource := range resources {
		if resource == nil {
			return errors.New("the resource cannot be null")
		}
		if resource.Metadata == nil {
			return errors.New("[tencent-tcr.PrepareForPush] the metadata of resource cannot be null")
		}
		if resource.Metadata.Repository == nil {
			return errors.New("[tencent-tcr.PrepareForPush] the namespace of resource cannot be null")
		}
		if len(resource.Metadata.Repository.Name) == 0 {
			return errors.New("[tencent-tcr.PrepareForPush] the name of the namespace cannot be null")
		}
		var paths = strings.Split(resource.Metadata.Repository.Name, "/")
		var namespace = paths[0]
		var repository = path.Join(paths[1:]...)

		log.Debugf("[tencent-tcr.PrepareForPush.createPrivateNamespace] namespace=%s", namespace)
		err = a.createPrivateNamespace(namespace)
		if err != nil {
			return
		}
		log.Debugf("[tencent-tcr.PrepareForPush.createRepository] namespace=%s, repository=%s", namespace, repository)
		err = a.createRepository(namespace, repository)
		if err != nil {
			return
		}
	}

	return
}
