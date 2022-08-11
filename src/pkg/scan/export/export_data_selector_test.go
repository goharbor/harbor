package export

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ExportDataSelectorTestSuite struct {
	suite.Suite
	exportDataSelector VulnerabilityDataSelector
}

func (suite *ExportDataSelectorTestSuite) SetupSuite() {
	suite.exportDataSelector = NewVulnerabilityDataSelector()
}

func (suite *ExportDataSelectorTestSuite) TestCVEFilter() {
	{
		dataRecords := suite.createDataRecords(10, 1)
		filtered, err := suite.exportDataSelector.Select(dataRecords, CVEIDMatches, "CVEId-1")
		suite.NoError(err)
		suite.Equal(1, len(filtered))
		suite.Equal("CVEId-1", filtered[0].CVEId)
	}
	{
		dataRecords := suite.createDataRecords(10, 1)
		filtered, err := suite.exportDataSelector.Select(dataRecords, CVEIDMatches, "")
		suite.NoError(err)
		suite.Equal(10, len(filtered))
	}
}

func (suite *ExportDataSelectorTestSuite) TestPackageFilter() {
	{
		dataRecords := suite.createDataRecords(10, 1)
		filtered, err := suite.exportDataSelector.Select(dataRecords, PackageMatches, "Package1")
		suite.NoError(err)
		suite.Equal(1, len(filtered))
		suite.Equal("Package1", filtered[0].Package)
	}
	{
		dataRecords := suite.createDataRecords(10, 1)
		filtered, err := suite.exportDataSelector.Select(dataRecords, PackageMatches, "")
		suite.NoError(err)
		suite.Equal(10, len(filtered))
	}
}

func (suite *ExportDataSelectorTestSuite) TestScannerNameFilter() {
	{
		dataRecords := suite.createDataRecords(10, 1)
		filtered, err := suite.exportDataSelector.Select(dataRecords, ScannerMatches, "TestScanner1")
		suite.NoError(err)
		suite.Equal(1, len(filtered))
		suite.Equal("TestScanner1", filtered[0].ScannerName)
	}
	{
		dataRecords := suite.createDataRecords(10, 1)
		filtered, err := suite.exportDataSelector.Select(dataRecords, ScannerMatches, "")
		suite.NoError(err)
		suite.Equal(10, len(filtered))
	}
}

func TestExportDataSelectorTestSuite(t *testing.T) {
	suite.Run(t, &ExportDataSelectorTestSuite{})
}

func (suite *ExportDataSelectorTestSuite) createDataRecords(numRecs int, ownerId int64) []Data {
	data := make([]Data, 0)
	for i := 1; i <= numRecs; i++ {
		dataRec := Data{
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
