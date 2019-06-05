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

package models

import (
	"time"
)

// ClairVulnTimestampTable is the name of the table that tracks the timestamp of vulnerability in Clair.
const ClairVulnTimestampTable = "clair_vuln_timestamp"

// ClairVulnTimestamp represents a record in DB that tracks the timestamp of vulnerability in Clair.
type ClairVulnTimestamp struct {
	ID            int64     `orm:"pk;auto;column(id)" json:"-"`
	Namespace     string    `orm:"column(namespace)" json:"namespace"`
	LastUpdate    time.Time `orm:"column(last_update)" json:"-"`
	LastUpdateUTC int64     `orm:"-" json:"last_update"`
}

// TableName is required by beego to map struct to table.
func (ct *ClairVulnTimestamp) TableName() string {
	return ClairVulnTimestampTable
}

// ClairLayer ...
type ClairLayer struct {
	Name           string            `json:"Name,omitempty"`
	NamespaceNames []string          `json:"NamespaceNames,omitempty"`
	Path           string            `json:"Path,omitempty"`
	Headers        map[string]string `json:"Headers,omitempty"`
	ParentName     string            `json:"ParentName,omitempty"`
	Format         string            `json:"Format,omitempty"`
	Features       []ClairFeature    `json:"Features,omitempty"`
}

// ClairFeature ...
type ClairFeature struct {
	Name            string               `json:"Name,omitempty"`
	NamespaceName   string               `json:"NamespaceName,omitempty"`
	VersionFormat   string               `json:"VersionFormat,omitempty"`
	Version         string               `json:"Version,omitempty"`
	Vulnerabilities []ClairVulnerability `json:"Vulnerabilities,omitempty"`
	AddedBy         string               `json:"AddedBy,omitempty"`
}

// ClairVulnerability ...
type ClairVulnerability struct {
	Name          string                 `json:"Name,omitempty"`
	NamespaceName string                 `json:"NamespaceName,omitempty"`
	Description   string                 `json:"Description,omitempty"`
	Link          string                 `json:"Link,omitempty"`
	Severity      string                 `json:"Severity,omitempty"`
	Metadata      map[string]interface{} `json:"Metadata,omitempty"`
	FixedBy       string                 `json:"FixedBy,omitempty"`
	FixedIn       []ClairFeature         `json:"FixedIn,omitempty"`
}

// ClairError ...
type ClairError struct {
	Message string `json:"Message,omitempty"`
}

// ClairLayerEnvelope ...
type ClairLayerEnvelope struct {
	Layer *ClairLayer `json:"Layer,omitempty"`
	Error *ClairError `json:"Error,omitempty"`
}

// ClairNotification ...
type ClairNotification struct {
	Name     string                        `json:"Name,omitempty"`
	Created  string                        `json:"Created,omitempty"`
	Notified string                        `json:"Notified,omitempty"`
	Deleted  string                        `json:"Deleted,omitempty"`
	Limit    int                           `json:"Limit,omitempty"`
	Page     string                        `json:"Page,omitempty"`
	NextPage string                        `json:"NextPage,omitempty"`
	Old      *ClairVulnerabilityWithLayers `json:"Old,omitempty"`
	New      *ClairVulnerabilityWithLayers `json:"New,omitempty"`
}

// ClairNotificationEnvelope ...
type ClairNotificationEnvelope struct {
	Notification *ClairNotification `json:"Notification,omitempty"`
	Error        *ClairError        `json:"Error,omitempty"`
}

// ClairVulnerabilityWithLayers ...
type ClairVulnerabilityWithLayers struct {
	Vulnerability                         *ClairVulnerability     `json:"Vulnerability,omitempty"`
	OrderedLayersIntroducingVulnerability []ClairOrderedLayerName `json:"OrderedLayersIntroducingVulnerability,omitempty"`
}

// ClairOrderedLayerName ...
type ClairOrderedLayerName struct {
	Index     int    `json:"Index"`
	LayerName string `json:"LayerName"`
}

// ClairVulnerabilityStatus reflects the readiness and freshness of vulnerability data in Clair,
// which will be returned in response of systeminfo API.
type ClairVulnerabilityStatus struct {
	OverallUTC int64                     `json:"overall_last_update,omitempty"`
	Details    []ClairNamespaceTimestamp `json:"details,omitempty"`
}

// ClairNamespaceTimestamp is a record to store the clair namespace and the timestamp,
// in practice different namespace in Clair maybe merged into one, e.g. ubuntu:14.04 and ubuntu:16.4 maybe merged into ubuntu and put into response.
type ClairNamespaceTimestamp struct {
	Namespace string `json:"namespace"`
	Timestamp int64  `json:"last_update"`
}

// ClairNamespace ...
type ClairNamespace struct {
	Name          string `json:"Name,omitempty"`
	VersionFormat string `json:"VersionFormat,omitempty"`
}

// ClairNamespaceEnvelope ...
type ClairNamespaceEnvelope struct {
	Namespaces *[]ClairNamespace `json:"Namespaces,omitempty"`
	Error      *ClairError       `json:"Error,omitempty"`
}
