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

// Package huggingface is a source-only registry adapter that serves Hugging Face Hub models as
// ModelPack model-spec OCI artifacts. The manifests are synthesized from the Hub API and are
// byte-deterministic for a given commit.
package huggingface

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const healthCheckTimeout = 30 * time.Second

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
	return newAdapter(r, defaultCache())
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return getAdapterPattern()
}

// getAdapterPattern suggests the public Hub but accepts any URL, e.g. a mirror in an air-gapped
// site. CredentialPattern stays nil: any credential pattern turns the secret field of the portal
// into a JSON editor. The access secret is the Hub token, the access key is ignored.
func getAdapterPattern() *model.AdapterPattern {
	return &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints: []*model.Endpoint{
				{
					Key:   "huggingface.co",
					Value: hub.DefaultEndpoint,
				},
			},
		},
	}
}

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

type adapter struct {
	registry *model.Registry
	hub      *hub.Client
	locator  *locator
}

func newAdapter(registry *model.Registry, c cache.Cache) (*adapter, error) {
	endpoint := strings.TrimSuffix(registry.URL, "/")
	if endpoint == "" {
		endpoint = hub.DefaultEndpoint
	}
	var token string
	if registry.Credential != nil {
		token = registry.Credential.AccessSecret
	}
	client, err := hub.New(endpoint, hub.Options{
		Token:         token,
		Insecure:      registry.Insecure,
		CACertificate: registry.CACertificate,
		Timeout:       config.RegistryHTTPClientTimeout(),
	})
	if err != nil {
		return nil, err
	}
	return &adapter{
		registry: registry,
		hub:      client,
		locator:  &locator{cache: c, registryID: registry.ID},
	}, nil
}

// Info ...
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type:                   model.RegistryTypeHuggingFace,
		SupportedResourceTypes: []string{model.ResourceTypeArtifact},
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
		SupportedCopyByChunk: true,
	}, nil
}

// HealthCheck validates the token when one is configured, otherwise that the model API answers.
func (a *adapter) HealthCheck() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
	defer cancel()
	var err error
	if a.hub.HasToken() {
		err = a.hub.WhoAmI(ctx)
	} else {
		err = a.hub.Ping(ctx)
	}
	if err != nil {
		log.Errorf("health check of hugging face registry %s failed: %v", a.hub.Endpoint(), err)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}

func errSourceOnly() error {
	return errors.New(nil).WithCode(errors.MethodNotAllowedCode).
		WithMessagef("the %s registry is source-only", model.RegistryTypeHuggingFace)
}

// PrepareForPush ...
func (a *adapter) PrepareForPush([]*model.Resource) error {
	return errSourceOnly()
}

// PushManifest ...
func (a *adapter) PushManifest(_, _, _ string, _ []byte) (string, error) {
	return "", errSourceOnly()
}

// DeleteManifest ...
func (a *adapter) DeleteManifest(_, _ string) error {
	return errSourceOnly()
}

// PushBlob ...
func (a *adapter) PushBlob(_, _ string, _ int64, _ io.Reader) error {
	return errSourceOnly()
}

// PushBlobChunk ...
func (a *adapter) PushBlobChunk(_, _ string, _ int64, _ io.Reader, _, _ int64, _ string) (string, int64, error) {
	return "", -1, errSourceOnly()
}

// MountBlob ...
func (a *adapter) MountBlob(_, _, _ string) error {
	return errSourceOnly()
}

// CanBeMount ...
func (a *adapter) CanBeMount(_ string) (bool, string, error) {
	return false, "", nil
}

// DeleteTag ...
func (a *adapter) DeleteTag(_, _ string) error {
	return errSourceOnly()
}
