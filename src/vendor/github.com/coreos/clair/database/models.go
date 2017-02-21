// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package database

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/coreos/clair/utils/types"
)

// ID is only meant to be used by database implementations and should never be used for anything else.
type Model struct {
	ID int
}

type Layer struct {
	Model

	Name          string
	EngineVersion int
	Parent        *Layer
	Namespace     *Namespace
	Features      []FeatureVersion
}

type Namespace struct {
	Model

	Name string
}

type Feature struct {
	Model

	Name      string
	Namespace Namespace
}

type FeatureVersion struct {
	Model

	Feature    Feature
	Version    types.Version
	AffectedBy []Vulnerability

	// For output purposes. Only make sense when the feature version is in the context of an image.
	AddedBy Layer
}

type Vulnerability struct {
	Model

	Name      string
	Namespace Namespace

	Description string
	Link        string
	Severity    types.Priority

	Metadata MetadataMap

	FixedIn                        []FeatureVersion
	LayersIntroducingVulnerability []Layer

	// For output purposes. Only make sense when the vulnerability
	// is already about a specific Feature/FeatureVersion.
	FixedBy types.Version `json:",omitempty"`
}

type MetadataMap map[string]interface{}

func (mm *MetadataMap) Scan(value interface{}) error {
	val, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(val, mm)
}

func (mm *MetadataMap) Value() (driver.Value, error) {
	json, err := json.Marshal(*mm)
	return string(json), err
}

type VulnerabilityNotification struct {
	Model

	Name string

	Created  time.Time
	Notified time.Time
	Deleted  time.Time

	OldVulnerability *Vulnerability
	NewVulnerability *Vulnerability
}

type VulnerabilityNotificationPageNumber struct {
	// -1 means that we reached the end already.
	OldVulnerability int
	NewVulnerability int
}

var VulnerabilityNotificationFirstPage = VulnerabilityNotificationPageNumber{0, 0}
var NoVulnerabilityNotificationPage = VulnerabilityNotificationPageNumber{-1, -1}
