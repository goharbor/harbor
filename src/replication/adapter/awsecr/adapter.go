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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsecrapi "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"net/http"
	"regexp"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeAwsEcr, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry)
	}); err != nil {
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
	reg, err := adp.NewDefaultImageRegistryWithCustomizedAuthorizer(registry, authorizer)
	if err != nil {
		return nil, err
	}
	return &adapter{
		registry:             registry,
		DefaultImageRegistry: reg,
	}, nil
}

func parseRegion(url string) (string, error) {
	pattern := "https://(?:api|\\d+\\.dkr)\\.ecr\\.([\\w\\-]+)\\.amazonaws\\.com"
	rs := regexp.MustCompile(pattern).FindStringSubmatch(url)
	if rs == nil {
		return "", errors.New("Bad aws url")
	}
	return rs[1], nil
}

type adapter struct {
	*adp.DefaultImageRegistry
	registry *model.Registry
}

var _ adp.Adapter = adapter{}

func (adapter) Info() (info *model.RegistryInfo, err error) {
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

// HealthCheck checks health status of a registry
func (d adapter) HealthCheck() (model.HealthStatus, error) {
	if d.registry.Credential == nil ||
		(len(d.registry.Credential.AccessKey) == 0 && len(d.registry.Credential.AccessSecret) == 0) {
		return model.Unhealthy, nil
	}
	if err := d.PingGet(); err != nil {
		log.Errorf("failed to ping registry %s: %v", d.registry.URL, err)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}

// PrepareForPush nothing need to do.
func (d adapter) PrepareForPush(resources []*model.Resource) error {
	for _, resource := range resources {
		if resource == nil {
			return errors.New("the resource cannot be null")
		}
		if resource.Metadata == nil {
			return errors.New("the metadata of resource cannot be null")
		}
		if resource.Metadata.Repository == nil {
			return errors.New("the namespace of resource cannot be null")
		}
		if len(resource.Metadata.Repository.Name) == 0 {
			return errors.New("the name of the namespace cannot be null")
		}

		err := d.createRepository(resource.Metadata.Repository.Name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d adapter) createRepository(repository string) error {
	cred := credentials.NewStaticCredentials(
		d.registry.Credential.AccessKey,
		d.registry.Credential.AccessSecret,
		"")
	region, err := parseRegion(d.registry.URL)
	if err != nil {
		return err
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: cred,
		Region:      &region,
		HTTPClient: &http.Client{
			Transport: registry.GetHTTPTransport(d.registry.Insecure),
		},
	}))

	svc := awsecrapi.New(sess)

	_, err = svc.CreateRepository(&awsecrapi.CreateRepositoryInput{
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
