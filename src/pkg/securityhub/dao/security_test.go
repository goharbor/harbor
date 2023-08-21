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
	"strings"
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
		`delete from tag`,
		`delete from artifact_accessory`,
		`delete from artifact`,
		`insert into scan_report(uuid, digest, registration_uuid, mime_type, critical_cnt, high_cnt, medium_cnt, low_cnt, unknown_cnt, fixable_cnt) values('uuid', 'digest1001', 'ruuid', 'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0', 50, 50, 50, 0, 0, 20)`,

		`insert into scan_report(uuid, digest, registration_uuid, mime_type, critical_cnt, high_cnt, medium_cnt, low_cnt, unknown_cnt, fixable_cnt) values('cnabreport1010', 'digest1010', 'ruuid', 'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0', 0, 0, 0, 0, 0, 20)`,
		`insert into scan_report(uuid, digest, registration_uuid, mime_type, critical_cnt, high_cnt, medium_cnt, low_cnt, unknown_cnt, fixable_cnt) values('cnabreport1011', 'digest1011', 'ruuid', 'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0', 0, 0, 0, 0, 0, 20)`,
		`insert into scan_report(uuid, digest, registration_uuid, mime_type, critical_cnt, high_cnt, medium_cnt, low_cnt, unknown_cnt, fixable_cnt) values('cnabreport1013', 'digest1013', 'ruuid', 'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0', 0, 0, 0, 0, 0, 20)`,
		`insert into scan_report(uuid, digest, registration_uuid, mime_type, critical_cnt, high_cnt, medium_cnt, low_cnt, unknown_cnt, fixable_cnt) values('cnabreport1014', 'digest1014', 'ruuid', 'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0', 0, 0, 0, 0, 0, 20)`,
		`insert into scan_report(uuid, digest, registration_uuid, mime_type, critical_cnt, high_cnt, medium_cnt, low_cnt, unknown_cnt, fixable_cnt) values('notag001', 'digest1017', 'ruuid', 'application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0', 0, 0, 0, 0, 0, 0)`,

		// image index
		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values  (1001, 1, 'library/hello-world', 'digest1001', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.docker.distribution.manifest.v2+json', 4452, '', null, '');`,
		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1002, 1, 'library/hello-world', 'digest1002', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,
		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1003, 1, 'library/hello-world', 'digest1003', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		// construct a cnab artifact
		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1010, 1, 'library/cnab_example', 'digest1010', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1011, 1, 'library/cnab_example', 'digest1011', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1012, 1, 'library/cnab_example', 'digest1012', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1013, 1, 'library/cnab_example', 'digest1013', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1014, 1, 'library/cnab_example', 'digest1014', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1015, 1, 'library/cnab_example', 'digest1015', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1016, 1, 'library/cnab_example', 'digest1016', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		// notag for this artifact
		`insert into artifact (id, project_id, repository_name, digest, type, pull_time, push_time, repository_id, media_type, manifest_media_type, size, extra_attrs, annotations, icon)
values (1017, 1, 'library/notag', 'digest1017', 'IMAGE', '2023-06-02 09:16:47.838778', '2023-06-02 01:45:55.050785', 1742, 'application/vnd.docker.container.image.v1+json', 'application/vnd.oci.image.config.v1+json', 4452, '', null, '');`,

		`insert into tag (id, repository_id, artifact_id, name, push_time, pull_time) values (1001, 1742, 1001, 'latest', '2023-06-02 01:45:55.050785', '2023-06-02 09:16:47.838778')`,

		`INSERT INTO artifact_accessory (id, artifact_id, subject_artifact_id, type, size, digest, creation_time, subject_artifact_digest, subject_artifact_repo) VALUES (1001, 1002, 1, 'signature.cosign', 2109, 'sha256:08c64c0de2667abcf3974b4b75b82903f294680b81584318adc4826d0dcb7a9c', '2023-08-03 04:54:32.102928', 'sha256:a97a153152fcd6410bdf4fb64f5622ecf97a753f07dcc89dab14509d059736cf', 'library/nuxeo')`,

		`INSERT INTO artifact_reference (id, parent_id, child_id, child_digest, platform, urls, annotations) VALUES (1001, 1001, 1003, 'sha256:d2b2f2980e9ccc570e5726b56b54580f23a018b7b7314c9eaff7e5e479c78657', '{"architecture":"amd64","os":"linux"}', '', null)`,

		`INSERT INTO artifact_reference (id, parent_id, child_id, child_digest, platform, urls, annotations) VALUES (1002, 1012, 1010, 'digest1010', '{"architecture":"amd64","os":"linux"}', '', null)`,
		`INSERT INTO artifact_reference (id, parent_id, child_id, child_digest, platform, urls, annotations) VALUES (1003, 1012, 1011, 'digest1011', '{"architecture":"amd64","os":"linux"}', '', null)`,
		`INSERT INTO artifact_reference (id, parent_id, child_id, child_digest, platform, urls, annotations) VALUES (1004, 1015, 1013, 'digest1013', '{"architecture":"amd64","os":"linux"}', '', null)`,
		`INSERT INTO artifact_reference (id, parent_id, child_id, child_digest, platform, urls, annotations) VALUES (1005, 1015, 1014, 'digest1014', '{"architecture":"amd64","os":"linux"}', '', null)`,
		`INSERT INTO artifact_reference (id, parent_id, child_id, child_digest, platform, urls, annotations) VALUES (1006, 1016, 1012, 'digest1012', '{"architecture":"amd64","os":"linux"}', '', null)`,
		`INSERT INTO artifact_reference (id, parent_id, child_id, child_digest, platform, urls, annotations) VALUES (1007, 1016, 1015, 'digest1015', '{"architecture":"amd64","os":"linux"}', '', null)`,

		`insert into scanner_registration (name, url, uuid, auth) values('trivy', 'https://www.vmware.com', 'ruuid', 'empty')`,
		`insert into vulnerability_record (id, cve_id, registration_uuid, cvss_score_v3) values (1, '2023-4567-12345', 'ruuid', 9.8)`,
		`insert into report_vulnerability_record (report_uuid, vuln_record_id) VALUES ('uuid', 1)`,
		`insert into report_vulnerability_record (report_uuid, vuln_record_id) VALUES ('cnabreport1010', 1)`,
		`insert into report_vulnerability_record (report_uuid, vuln_record_id) VALUES ('cnabreport1011', 1)`,
		`insert into report_vulnerability_record (report_uuid, vuln_record_id) VALUES ('cnabreport1013', 1)`,
		`insert into report_vulnerability_record (report_uuid, vuln_record_id) VALUES ('cnabreport1014', 1)`,
		`insert into report_vulnerability_record (report_uuid, vuln_record_id) VALUES ('notag001', 1)`,

		`INSERT INTO tag (repository_id, artifact_id, name) VALUES (1, (select id from artifact where repository_name = 'library/hello-world' limit 1), 'tag_test')`,

		// tag a cnab child, it should be considered as standalone artifact
		`INSERT INTO tag (repository_id, artifact_id, name) VALUES (2, 1014, 'tag_test')`,
	})

	testDao.ExecuteBatchSQL([]string{
		`INSERT INTO scanner_registration (name, url, uuid, auth) values('trivy2', 'https://www.trivy.com', 'uuid2', 'empty')`,
		`INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2021-44228', 'uuid2', 10, 'org.apache.logging.log4j:log4j-core');
		INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2021-21345', 'uuid2', 9.9, 'com.thoughtworks.xstream:xstream');
		INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2016-1585', 'uuid2', 9.8, 'libapparmor1');
		INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2023-0950', 'uuid2', 9.8, 'ure');
		INSERT INTO vulnerability_record(cve_id, registration_uuid, cvss_score_v3, package) VALUES ('CVE-2022-47629', 'uuid2', 9.8, 'libksba8');`,
		`INSERT INTO report_vulnerability_record(report_uuid, vuln_record_id) select 'uuid', id vuln_record_id from vulnerability_record where cve_id in ('CVE-2021-44228', 'CVE-2021-21345', 'CVE-2016-1585', 'CVE-2023-0950', 'CVE-2022-47629')`,
	})
}

func (suite *SecurityDaoTestSuite) TearDownTest() {
	testDao.ExecuteBatchSQL([]string{
		`delete from scan_report where uuid = 'uuid'`,
		`delete from tag`,
		`delete from artifact_accessory`,
		`delete from artifact_reference`,
		`delete from artifact`,
		`delete from scanner_registration`,
		`delete from vulnerability_record`,
		`delete from report_vulnerability_record`,
	})
}

func (suite *SecurityDaoTestSuite) TestGetSummary() {
	s, err := suite.dao.Summary(suite.Context(), "ruuid", 0, nil)
	suite.Require().NoError(err)
	suite.Equal(int64(50), s.CriticalCnt)
	suite.Equal(int64(50), s.HighCnt)
	suite.Equal(int64(50), s.MediumCnt)
	suite.Equal(int64(100), s.FixableCnt)
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
	suite.Equal(int64(4), count)
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
		{"happy_path", args{q.New(q.KeyWords{"sample": 1}), map[string]*filterMetaData{"sample": &filterMetaData{DataType: intType}}}, false},
		{"happy_path_cve_id", args{q.New(q.KeyWords{"cve_id": "CVE-2023-2345"}), map[string]*filterMetaData{"cve_id": &filterMetaData{DataType: stringType}}}, false},
		{"happy_path_severity", args{q.New(q.KeyWords{"severity": "Critical"}), map[string]*filterMetaData{"severity": &filterMetaData{DataType: stringType}}}, false},
		{"happy_path_cvss_score_v3", args{q.New(q.KeyWords{"cvss_score_v3": &q.Range{Min: 2.0, Max: 3.0}}), map[string]*filterMetaData{"cvss_score_v3": &filterMetaData{DataType: rangeType, FilterFunc: rangeFilter}}}, false},
		{"unhappy_path", args{q.New(q.KeyWords{"sample": 1}), map[string]*filterMetaData{"a": &filterMetaData{DataType: intType}}}, true},
		{"unhappy_path2", args{q.New(q.KeyWords{"cve_id": 1}), map[string]*filterMetaData{"cve_id": &filterMetaData{DataType: stringType}}}, true},
		{"unhappy_path3", args{q.New(q.KeyWords{"severity": &q.Range{Min: 2.0, Max: 10.0}}), map[string]*filterMetaData{"severity": &filterMetaData{DataType: stringType}}}, true},
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
		{"digest", args{suite.Context(), "digest", q.New(q.KeyWords{"digest": "digest123"})}, " and a.digest = ?", []interface{}{"digest123"}},
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

func (suite *SecurityDaoTestSuite) TestCountArtifact() {
	// hello-world  +1
	// cnab +1
	// cnab child with tag +1
	// top level artifact notag +1
	// = 4
	count, err := suite.dao.TotalArtifactsCount(suite.Context(), 0)
	suite.NoError(err)
	suite.Equal(int64(4), count)
}
func (suite *SecurityDaoTestSuite) TestCountVul() {
	count, err := suite.dao.CountVulnerabilities(suite.Context(), "ruuid", 0, true, nil)
	suite.NoError(err)
	suite.Equal(int64(6), count)
}

func (suite *SecurityDaoTestSuite) TestListVul() {
	vuls, err := suite.dao.ListVulnerabilities(suite.Context(), "ruuid", 0, nil)
	suite.NoError(err)
	suite.Equal(6, len(vuls))
}

func (suite *SecurityDaoTestSuite) TestTagFilter() {
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
		{"normal", args{suite.Context(), "tag", q.New(q.KeyWords{"tag": "tag_test"})}, " and a.id IN", nil},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			gotSQLStr, gotParams := tagFilter(tt.args.ctx, tt.args.key, tt.args.query)
			suite.True(strings.Contains(gotSQLStr, tt.wantSQLStr), "tagFilter() gotSqlStr = %v, want %v", gotSQLStr, tt.wantSQLStr)
			suite.Equal(gotParams, tt.wantParams, "tagFilter() gotParams = %v, want %v", gotParams, tt.wantParams)
		})
	}
}

func (suite *SecurityDaoTestSuite) TestApplyVulFilter() {
	type args struct {
		ctx    context.Context
		sqlStr string
		query  *q.Query
		params []interface{}
	}
	tests := []struct {
		name       string
		args       args
		wantSQLStr string
		wantParams []interface{}
	}{
		{"normal", args{suite.Context(), "select * from vulnerability_record", q.New(q.KeyWords{"tag": "tag_test"}), nil}, " and a.id IN", nil},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			gotSQLStr, gotParams := applyVulFilter(tt.args.ctx, tt.args.sqlStr, tt.args.query, tt.args.params)
			suite.True(strings.Contains(gotSQLStr, tt.wantSQLStr), "applyVulFilter() gotSqlStr = %v, want %v", gotSQLStr, tt.wantSQLStr)
			suite.Equal(gotParams, tt.wantParams, "applyVulFilter() gotParams = %v, want %v", gotParams, tt.wantParams)
		})
	}
}
