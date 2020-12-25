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

package postprocessors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

// NativeScanReportConverter is an interface that establishes the contract for the conversion process of a harbor native vulnerability report
// It is the responsibility of the implementation to store the report in a manner easily retrievable using the
// report UUID
type NativeScanReportConverter interface {
	ToRelationalSchema(ctx context.Context, reportUUID string, registrationUUID string, digest string, reportData string) (string, string, error)
	FromRelationalSchema(ctx context.Context, reportUUID string, artifactDigest string, reportSummary string) (string, error)
}

// nativeToRelationalSchemaConverter is responsible for converting the JSON scan report from the Harbor 1.0 format to
// the generic vulnerability format which follows a normalized storage schema.
type nativeToRelationalSchemaConverter struct {
	dao scan.VulnerabilityRecordDao
}

// NewNativeToRelationalSchemaConverter returns a new instance of a V1 report to V2 report converter
func NewNativeToRelationalSchemaConverter() NativeScanReportConverter {
	return &nativeToRelationalSchemaConverter{dao: scan.NewVulnerabilityRecordDao()}
}

// ToRelationalSchema converts the vulnerability report data present as JSON  to the new relational VulnerabilityRecord instance
func (c *nativeToRelationalSchemaConverter) ToRelationalSchema(ctx context.Context, reportUUID string, registrationUUID string, digest string, reportData string) (string, string, error) {

	if len(reportData) == 0 {
		log.Infof("There is no vulnerability report to toSchema for report UUID : %s", reportUUID)
		return reportUUID, "", nil
	}

	// parse the raw report with the V1 schema of the report to the normalized structures
	rawReport := new(vuln.Report)
	if err := json.Unmarshal([]byte(reportData), &rawReport); err != nil {
		return "", "", errors.Wrap(err, fmt.Sprintf("Error when toSchema V1 report to V2"))
	}

	if err := c.toSchema(ctx, reportUUID, registrationUUID, digest, reportData); err != nil {
		return "", "", errors.Wrap(err, fmt.Sprintf("Error when converting vulnerability report"))
	}
	rawReport.Vulnerabilities = nil
	data, err := json.Marshal(rawReport)
	if err != nil {
		return "", "", errors.Wrap(err, fmt.Sprintf("Error when persisting raw report summary"))
	}
	return reportUUID, string(data), nil

}

// FromRelationalSchema converts the generic vulnerability record stored in relational form to the
// native JSON blob.
func (c *nativeToRelationalSchemaConverter) FromRelationalSchema(ctx context.Context, reportUUID string, artifactDigest string, reportSummary string) (string, error) {
	vulns, err := c.dao.GetForReport(ctx, reportUUID)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error when toSchema generic vulnerability records for %s", reportUUID))
	}
	rp, err := c.fromSchema(ctx, reportUUID, artifactDigest, reportSummary, vulns)
	if err != nil {
		return "", err
	}
	return rp, nil
}

func (c *nativeToRelationalSchemaConverter) toSchema(ctx context.Context, reportUUID string, registrationUUID string, digest string, rawReportData string) error {

	var vulnReport vuln.Report
	err := json.Unmarshal([]byte(rawReportData), &vulnReport)
	if err != nil {
		return err
	}
	for _, v := range vulnReport.Vulnerabilities {
		vulnV2 := new(scan.VulnerabilityRecord)
		vulnV2.CVEID = v.ID
		vulnV2.Description = v.Description
		vulnV2.Package = v.Package
		vulnV2.PackageVersion = v.Version
		vulnV2.PackageType = "Unknown"
		vulnV2.Fix = v.FixVersion
		vulnV2.URLs = strings.Join(v.Links, "|")
		vulnV2.RegistrationUUID = registrationUUID
		vulnV2.Severity = v.Severity.String()

		// process the CVSS scores if the data is available
		if (vuln.CVSS{} != v.CVSSDetails) {
			vulnV2.CVE3Score = v.CVSSDetails.ScoreV3
			vulnV2.CVE2Score = v.CVSSDetails.ScoreV2
			vulnV2.CVSS3Vector = v.CVSSDetails.VectorV3
			vulnV2.CVSS2Vector = v.CVSSDetails.VectorV2
		}
		if len(v.CWEIds) > 0 {
			vulnV2.CWEIDs = strings.Join(v.CWEIds, ",")
		}

		// marshall the presented vendor attributes as a json string
		if len(v.VendorAttributes) > 0 {
			vendorAttributes, err := json.Marshal(v.VendorAttributes)
			// set the vendor attributes iff unmarshalling is successful
			if err == nil {
				vulnV2.VendorAttributes = string(vendorAttributes)
			}
		}

		_, err = c.dao.InsertForReport(ctx, reportUUID, vulnV2)
		if err != nil {
			log.Warningf("Could not insert vulnerability record -  report: %s, cve_id: %s, scanner: %s, package: %s, package_version: %s", reportUUID, v.ID, registrationUUID, v.Package, v.Version)
		}

	}
	log.Infof("Converted %d vulnerability records to the new schema for report ID %s and scanner Id %s", len(vulnReport.Vulnerabilities), reportUUID, registrationUUID)
	return nil
}

func (c *nativeToRelationalSchemaConverter) fromSchema(ctx context.Context, reportUUID string, artifactDigest string, reportSummary string, records []*scan.VulnerabilityRecord) (string, error) {
	if len(reportSummary) == 0 {
		return "", nil
	}
	vulnerabilityItems := make([]*vuln.VulnerabilityItem, 0)
	for _, record := range records {
		vi := new(vuln.VulnerabilityItem)
		vi.ID = record.CVEID
		vi.ArtifactDigest = artifactDigest
		vi.CVSSDetails.ScoreV2 = record.CVE2Score
		vi.CVSSDetails.ScoreV3 = record.CVE3Score
		vi.CVSSDetails.VectorV2 = record.CVSS2Vector
		vi.CVSSDetails.VectorV3 = record.CVSS3Vector
		cweIDs := strings.Split(record.CWEIDs, ",")
		for _, cweID := range cweIDs {
			vi.CWEIds = append(vi.CWEIds, cweID)
		}
		vi.CWEIds = cweIDs
		vi.Description = record.Description
		vi.FixVersion = record.Fix
		vi.Version = record.PackageVersion
		urls := strings.Split(record.URLs, "|")
		for _, url := range urls {
			vi.Links = append(vi.Links, url)
		}
		vi.Severity = vuln.ParseSeverityVersion3(record.Severity)
		vi.Package = record.Package
		var vendorAttributes map[string]interface{}
		_ = json.Unmarshal([]byte(record.VendorAttributes), &vendorAttributes)
		vi.VendorAttributes = vendorAttributes
		vulnerabilityItems = append(vulnerabilityItems, vi)
	}
	rp := new(vuln.Report)
	err := json.Unmarshal([]byte(reportSummary), rp)
	if err != nil {
		return "", err
	}
	if len(vulnerabilityItems) > 0 {
		rp.Vulnerabilities = make([]*vuln.VulnerabilityItem, 0)
		for _, v := range vulnerabilityItems {
			rp.Vulnerabilities = append(rp.Vulnerabilities, v)
		}
	}

	data, err := json.Marshal(rp)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GetNativeV1ReportFromResolvedData returns the native V1 scan report from the resolved
// interface data.
func (c *nativeToRelationalSchemaConverter) getNativeV1ReportFromResolvedData(ctx job.Context, rp interface{}) (*vuln.Report, error) {
	report, ok := rp.(*vuln.Report)
	if !ok {
		return nil, errors.New("Data cannot be converted to v1 report format")
	}
	ctx.GetLogger().Infof("Converted raw data to report. Count of Vulnerabilities in report : %d", len(report.Vulnerabilities))
	return report, nil
}
