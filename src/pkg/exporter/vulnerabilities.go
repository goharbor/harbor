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

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// VulnerabilityCollectorName is the name of the vulnerability collector
const VulnerabilityCollectorName = "VulnerabilityCollector"

// Severity label values exposed by the vulnerability metrics.
const (
	severityCritical = "critical"
	severityHigh     = "high"
	severityMedium   = "medium"
	severityLow      = "low"
	severityUnknown  = "unknown"
	severityNone     = "none"
)

var (
	// defaultScannerSQL resolves the UUID of the enabled default scanner.
	defaultScannerSQL = `SELECT uuid FROM scanner_registration
	WHERE is_default=TRUE AND disabled=FALSE
	LIMIT 1;`

	// projectVulnerabilitySQL aggregates the severity counters that are stored on the
	// scan reports of the given scanner per project. When an artifact has a report for
	// both supported vulnerability mime types, only one of them is taken into account
	// (the native report is preferred) so the counters are not doubled.
	projectVulnerabilitySQL = `WITH report AS (
		SELECT DISTINCT ON (s.digest) s.digest,
			COALESCE(s.critical_cnt, 0) AS critical_cnt,
			COALESCE(s.high_cnt, 0) AS high_cnt,
			COALESCE(s.medium_cnt, 0) AS medium_cnt,
			COALESCE(s.low_cnt, 0) AS low_cnt,
			COALESCE(s.unknown_cnt, 0) AS unknown_cnt,
			COALESCE(s.none_cnt, 0) AS none_cnt,
			COALESCE(s.fixable_cnt, 0) AS fixable_cnt
		FROM scan_report s
		WHERE s.registration_uuid = ? AND s.mime_type IN (?, ?)
		ORDER BY s.digest, CASE WHEN s.mime_type = ? THEN 0 ELSE 1 END
	)
	SELECT project.project_id, project.name,
		COUNT(artifact.id) AS scanned_total,
		SUM(report.critical_cnt) AS critical_cnt,
		SUM(report.high_cnt) AS high_cnt,
		SUM(report.medium_cnt) AS medium_cnt,
		SUM(report.low_cnt) AS low_cnt,
		SUM(report.unknown_cnt) AS unknown_cnt,
		SUM(report.none_cnt) AS none_cnt,
		SUM(report.fixable_cnt) AS fixable_cnt
	FROM project
	INNER JOIN artifact ON project.project_id = artifact.project_id
	INNER JOIN report ON artifact.digest = report.digest
	WHERE project.deleted = FALSE
	GROUP BY project.project_id, project.name;`
)

var (
	projectVulnerabilityTotal = typedDesc{
		desc:      newDescWithLables("", "project_vulnerability_total", "Total number of vulnerabilities found in the scanned artifacts of a project", "project_name", "severity"),
		valueType: prometheus.GaugeValue,
	}
	projectVulnerabilityFixableTotal = typedDesc{
		desc:      newDescWithLables("", "project_vulnerability_fixable_total", "Total number of fixable vulnerabilities found in the scanned artifacts of a project", "project_name"),
		valueType: prometheus.GaugeValue,
	}
	projectScannedArtifactTotal = typedDesc{
		desc:      newDescWithLables("", "project_scanned_artifact_total", "Total number of artifacts of a project that have a vulnerability report", "project_name"),
		valueType: prometheus.GaugeValue,
	}
)

// NewVulnerabilityCollector creates a collector for the vulnerability metrics
func NewVulnerabilityCollector() *VulnerabilityCollector {
	return &VulnerabilityCollector{}
}

// VulnerabilityCollector collects vulnerability metrics directly from the database
type VulnerabilityCollector struct{}

// GetName returns the name of the vulnerability collector
func (vc *VulnerabilityCollector) GetName() string {
	return VulnerabilityCollectorName
}

// Describe implements prometheus.Collector
func (vc *VulnerabilityCollector) Describe(c chan<- *prometheus.Desc) {
	c <- projectVulnerabilityTotal.Desc()
	c <- projectVulnerabilityFixableTotal.Desc()
	c <- projectScannedArtifactTotal.Desc()
}

// Collect implements prometheus.Collector
func (vc *VulnerabilityCollector) Collect(c chan<- prometheus.Metric) {
	for _, m := range vc.getVulnerabilityMetrics() {
		c <- m
	}
}

type projectVulnerabilityInfo struct {
	ProjectID    int64   `orm:"column(project_id)"`
	Name         string  `orm:"column(name)"`
	ScannedTotal float64 `orm:"column(scanned_total)"`
	CriticalCnt  float64 `orm:"column(critical_cnt)"`
	HighCnt      float64 `orm:"column(high_cnt)"`
	MediumCnt    float64 `orm:"column(medium_cnt)"`
	LowCnt       float64 `orm:"column(low_cnt)"`
	UnknownCnt   float64 `orm:"column(unknown_cnt)"`
	NoneCnt      float64 `orm:"column(none_cnt)"`
	FixableCnt   float64 `orm:"column(fixable_cnt)"`
}

// severityCounts returns the vulnerability counters of the project keyed by severity label
func (p *projectVulnerabilityInfo) severityCounts() map[string]float64 {
	return map[string]float64{
		severityCritical: p.CriticalCnt,
		severityHigh:     p.HighCnt,
		severityMedium:   p.MediumCnt,
		severityLow:      p.LowCnt,
		severityUnknown:  p.UnknownCnt,
		severityNone:     p.NoneCnt,
	}
}

type vulnerabilityOverview struct {
	Projects []*projectVulnerabilityInfo
}

func (vc *VulnerabilityCollector) getVulnerabilityMetrics() []prometheus.Metric {
	if CacheEnabled() {
		value, ok := CacheGet(VulnerabilityCollectorName)
		if ok {
			return value.([]prometheus.Metric)
		}
	}

	var result []prometheus.Metric
	scannerUUID, err := getDefaultScannerUUID()
	if err != nil {
		checkErr(err, "get default scanner from DB failure")
		return result
	}
	if scannerUUID == "" {
		log.Debug("no enabled default scanner found, skip collecting vulnerability metrics")
		return result
	}

	overview := getVulnerabilityInfo(scannerUUID)
	for _, p := range overview.Projects {
		for severity, count := range p.severityCounts() {
			result = append(result, projectVulnerabilityTotal.MustNewConstMetric(count, p.Name, severity))
		}
		result = append(result,
			projectVulnerabilityFixableTotal.MustNewConstMetric(p.FixableCnt, p.Name),
			projectScannedArtifactTotal.MustNewConstMetric(p.ScannedTotal, p.Name),
		)
	}

	if CacheEnabled() {
		CachePut(VulnerabilityCollectorName, result)
	}
	return result
}

// getDefaultScannerUUID returns the UUID of the enabled default scanner, or an empty
// string when no such scanner is configured
func getDefaultScannerUUID() (string, error) {
	var uuid string
	err := dao.GetOrmer().Raw(defaultScannerSQL).QueryRow(&uuid)
	if errors.Is(err, orm.ErrNoRows) {
		return "", nil
	}
	return uuid, err
}

// getVulnerabilityInfo reads the vulnerability data of the given scanner from the database
func getVulnerabilityInfo(scannerUUID string) *vulnerabilityOverview {
	overview := &vulnerabilityOverview{
		Projects: make([]*projectVulnerabilityInfo, 0),
	}

	_, err := dao.GetOrmer().Raw(projectVulnerabilitySQL,
		scannerUUID, v1.MimeTypeNativeReport, v1.MimeTypeGenericVulnerabilityReport, v1.MimeTypeNativeReport,
	).QueryRows(&overview.Projects)
	checkErr(err, "get project vulnerability data from DB failure")

	return overview
}
