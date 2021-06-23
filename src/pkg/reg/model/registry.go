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

package model

import (
	"time"
)

// const definition
const (
	RegistryTypeHarbor           = "harbor"
	RegistryTypeDockerHub        = "docker-hub"
	RegistryTypeDockerRegistry   = "docker-registry"
	RegistryTypeHuawei           = "huawei-SWR"
	RegistryTypeGoogleGcr        = "google-gcr"
	RegistryTypeAwsEcr           = "aws-ecr"
	RegistryTypeAzureAcr         = "azure-acr"
	RegistryTypeAliAcr           = "ali-acr"
	RegistryTypeJfrogArtifactory = "jfrog-artifactory"
	RegistryTypeQuay             = "quay"
	RegistryTypeGitLab           = "gitlab"
	RegistryTypeDTR              = "dtr"
	RegistryTypeTencentTcr       = "tencent-tcr"
	RegistryTypeGithubCR         = "github-ghcr"

	RegistryTypeHelmHub     = "helm-hub"
	RegistryTypeArtifactHub = "artifact-hub"

	FilterStyleTypeText  = "input"
	FilterStyleTypeRadio = "radio"
	FilterStyleTypeList  = "list"

	// CredentialTypeBasic indicates credential by user name, password
	CredentialTypeBasic = "basic"
	// CredentialTypeOAuth indicates credential by OAuth token
	CredentialTypeOAuth = "oauth"
	// CredentialTypeSecret is only used by the communication of Harbor internal components
	CredentialTypeSecret = "secret"

	// EndpointPatternTypeStandard ...
	EndpointPatternTypeStandard = "EndpointPatternTypeStandard"
	// EndpointPatternTypeFix ...
	EndpointPatternTypeFix = "EndpointPatternTypeFix"
	// EndpointPatternTypeList ...
	EndpointPatternTypeList = "EndpointPatternTypeList"

	// AccessKeyTypeStandard ...
	AccessKeyTypeStandard = "AccessKeyTypeStandard"
	// AccessKeyTypeFix ...
	AccessKeyTypeFix = "AccessKeyTypeFix"

	// AccessSecretTypeStandard ...
	AccessSecretTypeStandard = "AccessSecretTypePass"
	// AccessSecretTypeFile ...
	AccessSecretTypeFile = "AccessSecretTypeFile"

	// Healthy indicates registry is healthy
	Healthy = "healthy"
	// Unhealthy indicates registry is unhealthy
	Unhealthy = "unhealthy"

	RepositoryPathComponentTypeOnlyTwo    = "ONLY_TWO"
	RepositoryPathComponentTypeAtLeastTwo = "AT_LEAST_TWO"
)

// Credential keeps the access key and/or secret for the related registry
type Credential struct {
	// Type of the credential
	Type string `json:"type"`
	// The key of the access account, for OAuth token, it can be empty
	AccessKey string `json:"access_key"`
	// The secret or password for the key
	AccessSecret string `json:"access_secret"`
}

// Registry keeps the related info of registry
// Data required for the secure access way is not contained here.
// DAO layer is not considered here
type Registry struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	// TokenServiceURL is only used for local harbor instance to
	// avoid the requests passing through the external proxy for now
	TokenServiceURL string      `json:"token_service_url"`
	Credential      *Credential `json:"credential"`
	Insecure        bool        `json:"insecure"`
	Status          string      `json:"status"`
	CreationTime    time.Time   `json:"creation_time"`
	UpdateTime      time.Time   `json:"update_time"`
}

// FilterStyle ...
type FilterStyle struct {
	Type   string   `json:"type"`
	Style  string   `json:"style"`
	Values []string `json:"values,omitempty"`
}

// EndpointPattern ...
type EndpointPattern struct {
	EndpointType string      `json:"endpoint_type"`
	Endpoints    []*Endpoint `json:"endpoints"`
}

// Endpoint ...
type Endpoint struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CredentialPattern ...
type CredentialPattern struct {
	AccessKeyType    string `json:"access_key_type"`
	AccessKeyData    string `json:"access_key_data"`
	AccessSecretType string `json:"access_secret_type"`
	AccessSecretData string `json:"access_secret_data"`
}

// RegistryInfo provides base info and capability declarations of the registry
type RegistryInfo struct {
	Type                                 string         `json:"type"`
	Description                          string         `json:"description"`
	SupportedResourceTypes               []string       `json:"-"`
	SupportedResourceFilters             []*FilterStyle `json:"supported_resource_filters"`
	SupportedTriggers                    []string       `json:"supported_triggers"`
	SupportedRepositoryPathComponentType string         `json:"supported_repository_path_component_type"` // how many path components are allowed in the repository name
}

// AdapterPattern provides base info and capability declarations of the registry
type AdapterPattern struct {
	EndpointPattern   *EndpointPattern   `json:"endpoint_pattern"`
	CredentialPattern *CredentialPattern `json:"credential_pattern"`
}

// NewDefaultAdapterPattern ...
func NewDefaultAdapterPattern() *AdapterPattern {
	return &AdapterPattern{
		EndpointPattern:   NewDefaultEndpointPattern(),
		CredentialPattern: NewDefaultCredentialPattern(),
	}
}

// NewDefaultEndpointPattern ...
func NewDefaultEndpointPattern() *EndpointPattern {
	return &EndpointPattern{
		EndpointType: EndpointPatternTypeStandard,
	}
}

// NewDefaultCredentialPattern ...
func NewDefaultCredentialPattern() *CredentialPattern {
	return &CredentialPattern{
		AccessKeyType:    AccessKeyTypeStandard,
		AccessSecretType: AccessSecretTypeStandard,
	}
}
