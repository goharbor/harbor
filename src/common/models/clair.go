// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

//ClairLayer ...
type ClairLayer struct {
	Name           string            `json:"Name,omitempty"`
	NamespaceNames []string          `json:"NamespaceNames,omitempty"`
	Path           string            `json:"Path,omitempty"`
	Headers        map[string]string `json:"Headers,omitempty"`
	ParentName     string            `json:"ParentName,omitempty"`
	Format         string            `json:"Format,omitempty"`
	Features       []ClairFeature    `json:"Features,omitempty"`
}

//ClairFeature ...
type ClairFeature struct {
	Name            string               `json:"Name,omitempty"`
	NamespaceName   string               `json:"NamespaceName,omitempty"`
	VersionFormat   string               `json:"VersionFormat,omitempty"`
	Version         string               `json:"Version,omitempty"`
	Vulnerabilities []ClairVulnerability `json:"Vulnerabilities,omitempty"`
	AddedBy         string               `json:"AddedBy,omitempty"`
}

//ClairVulnerability ...
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

//ClairError ...
type ClairError struct {
	Message string `json:"Message,omitempty"`
}

//ClairLayerEnvelope ...
type ClairLayerEnvelope struct {
	Layer *ClairLayer `json:"Layer,omitempty"`
	Error *ClairError `json:"Error,omitempty"`
}
