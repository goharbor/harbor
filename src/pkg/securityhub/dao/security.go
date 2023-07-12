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

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/securityhub/model"
)

const (
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

	dangerousArtifactSQL = `select a.project_id project, a.repository_name repository, a.digest, s.critical_cnt, s.high_cnt, s.medium_cnt, s.low_cnt
from artifact a,
     scan_report s
where a.digest = s.digest
  and s.registration_uuid = ?
order by s.critical_cnt desc, s.high_cnt desc, s.medium_cnt desc, s.low_cnt desc
limit 5`

	scannedArtifactCountSQL = `select count(1) 
           from artifact a 
      left join scan_report s on a.digest = s.digest 
          where s.registration_uuid= ? and s.uuid is not null`

	dangerousCVESQL = `select vr.*
from vulnerability_record vr
where vr.cvss_score_v3 is not null
and vr.registration_uuid = ?
order by vr.cvss_score_v3 desc
limit 5`
)

// SecurityHubDao defines the interface to access security hub data.
type SecurityHubDao interface {
	// Summary returns the summary of the scan cve reports.
	Summary(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (*model.Summary, error)
	// DangerousCVEs get the top 5 most dangerous CVEs, return top 5 result
	DangerousCVEs(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*scan.VulnerabilityRecord, error)
	// DangerousArtifacts returns top 5 dangerous artifact for the given scanner. return top 5 result
	DangerousArtifacts(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) ([]*model.DangerousArtifact, error)
	// ScannedArtifactsCount return the count of scanned artifacts.
	ScannedArtifactsCount(ctx context.Context, scannerUUID string, projectID int64, query *q.Query) (int64, error)
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
