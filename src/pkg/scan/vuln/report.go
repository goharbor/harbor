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

package vuln

import (
	"encoding/json"

	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// Report model for vulnerability scan
type Report struct {
	// Time of generating this report
	GeneratedAt string `json:"generated_at"`
	// Scanner of generating this report
	Scanner *v1.Scanner `json:"scanner"`
	// A standard scale for measuring the severity of a vulnerability.
	Severity Severity `json:"severity"`
	// Vulnerability list
	Vulnerabilities []*VulnerabilityItem `json:"vulnerabilities"`
}

// MarshalJSON custom function to dump nil slice of Vulnerabilities as empty slice
// See https://github.com/goharbor/harbor/issues/11131 to get more details
func (report *Report) MarshalJSON() ([]byte, error) {
	type Alias Report

	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(report),
	}

	if aux.Vulnerabilities == nil {
		aux.Vulnerabilities = []*VulnerabilityItem{}
	}

	return json.Marshal(aux)
}

// Merge ...
func (report *Report) Merge(another *Report) *Report {
	generatedAt := report.GeneratedAt
	if another.GeneratedAt > generatedAt {
		generatedAt = another.GeneratedAt
	}

	vulnerabilities := report.Vulnerabilities
	if vulnerabilities == nil {
		vulnerabilities = another.Vulnerabilities
	} else {
		vulnerabilities = append(vulnerabilities, another.Vulnerabilities...)
	}

	r := &Report{
		GeneratedAt:     generatedAt,
		Scanner:         report.Scanner,
		Severity:        mergeSeverity(report.Severity, another.Severity),
		Vulnerabilities: vulnerabilities,
	}

	return r
}

// WithArtifactDigest set artifact digest for the report
func (report *Report) WithArtifactDigest(artifactDigest string) {
	for _, vul := range report.Vulnerabilities {
		vul.ArtifactDigest = artifactDigest
	}
}

// VulnerabilityItem represents one found vulnerability
type VulnerabilityItem struct {
	// The unique identifier of the vulnerability.
	// e.g: CVE-2017-8283
	ID string `json:"id"`
	// An operating system or software dependency package containing the vulnerability.
	// e.g: dpkg
	Package string `json:"package"`
	// The version of the package containing the vulnerability.
	// e.g: 1.17.27
	Version string `json:"version"`
	// The version of the package containing the fix if available.
	// e.g: 1.18.0
	FixVersion string `json:"fix_version"`
	// A standard scale for measuring the severity of a vulnerability.
	Severity Severity `json:"severity"`
	// example: dpkg-source in dpkg 1.3.0 through 1.18.23 is able to use a non-GNU patch program
	// and does not offer a protection mechanism for blank-indented diff hunks, which allows remote
	// attackers to conduct directory traversal attacks via a crafted Debian source package, as
	// demonstrated by using of dpkg-source on NetBSD.
	Description string `json:"description"`
	// The list of link to the upstream database with the full description of the vulnerability.
	// Format: URI
	// e.g: List [ "https://security-tracker.debian.org/tracker/CVE-2017-8283" ]
	Links []string `json:"links"`
	// The artifact digest which the vulnerability belonged
	// e.g: sha256@ee1d00c5250b5a886b09be2d5f9506add35dfb557f1ef37a7e4b8f0138f32956
	ArtifactDigest string `json:"artifact_digest"`
}
