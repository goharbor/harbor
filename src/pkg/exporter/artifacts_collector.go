package exporter

import (
	"errors"
	"strings"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/prometheus/client_golang/prometheus"
)

const ArtifactCollectorName = "ArtifactCollector"

const (
	severityLow    = "Low"
	severityMedium = "Medium"
	severityHigh   = "High"
)

var (
	// artifactReportSQL returns last scan reports per artifact.
	artifactReportSQL = `SELECT
		DISTINCT ON (artifact.id)
			scan_report.uuid AS report_uuid,
			project.name AS project_id,
			artifact.digest AS artifact_digest,
			artifact.size AS artifact_size,
			artifact.repository_name AS repository_id
		FROM scan_report 
		JOIN artifact ON artifact.digest = scan_report.digest
		JOIN project ON artifact.project_id = project.project_id
		
		ORDER BY artifact.id, scan_report.id DESC`

	// artifactVulnReportSQLPrefix returns vulnerabilities for artifacts.
	artifactVulnReportSQLPrefix = `SELECT
			report_vulnerability_record.report_uuid,
			vulnerability_record.severity,
			vulnerability_record.fixed_version
		FROM report_vulnerability_record
		JOIN vulnerability_record ON report_vulnerability_record.vuln_record_id = vulnerability_record.id
		WHERE report_vulnerability_record.report_uuid IN (`
)

var (
	artifactVulnerabilities = typedDesc{
		desc:      newDescWithLables("", "project_artifact_detected_vulnerabilities", "Discovered vulnerabilities count of an image per Level", "project_id", "repository_id", "artifact_id", "vulnerability_level"),
		valueType: prometheus.GaugeValue,
	}
	artifactFixableVulnerabilities = typedDesc{
		desc:      newDescWithLables("", "project_artifact_fixable_vulnerabilities", "Discovered fixable vulnerabilities count of an image per Level", "project_id", "repository_id", "artifact_id", "vulnerability_level"),
		valueType: prometheus.GaugeValue,
	}
	artifactSize = typedDesc{
		desc:      newDescWithLables("", "project_artifact_size", "Size of artifact in bytes", "project_id", "repository_id", "artifact_id"),
		valueType: prometheus.GaugeValue,
	}
)

var (
	errUnknownSeverity = errors.New("unknown severity")
)

func NewArtifactCollector() *ArtifactCollector {
	return &ArtifactCollector{}
}

type ArtifactCollector struct{}

// Describe implements prometheus.Collector
func (hc *ArtifactCollector) Describe(c chan<- *prometheus.Desc) {
	c <- artifactFixableVulnerabilities.Desc()
	c <- artifactVulnerabilities.Desc()
}

// Collect implements prometheus.Collector
func (hc *ArtifactCollector) Collect(c chan<- prometheus.Metric) {
	reportRecords, err := getVulnerabilitiesStats()
	if err != nil {
		return
	}

	// Make Metrics.
	for ri := range reportRecords {
		pr := &reportRecords[ri]

		// Vulnerabilities.
		// If Vulnerabilities == nil then there is no reports for given Artifact.
		if pr.Vulnerabilities != nil {
			c <- artifactVulnerabilities.MustNewConstMetric(float64(pr.Vulnerabilities.High), pr.ProjectName, pr.RepositoryName, pr.ArtifactDigest, severityHigh)
			c <- artifactVulnerabilities.MustNewConstMetric(float64(pr.Vulnerabilities.Medium), pr.ProjectName, pr.RepositoryName, pr.ArtifactDigest, severityMedium)
			c <- artifactVulnerabilities.MustNewConstMetric(float64(pr.Vulnerabilities.Low), pr.ProjectName, pr.RepositoryName, pr.ArtifactDigest, severityLow)

			// Fixable vulnerabilities.
			c <- artifactFixableVulnerabilities.MustNewConstMetric(float64(pr.Vulnerabilities.FixableHigh), pr.ProjectName, pr.RepositoryName, pr.ArtifactDigest, severityHigh)
			c <- artifactFixableVulnerabilities.MustNewConstMetric(float64(pr.Vulnerabilities.FixableMedium), pr.ProjectName, pr.RepositoryName, pr.ArtifactDigest, severityMedium)
			c <- artifactFixableVulnerabilities.MustNewConstMetric(float64(pr.Vulnerabilities.FixableLow), pr.ProjectName, pr.RepositoryName, pr.ArtifactDigest, severityLow)
		}
		// Artifact size.
		c <- artifactSize.MustNewConstMetric(float64(pr.ArtifactSize), pr.ProjectName, pr.RepositoryName, pr.ArtifactDigest)
	}
}

// artifactVulnStats statistics for one artifact.
type artifactVulnStats struct {
	Low    int64
	Medium int64
	High   int64

	FixableLow    int64
	FixableMedium int64
	FixableHigh   int64
}

// artifactReportRecord contains info about one artifact.
type artifactReportRecord struct {
	ReportUUID     string `orm:"column(report_uuid)"`
	ProjectName    string `orm:"column(project_id)"`
	ArtifactDigest string `orm:"column(artifact_digest)"`
	ArtifactSize   int64  `orm:"column(artifact_size)"`
	RepositoryName string `orm:"column(repository_id)"`

	Vulnerabilities *artifactVulnStats `orm:"column(-)"`
}

type artifactReportRecords []artifactReportRecord

type vulnerabilityRecord struct {
	ReportUUID   string `orm:"column(report_uuid)"`
	Severity     string `orm:"column(severity)"`
	FixedVersion string `orm:"column(fixed_version)"`
}

func getVulnerabilitiesInfo() (_ artifactReportRecords, _ []vulnerabilityRecord, err error) {
	// Start Transaction.
	dbORM := dao.GetOrmer()

	if err = dbORM.Begin(); err != nil {
		checkErr(err, "can not open DB transaction")

		return nil, nil, err
	}
	defer func() {
		if err == nil {
			err = dbORM.Commit()
		} else {
			_ = dbORM.Rollback()
		}
	}()

	// Query Scan Reports.
	var reportRecords artifactReportRecords

	if _, err = dbORM.Raw(artifactReportSQL).QueryRows(&reportRecords); err != nil {
		checkErr(err, "can not load Scan Reports")

		return nil, nil, err
	}

	// Query Vulnerabilities for the reports.
	var (
		// reportUUIDs stores UUIDs as interface{} to use in ORM WHERE IN.
		reportUUIDs = make([]interface{}, len(reportRecords))
	)
	for i := range reportRecords {
		reportUUIDs[i] = (reportRecords[i].ReportUUID)
	}

	// Build query.
	var qs = &strings.Builder{}

	_, _ = qs.WriteString(artifactVulnReportSQLPrefix)
	_, _ = qs.WriteString(strings.TrimRight(strings.Repeat("?,", len(reportUUIDs)), ","))
	_, _ = qs.WriteString(")")

	var vulns []vulnerabilityRecord

	if _, err = dbORM.Raw(qs.String(), reportUUIDs...).QueryRows(&vulns); err != nil {
		checkErr(err, "can not query vulnerabilities")

		return nil, nil, err
	}

	return reportRecords, vulns, nil
}

func computeVulnStats(artifacts artifactReportRecords, vulns []vulnerabilityRecord) error {
	var stats = make(map[string]*artifactVulnStats)

	for i := range vulns {
		vp := &vulns[i]

		stRec, ok := stats[vp.ReportUUID]
		if !ok {
			stRec = &artifactVulnStats{}

			stats[vp.ReportUUID] = stRec
		}

		switch vp.Severity {
		case severityLow:
			stRec.Low++

			if vp.FixedVersion != "" {
				stRec.FixableLow++
			}
		case severityMedium:
			stRec.Medium++

			if vp.FixedVersion != "" {
				stRec.FixableMedium++
			}
		case severityHigh:
			stRec.High++

			if vp.FixedVersion != "" {
				stRec.FixableHigh++
			}
		default:
			checkErr(errUnknownSeverity, "Severity: "+vp.Severity)

			return errUnknownSeverity
		}
	}

	for i := range artifacts {
		st, ok := stats[artifacts[i].ReportUUID]
		if ok {
			artifacts[i].Vulnerabilities = st
		}
	}

	return nil
}

func getVulnerabilitiesStats() (artifactReportRecords, error) {
	reportRecords, vulns, err := getVulnerabilitiesInfo()
	if err != nil {
		return nil, err
	}

	if err := computeVulnStats(reportRecords, vulns); err != nil {
		return nil, err
	}

	return reportRecords, nil
}
