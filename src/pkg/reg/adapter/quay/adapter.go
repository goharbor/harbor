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

package quay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	qauth "github.com/goharbor/harbor/src/pkg/reg/adapter/quay/auth"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)

type adapter struct {
	*native.Adapter
	autoCreateNs bool
	registry     *model.Registry
	client       *common_http.Client
}

func init() {
	err := adp.RegisterFactory(model.RegistryTypeQuay, new(factory))
	if err != nil {
		log.Errorf("failed to register factory for Quay: %v", err)
		return
	}
	log.Infof("the factory of Quay adapter was registered")
}

func newAdapter(registry *model.Registry) (*adapter, error) {
	var (
		modifiers                         []modifier.Modifier
		autoCreateNs                      bool
		tokenAuthorizer, apiKeyAuthorizer modifier.Modifier
	)

	if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
		var jsonCred cred
		err := json.Unmarshal([]byte(registry.Credential.AccessSecret), &jsonCred)
		if err != nil {
			return nil, err
		}
		tokenAuthorizer = qauth.NewAuthorizer(jsonCred.AccountName, jsonCred.DockerCliPassword, registry.Insecure)
		if len(jsonCred.OAuth2Token) != 0 {
			autoCreateNs = true
			apiKeyAuthorizer = qauth.NewAPIKeyAuthorizer("Authorization", fmt.Sprintf("Bearer %s", jsonCred.OAuth2Token), qauth.APIKeyInHeader)
		}
	}

	nativeRegistryAdapter := native.NewAdapterWithAuthorizer(registry, tokenAuthorizer)

	if apiKeyAuthorizer != nil {
		modifiers = append(modifiers, apiKeyAuthorizer)
	}

	return &adapter{
		Adapter:      nativeRegistryAdapter,
		autoCreateNs: autoCreateNs,
		registry:     registry,
		client: common_http.NewClient(
			&http.Client{
				Transport: common_http.GetHTTPTransport(common_http.WithInsecure(registry.Insecure)),
			},
			modifiers...,
		),
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
	info := &model.AdapterPattern{
		EndpointPattern: model.NewDefaultEndpointPattern(),
		CredentialPattern: &model.CredentialPattern{
			AccessKeyType:    model.AccessKeyTypeFix,
			AccessKeyData:    "json_file",
			AccessSecretType: model.AccessSecretTypeFile,
			AccessSecretData: "",
		},
	}
	return info
}

// Info returns information of the registry
func (a *adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeQuay,
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

// PrepareForPush does the prepare work that needed for pushing/uploading the resource
// eg: create the namespace or repository
func (a *adapter) PrepareForPush(resources []*model.Resource) error {
	if !a.autoCreateNs {
		return nil
	}
	namespaces := []string{}
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
		paths := strings.Split(resource.Metadata.Repository.Name, "/")
		namespace := paths[0]
		namespaces = append(namespaces, namespace)
	}

	for _, namespace := range namespaces {
		err := a.createNamespace(&model.Namespace{
			Name: namespace,
		})
		if err != nil {
			return fmt.Errorf("create namespace '%s' in Quay error: %v", namespace, err)
		}
		log.Debugf("namespace %s created", namespace)
	}
	return nil
}

// createNamespace creates a new namespace in Quay
func (a *adapter) createNamespace(namespace *model.Namespace) error {
	ns, err := a.getNamespace(namespace.Name)
	if err != nil {
		return fmt.Errorf("check existence of namespace '%s' error: %v", namespace.Name, err)
	}

	// If the namespace already exist, return succeeded directly.
	if ns != nil {
		log.Infof("Namespace %s already exist in Quay, skip it.", namespace.Name)
		return nil
	}

	org := &orgCreate{
		Name:  namespace.Name,
		Email: fmt.Sprintf("%s@quay.io", namespace.Name),
	}
	b, err := json.Marshal(org)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, buildOrgURL(a.registry.URL, ""), bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	log.Errorf("create namespace error: %d -- %s", resp.StatusCode, string(body))
	return fmt.Errorf("%d -- %s", resp.StatusCode, body)
}

// getNamespace get namespace from Quay, if the namespace not found, two nil would be returned.
func (a *adapter) getNamespace(namespace string) (*model.Namespace, error) {
	req, err := http.NewRequest(http.MethodGet, buildOrgURL(a.registry.URL, namespace), nil)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode/100 != 2 {
		log.Errorf("get namespace error: %d -- %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("%d -- %s", resp.StatusCode, body)
	}

	return &model.Namespace{
		Name: namespace,
	}, nil
}

// PushManifest ...
func (a *adapter) PushManifest(repository, reference, mediaType string, payload []byte) (string, error) {
	digest, err := a.Adapter.PushManifest(repository, reference, mediaType, payload)
	if err != nil {
		if comErr, ok := err.(*common_http.Error); ok {
			if comErr.Code == http.StatusAccepted {
				return digest, nil
			}
		}
	}
	return digest, err
}

// PullBlob ...
func (a *adapter) PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error) {
	size, blob, err = a.Adapter.PullBlob(repository, digest)
	if err != nil && blob != nil {
		if size == 0 {
			var data []byte
			defer blob.Close()
			data, err = ioutil.ReadAll(blob)
			if err != nil {
				return
			}
			size = int64(len(data))
			blob = ioutil.NopCloser(bytes.NewReader(data))
			return size, blob, nil
		}
	}
	return
}
