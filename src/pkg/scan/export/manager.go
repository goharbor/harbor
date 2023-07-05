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

package export

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	beego_orm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/lib/orm"
	q2 "github.com/goharbor/harbor/src/lib/q"
)

const (
	// This sql template aims to select vuln data from database,
	// which receive one parameter:
	// 1. artifacts id sets
	// consider for performance, the caller will slice the artifact ids to multi
	// groups if it's length over limit, so rowNum offset is designed to ensure the
	// final row id is sequence in the final output csv file.
	VulnScanReportQueryTemplate = `
select
    artifact.digest as artifact_digest,
    artifact.repository_id,
    artifact.repository_name,
    vulnerability_record.cve_id,
    vulnerability_record.package,
    vulnerability_record.severity,
    vulnerability_record.cwe_ids,
    vulnerability_record.package_version,
    vulnerability_record.fixed_version,
    to_jsonb(vulnerability_record.vendor_attributes)  as vendor_attributes,
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
    vulnerability_record.cwe_ids,
    vulnerability_record.package_version,
    vulnerability_record.fixed_version,
    to_jsonb(vulnerability_record.vendor_attributes),
    scanner_registration.id
	`
	JobModeExport = "export"
	JobModeKey    = "mode"
	JobID         = "JobId"
	JobRequest    = "Request"
)

var (
	Mgr = NewManager()
)

// Params specifies the filters for controlling the scan data export process
type Params struct {
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

	sql := fmt.Sprintf(VulnScanReportQueryTemplate, artIDs)
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
