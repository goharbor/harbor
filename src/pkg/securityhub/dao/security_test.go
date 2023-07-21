//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package dao

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	testDao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	htesting "github.com/goharbor/harbor/src/testing"
)

func TestDao(t *testing.T) {
	suite.Run(t, &SecurityDaoTestSuite{})
}

type SecurityDaoTestSuite struct {
	htesting.Suite
	dao SecurityHubDao
}

// SetupSuite prepares env for test suite.
func (suite *SecurityDaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = New()
}

// SetupTest prepares env for test case.
func (suite *SecurityDaoTestSuite) SetupTest() {
	testDao.ExecuteBatchSQL([]string{
		`insert into scan_report(uuid, digest, registration_uuid, mime_type, critical_cnt, high_cnt, medium_cnt, low_cnt, unknown_cnt, fixable_cnt) values('uuid', 'digest1001', 'ruuid', 'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0', 50, 50, 50, 0, 0, 20)`,
		`insert into artifact (project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values  (1, 'library/hello-world', 'digest1001', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.docker.distribution.manifest.v2+json', 4452, '{"architecture":"amd64","author":"","config":{"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["/hello"]},"created":"2023-05-04T17:37:03.872958712Z","os":"linux"}', null, '');`,
		`insert into scanner_registration (name, url, uuid, auth) values('trivy', 'https://www.vmware.com', 'ruuid', 'empty')`,
		`insert into vulnerability_record (id, cve_id, registration_uuid, cvss_score_v3) values (1, '2023-4567-12345', 'ruuid', 9.8)`,
		`insert into report_vulnerability_record (report_uuid, vuln_record_id) VALUES ('uuid', 1)`,
	})

	testDao.ExecuteBatchSQL([]string{
		`INSERT INTO scanner_registration (name, url, uuid, auth) values('trivy2', 'https://www.trivy.com', 'uuid2', 'empty')`,
		`INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2021-44228', 'uuid2', 10, 'org.apache.logging.log4j:log4j-core');
		INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2021-21345', 'uuid2', 9.9, 'com.thoughtworks.xstream:xstream');
		INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2016-1585', 'uuid2', 9.8, 'libapparmor1');
		INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2023-0950', 'uuid2', 9.8, 'ure');
		INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2022-47629', 'uuid2', 9.8, 'libksba8');
		`,
	})
}

func (suite *SecurityDaoTestSuite) TearDownTest() {
	testDao.ExecuteBatchSQL([]string{
		`delete from scan_report where uuid = 'uuid'`,
		`delete from artifact where digest = 'digest1001'`,
		`delete from scanner_registration where uuid='ruuid'`,
		`delete from scanner_registration where uuid='uuid2'`,
		`delete from vulnerability_record where cve_id='2023-4567-12345'`,
		`delete from report_vulnerability_record where report_uuid='ruuid'`,
		`delete from vulnerability_record where registration_uuid ='uuid2'`,
	})
}

func (suite *SecurityDaoTestSuite) TestGetSummary() {
	s, err := suite.dao.Summary(suite.Context(), "ruuid", 0, nil)
	suite.Require().NoError(err)
	suite.Equal(int64(50), s.CriticalCnt)
	suite.Equal(int64(50), s.HighCnt)
	suite.Equal(int64(50), s.MediumCnt)
	suite.Equal(int64(20), s.FixableCnt)
}
func (suite *SecurityDaoTestSuite) TestGetMostDangerousArtifact() {
	aList, err := suite.dao.DangerousArtifacts(orm.Context(), "ruuid", 0, nil)
	suite.Require().NoError(err)
	suite.Equal(1, len(aList))
	suite.Equal(int64(50), aList[0].CriticalCnt)
	suite.Equal(int64(50), aList[0].HighCnt)
	suite.Equal(int64(50), aList[0].MediumCnt)
	suite.Equal(int64(0), aList[0].LowCnt)
}

func (suite *SecurityDaoTestSuite) TestGetScannedArtifactCount() {
	count, err := suite.dao.ScannedArtifactsCount(orm.Context(), "ruuid", 0, nil)
	suite.Require().NoError(err)
	suite.Equal(int64(1), count)
}

func (suite *SecurityDaoTestSuite) TestGetDangerousCVEs() {
	records, err := suite.dao.DangerousCVEs(suite.Context(), `uuid2`, 0, nil)
	suite.NoError(err, "Error when fetching most dangerous artifact")
	suite.Equal(5, len(records))
}

func Test_checkQFilter(t *testing.T) {
	type args struct {
		query     *q.Query
		filterMap map[string]*filterMetaData
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"happy_path", args{q.New(q.KeyWords{"sample": 1}), map[string]*filterMetaData{"sample": &filterMetaData{intType, exactMatchFilter}}}, false},
		{"happy_path_cve_id", args{q.New(q.KeyWords{"cve_id": "CVE-2023-2345"}), map[string]*filterMetaData{"cve_id": &filterMetaData{stringType, exactMatchFilter}}}, false},
		{"happy_path_severity", args{q.New(q.KeyWords{"severity": "Critical"}), map[string]*filterMetaData{"severity": &filterMetaData{stringType, exactMatchFilter}}}, false},
		{"happy_path_cvss_score_v3", args{q.New(q.KeyWords{"cvss_score_v3": &q.Range{Min: 2.0, Max: 3.0}}), map[string]*filterMetaData{"cvss_score_v3": &filterMetaData{rangeType, rangeFilter}}}, false},
		{"unhappy_path", args{q.New(q.KeyWords{"sample": 1}), map[string]*filterMetaData{"a": &filterMetaData{DataType: intType}}}, true},
		{"unhappy_path2", args{q.New(q.KeyWords{"cve_id": 1}), map[string]*filterMetaData{"cve_id": &filterMetaData{stringType, exactMatchFilter}}}, true},
		{"unhappy_path3", args{q.New(q.KeyWords{"severity": &q.Range{Min: 2.0, Max: 10.0}}), map[string]*filterMetaData{"severity": &filterMetaData{stringType, exactMatchFilter}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkQFilter(tt.args.query, tt.args.filterMap); (err != nil) != tt.wantErr {
				t.Errorf("checkQFilter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func (suite *SecurityDaoTestSuite) TestExacthMatchFilter() {
	type args struct {
		ctx   context.Context
		key   string
		query *q.Query
	}
	tests := []struct {
		name       string
		args       args
		wantSQLStr string
		wantParams []interface{}
	}{
		{"normal", args{suite.Context(), "cve_id", q.New(q.KeyWords{"cve_id": "CVE-2023-2345"})}, " and cve_id = ?", []interface{}{"CVE-2023-2345"}},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			gotSQLStr, gotParams := exactMatchFilter(tt.args.ctx, tt.args.key, tt.args.query)
			suite.Equal(gotSQLStr, tt.wantSQLStr, "exactMatchFilter() gotSqlStr = %v, want %v", gotSQLStr, tt.wantSQLStr)
			suite.Equal(gotParams, tt.wantParams, "exactMatchFilter() gotParams = %v, want %v", gotParams, tt.wantParams)
		})
	}
}

func (suite *SecurityDaoTestSuite) TestRangeFilter() {
	type args struct {
		ctx   context.Context
		key   string
		query *q.Query
	}
	tests := []struct {
		name       string
		args       args
		wantSQLStr string
		wantParams []interface{}
	}{
		{"normal", args{suite.Context(), "cvss_score_v3", q.New(q.KeyWords{"cvss_score_v3": &q.Range{1.0, 2.0}})}, " and cvss_score_v3 between ? and ?", []interface{}{1.0, 2.0}},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			gotSQLStr, gotParams := rangeFilter(tt.args.ctx, tt.args.key, tt.args.query)
			suite.Equal(tt.wantSQLStr, gotSQLStr, "exactMatchFilter() gotSqlStr = %v, want %v", gotSQLStr, tt.wantSQLStr)
			suite.Equal(tt.wantParams, gotParams, "exactMatchFilter() gotParams = %v, want %v", gotParams, tt.wantParams)
		})
	}
}

func (suite *SecurityDaoTestSuite) TestCountVul() {
	count, err := suite.dao.CountVulnerabilities(suite.Context(), "ruuid", 0, true, nil)
	suite.NoError(err)
	suite.Equal(int64(1), count)
}

func (suite *SecurityDaoTestSuite) TestListVul() {
	vuls, err := suite.dao.ListVulnerabilities(suite.Context(), "ruuid", 0, nil)
	suite.NoError(err)
	suite.Equal(1, len(vuls))
}
