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

package exporter

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

const (
	vulnTestScannerUUID = "exporter-vuln-test-scanner"
)

// VulnerabilityCollectorTestSuite tests the vulnerability collector. It relies on the
// projects, repositories and artifacts created by setupTest in project_collector_test.go.
type VulnerabilityCollectorTestSuite struct {
	suite.Suite
	reportDao scan.DAO
	vulnDao   scan.VulnerabilityRecordDao
	reports   []string
}

func (v *VulnerabilityCollectorTestSuite) SetupSuite() {
	ctx := orm.Context()
	v.reportDao = scan.New()
	v.vulnDao = scan.NewVulnerabilityRecordDao()

	_, err := scanner.AddRegistration(ctx, &scanner.Registration{
		UUID:      vulnTestScannerUUID,
		Name:      "exporter-vuln-test-scanner",
		URL:       "http://exporter-vuln-test-scanner:8080",
		Disabled:  false,
		IsDefault: false,
	})
	v.Require().NoError(err)

	// art1 (project test1): a native report with 2 critical, 1 high, 3 low, 2 fixable
	v.createReport("exporter-vuln-report-1", art1.Digest, v1.MimeTypeNativeReport, &scan.Report{
		CriticalCnt: 2, HighCnt: 1, LowCnt: 3, FixableCnt: 2,
	})
	// art2 (project test2): a native report and a generic report for the same artifact,
	// only the native one must be counted
	v.createReport("exporter-vuln-report-2", art2.Digest, v1.MimeTypeNativeReport, &scan.Report{
		MediumCnt: 4, UnknownCnt: 1, NoneCnt: 2, FixableCnt: 1,
	})
	v.createReport("exporter-vuln-report-3", art2.Digest, v1.MimeTypeGenericVulnerabilityReport, &scan.Report{
		CriticalCnt: 100, HighCnt: 100, MediumCnt: 100, FixableCnt: 100,
	})
	// a report of an artifact that does not exist must be ignored
	v.createReport("exporter-vuln-report-4", "sha256:0000000000000000000000000000000000000000000000000000000000000000", v1.MimeTypeNativeReport, &scan.Report{
		CriticalCnt: 50,
	})

	// vulnerability records: CVE-1 is shared by both artifacts and must be counted once
	cve1 := v.createVulnRecord("CVE-2024-0001", "openssl", "1.0", "Critical")
	cve2 := v.createVulnRecord("CVE-2024-0002", "bash", "5.0", "Critical")
	cve3 := v.createVulnRecord("CVE-2024-0003", "zlib", "1.2", "Medium")
	cve4 := v.createVulnRecord("CVE-2024-0004", "curl", "7.0", "High")
	v.Require().NoError(v.vulnDao.InsertForReport(ctx, "exporter-vuln-report-1", cve1, cve2))
	v.Require().NoError(v.vulnDao.InsertForReport(ctx, "exporter-vuln-report-2", cve1, cve3))
	// only linked to the report of the missing artifact
	v.Require().NoError(v.vulnDao.InsertForReport(ctx, "exporter-vuln-report-4", cve4))
}

func (v *VulnerabilityCollectorTestSuite) TearDownSuite() {
	ctx := orm.Context()
	for _, uuid := range v.reports {
		_, err := v.vulnDao.DeleteForReport(ctx, uuid)
		v.NoError(err)
	}
	_, err := v.vulnDao.DeleteForScanner(ctx, vulnTestScannerUUID)
	v.NoError(err)
	_, err = dao.GetOrmer().Raw("DELETE FROM scan_report WHERE registration_uuid = ?", vulnTestScannerUUID).Exec()
	v.NoError(err)
	v.NoError(scanner.DeleteRegistration(ctx, vulnTestScannerUUID))
}

func (v *VulnerabilityCollectorTestSuite) createReport(uuid, digest, mimeType string, counts *scan.Report) {
	counts.UUID = uuid
	counts.Digest = digest
	counts.RegistrationUUID = vulnTestScannerUUID
	counts.MimeType = mimeType
	counts.Report = "{}"
	_, err := v.reportDao.Create(orm.Context(), counts)
	v.Require().NoError(err)
	v.reports = append(v.reports, uuid)
}

func (v *VulnerabilityCollectorTestSuite) createVulnRecord(cveID, pkg, version, severity string) int64 {
	id, err := v.vulnDao.Create(orm.Context(), &scan.VulnerabilityRecord{
		CVEID:            cveID,
		RegistrationUUID: vulnTestScannerUUID,
		Package:          pkg,
		PackageVersion:   version,
		PackageType:      "debian",
		Severity:         severity,
	})
	v.Require().NoError(err)
	return id
}

func (v *VulnerabilityCollectorTestSuite) TestGetVulnerabilityInfo() {
	overview := getVulnerabilityInfo(vulnTestScannerUUID)

	projects := make(map[int64]*projectVulnerabilityInfo)
	for _, p := range overview.Projects {
		projects[p.ProjectID] = p
	}
	v.Len(projects, 2, "only projects with scanned artifacts are expected")
	v.NotContains(projects, testPro3.ProjectID)

	p1, ok := projects[testPro1.ProjectID]
	if v.True(ok, "project test1 is missing") {
		v.Equal(testPro1.Name, p1.Name)
		v.Equal(float64(1), p1.ScannedTotal)
		v.Equal(map[string]float64{
			severityCritical: 2, severityHigh: 1, severityMedium: 0,
			severityLow: 3, severityUnknown: 0, severityNone: 0,
		}, p1.severityCounts())
		v.Equal(float64(2), p1.FixableCnt)
	}

	p2, ok := projects[testPro2.ProjectID]
	if v.True(ok, "project test2 is missing") {
		v.Equal(testPro2.Name, p2.Name)
		v.Equal(float64(1), p2.ScannedTotal, "an artifact with multiple reports must be counted once")
		v.Equal(map[string]float64{
			severityCritical: 0, severityHigh: 0, severityMedium: 4,
			severityLow: 0, severityUnknown: 1, severityNone: 2,
		}, p2.severityCounts(), "only the native report must be counted")
		v.Equal(float64(1), p2.FixableCnt)
	}

	cves := make(map[string]float64)
	for _, c := range overview.UniqueCVEs {
		cves[c.Severity] = c.CVETotal
	}
	v.Equal(map[string]float64{severityCritical: 2, severityMedium: 1}, cves)
}

func (v *VulnerabilityCollectorTestSuite) TestGetVulnerabilityInfoUnknownScanner() {
	overview := getVulnerabilityInfo("not-existing-scanner")
	v.Empty(overview.Projects)
	v.Empty(overview.UniqueCVEs)
}

func (v *VulnerabilityCollectorTestSuite) TestGetDefaultScannerUUID() {
	expected := ""
	err := dao.GetOrmer().Raw("SELECT uuid FROM scanner_registration WHERE is_default = TRUE AND disabled = FALSE LIMIT 1").QueryRow(&expected)
	if !errors.Is(err, orm.ErrNoRows) {
		v.Require().NoError(err)
	}

	uuid, err := getDefaultScannerUUID()
	v.NoError(err)
	v.Equal(expected, uuid)
	v.NotEqual(vulnTestScannerUUID, uuid, "the test scanner is not the default one")
}

func (v *VulnerabilityCollectorTestSuite) TestCollectorDescribe() {
	vc := NewVulnerabilityCollector()
	v.Equal(VulnerabilityCollectorName, vc.GetName())

	ch := make(chan *prometheus.Desc, 10)
	vc.Describe(ch)
	close(ch)
	v.Len(ch, 4)
}

func (v *VulnerabilityCollectorTestSuite) TestCollectorCollect() {
	// temporarily make the test scanner the default one and restore the original afterwards
	previous, err := getDefaultScannerUUID()
	v.Require().NoError(err)
	_, err = dao.GetOrmer().Raw("UPDATE scanner_registration SET is_default = (uuid = ?)", vulnTestScannerUUID).Exec()
	v.Require().NoError(err)
	defer func() {
		_, err := dao.GetOrmer().Raw("UPDATE scanner_registration SET is_default = (uuid = ?)", previous).Exec()
		v.NoError(err)
	}()

	vc := NewVulnerabilityCollector()
	ch := make(chan prometheus.Metric, 1000)
	v.NotPanics(func() {
		vc.Collect(ch)
	})
	close(ch)

	values := make(map[string]float64)
	for m := range ch {
		d := &dto.Metric{}
		v.Require().NoError(m.Write(d))
		key := m.Desc().String()
		for _, l := range d.GetLabel() {
			key += "|" + l.GetName() + "=" + l.GetValue()
		}
		values[key] = d.GetGauge().GetValue()
	}
	// 2 projects * (6 severities + fixable + scanned) + 2 unique CVE severities
	v.Len(values, 18)

	get := func(desc typedDesc, labels ...string) float64 {
		key := desc.Desc().String()
		for _, l := range labels {
			key += "|" + l
		}
		val, ok := values[key]
		v.Truef(ok, "metric %s not found", key)
		return val
	}
	v.Equal(float64(2), get(projectVulnerabilityTotal, "project_name="+testPro1.Name, "severity="+severityCritical))
	v.Equal(float64(4), get(projectVulnerabilityTotal, "project_name="+testPro2.Name, "severity="+severityMedium))
	v.Equal(float64(2), get(projectVulnerabilityFixableTotal, "project_name="+testPro1.Name))
	v.Equal(float64(1), get(projectScannedArtifactTotal, "project_name="+testPro2.Name))
	v.Equal(float64(2), get(vulnerabilityUniqueCVETotal, "severity="+severityCritical))
	v.Equal(float64(1), get(vulnerabilityUniqueCVETotal, "severity="+severityMedium))
}
