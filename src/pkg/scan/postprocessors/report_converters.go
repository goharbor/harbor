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
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

var (
	// Converter is the global native scan report converter
	Converter = NewNativeToRelationalSchemaConverter()
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
		log.G(ctx).Infof("There is no vulnerability report to toSchema for report UUID : %s", reportUUID)
		return reportUUID, "", nil
	}

	// parse the raw report with the V1 schema of the report to the normalized structures
	rawReport := new(vuln.Report)
	if err := json.Unmarshal([]byte(reportData), &rawReport); err != nil {
		return "", "", errors.Wrap(err, "Error when toSchema V1 report to V2")
	}

	if err := c.toSchema(ctx, reportUUID, registrationUUID, digest, reportData); err != nil {
		return "", "", errors.Wrap(err, "Error when converting vulnerability report")
	}

	if err := c.updateReport(ctx, rawReport.Vulnerabilities, reportUUID); err != nil {
		return "", "", errors.Wrap(err, "Error when updating report")
	}

	rawReport.Vulnerabilities = nil
	data, err := json.Marshal(rawReport)
	if err != nil {
		return "", "", errors.Wrap(err, "Error when persisting raw report summary")
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

func (c *nativeToRelationalSchemaConverter) toSchema(ctx context.Context, reportUUID string, registrationUUID string, _ string, rawReportData string) error {
	var vulnReport vuln.Report
	err := json.Unmarshal([]byte(rawReportData), &vulnReport)
	if err != nil {
		return err
	}

	var cveIDs []interface{}
	for _, v := range vulnReport.Vulnerabilities {
		v.Severity = vuln.ParseSeverityVersion3(v.Severity.String())
		cveIDs = append(cveIDs, v.ID)
	}

	records, err := c.dao.List(ctx, q.New(q.KeyWords{"cve_id": q.NewOrList(cveIDs), "registration_uuid": registrationUUID}))
	if err != nil {
		return err
	}

	l := vulnReport.GetVulnerabilityItemList()
	s := lib.Set{}

	var (
		outOfDateRecords []*scan.VulnerabilityRecord
		recordIDs        []int64
	)
	for _, record := range records {
		key := record.Key()

		v, ok := l.GetItem(key)
		if !ok {
			// skip the record which not in the vulnReport.Vulnerabilities
			continue
		}

		s.Add(key)

		recordIDs = append(recordIDs, record.ID)

		if record.Severity != v.Severity.String() {
			record.Severity = v.Severity.String()
			outOfDateRecords = append(outOfDateRecords, record)
		}
	}

	for _, record := range outOfDateRecords {
		// Update the severity of the record when it's changed in the scanner, closes #14745
		if err := c.dao.Update(ctx, record, "severity"); err != nil {
			return err
		}
	}

	if len(outOfDateRecords) > 0 {
		log.G(ctx).Infof("%d vulnerabilities' severity changed", len(outOfDateRecords))
	}

	var newRecords []*scan.VulnerabilityRecord
	for _, v := range vulnReport.Vulnerabilities {
		if !s.Exists(v.Key()) {
			newRecords = append(newRecords, toVulnerabilityRecord(ctx, v, registrationUUID))
		}
	}

	for _, record := range newRecords {
		recordID, err := c.dao.Create(ctx, record)
		if err != nil {
			fields := log.Fields{
				"error":          err,
				"report":         reportUUID,
				"cveID":          record.CVEID,
				"package":        record.Package,
				"packageVersion": record.PackageVersion,
			}
			log.G(ctx).WithFields(fields).Errorf("Could not insert vulnerability record")

			return err
		}

		recordIDs = append(recordIDs, recordID)
	}

	if err := c.dao.InsertForReport(ctx, reportUUID, recordIDs...); err != nil {
		fields := log.Fields{
			"error":  err,
			"report": reportUUID,
		}
		log.G(ctx).WithFields(fields).Errorf("Could not associate vulnerability records to the report")

		return err
	}

	fields := log.Fields{
		"report":               reportUUID,
		"scanner":              registrationUUID,
		"vulnerabilityRecords": len(vulnReport.Vulnerabilities),
	}
	log.G(ctx).WithFields(fields).Infof("Converted vulnerability records to the new schema")

	return nil
}

func (c *nativeToRelationalSchemaConverter) fromSchema(_ context.Context, _ string, artifactDigest string, reportSummary string, records []*scan.VulnerabilityRecord) (string, error) {
	if len(reportSummary) == 0 {
		return "", nil
	}
	vulnerabilityItems := make([]*vuln.VulnerabilityItem, 0)
	for _, record := range records {
		vulnerabilityItems = append(vulnerabilityItems, toVulnerabilityItem(record, artifactDigest))
	}

	rp := new(vuln.Report)
	err := json.Unmarshal([]byte(reportSummary), rp)
	if err != nil {
		return "", err
	}
	if len(vulnerabilityItems) > 0 {
		rp.Vulnerabilities = make([]*vuln.VulnerabilityItem, 0)
		rp.Vulnerabilities = append(rp.Vulnerabilities, vulnerabilityItems...)
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

func toVulnerabilityRecord(ctx context.Context, item *vuln.VulnerabilityItem, registrationUUID string) *scan.VulnerabilityRecord {
	record := new(scan.VulnerabilityRecord)

	record.CVEID = item.ID
	record.Description = item.Description
	record.Package = item.Package
	record.PackageVersion = item.Version
	record.PackageType = "Unknown"
	record.Fix = item.FixVersion
	record.URLs = strings.Join(item.Links, "|")
	record.RegistrationUUID = registrationUUID
	record.Severity = item.Severity.String()

	// process the CVSS scores if the data is available
	if (vuln.CVSS{} != item.CVSSDetails) {
		record.CVE3Score = item.CVSSDetails.ScoreV3
		record.CVE2Score = item.CVSSDetails.ScoreV2
		record.CVSS3Vector = item.CVSSDetails.VectorV3
		record.CVSS2Vector = item.CVSSDetails.VectorV2
	}
	if len(item.CWEIds) > 0 {
		record.CWEIDs = strings.Join(item.CWEIds, ",")
	}

	// marshall the presented vendor attributes as a json string
	if len(item.VendorAttributes) > 0 {
		vendorAttributes, err := json.Marshal(item.VendorAttributes)
		// set the vendor attributes iff unmarshalling is successful
		if err == nil {
			record.VendorAttributes = string(vendorAttributes)
		}

		// parse the NVD score from the vendor attributes
		nvdScore := parseScoreFromVendorAttribute(ctx, string(vendorAttributes))
		if record.CVE3Score == nil {
			record.CVE3Score = &nvdScore
		}
	}

	return record
}

func toVulnerabilityItem(record *scan.VulnerabilityRecord, artifactDigest string) *vuln.VulnerabilityItem {
	item := new(vuln.VulnerabilityItem)

	item.ID = record.CVEID
	item.ArtifactDigests = []string{artifactDigest}
	item.CVSSDetails.ScoreV2 = record.CVE2Score
	item.CVSSDetails.ScoreV3 = record.CVE3Score
	item.CVSSDetails.VectorV2 = record.CVSS2Vector
	item.CVSSDetails.VectorV3 = record.CVSS3Vector
	cweIDs := strings.Split(record.CWEIDs, ",")
	item.CWEIds = append(item.CWEIds, cweIDs...)
	item.Description = record.Description
	item.FixVersion = record.Fix
	item.Version = record.PackageVersion
	urls := strings.Split(record.URLs, "|")
	item.Links = append(item.Links, urls...)
	item.Severity = vuln.ParseSeverityVersion3(record.Severity)
	item.Package = record.Package
	var vendorAttributes map[string]interface{}
	_ = json.Unmarshal([]byte(record.VendorAttributes), &vendorAttributes)
	item.VendorAttributes = vendorAttributes

	return item
}

// updateReport updates the report summary with the vulnerability counts
func (c *nativeToRelationalSchemaConverter) updateReport(ctx context.Context, vulnerabilities []*vuln.VulnerabilityItem, reportUUID string) error {
	log.G(ctx).WithFields(log.Fields{"reportUUID": reportUUID}).Debugf("Update report summary for report")
	CriticalCnt := int64(0)
	HighCnt := int64(0)
	MediumCnt := int64(0)
	LowCnt := int64(0)
	NoneCnt := int64(0)
	UnknownCnt := int64(0)
	FixableCnt := int64(0)

	for _, v := range vulnerabilities {
		v.Severity = vuln.ParseSeverityVersion3(v.Severity.String())
		switch v.Severity {
		case vuln.Critical:
			CriticalCnt++
		case vuln.High:
			HighCnt++
		case vuln.Medium:
			MediumCnt++
		case vuln.Low:
			LowCnt++
		case vuln.None:
			NoneCnt++
		case vuln.Unknown:
			UnknownCnt++
		}
		if len(v.FixVersion) > 0 {
			FixableCnt++
		}
	}

	reports, err := report.Mgr.List(ctx, q.New(q.KeyWords{"uuid": reportUUID}))
	if err != nil {
		return err
	}
	if len(reports) == 0 {
		return errors.New(nil).WithMessage("report not found, uuid:%v", reportUUID)
	}
	r := reports[0]

	r.CriticalCnt = CriticalCnt
	r.HighCnt = HighCnt
	r.MediumCnt = MediumCnt
	r.LowCnt = LowCnt
	r.NoneCnt = NoneCnt
	r.FixableCnt = FixableCnt
	r.UnknownCnt = UnknownCnt

	return report.Mgr.Update(ctx, r, "CriticalCnt", "HighCnt", "MediumCnt", "LowCnt", "NoneCnt", "UnknownCnt", "FixableCnt")
}

// CVS ...
type CVS struct {
	CVSS map[string]map[string]interface{} `json:"CVSS"`
}

func parseScoreFromVendorAttribute(ctx context.Context, vendorAttribute string) float64 {
	var data CVS
	err := json.Unmarshal([]byte(vendorAttribute), &data)
	if err != nil {
		log.G(ctx).Errorf("failed to parse vendor_attribute, error %v", err)
		return 0
	}

	// set the nvd as the first priority, if it's unavailable, return the first V3Score available.
	if val, ok := data.CVSS["nvd"]["V3Score"]; ok {
		return val.(float64)
	}

	for vendor := range data.CVSS {
		if val, ok := data.CVSS[vendor]["V3Score"]; ok {
			return val.(float64)
		}
	}
	return 0
}
