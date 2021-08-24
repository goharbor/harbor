package scandataexport

import (
	"fmt"
	"github.com/goharbor/harbor/src/controller/scandataexport"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	htesting "github.com/goharbor/harbor/src/testing"
	exporttesting "github.com/goharbor/harbor/src/testing/controller/scan/export"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/opencontainers/go-digest"
	testifymock "github.com/stretchr/testify/mock"
	"os"
	"strconv"

	"github.com/stretchr/testify/suite"
	"testing"
)

const JobId = float64(100)
const MockDigest = "mockDigest"

type ScanDataExportJobTestSuite struct {
	htesting.Suite
	execMgr          *tasktesting.ExecutionManager
	job              *ScanDataExport
	exportMgr        *exporttesting.Manager
	regCli           *exporttesting.RegistryClient
	digestCalculator *exporttesting.DigestCalculator
}

func (suite *ScanDataExportJobTestSuite) SetupSuite() {
	suite.execMgr = &tasktesting.ExecutionManager{}
	suite.exportMgr = &exporttesting.Manager{}
	suite.regCli = &exporttesting.RegistryClient{}
	suite.digestCalculator = &exporttesting.DigestCalculator{}
	suite.job = &ScanDataExport{
		execMgr:               suite.execMgr,
		exportMgr:             suite.exportMgr,
		scanDataExportDirPath: "/tmp",
		regCli:                suite.regCli,
		digestCalculator:      suite.digestCalculator,
	}
}

func (suite *ScanDataExportJobTestSuite) SetupTest() {
	suite.execMgr.On("UpdateExtraAttrs", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// all BLOB related operations succeed
	mock.OnAnything(suite.regCli, "PushBlob").Return(nil).Once()
}

func (suite *ScanDataExportJobTestSuite) TestRun() {
	data := suite.createDataRecords(3, 1)
	mock.OnAnything(suite.exportMgr, "Fetch").Return(data, nil).Once()
	mock.OnAnything(suite.exportMgr, "Fetch").Return(make([]export.Data, 0), nil).Once()
	mock.OnAnything(suite.digestCalculator, "Calculate").Return(digest.Digest(MockDigest), nil)
	params := job.Parameters{}
	params["JobId"] = JobId
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.NoError(err)
	// assert that the delete method has been called on the repository
	suite.regCli.AssertCalled(suite.T(), "PushBlob", "scandata_export_100", MockDigest, mock.Anything, mock.Anything)

	m := make(map[string]interface{})
	m[scandataexport.DigestKey] = MockDigest
	m[CreateTimestampKey] = mock.Anything

	extraAttrsMatcher := testifymock.MatchedBy(func(attrsMap map[string]interface{}) bool {
		_, ok := m[CreateTimestampKey]
		return attrsMap[scandataexport.DigestKey] == MockDigest && ok
	})
	suite.execMgr.AssertCalled(suite.T(), "UpdateExtraAttrs", mock.Anything, int64(JobId), extraAttrsMatcher)
}

func (suite *ScanDataExportJobTestSuite) TearDownTest() {
	path := fmt.Sprintf("/tmp/scandata_export_%v.csv", JobId)
	err := os.Remove(path)
	suite.NoError(err)
}

func (suite *ScanDataExportJobTestSuite) createDataRecords(numRecs int, ownerId int64) []export.Data {
	data := make([]export.Data, 0)
	for i := 1; i <= numRecs; i++ {
		dataRec := export.Data{
			Id:           int64(i),
			ProjectName:  fmt.Sprintf("TestProject%d", i),
			ProjectOwner: strconv.FormatInt(ownerId, 10),
			ScannerName:  fmt.Sprintf("TestScanner%d", i),
			CVEId:        fmt.Sprintf("CVEId-%d", i),
			Package:      fmt.Sprintf("Package%d", i),
			Severity:     fmt.Sprintf("Severity%d", i),
			CVSSScoreV3:  fmt.Sprintf("3.0"),
			CVSSScoreV2:  fmt.Sprintf("2.0"),
			CVSSVectorV3: fmt.Sprintf("TestCVSSVectorV3%d", i),
			CVSSVectorV2: fmt.Sprintf("TestCVSSVectorV2%d", i),
			CWEIds:       "",
		}
		data = append(data, dataRec)
	}
	return data
}
func TestScanDataExportJobSuite(t *testing.T) {
	suite.Run(t, &ScanDataExportJobTestSuite{})
}
