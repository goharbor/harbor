package export

import (
	"context"
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

const (
	MockDigestValue = "mockDigest"
)

type ArtifactCleanupTestSuite struct {
	suite.Suite
	execMgr *tasktesting.ExecutionManager
	regCli  *registry.FakeClient
	ac      *exportOCIArtifactCleanupManager
}

func (suite *ArtifactCleanupTestSuite) SetupSuite() {

}

func (suite *ArtifactCleanupTestSuite) SetupTest() {
	suite.execMgr = &tasktesting.ExecutionManager{}
	suite.regCli = &registry.FakeClient{}
	suite.ac = &exportOCIArtifactCleanupManager{
		execMgr:           suite.execMgr,
		regCli:            suite.regCli,
		timeWindowMinutes: 0,
		pageSize:          0,
	}
}

func (suite *ArtifactCleanupTestSuite) TestCleanup() {
	// cleanup with default pagesize and default time range
	{
		createTsTimeHistory := time.Now().Add(-time.Duration(24) * time.Hour)
		mock.OnAnything(suite.regCli, "DeleteBlob").Return(nil).Once()
		execs := []*task.Execution{{
			ID:            int64(1),
			VendorType:    job.ScanDataExport,
			VendorID:      -1,
			Status:        "Success",
			StatusMessage: "",
			Metrics:       nil,
			Trigger:       "Schedule",
			ExtraAttrs:    map[string]interface{}{"artifact_digest": MockDigestValue, "create_ts": float64(createTsTimeHistory.Unix())},
			StartTime:     time.Time{},
			UpdateTime:    time.Time{},
			EndTime:       time.Time{},
		}}

		mock.OnAnything(suite.execMgr, "List").Return(execs, nil).Once()
		mock.OnAnything(suite.execMgr, "List").Return(make([]*task.Execution, 0), nil).Once()
		suite.ac.Execute(context.TODO())

		queryMatcher := testifymock.MatchedBy(func(query *q.Query) bool {
			return query.PageSize == DefaultPageSize
		})
		suite.execMgr.AssertCalled(suite.T(), "List", mock.Anything, queryMatcher)
		suite.regCli.AssertCalled(suite.T(), "DeleteBlob", fmt.Sprintf("scandata_export_%v", execs[0].ID), MockDigestValue)
	}

}

func (suite *ArtifactCleanupTestSuite) TestCleanupArtifactWithRecentArtifactTimestamp() {
	{
		createTsTimeFuture := time.Now().Add(time.Duration(24) * time.Hour)
		mock.OnAnything(suite.regCli, "DeleteBlob").Return(nil).Once()
		execs := []*task.Execution{{
			ID:            int64(1),
			VendorType:    job.ScanDataExport,
			VendorID:      -1,
			Status:        "Success",
			StatusMessage: "",
			Metrics:       nil,
			Trigger:       "Schedule",
			ExtraAttrs:    map[string]interface{}{"artifact_digest": MockDigestValue, "create_ts": float64(createTsTimeFuture.Unix())},
			StartTime:     time.Time{},
			UpdateTime:    time.Time{},
			EndTime:       time.Time{},
		}}

		mock.OnAnything(suite.execMgr, "List").Return(execs, nil).Once()
		mock.OnAnything(suite.execMgr, "List").Return(make([]*task.Execution, 0), nil).Once()
		suite.ac.Execute(context.TODO())

		queryMatcher := testifymock.MatchedBy(func(query *q.Query) bool {
			return query.PageSize == DefaultPageSize
		})
		suite.execMgr.AssertCalled(suite.T(), "List", mock.Anything, queryMatcher)
		suite.regCli.AssertNotCalled(suite.T(), "DeleteBlob", fmt.Sprintf("scandata_export_%v", execs[0].ID), MockDigestValue)
	}
}

func (suite *ArtifactCleanupTestSuite) TestCleanupArtifactWithNonDefaultPageSize() {
	{
		suite.ac = &exportOCIArtifactCleanupManager{
			execMgr:           suite.execMgr,
			regCli:            suite.regCli,
			timeWindowMinutes: 0,
			pageSize:          10,
		}
		createTsTimeFuture := time.Now().Add(time.Duration(24) * time.Hour)
		mock.OnAnything(suite.regCli, "DeleteBlob").Return(nil).Once()
		execs := []*task.Execution{{
			ID:            int64(1),
			VendorType:    job.ScanDataExport,
			VendorID:      -1,
			Status:        "Success",
			StatusMessage: "",
			Metrics:       nil,
			Trigger:       "Schedule",
			ExtraAttrs:    map[string]interface{}{"artifact_digest": MockDigestValue, "create_ts": float64(createTsTimeFuture.Unix())},
			StartTime:     time.Time{},
			UpdateTime:    time.Time{},
			EndTime:       time.Time{},
		}}

		mock.OnAnything(suite.execMgr, "List").Return(execs, nil).Once()
		mock.OnAnything(suite.execMgr, "List").Return(make([]*task.Execution, 0), nil).Once()
		suite.ac.Execute(context.TODO())

		queryMatcher := testifymock.MatchedBy(func(query *q.Query) bool {
			return query.PageSize == 10
		})
		suite.execMgr.AssertCalled(suite.T(), "List", mock.Anything, queryMatcher)
		suite.regCli.AssertNotCalled(suite.T(), "DeleteBlob", fmt.Sprintf("scandata_export_%v", execs[0].ID), MockDigestValue)
	}
}

func (suite *ArtifactCleanupTestSuite) TestCleanupArtifactExecListErrors() {
	{
		suite.ac = &exportOCIArtifactCleanupManager{
			execMgr:           suite.execMgr,
			regCli:            suite.regCli,
			timeWindowMinutes: 0,
			pageSize:          10,
		}
		createTsTimeFuture := time.Now().Add(time.Duration(24) * time.Hour)
		mock.OnAnything(suite.regCli, "DeleteBlob").Return(nil).Once()
		execs := []*task.Execution{{
			ID:            int64(1),
			VendorType:    job.ScanDataExport,
			VendorID:      -1,
			Status:        "Success",
			StatusMessage: "",
			Metrics:       nil,
			Trigger:       "Schedule",
			ExtraAttrs:    map[string]interface{}{"artifact_digest": MockDigestValue, "create_ts": float64(createTsTimeFuture.Unix())},
			StartTime:     time.Time{},
			UpdateTime:    time.Time{},
			EndTime:       time.Time{},
		}}

		mock.OnAnything(suite.execMgr, "List").Return(nil, errors.New("test error")).Once()
		err := suite.ac.Execute(context.TODO())
		suite.Error(err)
		queryMatcher := testifymock.MatchedBy(func(query *q.Query) bool {
			return query.PageSize == 10
		})
		suite.execMgr.AssertCalled(suite.T(), "List", mock.Anything, queryMatcher)
		suite.regCli.AssertNotCalled(suite.T(), "DeleteBlob", fmt.Sprintf("scandata_export_%v", execs[0].ID), MockDigestValue)
	}
}

func (suite *ArtifactCleanupTestSuite) TestCleanupSettings() {
	settings := NewCleanupSettings()
	{
		settings.Set("key1", 10)
		settings.Set("key2", "value2")
		suite.Equal("value2", settings.Get("key2"))
		suite.Equal(10, settings.Get("key1"))
		suite.Nil(settings.Get("key3"))
	}

}

func TestArtifactCleanupTestSuite(t *testing.T) {
	suite.Run(t, &ArtifactCleanupTestSuite{})
}
