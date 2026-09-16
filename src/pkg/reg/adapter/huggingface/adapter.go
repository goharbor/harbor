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

// Package huggingface provides the registry adapter for the Hugging Face model hub.
//
// The adapter only implements the base adapter surface (info and health check) so
// that a Hugging Face endpoint can be registered through registry management and
// used as the source of model sync policies. It advertises no replication-supported
// resource types. Replication policy validation and the rule editor exclude
// Hugging Face endpoints from replication.
package huggingface

import (
	"fmt"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const (
	// DefaultEndpoint is the public Hugging Face hub endpoint.
	DefaultEndpoint = "https://huggingface.co"
	// AccessKeyData is the fixed access key shown in the registry management UI, the
	// token is stored as access secret.
	AccessKeyData = "token"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeHuggingFace, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeHuggingFace, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeHuggingFace)
}

type factory struct{}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return New(r), nil
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints: []*model.Endpoint{
				{
					Key:   "huggingface.co",
					Value: DefaultEndpoint,
				},
			},
		},
		CredentialPattern: &model.CredentialPattern{
			AccessKeyType:    model.AccessKeyTypeFix,
			AccessKeyData:    AccessKeyData,
			AccessSecretType: model.AccessSecretTypeStandard,
		},
	}
}

var _ adp.Adapter = (*Adapter)(nil)

// Adapter is the registry adapter for Hugging Face.
type Adapter struct {
	registry *model.Registry
	client   *http.Client
}

// New creates a Hugging Face registry adapter.
func New(registry *model.Registry) *Adapter {
	return &Adapter{
		registry: registry,
		client: &http.Client{
			Transport: common_http.GetHTTPTransport(
				common_http.WithInsecure(registry.Insecure),
				common_http.WithCACert(registry.CACertificate),
			),
			Timeout: config.RegistryHTTPClientTimeout(),
		},
	}
}

// Endpoint returns the normalized endpoint of the hub.
func Endpoint(registry *model.Registry) string {
	url := strings.TrimSpace(registry.URL)
	if url == "" {
		url = DefaultEndpoint
	}
	return strings.TrimSuffix(url, "/")
}

// Token returns the access token configured for the registry, empty if none.
func Token(registry *model.Registry) string {
	if registry == nil || registry.Credential == nil {
		return ""
	}
	return strings.TrimSpace(registry.Credential.AccessSecret)
}

// Info returns the adapter information. No replication resource types are advertised.
func (a *Adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type:                     model.RegistryTypeHuggingFace,
		Description:              "Hugging Face model hub, used as the source of model sync policies",
		SupportedResourceTypes:   []string{},
		SupportedResourceFilters: []*model.FilterStyle{},
		SupportedTriggers: []string{
			model.TriggerTypeManual,
			model.TriggerTypeScheduled,
		},
	}, nil
}

// PrepareForPush is not supported for Hugging Face.
func (a *Adapter) PrepareForPush([]*model.Resource) error {
	return fmt.Errorf("pushing to %s is not supported", model.RegistryTypeHuggingFace)
}

// HealthCheck verifies endpoint reachability and, when a token is configured, its
// validity via the whoami API.
func (a *Adapter) HealthCheck() (string, error) {
	endpoint := Endpoint(a.registry)
	token := Token(a.registry)

	path := "/api/models?limit=1"
	if token != "" {
		path = "/api/whoami-v2"
	}
	req, err := http.NewRequest(http.MethodGet, endpoint+path, nil)
	if err != nil {
		return model.Unhealthy, err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := a.client.Do(req)
	if err != nil {
		log.Errorf("failed to ping hugging face endpoint %s: %v", endpoint, err)
		return model.Unhealthy, nil
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Errorf("unexpected status code %d when pinging hugging face endpoint %s", resp.StatusCode, endpoint)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}
