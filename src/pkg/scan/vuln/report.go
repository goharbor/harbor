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
	Links []string
}
