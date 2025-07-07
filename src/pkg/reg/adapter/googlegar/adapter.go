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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/opencontainers/go-digest"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/googlegar/auth"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var (
	// garEndpointPattern matches Google Artifact Registry endpoints
	// Format: [region]-docker.pkg.dev (or other artifact types like maven, npm, etc.)
	garEndpointPattern = regexp.MustCompile(`^https://[a-z0-9-]+-[a-z]+\.pkg\.dev/?$`)
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeGoogleGar, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeGoogleGar, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeGoogleGar)
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	ctx := context.Background()

	var authorizer *auth.OAuth2Authorizer
	var err error

	// Check if we have explicit credentials (JSON service account key)
	if registry.Credential != nil &&
		registry.Credential.AccessSecret != "" &&
		strings.Contains(registry.Credential.AccessSecret, "private_key") {
		// Use explicit credentials
		authorizer, err = auth.NewOAuth2AuthorizerWithCredentials(
			ctx,
			[]byte(registry.Credential.AccessSecret),
			[]string{auth.StorageScope, auth.CloudPlatformScope},
		)
		if err != nil {
			log.Errorf("failed to create OAuth2 authorizer with credentials for %s: %v", registry.URL, err)
			return nil, err
		}
	} else {
		// Use default credential chain (metadata server, gcloud, etc.)
		authorizer, err = auth.NewOAuth2Authorizer(
			ctx,
			[]string{auth.StorageScope, auth.CloudPlatformScope},
		)
		if err != nil {
			log.Errorf("failed to create OAuth2 authorizer for %s: %v", registry.URL, err)
			return nil, err
		}
	}

	return &adapter{
		registry: registry,
		Adapter:  native.NewAdapterWithAuthorizer(registry, authorizer),
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

	// Google Container Registry endpoints
	endpoints = append(endpoints, &model.Endpoint{
		Key:   "gcr.io",
		Value: "https://gcr.io",
	})
	endpoints = append(endpoints, &model.Endpoint{
		Key:   "us.gcr.io",
		Value: "https://us.gcr.io",
	})
	endpoints = append(endpoints, &model.Endpoint{
		Key:   "eu.gcr.io",
		Value: "https://eu.gcr.io",
	})
	endpoints = append(endpoints, &model.Endpoint{
		Key:   "asia.gcr.io",
		Value: "https://asia.gcr.io",
	})

	// Google Artifact Registry endpoints
	// https://cloud.google.com/compute/docs/regions-zones#available
	for _, region := range []string{
		"us-central1",
		"us-east1",
		"us-east4",
		"us-west1",
		"us-west2",
		"us-west3",
		"us-west4",
		"northamerica-northeast1",
		"northamerica-northeast2",
		"southamerica-east1",
		"southamerica-west1",
		"europe-north1",
		"europe-west1",
		"europe-west2",
		"europe-west3",
		"europe-west4",
		"europe-west6",
		"europe-west8",
		"europe-west9",
		"europe-central2",
		"europe-southwest1",
		"asia-east1",
		"asia-east2",
		"asia-northeast1",
		"asia-northeast2",
		"asia-northeast3",
		"asia-south1",
		"asia-south2",
		"asia-southeast1",
		"asia-southeast2",
		"australia-southeast1",
		"australia-southeast2",
	} {
		endpoints = append(endpoints, &model.Endpoint{
			Key:   region + "-docker.pkg.dev",
			Value: "https://" + region + "-docker.pkg.dev",
		})
	}

	// Allow custom endpoint
	endpoints = append(endpoints, &model.Endpoint{
		Key:   "custom",
		Value: "",
	})

	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints:    endpoints,
		},
		CredentialPattern: &model.CredentialPattern{
			AccessKeyType:    model.AccessKeyTypeStandard,
			AccessKeyData:    "oauth2",
			AccessSecretType: model.AccessSecretTypeFile,
			AccessSecretData: "Service Account JSON (optional - uses default credentials if empty)",
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

// isGoogleArtifactRegistry determines if the registry URL is for Google Artifact Registry
func (a *adapter) isGoogleArtifactRegistry() bool {
	return garEndpointPattern.MatchString(a.registry.URL)
}

// DeleteManifest deletes the manifest specified by reference; the reference can be digest or tag
func (a *adapter) DeleteManifest(repository, reference string) error {
	// For Google Container Registry, we need to use their specific API
	if strings.Contains(a.registry.URL, "gcr.io") {
		return a.deleteGcrManifest(repository, reference)
	}

	// For Google Artifact Registry, use standard Docker API
	return a.Adapter.DeleteManifest(repository, reference)
}

// deleteGcrManifest handles deletion for Google Container Registry using their tags API
func (a *adapter) deleteGcrManifest(repository, reference string) error {
	tags, d, err := a.listGcrTagsByRef(repository, reference)
	if err != nil {
		return err
	}

	if d == "" {
		return errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("%s:%s not found", repository, reference)
	}

	for _, t := range append(tags, d) {
		req, err := http.NewRequest(http.MethodDelete, buildManifestURL(a.registry.URL, repository, t), nil)
		if err != nil {
			return err
		}
		resp, err := a.Client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	return nil
}

// listGcrTagsByRef lists tags for GCR (similar to existing googlegcr adapter)
func (a *adapter) listGcrTagsByRef(repository, reference string) ([]string, string, error) {
	u := buildTagListURL(a.registry.URL, repository)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := a.Client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	tgs := struct {
		Name     string   `json:"name"`
		Tags     []string `json:"tags"`
		Manifest map[string]struct {
			Tag       []string `json:"tag"`
			MediaType string   `json:"mediaType"`
		} `json:"manifest"`
	}{}

	if err = json.Unmarshal(body, &tgs); err != nil {
		return nil, "", err
	}

	_, err = digest.Parse(reference)
	if err == nil {
		// for sha256 as reference
		if m, ok := tgs.Manifest[reference]; ok {
			return m.Tag, reference, nil
		}
		return nil, reference, nil
	}

	// for tag as reference
	for d, m := range tgs.Manifest {
		for _, t := range m.Tag {
			if t == reference {
				return m.Tag, d, nil
			}
		}
	}
	return nil, "", nil
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

func buildTagListURL(endpoint, repository string) string {
	return fmt.Sprintf("%s/v2/%s/tags/list", endpoint, repository)
}

func buildManifestURL(endpoint, repository, reference string) string {
	return fmt.Sprintf("%s/v2/%s/manifests/%s", endpoint, repository, reference)
}
