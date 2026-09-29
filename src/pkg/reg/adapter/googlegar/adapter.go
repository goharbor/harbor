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

package googlegar

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/oauth2/google"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const (
	// CloudPlatformScope is the Google Cloud Platform OAuth2 scope for Artifact Registry access
	CloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

	jsonKeyUsername       = "_json_key"
	oauth2AccessTokenUser = "oauth2accesstoken"
)

var (
	// garEndpointPattern matches Google Artifact Registry endpoints
	// Format: https://[location]-docker.pkg.dev (or other artifact types under pkg.dev)
	garEndpointPattern = regexp.MustCompile(`^https://[a-z0-9-]+-[a-z]+\.pkg\.dev/?$`)

	defaultTokenSource = google.DefaultTokenSource
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeGoogleGar, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeGoogleGar, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeGoogleGar)
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	if registry == nil || !isGoogleArtifactRegistry(registry.URL) {
		var u string
		if registry != nil {
			u = registry.URL
		}
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("invalid Google Artifact Registry URL: %s", u)
	}

	reg := *registry
	if registry.Credential != nil && strings.TrimSpace(registry.Credential.AccessSecret) != "" {
		accessKey := strings.TrimSpace(registry.Credential.AccessKey)
		if accessKey == "" || accessKey == "oauth2" {
			accessKey = jsonKeyUsername
		}
		reg.Credential = &model.Credential{
			Type:         registry.Credential.Type,
			AccessKey:    accessKey,
			AccessSecret: registry.Credential.AccessSecret,
		}
	} else {
		tokenSource, err := defaultTokenSource(context.Background(), CloudPlatformScope)
		if err != nil {
			log.Errorf("failed to create default token source for %s: %v", registry.URL, err)
			return nil, fmt.Errorf("failed to create default token source: %w", err)
		}
		token, err := tokenSource.Token()
		if err != nil {
			log.Errorf("failed to get OAuth2 token for %s: %v", registry.URL, err)
			return nil, fmt.Errorf("failed to get OAuth2 token: %w", err)
		}
		if token == nil || token.AccessToken == "" {
			return nil, errors.New("empty OAuth2 access token")
		}
		reg.Credential = &model.Credential{
			AccessKey:    oauth2AccessTokenUser,
			AccessSecret: token.AccessToken,
		}
	}

	return &adapter{
		registry: registry,
		Adapter:  native.NewAdapter(&reg),
	}, nil
}

type factory struct {
}

// Create creates an adapter
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern returns the adapter pattern info
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
}

// Info returns the basic information about the adapter
func (a *adapter) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeGoogleGar,
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
	for _, ep := range []string{
		"us-docker.pkg.dev",
		"europe-docker.pkg.dev",
		"asia-docker.pkg.dev",
		"us-central1-docker.pkg.dev",
		"us-east1-docker.pkg.dev",
		"us-west1-docker.pkg.dev",
		"europe-west1-docker.pkg.dev",
		"asia-east1-docker.pkg.dev",
	} {
		endpoints = append(endpoints, &model.Endpoint{
			Key:   ep,
			Value: fmt.Sprintf("https://%s", ep),
		})
	}
	return &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints:    endpoints,
		},
		CredentialPattern: &model.CredentialPattern{
			AccessKeyType:    model.AccessKeyTypeFix,
			AccessKeyData:    jsonKeyUsername,
			AccessSecretType: model.AccessSecretTypeFile,
			AccessSecretData: "No Change",
		},
	}
}

// HealthCheck checks health status of a registry
func (a *adapter) HealthCheck() (string, error) {
	if err := a.Ping(); err != nil {
		log.Errorf("failed to ping registry %s: %v", a.registry.URL, err)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}

// isGoogleArtifactRegistry determines if the registry URL is a valid Google Artifact Registry endpoint
func isGoogleArtifactRegistry(u string) bool {
	return garEndpointPattern.MatchString(u)
}

// DeleteTag deletes the specified tag
func (a *adapter) DeleteTag(repository, tag string) error {
	req, err := http.NewRequest(http.MethodDelete, buildManifestURL(a.registry.URL, repository, tag), nil)
	if err != nil {
		return err
	}
	resp, err := a.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func buildManifestURL(endpoint, repository, reference string) string {
	return fmt.Sprintf("%s/v2/%s/manifests/%s", strings.TrimSuffix(endpoint, "/"), repository, reference)
}
