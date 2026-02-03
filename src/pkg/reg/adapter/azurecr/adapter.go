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

package azurecr

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeAzureAcr, new(factory)); err != nil {
		log.Errorf("Register adapter factory for %s error: %v", model.RegistryTypeAzureAcr, err)
		return
	}
	log.Infof("Factory for adapter %s registered", model.RegistryTypeAzureAcr)
}

func newAdapter(registry *model.Registry) (adp.Adapter, error) {
	return &adapter{
		Adapter:     native.NewAdapterWithAuthorizer(registry, newAuthorizer(registry)),
		registryURL: registry.URL,
	}, nil
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

type adapter struct {
	*native.Adapter
	registryURL string
}

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeAzureAcr,
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

// DeleteTag deletes a tag from a repository in Azure Container Registry. it is implemented by untag api operation
func (a *adapter) DeleteTag(repository, tag string) error {
	// Azure Container Registry uses the standard Docker Registry API for tag deletion
	// The "untag" operation is implemented by deleting the manifest by tag
	// This follows the same pattern as other registry adapters

	// Build the manifest deletion URL for Azure Container Registry
	// Use the standard /v2/{repository}/manifests/{tag} endpoint
	manifestURL, err := url.Parse(fmt.Sprintf("%s/v2/%s/manifests/%s", a.registryURL, repository, tag))
	if err != nil {
		return fmt.Errorf("failed to build manifest deletion URL: %v", err)
	}

	// Create DELETE request
	req, err := http.NewRequest(http.MethodDelete, manifestURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %v", err)
	}

	// Use the existing client with authorizer to make the request
	resp, err := a.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete tag %s from repository %s: %v", tag, repository, err)
	}
	defer resp.Body.Close()

	// Check response status - Azure CR returns 202 Accepted for successful deletions
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("failed to delete tag %s from repository %s: HTTP %d", tag, repository, resp.StatusCode)
	}

	log.Infof("Successfully deleted tag %s from repository %s", tag, repository)
	return nil
}
