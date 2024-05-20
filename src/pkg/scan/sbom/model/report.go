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

import v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"

// Report sbom report.
// Identified by the `artifact_id`, `registration_uuid` and `mime_type`.
type Report struct {
	ID               int64  `orm:"pk;auto;column(id)"`
	UUID             string `orm:"unique;column(uuid)"`
	ArtifactID       int64  `orm:"column(artifact_id)"`
	RegistrationUUID string `orm:"column(registration_uuid)"`
	MimeType         string `orm:"column(mime_type)"`
	MediaType        string `orm:"column(media_type)"`
	ReportSummary    string `orm:"column(report);type(json)"`
}

// TableName for sbom report
func (r *Report) TableName() string {
	return "sbom_report"
}

// RawSBOMReport the original report of the sbom report get from scanner
type RawSBOMReport struct {
	// Time of generating this report
	GeneratedAt string `json:"generated_at"`
	// Scanner of generating this report
	Scanner *v1.Scanner `json:"scanner"`
	// MediaType the media type of the report, e.g. application/spdx+json
	MediaType string `json:"media_type"`
	// SBOM sbom content
	SBOM map[string]interface{} `json:"sbom,omitempty"`
}
