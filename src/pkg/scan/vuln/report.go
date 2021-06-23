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
	"fmt"

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

	vulnerabilityItemList *VulnerabilityItemList
}

// GetVulnerabilityItemList returns VulnerabilityItemList from the Vulnerabilities of report
func (report *Report) GetVulnerabilityItemList() *VulnerabilityItemList {
	l := report.vulnerabilityItemList
	if l == nil {
		l = &VulnerabilityItemList{}
		l.Add(report.Vulnerabilities...)

		report.vulnerabilityItemList = l
	}

	return l
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
	scanner := report.Scanner
	generatedAt := report.GeneratedAt
	if another.GeneratedAt > report.GeneratedAt {
		generatedAt = another.GeneratedAt

		// choose the scanner from the newer summary
		// because the generatedAt of the report is from the newer report
		scanner = another.Scanner
	}

	l := report.GetVulnerabilityItemList()
	l.Add(another.Vulnerabilities...)

	r := &Report{
		GeneratedAt:           generatedAt,
		Scanner:               scanner,
		Severity:              mergeSeverity(report.Severity, another.Severity),
		Vulnerabilities:       l.Items(),
		vulnerabilityItemList: l,
	}

	return r
}

// WithArtifactDigest set artifact digest for the report
func (report *Report) WithArtifactDigest(artifactDigest string) {
	for _, vul := range report.Vulnerabilities {
		vul.ArtifactDigests = []string{artifactDigest}
	}
}

// NewVulnerabilityItemList returns VulnerabilityItemList from lists
func NewVulnerabilityItemList(lists ...*VulnerabilityItemList) *VulnerabilityItemList {
	var availableLists []*VulnerabilityItemList
	for _, li := range lists {
		if li != nil {
			availableLists = append(availableLists, li)
		}
	}

	if len(availableLists) == 0 {
		return nil
	}

	l := &VulnerabilityItemList{}
	for _, li := range availableLists {
		l.Add(li.Items()...)
	}

	return l
}

// VulnerabilityItemList the list can skip the VulnerabilityItem exists in the list when adding
type VulnerabilityItemList struct {
	items   []*VulnerabilityItem
	indexed map[string]*VulnerabilityItem
}

// Items returns the vulnerabilities in the l
func (l *VulnerabilityItemList) Items() []*VulnerabilityItem {
	return l.items
}

// Add add item to the list when the item not exists in list
func (l *VulnerabilityItemList) Add(items ...*VulnerabilityItem) {
	if l.indexed == nil {
		l.indexed = map[string]*VulnerabilityItem{}
	}

	for _, item := range items {
		key := item.Key()
		if v, ok := l.indexed[key]; ok {
			v.ArtifactDigests = append(v.ArtifactDigests, item.ArtifactDigests...)
		} else {
			l.items = append(l.items, item)
			l.indexed[key] = item
		}
	}
}

// GetItem returns VulnerabilityItem by key
func (l *VulnerabilityItemList) GetItem(key string) (*VulnerabilityItem, bool) {
	item, ok := l.indexed[key]

	return item, ok
}

// GetSeveritySummary returns the severity and summary of l
func (l *VulnerabilityItemList) GetSeveritySummary() (Severity, *VulnerabilitySummary) {
	if l == nil {
		return Severity(""), nil
	}

	sum := &VulnerabilitySummary{
		Total:   len(l.Items()),
		Summary: make(SeveritySummary),
	}

	severity := None
	for _, v := range l.Items() {
		if num, ok := sum.Summary[v.Severity]; ok {
			sum.Summary[v.Severity] = num + 1
		} else {
			sum.Summary[v.Severity] = 1
		}

		// Update the overall severity if necessary
		if v.Severity.Code() > severity.Code() {
			severity = v.Severity
		}
		// If the CVE item has a fixable version
		if len(v.FixVersion) > 0 {
			sum.Fixable++
		}
	}

	return severity, sum
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
	// The artifact digests which the vulnerability belonged
	// e.g: sha256@ee1d00c5250b5a886b09be2d5f9506add35dfb557f1ef37a7e4b8f0138f32956
	ArtifactDigests []string `json:"artifact_digests"`
	// The CVSS3 and CVSS2 based scores and attack vector for the vulnerability item
	CVSSDetails CVSS `json:"preferred_cvss"`
	// A separated list of CWE Ids associated with this vulnerability
	// e.g. CWE-465,CWE-124
	CWEIds []string `json:"cwe_ids"`
	// A collection of vendor specific attributes for the vulnerability item
	// with each attribute represented as a key-value pair.
	VendorAttributes map[string]interface{} `json:"vendor_attributes"`
}

// Key returns the uniq key for the item
func (item *VulnerabilityItem) Key() string {
	return fmt.Sprintf("%s-%s-%s", item.ID, item.Package, item.Version)
}

// CVSS holds the score and attack vector for the vulnerability based on the CVSS3 and CVSS2 standards
type CVSS struct {
	// The CVSS-3 score for the vulnerability
	// e.g. 2.5
	ScoreV3 *float64 `json:"score_v3"`
	// The CVSS-3 score for the vulnerability
	// e.g. 2.5
	ScoreV2 *float64 `json:"score_v2"`
	// The CVSS-3 attack vector.
	// e.g. CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N
	VectorV3 string `json:"vector_v3"`
	// The CVSS-3 attack vector.
	// e.g. AV:L/AC:M/Au:N/C:P/I:N/A:N
	VectorV2 string `json:"vector_v2"`
}
