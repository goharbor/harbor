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

package dao

import (
	"encoding/json"
	"time"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	v1 "github.com/goharbor/harbor/src/pkg/optimizer/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/rest/auth"
)

const (
	// authorizationType is the property key an adapter can use to declare which
	// registry authorization scheme it expects when pulling artifacts.
	// Note: the auth pkg reused from the scanner framework sends API keys via the
	// X-ScannerAdapter-API-Key header; the header name is cosmetic and kept as-is.
	authorizationType   = "harbor.optimizer-adapter/registry-authorization-type"
	authorizationBearer = "Bearer"
	authorizationBasic  = "Basic"
)

// Registration represents a named configuration for invoking an optimizer via its adapter.
// UUID is used to track the registration as unique ID.
type Registration struct {
	// Basic information
	ID          int64  `orm:"pk;auto;column(id)" json:"-"`
	UUID        string `orm:"unique;column(uuid)" json:"uuid"`
	Name        string `orm:"unique;column(name);size(128)" json:"name"`
	Description string `orm:"column(description);null;size(1024)" json:"description"`
	URL         string `orm:"column(url);unique;size(512)" json:"url"`
	Disabled    bool   `orm:"column(disabled);default(true)" json:"disabled"`
	IsDefault   bool   `orm:"column(is_default);default(false)" json:"is_default"`
	Health      string `orm:"-" json:"health,omitempty"`

	// Authentication settings
	// "","Basic", "Bearer" and api key header can be supported
	Auth             string `orm:"column(auth);size(16)" json:"auth"`
	AccessCredential string `orm:"column(access_cred);null;size(512)" json:"access_credential,omitempty"`

	// Http connection settings
	SkipCertVerify bool `orm:"column(skip_cert_verify);default(false)" json:"skip_certVerify"`

	// Indicate whether use internal registry addr for the optimizer to pull content
	UseInternalAddr bool `orm:"column(use_internal_addr);default(false)" json:"use_internal_addr"`

	// Indicate if the registration is immutable which is not allowed to remove
	Immutable bool `orm:"column(immutable);default(false)" json:"-"`

	// Optional properties for describing the adapter
	Adapter string `orm:"-" json:"adapter,omitempty"`
	Vendor  string `orm:"-" json:"vendor,omitempty"`
	Version string `orm:"-" json:"version,omitempty"`

	Metadata *v1.OptimizerAdapterMetadata `orm:"-" json:"-"`

	// Timestamps
	CreateTime time.Time `orm:"column(create_time);auto_now_add;type(datetime)" json:"create_time"`
	UpdateTime time.Time `orm:"column(update_time);auto_now;type(datetime)" json:"update_time"`
}

// TableName for Registration
func (r *Registration) TableName() string {
	return "optimizer_registration"
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

	url, err := lib.ValidateHTTPURL(r.URL)
	if err != nil {
		return errors.Wrap(err, "optimizer registration validate")
	}
	r.URL = url

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

// HasCapability returns true when the mime type of the artifact is supported by the optimizer
func (r *Registration) HasCapability(manifestMimeType string) bool {
	if r.Metadata == nil {
		return false
	}

	return r.Metadata.HasCapability(manifestMimeType)
}

// GetRegistryAuthorizationType returns the registry authorization type of the optimizer
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
