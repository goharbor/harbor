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
	"fmt"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/awserr"
	awsecrapi "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
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
	svc, err := getAwsSvc(
		region, registry.Credential.AccessKey, registry.Credential.AccessSecret, registry.Insecure, nil)
	if err != nil {
		return nil, err
	}
	authorizer := NewAuth(registry.Credential.AccessKey, svc)
	return &adapter{
		registry: registry,
		Adapter:  native.NewAdapterWithAuthorizer(registry, authorizer),
		cacheSvc: svc,
	}, nil
}

func parseRegion(url string) (string, error) {
	rs := regionRegexp.FindStringSubmatch(url)
	if rs == nil {
		return "", errors.New("bad aws url")
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
	registry *model.Registry
	cacheSvc *awsecrapi.ECR
}

func (*adapter) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeAwsEcr,
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
	}, nil
}

func getAdapterInfo() *model.AdapterPattern {
	var endpoints []*model.Endpoint
	// https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints
	for _, e := range []string{
		"us-east-2",
		"us-east-1",
		"us-west-1",
		"us-west-2",
		"af-south-1",
		"ap-east-1",
		"ap-south-1",
		"ap-northeast-3",
		"ap-northeast-2",
		"ap-southeast-1",
		"ap-southeast-2",
		"ap-northeast-1",
		"ca-central-1",
		"cn-north-1",
		"cn-northwest-1",
		"eu-central-1",
		"eu-west-1",
		"eu-west-2",
		"eu-south-1",
		"eu-west-3",
		"eu-north-1",
		"me-south-1",
		"sa-east-1",
	} {
		endpoints = append(endpoints, &model.Endpoint{
			Key:   e,
			Value: fmt.Sprintf("https://api.ecr.%s.amazonaws.com", e),
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

// HealthCheck checks health status of a registry
func (a *adapter) HealthCheck() (string, error) {
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

		exist, err := a.checkRepository(resource.Metadata.Repository.Name)
		if err != nil {
			return err
		}
		if exist {
			log.Infof("Namespace %s already exist in AWS ECR, skip it.", resource.Metadata.Repository.Name)
			continue
		}
		err = a.createRepository(resource.Metadata.Repository.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *adapter) checkRepository(repository string) (exists bool, err error) {
	out, err := a.cacheSvc.DescribeRepositories(&awsecrapi.DescribeRepositoriesInput{
		RepositoryNames: []*string{&repository},
	})
	if err != nil {
		if e, ok := err.(awserr.Error); ok {
			if e.Code() == awsecrapi.ErrCodeRepositoryNotFoundException {
				return false, nil
			}
		}
		return false, err
	}
	return len(out.Repositories) > 0, nil
}

func (a *adapter) createRepository(repository string) error {
	_, err := a.cacheSvc.CreateRepository(&awsecrapi.CreateRepositoryInput{
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
	_, err := a.cacheSvc.BatchDeleteImage(&awsecrapi.BatchDeleteImageInput{
		RepositoryName: &repository,
		ImageIds:       []*awsecrapi.ImageIdentifier{{ImageTag: &reference}},
	})
	return err
}
