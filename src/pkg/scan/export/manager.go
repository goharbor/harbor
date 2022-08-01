package export

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	beego_orm "github.com/beego/beego/orm"

	"github.com/goharbor/harbor/src/lib/orm"
	q2 "github.com/goharbor/harbor/src/lib/q"
)

const (
	// This sql template aims to select vuln data from database,
	// which receive two parameters:
	// 1. rowNum offset
	// 2. artifacts id sets
	// consider for performance, the caller will slice the artifact ids to multi
	// groups if it's length over limit, so rowNum offset is designed to ensure the
	// final row id is sequence in the final output csv file.
	VulnScanReportQueryTemplate = `
select
    row_number() over() + %d as result_row_id,
    artifact.digest as artifact_digest,
    artifact.repository_id,
    artifact.repository_name,
    vulnerability_record.cve_id,
    vulnerability_record.package,
    vulnerability_record.severity,
    vulnerability_record.cvss_score_v3,
    vulnerability_record.cvss_score_v2,
    vulnerability_record.cvss_vector_v3,
    vulnerability_record.cvss_vector_v2,
    vulnerability_record.cwe_ids,
    vulnerability_record.package_version,
    vulnerability_record.fixed_version,
    scanner_registration."name" as scanner_name
from
    report_vulnerability_record
    inner join scan_report on report_vulnerability_record.report_uuid = scan_report.uuid
    inner join artifact on scan_report.digest = artifact.digest
    left outer join artifact_reference on artifact.id = artifact_reference.child_id
    inner join vulnerability_record on report_vulnerability_record.vuln_record_id = vulnerability_record.id
    inner join scanner_registration on scan_report.registration_uuid = scanner_registration.uuid
and artifact.id in (%s)

group by
    package,
    vulnerability_record.severity,
    vulnerability_record.cve_id,
    artifact.digest,
    artifact.repository_id,
    artifact.repository_name,
    vulnerability_record.cvss_score_v3,
    vulnerability_record.cvss_score_v2,
    vulnerability_record.cvss_vector_v3,
    vulnerability_record.cvss_vector_v2,
    vulnerability_record.cwe_ids,
    vulnerability_record.package_version,
    vulnerability_record.fixed_version,
    scanner_registration.id
	`
	JobModeExport = "export"
	JobModeKey    = "mode"
)

var (
	Mgr = NewManager()
)

// Params specifies the filters for controlling the scan data export process
type Params struct {
	// rowNumber offset
	RowNumOffset int64

	// cve ids
	CVEIds string

	// artifact ids
	ArtifactIDs []int64

	// PageNumber
	PageNumber int64

	// PageSize
	PageSize int64
}

// FromJSON parses robot from json data
func (p *Params) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), p)
}

// ToJSON marshals Robot to JSON data
func (p *Params) ToJSON() (string, error) {
	data, err := json.Marshal(p)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type Manager interface {
	Fetch(ctx context.Context, params Params) ([]Data, error)
}

type exportManager struct {
	exportDataFilter VulnerabilityDataSelector
}

func NewManager() Manager {
	return &exportManager{exportDataFilter: NewVulnerabilityDataSelector()}
}

func (em *exportManager) Fetch(ctx context.Context, params Params) ([]Data, error) {
	exportData := make([]Data, 0)
	rawSeter, _ := em.buildQuery(ctx, params)
	_, err := rawSeter.QueryRows(&exportData)
	if err != nil {
		return nil, err
	}

	exportData, err = em.exportDataFilter.Select(exportData, CVEIDMatches, params.CVEIds)
	if err != nil {
		return nil, err
	}

	return exportData, nil
}

func (em *exportManager) buildQuery(ctx context.Context, params Params) (beego_orm.RawSeter, error) {
	artIDs := ""
	for _, artID := range params.ArtifactIDs {
		if len(artIDs) == 0 {
			artIDs += fmt.Sprintf("%d", artID)
		} else {
			artIDs += fmt.Sprintf(",%d", artID)
		}
	}

	sql := fmt.Sprintf(VulnScanReportQueryTemplate, params.RowNumOffset, artIDs)
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	pageSize := params.PageSize
	q := &q2.Query{
		Keywords:   nil,
		Sorts:      nil,
		PageNumber: params.PageNumber,
		PageSize:   pageSize,
		Sorting:    "",
	}
	paginationParams := make([]interface{}, 0)
	query, pageLimits := orm.PaginationOnRawSQL(q, sql, paginationParams)
	// user can open ORM_DEBUG for log the sql
	return ormer.Raw(query, pageLimits), nil
}
