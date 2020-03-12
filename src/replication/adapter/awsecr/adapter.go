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

package awsecr

import (
	"errors"
	"github.com/goharbor/harbor/src/internal"
	"net/http"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsecrapi "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/native"
	"github.com/goharbor/harbor/src/replication/model"
)

const (
	regionPattern = "https://(?:api|\\d+\\.dkr)\\.ecr\\.([\\w\\-]+)\\.amazonaws\\.com"
)

var (
	regionRegexp = regexp.MustCompile(regionPattern)
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeAwsEcr, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeAwsEcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeAwsEcr)
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	region, err := parseRegion(registry.URL)
	if err != nil {
		return nil, err
	}
	authorizer := NewAuth(region, registry.Credential.AccessKey, registry.Credential.AccessSecret, registry.Insecure)
	return &adapter{
		registry: registry,
		Adapter:  native.NewAdapterWithAuthorizer(registry, authorizer),
		region:   region,
	}, nil
}

func parseRegion(url string) (string, error) {
	rs := regionRegexp.FindStringSubmatch(url)
	if rs == nil {
		return "", errors.New("Bad aws url")
	}
	return rs[1], nil
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

type adapter struct {
	*native.Adapter
	registry      *model.Registry
	region        string
	forceEndpoint *string
}

func (*adapter) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeAwsEcr,
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

func getAdapterInfo() *model.AdapterPattern {
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints: []*model.Endpoint{
				{
					Key:   "ap-northeast-1",
					Value: "https://api.ecr.ap-northeast-1.amazonaws.com",
				},
				{
					Key:   "us-east-1",
					Value: "https://api.ecr.us-east-1.amazonaws.com",
				},
				{
					Key:   "us-east-2",
					Value: "https://api.ecr.us-east-2.amazonaws.com",
				},
				{
					Key:   "us-west-1",
					Value: "https://api.ecr.us-west-1.amazonaws.com",
				},
				{
					Key:   "us-west-2",
					Value: "https://api.ecr.us-west-2.amazonaws.com",
				},
				{
					Key:   "ap-east-1",
					Value: "https://api.ecr.ap-east-1.amazonaws.com",
				},
				{
					Key:   "ap-south-1",
					Value: "https://api.ecr.ap-south-1.amazonaws.com",
				},
				{
					Key:   "ap-northeast-2",
					Value: "https://api.ecr.ap-northeast-2.amazonaws.com",
				},
				{
					Key:   "ap-southeast-1",
					Value: "https://api.ecr.ap-southeast-1.amazonaws.com",
				},
				{
					Key:   "ap-southeast-2",
					Value: "https://api.ecr.ap-southeast-2.amazonaws.com",
				},
				{
					Key:   "ca-central-1",
					Value: "https://api.ecr.ca-central-1.amazonaws.com",
				},
				{
					Key:   "eu-central-1",
					Value: "https://api.ecr.eu-central-1.amazonaws.com",
				},
				{
					Key:   "eu-west-1",
					Value: "https://api.ecr.eu-west-1.amazonaws.com",
				},
				{
					Key:   "eu-west-2",
					Value: "https://api.ecr.eu-west-2.amazonaws.com",
				},
				{
					Key:   "eu-west-3",
					Value: "https://api.ecr.eu-west-3.amazonaws.com",
				},
				{
					Key:   "eu-north-1",
					Value: "https://api.ecr.eu-north-1.amazonaws.com",
				},
				{
					Key:   "sa-east-1",
					Value: "https://api.ecr.sa-east-1.amazonaws.com",
				},
			},
		},
	}
	return info
}

// HealthCheck checks health status of a registry
func (a *adapter) HealthCheck() (model.HealthStatus, error) {
	if a.registry.Credential == nil ||
		len(a.registry.Credential.AccessKey) == 0 || len(a.registry.Credential.AccessSecret) == 0 {
		log.Errorf("no credential to ping registry %s", a.registry.URL)
		return model.Unhealthy, nil
	}
	if err := a.Ping(); err != nil {
		log.Errorf("failed to ping registry %s: %v", a.registry.URL, err)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}

// PrepareForPush nothing need to do.
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	for _, resource := range resources {
		if resource == nil {
			return errors.New("the resource cannot be nil")
		}
		if resource.Metadata == nil {
			return errors.New("the metadata of resource cannot be nil")
		}
		if resource.Metadata.Repository == nil {
			return errors.New("the namespace of resource cannot be nil")
		}
		if len(resource.Metadata.Repository.Name) == 0 {
			return errors.New("the name of the namespace cannot be nil")
		}

		err := a.createRepository(resource.Metadata.Repository.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *adapter) createRepository(repository string) error {
	if a.registry.Credential == nil ||
		len(a.registry.Credential.AccessKey) == 0 || len(a.registry.Credential.AccessSecret) == 0 {
		return errors.New("no credential ")
	}
	cred := credentials.NewStaticCredentials(
		a.registry.Credential.AccessKey,
		a.registry.Credential.AccessSecret,
		"")
	if a.region == "" {
		return errors.New("no region parsed")
	}
	config := &aws.Config{
		Credentials: cred,
		Region:      &a.region,
		HTTPClient: &http.Client{
			Transport: internal.GetHTTPTransport(a.registry.Insecure),
		},
	}
	if a.forceEndpoint != nil {
		config.Endpoint = a.forceEndpoint
	}
	sess := session.Must(session.NewSession(config))

	svc := awsecrapi.New(sess)

	_, err := svc.CreateRepository(&awsecrapi.CreateRepositoryInput{
		RepositoryName: &repository,
	})
	if err != nil {
		if e, ok := err.(awserr.Error); ok {
			if e.Code() == awsecrapi.ErrCodeRepositoryAlreadyExistsException {
				return nil
			}
		}
		return err
	}
	return nil
}

// DeleteManifest ...
func (a *adapter) DeleteManifest(repository, reference string) error {
	// AWS doesn't implement standard OCI delete manifest API, so use it's sdk.
	if a.registry.Credential == nil ||
		len(a.registry.Credential.AccessKey) == 0 || len(a.registry.Credential.AccessSecret) == 0 {
		return errors.New("no credential ")
	}
	cred := credentials.NewStaticCredentials(
		a.registry.Credential.AccessKey,
		a.registry.Credential.AccessSecret,
		"")
	if a.region == "" {
		return errors.New("no region parsed")
	}
	config := &aws.Config{
		Credentials: cred,
		Region:      &a.region,
		HTTPClient: &http.Client{
			Transport: internal.GetHTTPTransport(a.registry.Insecure),
		},
	}
	if a.forceEndpoint != nil {
		config.Endpoint = a.forceEndpoint
	}
	sess := session.Must(session.NewSession(config))

	svc := awsecrapi.New(sess)

	_, err := svc.BatchDeleteImage(&awsecrapi.BatchDeleteImageInput{
		RepositoryName: &repository,
		ImageIds:       []*awsecrapi.ImageIdentifier{{ImageTag: &reference}},
	})
	return err
}
