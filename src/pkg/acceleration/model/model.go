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

const (
	// AccelerationTypeNydu ...
	AccelerationTypeNydus = "nydus"

	// Healthy indicates registry is healthy
	Healthy = "healthy"
	// Unhealthy indicates registry is unhealthy
	Unhealthy = "unhealthy"

	// CredentialTypeBasic indicates credential by user name, password
	CredentialTypeBasic = "basic"
	// CredentialTypeOAuth indicates credential by OAuth token
	CredentialTypeOAuth = "oauth"
	// CredentialTypeSecret is only used by the communication of Harbor internal components
	CredentialTypeSecret = "secret"
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

// AccelerationService ...
type AccelerationService struct {
	ID           int64       `json:"id"`
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	URL          string      `json:"url"`
	Credential   *Credential `json:"credential"`
	Insecure     bool        `json:"insecure"`
	Status       string      `json:"status"`
	CreationTime time.Time   `json:"creation_time"`
	UpdateTime   time.Time   `json:"update_time"`
}
