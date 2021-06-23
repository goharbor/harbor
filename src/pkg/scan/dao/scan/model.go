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

package scan

import (
	"fmt"
	"time"
)

// Report of the scan.
// Identified by the `digest`, `registration_uuid` and `mime_type`.
type Report struct {
	ID               int64  `orm:"pk;auto;column(id)"`
	UUID             string `orm:"unique;column(uuid)"`
	Digest           string `orm:"column(digest)"`
	RegistrationUUID string `orm:"column(registration_uuid)"`
	MimeType         string `orm:"column(mime_type)"`
	Report           string `orm:"column(report);type(json)"`

	Status    string    `orm:"-"`
	StartTime time.Time `orm:"-"`
	EndTime   time.Time `orm:"-"`
}

// TableName for Report
func (r *Report) TableName() string {
	return "scan_report"
}

// TableUnique for Report
func (r *Report) TableUnique() [][]string {
	return [][]string{
		{"uuid"},
		{"digest", "registration_uuid", "mime_type"},
	}
}

// VulnerabilityRecord of an individual vulnerability. Identifies an  individual vulnerability item in the scan.
// Since multiple scanners could be registered with the projects, each scanner
// would have it's own definition for the same CVE ID. Hence a CVE ID is qualified along
// with the ID of the scanner that owns the CVE record definition.
// The scanner ID would be the same as the RegistrationUUID field of Report.
// Identified by the `cve_id` and `registration_uuid`.
// Relates to the image using the `digest` and to the report using the `report UUID` field
type VulnerabilityRecord struct {
	ID               int64    `orm:"pk;auto;column(id)"`
	CVEID            string   `orm:"column(cve_id)"`
	RegistrationUUID string   `orm:"column(registration_uuid)"`
	Package          string   `orm:"column(package)"`
	PackageVersion   string   `orm:"column(package_version)"`
	PackageType      string   `orm:"column(package_type)"`
	Severity         string   `orm:"column(severity)"`
	Fix              string   `orm:"column(fixed_version);null"`
	URLs             string   `orm:"column(urls);null"`
	CVE3Score        *float64 `orm:"column(cvss_score_v3);null"`
	CVE2Score        *float64 `orm:"column(cvss_score_v2);null"`
	CVSS3Vector      string   `orm:"column(cvss_vector_v3);null"` // e.g. CVSS:3.0/AV:L/AC:L/PR:L/UI:N/S:U/C:H/I:N/A:N
	CVSS2Vector      string   `orm:"column(cvss_vector_v2);null"` // e.g. AV:L/AC:M/Au:N/C:P/I:N/A:N
	Description      string   `orm:"column(description);null"`
	CWEIDs           string   `orm:"column(cwe_ids);null"` // e.g. CWE-476,CWE-123,CWE-234
	VendorAttributes string   `orm:"column(vendor_attributes);type(json);null"`
}

// TableName for VulnerabilityRecord
func (vr *VulnerabilityRecord) TableName() string {
	return "vulnerability_record"
}

// TableUnique for VulnerabilityRecord
func (vr *VulnerabilityRecord) TableUnique() [][]string {
	return [][]string{
		{"cve_id", "registration_uuid", "package", "package_version"},
	}
}

// GetID returns the ID of the record
func (vr *VulnerabilityRecord) GetID() int64 {
	return vr.ID
}

// Key returns the uniq key of the vuln
func (vr *VulnerabilityRecord) Key() string {
	return fmt.Sprintf("%s-%s-%s", vr.CVEID, vr.Package, vr.PackageVersion)
}

// ReportVulnerabilityRecord is relation table required to optimize data storage for both the
// vulnerability records and the scan report.
// identified by composite key (ID, Report)
// Since each scan report has a separate UUID, the composite key
// would ensure that the immutability of the historical scan reports is guaranteed.
// It is sufficient to store the int64 VulnerabilityRecord Id since the vulnerability records
// are uniquely identified in the table based on the ScannerID and the CVEID
type ReportVulnerabilityRecord struct {
	ID           int64  `orm:"pk;auto;column(id)"`
	Report       string `orm:"column(report_uuid);"`
	VulnRecordID int64  `orm:"column(vuln_record_id);"`
}

// TableName for ReportVulnerabilityRecord
func (rvr *ReportVulnerabilityRecord) TableName() string {
	return "report_vulnerability_record"
}

// TableUnique for ReportVulnerabilityRecord
func (rvr *ReportVulnerabilityRecord) TableUnique() [][]string {
	return [][]string{
		{"report_uuid", "vuln_record_id"},
	}
}

// GetID returns the ID of the record
func (rvr *ReportVulnerabilityRecord) GetID() int64 {
	return rvr.ID
}
