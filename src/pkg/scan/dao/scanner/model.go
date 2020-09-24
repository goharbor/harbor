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

package scanner

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/scan/rest/auth"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

const (
	authorizationType   = "harbor.scanner-adapter/registry-authorization-type"
	authorizationBearer = "Bearer"
	authorizationBasic  = "Basic"
)

// Registration represents a named configuration for invoking a scanner via its adapter.
// UUID will be used to track the scanner.Endpoint as unique ID
type Registration struct {
	// Basic information
	// int64 ID is kept for being aligned with previous DB schema
	ID          int64  `orm:"pk;auto;column(id)" json:"-"`
	UUID        string `orm:"unique;column(uuid)" json:"uuid"`
	Name        string `orm:"unique;column(name);size(128)" json:"name"`
	Description string `orm:"column(description);null;size(1024)" json:"description"`
	URL         string `orm:"column(url);unique;size(512)" json:"url"`
	Disabled    bool   `orm:"column(disabled);default(true)" json:"disabled"`
	IsDefault   bool   `orm:"column(is_default);default(false)" json:"is_default"`
	Health      string `orm:"-" json:"health,omitempty"`

	// Authentication settings
	// "","Basic", "Bearer" and api key header "X-ScannerAdapter-API-Key" can be supported
	Auth             string `orm:"column(auth);size(16)" json:"auth"`
	AccessCredential string `orm:"column(access_cred);null;size(512)" json:"access_credential,omitempty"`

	// Http connection settings
	SkipCertVerify bool `orm:"column(skip_cert_verify);default(false)" json:"skip_certVerify"`

	// Indicate whether use internal registry addr for the scanner to pull content
	UseInternalAddr bool `orm:"column(use_internal_addr);default(false)" json:"use_internal_addr"`

	// Indicate if the registration is immutable which is not allowed to remove
	Immutable bool `orm:"column(immutable);default(false)" json:"-"`

	// Optional properties for describing the adapter
	Adapter string `orm:"-" json:"adapter,omitempty"`
	Vendor  string `orm:"-" json:"vendor,omitempty"`
	Version string `orm:"-" json:"version,omitempty"`

	Metadata *v1.ScannerAdapterMetadata `orm:"-" json:"-"`

	// Timestamps
	CreateTime time.Time `orm:"column(create_time);auto_now_add;type(datetime)" json:"create_time"`
	UpdateTime time.Time `orm:"column(update_time);auto_now;type(datetime)" json:"update_time"`
}

// TableName for Endpoint
func (r *Registration) TableName() string {
	return "scanner_registration"
}

// FromJSON parses registration from json data
func (r *Registration) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), r)
}

// ToJSON marshals registration to JSON data
func (r *Registration) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Validate registration
func (r *Registration) Validate(checkUUID bool) error {
	if checkUUID && len(r.UUID) == 0 {
		return errors.New("malformed endpoint")
	}

	if len(r.Name) == 0 {
		return errors.New("missing registration name")
	}

	err := checkURL(r.URL)
	if err != nil {
		return errors.Wrap(err, "scanner registration validate")
	}

	if len(r.Auth) > 0 &&
		r.Auth != auth.Basic &&
		r.Auth != auth.Bearer &&
		r.Auth != auth.APIKey {
		return errors.Errorf("auth type %s is not supported", r.Auth)
	}

	if len(r.Auth) > 0 && len(r.AccessCredential) == 0 {
		return errors.Errorf("access_credential is required for auth type %s", r.Auth)
	}

	return nil
}

// Client returns client of registration
func (r *Registration) Client(pool v1.ClientPool) (v1.Client, error) {
	if err := r.Validate(false); err != nil {
		return nil, err
	}

	return pool.Get(r.URL, r.Auth, r.AccessCredential, r.SkipCertVerify)
}

// HasCapability returns true when mime type of the artifact support by the scanner
func (r *Registration) HasCapability(manifestMimeType string) bool {
	if r.Metadata == nil {
		return false
	}

	for _, capability := range r.Metadata.Capabilities {
		for _, mt := range capability.ConsumesMimeTypes {
			if mt == manifestMimeType {
				return true
			}
		}
	}

	return false
}

// GetProducesMimeTypes returns produces mime types for the artifact
func (r *Registration) GetProducesMimeTypes(mimeType string) []string {
	if r.Metadata == nil {
		return nil
	}

	for _, capability := range r.Metadata.Capabilities {
		for _, mt := range capability.ConsumesMimeTypes {
			if mt == mimeType {
				return capability.ProducesMimeTypes
			}
		}
	}

	return nil
}

// GetCapability returns capability for the mime type
func (r *Registration) GetCapability(mimeType string) *v1.ScannerCapability {
	if r.Metadata == nil {
		return nil
	}

	for _, capability := range r.Metadata.Capabilities {
		for _, mt := range capability.ConsumesMimeTypes {
			if mt == mimeType {
				return capability
			}
		}
	}

	return nil
}

// GetRegistryAuthorizationType returns the registry authorization type of the scanner
func (r *Registration) GetRegistryAuthorizationType() string {
	var auth string
	if r.Metadata != nil && r.Metadata.Properties != nil {
		if v, ok := r.Metadata.Properties[authorizationType]; ok {
			auth = v
		}
	}

	if auth != authorizationBasic && auth != authorizationBearer {
		auth = authorizationBasic
	}

	return auth
}

// Check the registration URL with url package
func checkURL(u string) error {
	if len(strings.TrimSpace(u)) == 0 {
		return errors.New("empty url")
	}

	uri, err := url.Parse(u)
	if err == nil {
		if uri.Scheme != "http" && uri.Scheme != "https" {
			err = errors.New("invalid scheme")
		}
	}

	return err
}
