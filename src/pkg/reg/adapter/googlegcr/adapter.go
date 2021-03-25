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

package googlegcr

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/opencontainers/go-digest"
	"io/ioutil"
	"net/http"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeGoogleGcr, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeGoogleGcr, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeGoogleGcr)
}

func newAdapter(registry *model.Registry) *adapter {
	return &adapter{
		registry: registry,
		Adapter:  native.NewAdapter(registry),
	}
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r), nil
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
}

var _ adp.Adapter = adapter{}

func (adapter) Info() (info *model.RegistryInfo, err error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeGoogleGcr,
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
	info := &model.AdapterPattern{
		EndpointPattern: &model.EndpointPattern{
			EndpointType: model.EndpointPatternTypeList,
			Endpoints: []*model.Endpoint{
				{
					Key:   "gcr.io",
					Value: "https://gcr.io",
				},
				{
					Key:   "us.gcr.io",
					Value: "https://us.gcr.io",
				},
				{
					Key:   "eu.gcr.io",
					Value: "https://eu.gcr.io",
				},
				{
					Key:   "asia.gcr.io",
					Value: "https://asia.gcr.io",
				},
			},
		},
		CredentialPattern: &model.CredentialPattern{
			AccessKeyType:    model.AccessKeyTypeFix,
			AccessKeyData:    "_json_key",
			AccessSecretType: model.AccessSecretTypeFile,
			AccessSecretData: "No Change",
		},
	}
	return info
}

// HealthCheck checks health status of a registry
func (a adapter) HealthCheck() (string, error) {
	var err error
	if a.registry.Credential == nil ||
		len(a.registry.Credential.AccessKey) == 0 || len(a.registry.Credential.AccessSecret) == 0 {
		log.Errorf("no credential to ping registry %s", a.registry.URL)
		return model.Unhealthy, nil
	}
	if err = a.Ping(); err != nil {
		log.Errorf("failed to ping registry %s: %v", a.registry.URL, err)
		return model.Unhealthy, nil
	}
	return model.Healthy, nil
}

/*
{
	"child": [],
	"manifest": {
		"sha256:400ee2ed939df769d4681023810d2e4fb9479b8401d97003c710d0e20f7c49c6": {
			"imageSizeBytes": "763789",
			"layerId": "",
			"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
			"tag": ["another", "latest"],
			"timeCreatedMs": "1595895577054",
			"timeUploadedMs": "1597767277119"
		}
	},
	"name": "eminent-nation-87317/testgcr/busybox",
	"tags": ["another", "latest"]
}
*/
func (a adapter) listGcrTagsByRef(repository, reference string) ([]string, string, error) {
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

	body, err := ioutil.ReadAll(resp.Body)
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

func (a adapter) DeleteManifest(repository, reference string) error {
	tags, d, err := a.listGcrTagsByRef(repository, reference)
	if err != nil {
		return err
	}

	if d == "" {
		return errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessage("%s:%s not found", repository, reference)
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

func buildTagListURL(endpoint, repository string) string {
	return fmt.Sprintf("%s/v2/%s/tags/list", endpoint, repository)
}

func buildManifestURL(endpoint, repository, reference string) string {
	return fmt.Sprintf("%s/v2/%s/manifests/%s", endpoint, repository, reference)
}

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
