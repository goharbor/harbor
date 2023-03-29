package scandataexport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/opencontainers/go-digest"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/selector"
	artpkg "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	"github.com/goharbor/harbor/src/pkg/systemartifact/model"
	"github.com/goharbor/harbor/src/pkg/task"
	htesting "github.com/goharbor/harbor/src/testing"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	export2 "github.com/goharbor/harbor/src/testing/pkg/scan/export"
	systemartifacttesting "github.com/goharbor/harbor/src/testing/pkg/systemartifact"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
)

const (
	ExecID     int64 = 1000000
	JobId            = "1000000"
	MockDigest       = "mockDigest"
)

type ScanDataExportJobTestSuite struct {
	htesting.Suite
	execMgr          *tasktesting.ExecutionManager
	job              *ScanDataExport
	exportMgr        *export2.Manager
	digestCalculator *export2.ArtifactDigestCalculator
	filterProcessor  *export2.FilterProcessor
	projectMgr       *project.Manager
	sysArtifactMgr   *systemartifacttesting.Manager
}

func (suite *ScanDataExportJobTestSuite) SetupSuite() {
}

func (suite *ScanDataExportJobTestSuite) SetupTest() {
	suite.execMgr = &tasktesting.ExecutionManager{}
	suite.exportMgr = &export2.Manager{}
	suite.digestCalculator = &export2.ArtifactDigestCalculator{}
	suite.filterProcessor = &export2.FilterProcessor{}
	suite.projectMgr = &project.Manager{}
	suite.sysArtifactMgr = &systemartifacttesting.Manager{}
	suite.job = &ScanDataExport{
		execMgr:               suite.execMgr,
		exportMgr:             suite.exportMgr,
		scanDataExportDirPath: "/tmp",
		digestCalculator:      suite.digestCalculator,
		filterProcessor:       suite.filterProcessor,
		sysArtifactMgr:        suite.sysArtifactMgr,
	}

	suite.execMgr.On("UpdateExtraAttrs", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// all BLOB related operations succeed
	suite.sysArtifactMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
}

func (suite *ScanDataExportJobTestSuite) TestRun() {

	data := suite.createDataRecords(3)
	mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
	mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
	mock.OnAnything(suite.filterProcessor, "ProcessRepositoryFilter").Return([]int64{1}, nil).Once()
	mock.OnAnything(suite.filterProcessor, "ProcessTagFilter").Return([]*artifact.Artifact{{Artifact: artpkg.Artifact{ID: 1}}}, nil).Once()
	mock.OnAnything(suite.filterProcessor, "ProcessLabelFilter").Return([]*artifact.Artifact{{Artifact: artpkg.Artifact{ID: 1}}}, nil).Once()

	execAttrs := make(map[string]interface{})
	execAttrs[export.JobNameAttribute] = "test-job"
	execAttrs[export.UserNameAttribute] = "test-user"
	mock.OnAnything(suite.execMgr, "Get").Return(&task.Execution{ID: ExecID, ExtraAttrs: execAttrs}, nil)

	params := job.Parameters{}
	params[export.JobModeKey] = export.JobModeExport
	params["JobId"] = JobId
	params["Request"] = map[string]interface{}{
		"projects": []int64{1},
	}
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.NoError(err)
	sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
		return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
	})
	suite.sysArtifactMgr.AssertCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)

	m := make(map[string]interface{})
	m[export.DigestKey] = MockDigest
	m[export.CreateTimestampKey] = mock.Anything

	extraAttrsMatcher := testifymock.MatchedBy(func(attrsMap map[string]interface{}) bool {
		_, ok := m[export.CreateTimestampKey]
		return attrsMap[export.DigestKey] == MockDigest && ok && attrsMap[export.JobNameAttribute] == "test-job" && attrsMap[export.UserNameAttribute] == "test-user"
	})
	suite.execMgr.AssertCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, extraAttrsMatcher)
	_, err = os.Stat("/tmp/scandata_export_1000000.csv")
	suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")

}

func (suite *ScanDataExportJobTestSuite) TestRunWithEmptyData() {
	var data []export.Data
	mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
	mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)

	execAttrs := make(map[string]interface{})
	execAttrs[export.JobNameAttribute] = "test-job"
	execAttrs[export.UserNameAttribute] = "test-user"
	mock.OnAnything(suite.execMgr, "Get").Return(&task.Execution{ID: ExecID, ExtraAttrs: execAttrs}, nil)

	params := job.Parameters{}
	params[export.JobModeKey] = export.JobModeExport
	params["JobId"] = JobId
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.NoError(err)

	extraAttrsMatcher := testifymock.MatchedBy(func(attrsMap map[string]interface{}) bool {
		return attrsMap["status_message"] == "No vulnerabilities found or matched" && attrsMap[export.JobNameAttribute] == "test-job" && attrsMap[export.UserNameAttribute] == "test-user"
	})
	suite.execMgr.AssertCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, extraAttrsMatcher)
}

func (suite *ScanDataExportJobTestSuite) TestRunAttributeUpdateError() {

	data := suite.createDataRecords(3)
	mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
	mock.OnAnything(suite.filterProcessor, "ProcessRepositoryFilter").Return([]int64{1}, nil).Once()
	mock.OnAnything(suite.filterProcessor, "ProcessTagFilter").Return([]*artifact.Artifact{{Artifact: artpkg.Artifact{ID: 1}}}, nil).Once()
	mock.OnAnything(suite.filterProcessor, "ProcessLabelFilter").Return([]*artifact.Artifact{{Artifact: artpkg.Artifact{ID: 1}}}, nil).Once()
	mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)

	execAttrs := make(map[string]interface{})
	execAttrs[export.JobNameAttribute] = "test-job"
	execAttrs[export.UserNameAttribute] = "test-user"
	mock.OnAnything(suite.execMgr, "Get").Return(nil, errors.New("test-error"))

	params := job.Parameters{}
	params[export.JobModeKey] = export.JobModeExport
	params["JobId"] = JobId
	params["Request"] = map[string]interface{}{
		"projects": []int{1},
	}
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.Error(err)
	sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
		return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
	})
	suite.sysArtifactMgr.AssertCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)

	m := make(map[string]interface{})
	m[export.DigestKey] = MockDigest
	m[export.CreateTimestampKey] = mock.Anything

	extraAttrsMatcher := testifymock.MatchedBy(func(attrsMap map[string]interface{}) bool {
		_, ok := m[export.CreateTimestampKey]
		return attrsMap[export.DigestKey] == MockDigest && ok && attrsMap[export.JobNameAttribute] == "test-job" && attrsMap[export.UserNameAttribute] == "test-user"
	})
	suite.execMgr.AssertNotCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, extraAttrsMatcher)
	_, err = os.Stat("/tmp/scandata_export_1000000.csv")
	suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")

}

func (suite *ScanDataExportJobTestSuite) TestExtractCriteria() {
	// empty request should return error
	_, err := suite.job.extractCriteria(job.Parameters{})
	suite.Error(err)
	// invalid request should return error
	_, err = suite.job.extractCriteria(job.Parameters{"Request": ""})
	suite.Error(err)
	// valid request should not return error and trim space
	c, err := suite.job.extractCriteria(job.Parameters{"Request": map[string]interface{}{
		"CVEIds":       "CVE-123, CVE-456 ",
		"Repositories": " test-repo1 ",
		"Tags":         "test-tag1, test-tag2",
	}})
	suite.NoError(err)
	suite.Equal("CVE-123,CVE-456", c.CVEIds)
	suite.Equal("test-repo1", c.Repositories)
	suite.Equal("test-tag1,test-tag2", c.Tags)
}

func (suite *ScanDataExportJobTestSuite) TestRunWithCriteria() {
	{
		data := suite.createDataRecords(3)

		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		execAttrs := make(map[string]interface{})
		execAttrs[export.JobNameAttribute] = "test-job"
		execAttrs[export.UserNameAttribute] = "test-user"
		mock.OnAnything(suite.execMgr, "Get").Return(&task.Execution{ID: ExecID, ExtraAttrs: execAttrs}, nil).Once()

		repoCandidates := []int64{1}
		artCandidates := []*artifact.Artifact{{Artifact: artpkg.Artifact{ID: 1, Digest: "digest1"}}}
		mock.OnAnything(suite.filterProcessor, "ProcessRepositoryFilter").Return(repoCandidates, nil)
		mock.OnAnything(suite.filterProcessor, "ProcessTagFilter").Return(artCandidates, nil)
		mock.OnAnything(suite.filterProcessor, "ProcessLabelFilter").Return(artCandidates, nil)

		criteria := export.Request{
			CVEIds:       "CVE-123",
			Labels:       []int64{1},
			Projects:     []int64{1},
			Repositories: "test-repo",
			Tags:         "test-tag",
		}
		criteriaMap := make(map[string]interface{})
		bytes, _ := json.Marshal(criteria)
		json.Unmarshal(bytes, &criteriaMap)
		params := job.Parameters{}
		params[export.JobModeKey] = export.JobModeExport
		params["JobId"] = JobId
		params["Request"] = criteriaMap

		ctx := &mockjobservice.MockJobContext{}
		ctx.On("SystemContext").Return(context.TODO()).Once()

		err := suite.job.Run(ctx, params)
		suite.NoError(err)
		sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
			return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
		})
		suite.sysArtifactMgr.AssertCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)

		m := make(map[string]interface{})
		m[export.DigestKey] = MockDigest
		m[export.CreateTimestampKey] = mock.Anything

		extraAttrsMatcher := testifymock.MatchedBy(func(attrsMap map[string]interface{}) bool {
			_, ok := m[export.CreateTimestampKey]
			return attrsMap[export.DigestKey] == MockDigest && ok
		})
		suite.execMgr.AssertCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, extraAttrsMatcher)
		_, err = os.Stat("/tmp/scandata_export_1000000.csv")

		exportParamsMatcher := testifymock.MatchedBy(func(params export.Params) bool {
			return reflect.DeepEqual(params.CVEIds, criteria.CVEIds)
		})
		suite.exportMgr.AssertCalled(suite.T(), "Fetch", mock.Anything, exportParamsMatcher)

		suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")
	}

	{
		mock.OnAnything(suite.sysArtifactMgr, "Create").Return(int64(1), nil).Once()
		data := suite.createDataRecords(3)
		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		execAttrs := make(map[string]interface{})
		execAttrs[export.JobNameAttribute] = "test-job"
		execAttrs[export.UserNameAttribute] = "test-user"
		mock.OnAnything(suite.execMgr, "Get").Return(&task.Execution{ID: ExecID, ExtraAttrs: execAttrs}, nil).Once()

		repoCandidate1 := &selector.Candidate{NamespaceID: 1}
		repoCandidates := []*selector.Candidate{repoCandidate1}
		mock.OnAnything(suite.filterProcessor, "ProcessRepositoryFilter").Return(repoCandidates, nil)
		mock.OnAnything(suite.filterProcessor, "ProcessTagFilter").Return(repoCandidates, nil)

		criteria := export.Request{
			CVEIds:   "CVE-123",
			Labels:   []int64{1},
			Projects: []int64{1},
			Tags:     "test-tag",
		}
		criteriaMap := make(map[string]interface{})
		bytes, _ := json.Marshal(criteria)
		json.Unmarshal(bytes, &criteriaMap)
		params := job.Parameters{}
		params[export.JobModeKey] = export.JobModeExport
		params["JobId"] = JobId
		params["Request"] = criteriaMap

		ctx := &mockjobservice.MockJobContext{}
		ctx.On("SystemContext").Return(context.TODO()).Times(3)

		err := suite.job.Run(ctx, params)
		suite.NoError(err)
		sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
			return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
		})
		suite.sysArtifactMgr.AssertCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)
		m := make(map[string]interface{})
		m[export.DigestKey] = MockDigest
		m[export.CreateTimestampKey] = mock.Anything

		extraAttrsMatcher := testifymock.MatchedBy(func(attrsMap map[string]interface{}) bool {
			_, ok := m[export.CreateTimestampKey]
			return attrsMap[export.DigestKey] == MockDigest && ok
		})
		suite.execMgr.AssertCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, extraAttrsMatcher)
		_, err = os.Stat("/tmp/scandata_export_1000000.csv")

		exportParamsMatcher := testifymock.MatchedBy(func(params export.Params) bool {
			return reflect.DeepEqual(params.CVEIds, criteria.CVEIds)
		})
		suite.exportMgr.AssertCalled(suite.T(), "Fetch", mock.Anything, exportParamsMatcher)

		suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")
	}
}

func (suite *ScanDataExportJobTestSuite) TestRunWithCriteriaForRepositoryIdFilter() {
	{
		data := suite.createDataRecords(3)

		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		execAttrs := make(map[string]interface{})
		execAttrs[export.JobNameAttribute] = "test-job"
		execAttrs[export.UserNameAttribute] = "test-user"
		mock.OnAnything(suite.execMgr, "Get").Return(&task.Execution{ID: ExecID, ExtraAttrs: execAttrs}, nil).Once()

		mock.OnAnything(suite.filterProcessor, "ProcessRepositoryFilter").Return([]int64{1}, errors.New("test error")).Once()
		mock.OnAnything(suite.filterProcessor, "ProcessTagFilter").Return([]*artifact.Artifact{{Artifact: artpkg.Artifact{ID: 1}}}, nil).Once()

		criteria := export.Request{
			CVEIds:       "CVE-123",
			Labels:       []int64{1},
			Projects:     []int64{1},
			Repositories: "test-repo",
			Tags:         "test-tag",
		}
		criteriaMap := make(map[string]interface{})
		bytes, _ := json.Marshal(criteria)
		json.Unmarshal(bytes, &criteriaMap)
		params := job.Parameters{}
		params[export.JobModeKey] = export.JobModeExport
		params["JobId"] = JobId
		params["Request"] = criteriaMap

		ctx := &mockjobservice.MockJobContext{}
		ctx.On("SystemContext").Return(context.TODO()).Once()

		err := suite.job.Run(ctx, params)
		suite.Error(err)
		sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
			return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
		})
		suite.sysArtifactMgr.AssertNotCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)
		suite.execMgr.AssertNotCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, mock.Anything)
		_, err = os.Stat("/tmp/scandata_export_1000000.csv")

		suite.exportMgr.AssertNotCalled(suite.T(), "Fetch", mock.Anything, mock.Anything)

		suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")
	}

	// empty list of repo ids
	{
		data := suite.createDataRecords(3)

		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		execAttrs := make(map[string]interface{})
		execAttrs[export.JobNameAttribute] = "test-job"
		execAttrs[export.UserNameAttribute] = "test-user"
		mock.OnAnything(suite.execMgr, "Get").Return(&task.Execution{ID: ExecID, ExtraAttrs: execAttrs}, nil).Once()

		mock.OnAnything(suite.filterProcessor, "ProcessRepositoryFilter").Return([]int64{}, nil).Once()
		mock.OnAnything(suite.filterProcessor, "ProcessTagFilter").Return([]*artifact.Artifact{}, nil).Once()

		criteria := export.Request{
			CVEIds:       "CVE-123",
			Labels:       []int64{1},
			Projects:     []int64{1},
			Repositories: "test-repo",
			Tags:         "test-tag",
		}
		criteriaMap := make(map[string]interface{})
		bytes, _ := json.Marshal(criteria)
		json.Unmarshal(bytes, &criteriaMap)
		params := job.Parameters{}
		params[export.JobModeKey] = export.JobModeExport
		params["JobId"] = JobId
		params["Request"] = criteriaMap

		ctx := &mockjobservice.MockJobContext{}
		ctx.On("SystemContext").Return(context.TODO()).Once()

		err := suite.job.Run(ctx, params)
		suite.NoError(err)
		sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
			return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
		})
		suite.sysArtifactMgr.AssertNotCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)
		suite.execMgr.AssertCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, mock.Anything)
		_, err = os.Stat("/tmp/scandata_export_1000000.csv")

		suite.exportMgr.AssertNotCalled(suite.T(), "Fetch", mock.Anything, mock.Anything)

		suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")
	}

}

func (suite *ScanDataExportJobTestSuite) TestRunWithCriteriaForRepositoryIdWithTagFilter() {
	{
		data := suite.createDataRecords(3)

		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		execAttrs := make(map[string]interface{})
		execAttrs[export.JobNameAttribute] = "test-job"
		execAttrs[export.UserNameAttribute] = "test-user"
		mock.OnAnything(suite.execMgr, "Get").Return(&task.Execution{ID: ExecID, ExtraAttrs: execAttrs}, nil).Once()

		mock.OnAnything(suite.filterProcessor, "ProcessRepositoryFilter").Return([]int64{1}, nil).Once()
		mock.OnAnything(suite.filterProcessor, "ProcessTagFilter").Return(nil, errors.New("test error")).Once()

		criteria := export.Request{
			CVEIds:       "CVE-123",
			Labels:       []int64{1},
			Projects:     []int64{1},
			Repositories: "test-repo",
			Tags:         "test-tag",
		}
		criteriaMap := make(map[string]interface{})
		bytes, _ := json.Marshal(criteria)
		json.Unmarshal(bytes, &criteriaMap)
		params := job.Parameters{}
		params[export.JobModeKey] = export.JobModeExport
		params["JobId"] = JobId
		params["Request"] = criteriaMap

		ctx := &mockjobservice.MockJobContext{}
		ctx.On("SystemContext").Return(context.TODO()).Once()

		err := suite.job.Run(ctx, params)
		suite.Error(err)
		sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
			return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
		})
		suite.sysArtifactMgr.AssertNotCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)
		suite.execMgr.AssertNotCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, mock.Anything)
		_, err = os.Stat("/tmp/scandata_export_1000000.csv")

		suite.exportMgr.AssertNotCalled(suite.T(), "Fetch", mock.Anything, mock.Anything)

		suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")
	}

	// empty list of repo ids after applying tag filters
	{
		data := suite.createDataRecords(3)

		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
		mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
		mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
		execAttrs := make(map[string]interface{})
		execAttrs[export.JobNameAttribute] = "test-job"
		execAttrs[export.UserNameAttribute] = "test-user"
		mock.OnAnything(suite.execMgr, "Get").Return(&task.Execution{ID: ExecID, ExtraAttrs: execAttrs}, nil).Once()

		mock.OnAnything(suite.filterProcessor, "ProcessRepositoryFilter").Return([]int64{}, nil).Once()
		mock.OnAnything(suite.filterProcessor, "ProcessTagFilter").Return(nil, nil).Once()

		criteria := export.Request{
			CVEIds:       "CVE-123",
			Labels:       []int64{1},
			Projects:     []int64{1},
			Repositories: "test-repo",
			Tags:         "test-tag",
		}
		criteriaMap := make(map[string]interface{})
		bytes, _ := json.Marshal(criteria)
		json.Unmarshal(bytes, &criteriaMap)
		params := job.Parameters{}
		params[export.JobModeKey] = export.JobModeExport
		params["JobId"] = JobId
		params["Request"] = criteriaMap

		ctx := &mockjobservice.MockJobContext{}
		ctx.On("SystemContext").Return(context.TODO()).Once()

		err := suite.job.Run(ctx, params)
		suite.NoError(err)
		sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
			return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
		})
		suite.sysArtifactMgr.AssertNotCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)
		suite.execMgr.AssertCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, ExecID, mock.Anything)
		_, err = os.Stat("/tmp/scandata_export_1000000.csv")

		suite.exportMgr.AssertNotCalled(suite.T(), "Fetch", mock.Anything, mock.Anything)

		suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")
	}

}

func (suite *ScanDataExportJobTestSuite) TestExportDigestCalculationErrorsOut() {
	data := suite.createDataRecords(3)
	mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
	mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
	mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(""), errors.New("test error"))
	params := job.Parameters{}
	params[export.JobModeKey] = export.JobModeExport
	params["JobId"] = JobId
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.Error(err)
	sysArtifactRecordMatcher := testifymock.MatchedBy(func(sa *model.SystemArtifact) bool {
		return sa.Repository == "scandata_export_1000000" && sa.Vendor == strings.ToLower(export.Vendor) && sa.Digest == MockDigest
	})
	suite.sysArtifactMgr.AssertNotCalled(suite.T(), "Create", mock.Anything, sysArtifactRecordMatcher, mock.Anything)
	suite.execMgr.AssertNotCalled(suite.T(), "UpdateExtraAttrs")
	_, err = os.Stat("/tmp/scandata_export_1000000.csv")
	suite.Truef(os.IsNotExist(err), "Expected CSV file to be deleted")
}

func (suite *ScanDataExportJobTestSuite) TearDownTest() {
	path := fmt.Sprintf("/tmp/scandata_export_%v.csv", JobId)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return
	}
	err = os.Remove(path)
	suite.NoError(err)
}

func (suite *ScanDataExportJobTestSuite) createDataRecords(numRecs int) []export.Data {
	data := make([]export.Data, 0)
	for i := 1; i <= numRecs; i++ {
		dataRec := export.Data{
			ScannerName:    fmt.Sprintf("TestScanner%d", i),
			Repository:     fmt.Sprintf("Repository%d", i),
			ArtifactDigest: fmt.Sprintf("Digest%d", i),
			CVEId:          fmt.Sprintf("CVEId-%d", i),
			Package:        fmt.Sprintf("Package%d", i),
			Version:        fmt.Sprintf("Version%d", i),
			FixVersion:     fmt.Sprintf("FixVersion%d", i),
			Severity:       fmt.Sprintf("Severity%d", i),
			CWEIds:         "",
		}
		data = append(data, dataRec)
	}
	return data
}
func TestScanDataExportJobSuite(t *testing.T) {
	suite.Run(t, &ScanDataExportJobTestSuite{})
}
