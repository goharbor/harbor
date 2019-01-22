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

// RegistryType indicates the type of registry
type RegistryType string

// CredentialType represents the supported credential types
// e.g: u/p, OAuth token
type CredentialType string

// Credential keeps the access key and/or secret for the related registry
type Credential struct {
	// Type of the credential
	Type CredentialType `json:"type"`
	// The key of the access account, for OAuth token, it can be empty
	AccessKey string `json:"access_key"`
	// The secret or password for the key
	AccessSecret string `json:"access_secret"`
}

// Registry keeps the related info of registry
// Data required for the secure access way is not contained here.
// DAO layer is not considered here
type Registry struct {
	ID          int64        `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Type        RegistryType `json:"type"`
	URL         string       `json:"url"`
	Credential  *Credential  `json:"credential"`
	Insecure    bool         `json:"insecure"`
	Status      string       `json:"status"`
}
