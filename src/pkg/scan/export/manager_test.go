package export

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/q"
	artifactDao "github.com/goharbor/harbor/src/pkg/artifact/dao"
	labelDao "github.com/goharbor/harbor/src/pkg/label/dao"
	labelModel "github.com/goharbor/harbor/src/pkg/label/model"
	projectDao "github.com/goharbor/harbor/src/pkg/project/dao"
	repoDao "github.com/goharbor/harbor/src/pkg/repository/dao"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	daoscan "github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	tagDao "github.com/goharbor/harbor/src/pkg/tag/dao"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	userDao "github.com/goharbor/harbor/src/pkg/user/dao"
	htesting "github.com/goharbor/harbor/src/testing"
)

const RegistrationUUID = "scannerIdExportData"
const ReportUUD = "reportUUId"

type ExportManagerSuite struct {
	htesting.Suite
	artifactDao   artifactDao.DAO
	projectDao    projectDao.DAO
	userDao       userDao.DAO
	repositoryDao repoDao.DAO
	tagDao        tagDao.DAO
	scanDao       daoscan.DAO
	vulnDao       daoscan.VulnerabilityRecordDao
	labelDao      labelDao.DAO
	exportManager Manager
	testDataId    testDataIds
}

type testDataIds struct {
	artifactId   []int64
	repositoryId []int64
	artRefId     []int64
	reportId     []int64
	tagId        []int64
	vulnRecs     []int64
	labelId      []int64
	labelRefId   []int64
}

func (suite *ExportManagerSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.artifactDao = artifactDao.New()
	suite.projectDao = projectDao.New()
	suite.userDao = userDao.New()
	suite.repositoryDao = repoDao.New()
	suite.tagDao = tagDao.New()
	suite.scanDao = daoscan.New()
	suite.vulnDao = daoscan.NewVulnerabilityRecordDao()
	suite.labelDao = labelDao.New()
	suite.exportManager = NewManager()
	suite.setupTestData()
}

func (suite *ExportManagerSuite) TearDownSuite() {
	suite.clearTestData()
}

func (suite *ExportManagerSuite) SetupTest() {

}

func (suite *ExportManagerSuite) TearDownTest() {

}

func (suite *ExportManagerSuite) clearTestData() {

	for _, labelRefId := range suite.testDataId.labelRefId {
		err := suite.labelDao.DeleteReference(suite.Context(), labelRefId)
		suite.NoError(err)
	}

	// delete labels and label references
	for _, labelId := range suite.testDataId.labelId {
		err := suite.labelDao.Delete(suite.Context(), labelId)
		suite.NoError(err)
	}

	for _, artRefId := range suite.testDataId.artRefId {
		err := suite.artifactDao.DeleteReference(suite.Context(), artRefId)
		suite.NoError(err)
	}

	for _, repoId := range suite.testDataId.repositoryId {
		err := suite.repositoryDao.Delete(suite.Context(), repoId)
		suite.NoError(err)
	}

	for _, tagId := range suite.testDataId.tagId {
		err := suite.tagDao.Delete(suite.Context(), tagId)
		suite.NoError(err)
	}

	for _, artId := range suite.testDataId.artifactId {
		err := suite.artifactDao.Delete(suite.Context(), artId)
		suite.NoError(err)
	}

	err := scanner.DeleteRegistration(suite.Context(), RegistrationUUID)
	suite.NoError(err, "Error when cleaning up scanner registrations")

	suite.cleanUpAdditionalData(ReportUUD, RegistrationUUID)
}

func (suite *ExportManagerSuite) TestExport() {
	{
		data, err := suite.exportManager.Fetch(suite.Context(), Params{ArtifactIDs: []int64{1}})
		suite.NoError(err)
		suite.Equal(10, len(data))
		for _, datum := range data {
			suite.Equal("{\"CVSS\": {\"nvd\": {\"V2Score\": \"4.3\"}}}", datum.AdditionalData)
		}
	}
}

func (suite *ExportManagerSuite) TestExportWithCVEFilter() {
	{
		p := Params{
			ArtifactIDs: []int64{1},
			CVEIds:      "CVE-ID2",
		}
		data, err := suite.exportManager.Fetch(suite.Context(), p)
		suite.NoError(err)
		suite.Equal(1, len(data))
		suite.Equal(p.CVEIds, data[0].CVEId)
	}
}

func (suite *ExportManagerSuite) registerScanner(registrationUUID string) {
	r := &scanner.Registration{
		UUID:        registrationUUID,
		Name:        registrationUUID,
		Description: "sample registration",
		URL:         fmt.Sprintf("https://sample.scanner.com/%s", registrationUUID),
	}

	_, err := scanner.AddRegistration(suite.Context(), r)
	suite.NoError(err, "add new registration")
}

func (suite *ExportManagerSuite) generateVulnerabilityRecordsForReport(registrationUUID string, numRecords int) []*daoscan.VulnerabilityRecord {
	vulns := make([]*daoscan.VulnerabilityRecord, 0)
	for i := 1; i <= numRecords; i++ {
		vulnV2 := new(daoscan.VulnerabilityRecord)
		vulnV2.CVEID = fmt.Sprintf("CVE-ID%d", i)
		vulnV2.Package = fmt.Sprintf("Package%d", i)
		vulnV2.PackageVersion = "Package-0.9.0"
		vulnV2.PackageType = "Unknown"
		vulnV2.Fix = "1.0.0"
		vulnV2.URLs = "url1"
		vulnV2.RegistrationUUID = registrationUUID
		if i%2 == 0 {
			vulnV2.Severity = "High"
		} else if i%3 == 0 {
			vulnV2.Severity = "Medium"
		} else if i%4 == 0 {
			vulnV2.Severity = "Critical"
		} else {
			vulnV2.Severity = "Low"
		}
		var vendorAttributes = make(map[string]interface{})
		vendorAttributes["CVSS"] = map[string]interface{}{"nvd": map[string]interface{}{"V2Score": "4.3"}}
		data, _ := json.Marshal(vendorAttributes)
		vulnV2.VendorAttributes = string(data)
		vulns = append(vulns, vulnV2)
	}

	return vulns
}

func (suite *ExportManagerSuite) insertVulnRecordForReport(reportUUID string, vr *daoscan.VulnerabilityRecord) {
	id, err := suite.vulnDao.Create(suite.Context(), vr)
	suite.NoError(err, "Failed to create vulnerability record")
	suite.testDataId.vulnRecs = append(suite.testDataId.vulnRecs, id)

	err = suite.vulnDao.InsertForReport(suite.Context(), reportUUID, id)
	suite.NoError(err, "Failed to insert vulnerability record row for report %s", reportUUID)
}

func (suite *ExportManagerSuite) cleanUpAdditionalData(reportID string, scannerID string) {
	_, err := suite.scanDao.DeleteMany(suite.Context(), q.Query{Keywords: q.KeyWords{"uuid": reportID}})

	suite.NoError(err)
	_, err = suite.vulnDao.DeleteForReport(suite.Context(), reportID)
	suite.NoError(err, "Failed to cleanup records")
	_, err = suite.vulnDao.DeleteForScanner(suite.Context(), scannerID)
	suite.NoError(err, "Failed to delete vulnerability records")
}

func (suite *ExportManagerSuite) setupTestData() {
	// create repositories
	repoRecord := &model.RepoRecord{
		Name:         "library/ubuntu",
		ProjectID:    1,
		Description:  "",
		PullCount:    1,
		StarCount:    0,
		CreationTime: time.Time{},
		UpdateTime:   time.Time{},
	}
	repoId, err := suite.repositoryDao.Create(suite.Context(), repoRecord)
	suite.NoError(err)
	suite.testDataId.repositoryId = append(suite.testDataId.repositoryId, repoId)

	// create artifacts for repositories
	art := &artifactDao.Artifact{
		ID:                1,
		Type:              "IMAGE",
		MediaType:         "application/vnd.docker.container.image.v1+json",
		ManifestMediaType: "application/vnd.docker.distribution.manifest.v2+json",
		ProjectID:         1,
		RepositoryID:      repoId,
		RepositoryName:    "library/ubuntu",
		Digest:            "sha256:e3d7ff9efd8431d9ef39a144c45992df5502c995b9ba3c53ff70c5b52a848d9c",
		Size:              28573056,
		Icon:              "",
		PushTime:          time.Time{},
		PullTime:          time.Time{}.Add(-10 * time.Minute),
		ExtraAttrs:        `{"architecture":"amd64","author":"","config":{"Env":["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"],"Cmd":["/bin/bash"]},"created":"2021-03-04T02:24:42.927713926Z","os":"linux"}`,
		Annotations:       "",
	}
	artId, err := suite.artifactDao.Create(suite.Context(), art)
	suite.NoError(err)
	suite.testDataId.artifactId = append(suite.testDataId.artifactId, artId)

	// create a tag and associate with the repository
	t := &tag.Tag{
		RepositoryID: repoId,
		ArtifactID:   artId,
		Name:         "latest",
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	tagId, err := suite.tagDao.Create(suite.Context(), t)
	suite.NoError(err)
	suite.testDataId.tagId = append(suite.testDataId.tagId, tagId)

	// create an artifact reference
	artReference := &artifactDao.ArtifactReference{
		ParentID:    artId,
		ChildID:     artId,
		ChildDigest: "sha256:e3d7ff9efd8431d9ef39a144c45992df5502c995b9ba3c53ff70c5b52a848d9c",
		Platform:    `{"architecture":"amd64","os":"linux"}`,
		URLs:        "",
		Annotations: "",
	}
	artRefId, err := suite.artifactDao.CreateReference(suite.Context(), artReference)
	suite.NoError(err)
	suite.testDataId.artRefId = append(suite.testDataId.artRefId, artRefId)

	// create a label
	l := labelModel.Label{
		Name:         "TestLabel",
		Description:  "",
		Color:        "Green",
		Level:        "",
		Scope:        "",
		ProjectID:    1,
		CreationTime: time.Time{},
		UpdateTime:   time.Time{},
		Deleted:      false,
	}
	labelId, err := suite.labelDao.Create(suite.Context(), &l)
	suite.NoError(err)
	suite.testDataId.labelId = append(suite.testDataId.labelId, labelId)

	lRef := labelModel.Reference{
		ID:           0,
		LabelID:      labelId,
		ArtifactID:   artId,
		CreationTime: time.Time{},
		UpdateTime:   time.Time{},
	}
	lRefId, err := suite.labelDao.CreateReference(suite.Context(), &lRef)
	suite.NoError(err)
	suite.testDataId.labelRefId = append(suite.testDataId.labelRefId, lRefId)

	// register a scanner
	suite.registerScanner(RegistrationUUID)

	// create a vulnerability scan report
	r := &daoscan.Report{
		UUID:             ReportUUD,
		Digest:           "sha256:e3d7ff9efd8431d9ef39a144c45992df5502c995b9ba3c53ff70c5b52a848d9c",
		RegistrationUUID: RegistrationUUID,
		MimeType:         v1.MimeTypeGenericVulnerabilityReport,
		Status:           job.PendingStatus.String(),
		Report:           "",
	}
	reportId, err := suite.scanDao.Create(suite.Context(), r)
	suite.NoError(err)
	suite.testDataId.reportId = append(suite.testDataId.reportId, reportId)

	// generate vulnerability records for the report
	vulns := suite.generateVulnerabilityRecordsForReport(RegistrationUUID, 10)
	suite.NotEmpty(vulns)

	for _, vuln := range vulns {
		suite.insertVulnRecordForReport(ReportUUD, vuln)
	}
}

func TestExportManager(t *testing.T) {
	suite.Run(t, &ExportManagerSuite{})
}
