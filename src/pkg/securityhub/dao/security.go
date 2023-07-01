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

package dao

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/securityhub/model"
)

const (
	// sql to query the security summary
	summarySQL = `select sum(s.critical_cnt) critical_cnt,
       sum(s.high_cnt)     high_cnt,
       sum(s.medium_cnt)   medium_cnt,
       sum(s.low_cnt)      low_cnt,
       sum(s.none_cnt)     none_cnt,
       sum(s.unknown_cnt)  unknown_cnt,
       sum(s.fixable_cnt)  fixable_cnt
from artifact a
         left join scan_report s on a.digest = s.digest
         where s.registration_uuid = ?`
	// sql to query the dangerous artifact
	dangerousArtifactSQL = `select a.project_id project, a.repository_name repository, a.digest, s.critical_cnt, s.high_cnt, s.medium_cnt, s.low_cnt
from artifact a,
     scan_report s
where a.digest = s.digest
  and s.registration_uuid = ?
order by s.critical_cnt desc, s.high_cnt desc, s.medium_cnt desc, s.low_cnt desc
limit 5`

	// sql to query the scanned artifact count
	scannedArtifactCountSQL = `select count(1) 
           from artifact a 
      left join scan_report s on a.digest = s.digest 
          where s.registration_uuid= ? and s.uuid is not null`

	// sql to query the dangerous CVEs
	dangerousCVESQL = `select vr.*
from vulnerability_record vr
where vr.cvss_score_v3 is not null
and vr.registration_uuid = ?
order by vr.cvss_score_v3 desc
limit 5`

	// sql to query vulnerabilities
	vulnerabilitySQL = `select  vr.cve_id, vr.cvss_score_v3, vr.package, a.repository_name repository, a.id artifact_id, a.digest, vr.package, vr.package_version, vr.severity, vr.fixed_version, vr.description, vr.urls, a.project_id
from artifact a,
     scan_report s,
     report_vulnerability_record rvr,
     vulnerability_record vr
where a.digest = s.digest
  and s.uuid = rvr.report_uuid
  and rvr.vuln_record_id = vr.id
  and rvr.report_uuid is not null
  and vr.registration_uuid = ? `

	stringType = "string"
	intType    = "int"
	rangeType  = "range"
)

var allowedFieldTypeMap = map[string]string{
	"cve_id":        stringType,
	"severity":      stringType,
	"cvss_score_v3": rangeType,
	"project_id":    stringType,
	"repository":    stringType,
	"package":       stringType,
	"tag":           stringType,
}

var filterFuncMap = map[string]func(ctx context.Context, key string, query *q.Query) (sqlStr string, params []interface{}){
	"cve_id":        exactMatchFilter,
	"severity":      exactMatchFilter,
	"cvss_score_v3": rangeFilter,
	"project_id":    exactMatchFilter,
	"repository":    repositoryFilter,
	"package":       exactMatchFilter,
	"tag":           tagFilter,
}

var applyFilterFunc func(ctx context.Context, key string, query *q.Query) (sqlStr string, params []interface{})

func exactMatchFilter(ctx context.Context, key string, query *q.Query) (sqlStr string, params []interface{}) {
	if query == nil {
		return
	}
	if val, ok := query.Keywords[key]; ok {
		sqlStr = fmt.Sprintf(" and %v = ?", key)
		params = append(params, val)
		return
	}
	return
}

func rangeFilter(ctx context.Context, key string, query *q.Query) (sqlStr string, params []interface{}) {
	if query == nil {
		return
	}
	if val, ok := query.Keywords[key]; ok {
		if r, ok := val.(q.Range); ok {
			sqlStr = fmt.Sprintf(" and %v between ? and ?", key)
			params = append(params, r.Min, r.Max)
		}
	}
	return
}

func repositoryFilter(ctx context.Context, key string, query *q.Query) (sqlStr string, params []interface{}) {
	if query == nil {
		return
	}
	if val, ok := query.Keywords["repository"]; ok {
		sqlStr = " and a.repository_name = ?"
		params = append(params, val)
	}
	return
}

func tagFilter(ctx context.Context, key string, query *q.Query) (sqlStr string, params []interface{}) {
	if query == nil {
		return
	}
	if val, ok := query.Keywords["tag"]; ok {
		inClause, err := orm.CreateInClause(ctx, `SELECT artifact_id FROM tag 
			WHERE tag.name = ?`, val)
		if err != nil {
			log.Errorf("failed to create in clause: %v, skip this condition", err)
		} else {
			sqlStr = " and a.id " + inClause
		}
	}
	return
}

// SecurityHubDao defines the interface to access security hub data.
type SecurityHubDao interface {
	// Summary returns the summary of the scan cve reports.
	Summary(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (*model.Summary, error)
	// DangerousCVEs get the top 5 most dangerous CVEs
	DangerousCVEs(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*scan.VulnerabilityRecord, error)
	// DangerousArtifacts returns top 5 dangerous artifact for the given scanner.
	DangerousArtifacts(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*model.DangerousArtifact, error)
	// ScannedArtifactsCount return the count of scanned artifacts.
	ScannedArtifactsCount(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (int64, error)
	// ListVulnerabilities search vulnerability record by cveID
	ListVulnerabilities(ctx context.Context, registrationUUID string, projectID int64, query *q.Query) ([]*model.VulnerabilityItem, error)
	// CountVulnerabilities count the total vulnerabilities
	CountVulnerabilities(ctx context.Context, registrationUUID string, projectID int64, tuneCount bool, query *q.Query) (int64, error)
}

// New creates a new SecurityHubDao instance.
func New() SecurityHubDao {
	return &dao{}
}

type dao struct {
}

func (d *dao) Summary(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (*model.Summary, error) {
	if len(scannerUUID) == 0 || projectID != 0 {
		return nil, nil
	}
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	var sum model.Summary
	err = o.Raw(summarySQL, scannerUUID).QueryRow(&sum.CriticalCnt,
		&sum.HighCnt,
		&sum.MediumCnt,
		&sum.LowCnt,
		&sum.NoneCnt,
		&sum.UnknownCnt,
		&sum.FixableCnt)
	return &sum, err
}
func (d *dao) DangerousArtifacts(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*model.DangerousArtifact, error) {
	if len(scannerUUID) == 0 || projectID != 0 {
		return nil, nil
	}
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	var artifacts []*model.DangerousArtifact
	_, err = o.Raw(dangerousArtifactSQL, scannerUUID).QueryRows(&artifacts)
	return artifacts, err
}

func (d *dao) ScannedArtifactsCount(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (int64, error) {
	if len(scannerUUID) == 0 || projectID != 0 {
		return 0, nil
	}
	var cnt int64
	o, err := orm.FromContext(ctx)
	if err != nil {
		return cnt, err
	}
	err = o.Raw(scannedArtifactCountSQL, scannerUUID).QueryRow(&cnt)
	return cnt, err
}
func (d *dao) DangerousCVEs(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*scan.VulnerabilityRecord, error) {
	if len(scannerUUID) == 0 || projectID != 0 {
		return nil, nil
	}
	cves := make([]*scan.VulnerabilityRecord, 0)
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	_, err = o.Raw(dangerousCVESQL, scannerUUID).QueryRows(&cves)
	return cves, err
}

func countSQL(strSQL string) string {
	return fmt.Sprintf(`select count(1) cnt from (%v) as t`, strSQL)
}

func (d *dao) CountVulnerabilities(ctx context.Context, registrationUUID string, projectID int64, tuneCount bool, query *q.Query) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	sqlStr := vulnerabilitySQL
	params := []interface{}{registrationUUID}
	if err := checkQFilter(query, allowedFieldTypeMap); err != nil {
		return 0, err
	}
	sqlStr, params = applyVulFilter(ctx, sqlStr, query, params)
	if tuneCount {
		exceedLimit, err := d.countExceedLimit(ctx, sqlStr, params)
		if err != nil {
			return 0, err
		}
		log.Infof("exceedLimit: %v", exceedLimit)
		if exceedLimit {
			return -1, nil
		}
	}
	var cnt int64
	err = o.Raw(countSQL(sqlStr), params).QueryRow(&cnt)
	return cnt, err
}

func (d *dao) countExceedLimit(ctx context.Context, sqlStr string, params []interface{}) (bool, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return true, err
	}
	queryExceed := fmt.Sprintf(`SELECT EXISTS (%v LIMIT 1 OFFSET 1000)`, sqlStr)
	var exceed bool
	err = o.Raw(queryExceed, params).QueryRow(&exceed)
	if err != nil {
		return true, err
	}
	return exceed, nil
}

func (d *dao) ListVulnerabilities(ctx context.Context, registrationUUID string, projectID int64, query *q.Query) ([]*model.VulnerabilityItem, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	sqlStr := vulnerabilitySQL
	params := []interface{}{registrationUUID}
	if err := checkQFilter(query, allowedFieldTypeMap); err != nil {
		return nil, err
	}
	sqlStr, params = applyVulFilter(ctx, sqlStr, query, params)
	sqlStr = applyVulPagination(sqlStr, query)
	vulnRecs := make([]*model.VulnerabilityItem, 0)
	_, err = o.Raw(sqlStr, params).QueryRows(&vulnRecs)
	return vulnRecs, err
}

func applyVulFilter(ctx context.Context, sqlStr string, query *q.Query, params []interface{}) (queryStr string, newParam []interface{}) {
	queryStr = sqlStr
	newParam = params
	for k, f := range filterFuncMap {
		s, p := f(ctx, k, query)
		queryStr = queryStr + s
		newParam = append(newParam, p...)
	}
	return queryStr, newParam
}

func applyVulPagination(sqlStr string, query *q.Query) string {
	offSet := int64(0)
	pageSize := int64(15)
	if query != nil && query.PageNumber > 1 {
		offSet = (query.PageNumber - 1) * query.PageSize
	}
	if query != nil && query.PageSize > 0 {
		pageSize = query.PageSize
	}

	return fmt.Sprintf("%v order by repository asc limit %d offset %d", sqlStr, pageSize, offSet)
}

func checkQFilter(query *q.Query, expectedTypeMap map[string]string) error {
	if query == nil {
		return nil
	}
	if len(query.Keywords) == 0 {
		return nil
	}
	for k := range query.Keywords {
		if typeName, exist := expectedTypeMap[k]; exist {
			switch typeName {
			case rangeType:
				if _, ok := query.Keywords[k].(*q.Range); !ok {
					return errors.BadRequestError(fmt.Errorf("keyword: %v, the query type is not allowed", k))
				}
			case stringType:
				if _, ok := query.Keywords[k].(string); !ok {
					return errors.BadRequestError(fmt.Errorf("keyword: %v, the query type is not allowed", k))
				}
			case intType:
				if _, ok := query.Keywords[k].(int); !ok {
					return errors.BadRequestError(fmt.Errorf("keyword: %v, the query type is not allowed", k))
				}
			}
		} else {
			return errors.BadRequestError(fmt.Errorf("keyword: %v is not allowed", k))
		}
	}
	return nil
}
