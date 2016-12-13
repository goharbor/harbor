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

package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/coreos/clair/database"
	"github.com/coreos/clair/utils/types"
	"github.com/coreos/pkg/capnslog"
	"github.com/fernet/fernet-go"
)

var log = capnslog.NewPackageLogger("github.com/coreos/clair", "v1")

type Error struct {
	Message string `json:"Layer`
}

type Layer struct {
	Name             string            `json:"Name,omitempty"`
	NamespaceName    string            `json:"NamespaceName,omitempty"`
	Path             string            `json:"Path,omitempty"`
	Headers          map[string]string `json:"Headers,omitempty"`
	ParentName       string            `json:"ParentName,omitempty"`
	Format           string            `json:"Format,omitempty"`
	IndexedByVersion int               `json:"IndexedByVersion,omitempty"`
	Features         []Feature         `json:"Features,omitempty"`
}

func LayerFromDatabaseModel(dbLayer database.Layer, withFeatures, withVulnerabilities bool) Layer {
	layer := Layer{
		Name:             dbLayer.Name,
		IndexedByVersion: dbLayer.EngineVersion,
	}

	if dbLayer.Parent != nil {
		layer.ParentName = dbLayer.Parent.Name
	}

	if dbLayer.Namespace != nil {
		layer.NamespaceName = dbLayer.Namespace.Name
	}

	if withFeatures || withVulnerabilities && dbLayer.Features != nil {
		for _, dbFeatureVersion := range dbLayer.Features {
			feature := Feature{
				Name:          dbFeatureVersion.Feature.Name,
				NamespaceName: dbFeatureVersion.Feature.Namespace.Name,
				Version:       dbFeatureVersion.Version.String(),
				AddedBy:       dbFeatureVersion.AddedBy.Name,
			}

			for _, dbVuln := range dbFeatureVersion.AffectedBy {
				vuln := Vulnerability{
					Name:          dbVuln.Name,
					NamespaceName: dbVuln.Namespace.Name,
					Description:   dbVuln.Description,
					Link:          dbVuln.Link,
					Severity:      string(dbVuln.Severity),
					Metadata:      dbVuln.Metadata,
				}

				if dbVuln.FixedBy != types.MaxVersion {
					vuln.FixedBy = dbVuln.FixedBy.String()
				}
				feature.Vulnerabilities = append(feature.Vulnerabilities, vuln)
			}
			layer.Features = append(layer.Features, feature)
		}
	}

	return layer
}

type Namespace struct {
	Name string `json:"Name,omitempty"`
}

type Vulnerability struct {
	Name          string                 `json:"Name,omitempty"`
	NamespaceName string                 `json:"NamespaceName,omitempty"`
	Description   string                 `json:"Description,omitempty"`
	Link          string                 `json:"Link,omitempty"`
	Severity      string                 `json:"Severity,omitempty"`
	Metadata      map[string]interface{} `json:"Metadata,omitempty"`
	FixedBy       string                 `json:"FixedBy,omitempty"`
	FixedIn       []Feature              `json:"FixedIn,omitempty"`
}

func (v Vulnerability) DatabaseModel() (database.Vulnerability, error) {
	severity := types.Priority(v.Severity)
	if !severity.IsValid() {
		return database.Vulnerability{}, errors.New("Invalid severity")
	}

	var dbFeatures []database.FeatureVersion
	for _, feature := range v.FixedIn {
		dbFeature, err := feature.DatabaseModel()
		if err != nil {
			return database.Vulnerability{}, err
		}

		dbFeatures = append(dbFeatures, dbFeature)
	}

	return database.Vulnerability{
		Name:        v.Name,
		Namespace:   database.Namespace{Name: v.NamespaceName},
		Description: v.Description,
		Link:        v.Link,
		Severity:    severity,
		Metadata:    v.Metadata,
		FixedIn:     dbFeatures,
	}, nil
}

func VulnerabilityFromDatabaseModel(dbVuln database.Vulnerability, withFixedIn bool) Vulnerability {
	vuln := Vulnerability{
		Name:          dbVuln.Name,
		NamespaceName: dbVuln.Namespace.Name,
		Description:   dbVuln.Description,
		Link:          dbVuln.Link,
		Severity:      string(dbVuln.Severity),
		Metadata:      dbVuln.Metadata,
	}

	if withFixedIn {
		for _, dbFeatureVersion := range dbVuln.FixedIn {
			vuln.FixedIn = append(vuln.FixedIn, FeatureFromDatabaseModel(dbFeatureVersion))
		}
	}

	return vuln
}

type Feature struct {
	Name            string          `json:"Name,omitempty"`
	NamespaceName   string          `json:"NamespaceName,omitempty"`
	Version         string          `json:"Version,omitempty"`
	Vulnerabilities []Vulnerability `json:"Vulnerabilities,omitempty"`
	AddedBy         string          `json:"AddedBy,omitempty"`
}

func FeatureFromDatabaseModel(dbFeatureVersion database.FeatureVersion) Feature {
	versionStr := dbFeatureVersion.Version.String()
	if versionStr == types.MaxVersion.String() {
		versionStr = "None"
	}

	return Feature{
		Name:          dbFeatureVersion.Feature.Name,
		NamespaceName: dbFeatureVersion.Feature.Namespace.Name,
		Version:       versionStr,
		AddedBy:       dbFeatureVersion.AddedBy.Name,
	}
}

func (f Feature) DatabaseModel() (database.FeatureVersion, error) {
	var version types.Version
	if f.Version == "None" {
		version = types.MaxVersion
	} else {
		var err error
		version, err = types.NewVersion(f.Version)
		if err != nil {
			return database.FeatureVersion{}, err
		}
	}

	return database.FeatureVersion{
		Feature: database.Feature{
			Name:      f.Name,
			Namespace: database.Namespace{Name: f.NamespaceName},
		},
		Version: version,
	}, nil
}

type Notification struct {
	Name     string                   `json:"Name,omitempty"`
	Created  string                   `json:"Created,omitempty"`
	Notified string                   `json:"Notified,omitempty"`
	Deleted  string                   `json:"Deleted,omitempty"`
	Limit    int                      `json:"Limit,omitempty"`
	Page     string                   `json:"Page,omitempty"`
	NextPage string                   `json:"NextPage,omitempty"`
	Old      *VulnerabilityWithLayers `json:"Old,omitempty"`
	New      *VulnerabilityWithLayers `json:"New,omitempty"`
}

func NotificationFromDatabaseModel(dbNotification database.VulnerabilityNotification, limit int, pageToken string, nextPage database.VulnerabilityNotificationPageNumber, key string) Notification {
	var oldVuln *VulnerabilityWithLayers
	if dbNotification.OldVulnerability != nil {
		v := VulnerabilityWithLayersFromDatabaseModel(*dbNotification.OldVulnerability)
		oldVuln = &v
	}

	var newVuln *VulnerabilityWithLayers
	if dbNotification.NewVulnerability != nil {
		v := VulnerabilityWithLayersFromDatabaseModel(*dbNotification.NewVulnerability)
		newVuln = &v
	}

	var nextPageStr string
	if nextPage != database.NoVulnerabilityNotificationPage {
		nextPageBytes, _ := tokenMarshal(nextPage, key)
		nextPageStr = string(nextPageBytes)
	}

	var created, notified, deleted string
	if !dbNotification.Created.IsZero() {
		created = fmt.Sprintf("%d", dbNotification.Created.Unix())
	}
	if !dbNotification.Notified.IsZero() {
		notified = fmt.Sprintf("%d", dbNotification.Notified.Unix())
	}
	if !dbNotification.Deleted.IsZero() {
		deleted = fmt.Sprintf("%d", dbNotification.Deleted.Unix())
	}

	// TODO(jzelinskie): implement "changed" key
	fmt.Println(dbNotification.Deleted.IsZero())
	return Notification{
		Name:     dbNotification.Name,
		Created:  created,
		Notified: notified,
		Deleted:  deleted,
		Limit:    limit,
		Page:     pageToken,
		NextPage: nextPageStr,
		Old:      oldVuln,
		New:      newVuln,
	}
}

type VulnerabilityWithLayers struct {
	Vulnerability                  *Vulnerability `json:"Vulnerability,omitempty"`
	LayersIntroducingVulnerability []string       `json:"LayersIntroducingVulnerability,omitempty"`
}

func VulnerabilityWithLayersFromDatabaseModel(dbVuln database.Vulnerability) VulnerabilityWithLayers {
	vuln := VulnerabilityFromDatabaseModel(dbVuln, true)

	var layers []string
	for _, layer := range dbVuln.LayersIntroducingVulnerability {
		layers = append(layers, layer.Name)
	}

	return VulnerabilityWithLayers{
		Vulnerability:                  &vuln,
		LayersIntroducingVulnerability: layers,
	}
}

type LayerEnvelope struct {
	Layer *Layer `json:"Layer,omitempty"`
	Error *Error `json:"Error,omitempty"`
}

type NamespaceEnvelope struct {
	Namespaces *[]Namespace `json:"Namespaces,omitempty"`
	Error      *Error       `json:"Error,omitempty"`
}

type VulnerabilityEnvelope struct {
	Vulnerability   *Vulnerability   `json:"Vulnerability,omitempty"`
	Vulnerabilities *[]Vulnerability `json:"Vulnerabilities,omitempty"`
	NextPage        string           `json:"NextPage,omitempty"`
	Error           *Error           `json:"Error,omitempty"`
}

type NotificationEnvelope struct {
	Notification *Notification `json:"Notification,omitempty"`
	Error        *Error        `json:"Error,omitempty"`
}

type FeatureEnvelope struct {
	Feature  *Feature   `json:"Feature,omitempty"`
	Features *[]Feature `json:"Features,omitempty"`
	Error    *Error     `json:"Error,omitempty"`
}

func tokenUnmarshal(token string, key string, v interface{}) error {
	k, _ := fernet.DecodeKey(key)
	msg := fernet.VerifyAndDecrypt([]byte(token), time.Hour, []*fernet.Key{k})
	if msg == nil {
		return errors.New("invalid or expired pagination token")
	}

	return json.NewDecoder(bytes.NewBuffer(msg)).Decode(&v)
}

func tokenMarshal(v interface{}, key string) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(v)
	if err != nil {
		return nil, err
	}

	k, _ := fernet.DecodeKey(key)
	return fernet.EncryptAndSign(buf.Bytes(), k)
}
