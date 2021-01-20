package exporter

import (
	"context"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	proctl "github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/stretchr/testify/suite"
)

type TestArtifactCollectorSuite struct {
	suite.Suite

	dbState struct {
		testPro1 models.Project
		testPro2 models.Project

		repo1 models.RepoRecord
		repo2 models.RepoRecord

		art1 artifact.Artifact
		art2 artifact.Artifact

		vulnScanner1 scanner.Registration

		vuln1 scan.VulnerabilityRecord
		vuln2 scan.VulnerabilityRecord
		vuln3 scan.VulnerabilityRecord

		vulnReport1 scan.Report // Replaced with vulnReport2 - the new record is in the DB.
		vulnReport2 scan.Report
		vulnReport3 scan.Report

		vulnRecord1 scan.ReportVulnerabilityRecord
		vulnRecord2 scan.ReportVulnerabilityRecord
		vulnRecord3 scan.ReportVulnerabilityRecord
		vulnRecord4 scan.ReportVulnerabilityRecord
		vulnRecord5 scan.ReportVulnerabilityRecord
	}

	dbContext context.Context
}

func (t *TestArtifactCollectorSuite) SetupTest() {
	// Projects.
	t.dbState.testPro1 = models.Project{OwnerID: 1, Name: "test1", Metadata: map[string]string{"public": "true"}}
	t.dbState.testPro2 = models.Project{OwnerID: 1, Name: "test2", Metadata: map[string]string{"public": "false"}}

	// Repositories.
	t.dbState.repo1 = models.RepoRecord{Name: "repo1"}
	t.dbState.repo2 = models.RepoRecord{Name: "repo2"}

	// Artifacts.
	t.dbState.art1 = artifact.Artifact{
		RepositoryName: repo1.Name,
		Type:           "IMAGE",
		Digest:         "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
		Size:           100,
	}
	t.dbState.art2 = artifact.Artifact{
		RepositoryName: repo2.Name,
		Type:           "IMAGE",
		Digest:         "sha256:3198b18471892718923712837192831287312893712893712897312db1a3bc73",
		Size:           101,
	}

	// Scanners
	t.dbState.vulnScanner1 = scanner.Registration{
		UUID:       "98777f18-e096-4953-98f2-1adc77eed4d8",
		Name:       "scanner-1",
		IsDefault:  true,
		CreateTime: time.Now(),
	}

	// Vulnerabilities.
	t.dbState.vuln1 = scan.VulnerabilityRecord{
		CVEID:            "CVE-001",
		Severity:         "High",
		Fix:              "v0.0.2",
		PackageVersion:   "v0.0.1",
		Package:          "vuln-package1",
		RegistrationUUID: "98777f18-e096-4953-98f2-1adc77eed4d8",
	}
	t.dbState.vuln2 = scan.VulnerabilityRecord{
		CVEID:            "CVE-002",
		Severity:         "Medium",
		Fix:              "v0.0.2",
		PackageVersion:   "v0.0.1",
		Package:          "vuln-package2",
		RegistrationUUID: "98777f18-e096-4953-98f2-1adc77eed4d8",
	}
	t.dbState.vuln3 = scan.VulnerabilityRecord{
		CVEID:            "CVE-003",
		Severity:         "Low",
		Fix:              "",
		PackageVersion:   "v0.0.1",
		Package:          "vuln-package3",
		RegistrationUUID: "98777f18-e096-4953-98f2-1adc77eed4d8",
	}

	// Vulnerabilities Reports.
	// vulnReport1 - replaced by vulnReport2.
	t.dbState.vulnReport1 = scan.Report{
		UUID:             "ab1ff7ee-5cfd-492a-925c-ed78d0c9829b",
		RegistrationUUID: "7aa2e74a-6eeb-4e09-8325-f70d64bf7ded",
		Digest:           "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180", // art1.
		Report:           "{}",                                                                      // Left empty because It is not used for now.
		Status:           "Success",
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(time.Minute),
		MimeType:         "scanner/scanner-1",
	}

	// vulnReport2 - replaces vulnReport1.
	// The fact of replacing is emulated by adding vulnReport1 and then vulnReport2
	// so that the vulnReport2 ID is expected to be larger than vulnReport1.
	t.dbState.vulnReport2 = scan.Report{
		UUID:             "12850564-e4cb-4400-8504-6a54af079f6b",
		RegistrationUUID: "98777f18-e096-4953-98f2-1adc77eed4d8",
		Digest:           "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180", // art1.
		Report:           "{}",
		Status:           "Success",
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(time.Minute),
		MimeType:         "scanner/scanner-2",
	}

	// Failed scan.
	t.dbState.vulnReport3 = scan.Report{
		UUID:             "5ada708c-cf09-448b-9c79-6f7ffffbc16f",
		RegistrationUUID: "98777f18-e096-4953-98f2-1adc77eed4d8",
		Digest:           "sha256:3198b18471892718923712837192831287312893712893712897312db1a3bc73", // art2.
		Report:           "{}",
		Status:           "Failure", // To emulate failed scan.
		StartTime:        time.Now(),
		EndTime:          time.Now().Add(time.Minute),
		MimeType:         "scanner/scanner-1",
	}

	initDBOnce()

	ctx := orm.NewContext(context.Background(), dao.GetOrmer())
	t.dbContext = ctx

	// Projects.
	if _, err := proctl.Ctl.Create(ctx, &t.dbState.testPro1); err != nil {
		t.FailNow("project creating", err)
	}
	if _, err := proctl.Ctl.Create(ctx, &t.dbState.testPro2); err != nil {
		t.FailNow("project creating", err)
	}

	// Add repo to project.
	t.dbState.repo1.ProjectID = t.dbState.testPro1.ProjectID
	repo1ID, err := repository.Mgr.Create(ctx, &t.dbState.repo1)
	if err != nil {
		t.FailNow("add repo error", err)
	}
	repo1.RepositoryID = repo1ID

	t.dbState.repo2.ProjectID = t.dbState.testPro2.ProjectID
	repo2ID, err := repository.Mgr.Create(ctx, &t.dbState.repo2)
	t.dbState.repo2.RepositoryID = repo2ID
	if err != nil {
		t.FailNow("add repo error", err)
	}

	// Add artifacts.
	t.dbState.art1.ProjectID = t.dbState.testPro1.ProjectID
	t.dbState.art1.RepositoryID = repo1ID
	t.dbState.art1.PushTime = time.Now()
	_, err = artifact.Mgr.Create(ctx, &t.dbState.art1)
	if err != nil {
		t.FailNow("add repo error", err)
	}

	t.dbState.art2.ProjectID = t.dbState.testPro2.ProjectID
	t.dbState.art2.RepositoryID = repo2ID
	t.dbState.art2.PushTime = time.Now()
	_, err = artifact.Mgr.Create(ctx, &t.dbState.art2)
	if err != nil {
		t.FailNow("add repo error", err)
	}

	// Scanners.
	_, err = scanner.AddRegistration(&t.dbState.vulnScanner1)
	if err != nil {
		t.FailNow("add vulnerability scanner error", err)
	}

	// Vulnerabilities.
	_, err = scan.NewVulnerabilityRecordDao().Create(ctx, &t.dbState.vuln1)
	if err != nil {
		t.FailNow("add vulnerability error", err)
	}
	_, err = scan.NewVulnerabilityRecordDao().Create(ctx, &t.dbState.vuln2)
	if err != nil {
		t.FailNow("add vulnerability error", err)
	}
	_, err = scan.NewVulnerabilityRecordDao().Create(ctx, &t.dbState.vuln3)
	if err != nil {
		t.FailNow("add vulnerability error", err)
	}

	// Vulnerability Reports.
	_, err = scan.New().Create(ctx, &t.dbState.vulnReport1)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}
	_, err = scan.New().Create(ctx, &t.dbState.vulnReport2)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}
	_, err = scan.New().Create(ctx, &t.dbState.vulnReport3)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}

	// Vuln Report Records.
	vRepDAO := scan.NewVulnerabilityRecordDao()
	// Report 1.
	_, err = vRepDAO.InsertForReport(ctx, t.dbState.vulnReport1.UUID, &t.dbState.vuln1)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}
	_, err = vRepDAO.InsertForReport(ctx, t.dbState.vulnReport1.UUID, &t.dbState.vuln2)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}
	// Report 2.
	_, err = vRepDAO.InsertForReport(ctx, t.dbState.vulnReport2.UUID, &t.dbState.vuln1)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}
	// Report 3.
	_, err = vRepDAO.InsertForReport(ctx, t.dbState.vulnReport3.UUID, &t.dbState.vuln1)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}
	_, err = vRepDAO.InsertForReport(ctx, t.dbState.vulnReport3.UUID, &t.dbState.vuln2)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}
	_, err = vRepDAO.InsertForReport(ctx, t.dbState.vulnReport3.UUID, &t.dbState.vuln3)
	if err != nil {
		t.FailNow("add vulnerability scan report error", err)
	}
}

func (t *TestArtifactCollectorSuite) TearDownTest() {
	ormer, err := orm.FromContext(t.dbContext)
	if err != nil {
		t.FailNow("can not get DB from dbContext")
	}

	ormer.Raw("delete from project_metadata where project_id in (?, ?)", []int64{t.dbState.testPro1.ProjectID, t.dbState.testPro2.ProjectID}).Exec()
	ormer.Raw("delete from project where project_id in (?, ?)", []int64{t.dbState.testPro1.ProjectID, t.dbState.testPro2.ProjectID}).Exec()
	ormer.Raw("delete from artifact where project_id in (?, ?)", []int64{t.dbState.testPro1.ProjectID, t.dbState.testPro2.ProjectID}).Exec()
	ormer.Raw("delete from repository where project_id in (?, ?)", []int64{t.dbState.testPro1.ProjectID, t.dbState.testPro2.ProjectID}).Exec()
	ormer.Raw("delete from scan_report where id in (?, ?, ?)", []int64{t.dbState.vulnReport1.ID, t.dbState.vulnReport2.ID, t.dbState.vulnReport3.ID}).Exec()
	ormer.Raw("delete from scanner_registration where id in (?)", []int64{t.dbState.vulnScanner1.ID}).Exec()
	ormer.Raw("delete from vulnerability_record where id in (?, ?, ?)", []int64{t.dbState.vuln1.ID, t.dbState.vuln2.ID, t.dbState.vuln3.ID}).Exec()
	ormer.Raw("delete from report_vulnerability_record where report_uuid in (?, ?, ?)", []string{t.dbState.vulnReport1.UUID, t.dbState.vulnReport2.UUID, t.dbState.vulnReport3.UUID}).Exec()
}

func (t *TestArtifactCollectorSuite) TestVulnerabilitiesMetrics() {
	reportRecords, err := getVulnerabilitiesStats()

	if !t.NoError(err) {
		t.FailNow("error getting data")
	}

	// Prepare result.
	rMap := make(map[string]artifactReportRecord)

	for i := range reportRecords {
		rMap[reportRecords[i].ReportUUID] = reportRecords[i]
	}

	// Assert.
	// vulnReport1 must not appear in the stats.
	_, ok := rMap[t.dbState.vulnReport1.UUID]
	if ok {
		t.FailNow("vulnReport1 must not appear in the result")
	}

	// vulnReport2.
	vulnReport2Stats, ok := rMap[t.dbState.vulnReport2.UUID]
	if !ok {
		t.FailNowf("can not find statistics for report ", "%q", t.dbState.vulnReport2.UUID)
	}

	// Checks.
	t.EqualValues(t.dbState.art1.Digest, vulnReport2Stats.ArtifactDigest, "Incorrect Artifact 1 in Vulnerabilities Report 2")
	t.EqualValues(t.dbState.art1.Size, vulnReport2Stats.ArtifactSize, "Incorrect Artifact 1 size")
	t.EqualValues(t.dbState.testPro1.Name, vulnReport2Stats.ProjectName, "Artifact 1 must be in t.dbState.testPro1")
	t.EqualValues(t.dbState.vulnReport2.UUID, vulnReport2Stats.ReportUUID, "Artifact 1: possibly wrong report loaded from DB")
	t.EqualValues(t.dbState.art1.RepositoryName, vulnReport2Stats.RepositoryName, "Artifact 1: must be in repository t.dbState.repo1")

	if !t.NotNil(vulnReport2Stats.Vulnerabilities) {
		t.FailNow("Vulnerabilities statistics is nil - not computed")
	}

	t.EqualValues(1, vulnReport2Stats.Vulnerabilities.High, "Artifact 1: Must be one CVE-001")
	t.EqualValues(0, vulnReport2Stats.Vulnerabilities.Medium, "Artifact 1: Must not be vulnerabilities except one High CVE-001")
	t.EqualValues(0, vulnReport2Stats.Vulnerabilities.Low, "Artifact 1: Must not be vulnerabilities except one High CVE-001")
	t.EqualValues(1, vulnReport2Stats.Vulnerabilities.FixableHigh, "Artifact 1: High CVE-001 must be 'fixable' - has 'Fix' version")
	t.EqualValues(0, vulnReport2Stats.Vulnerabilities.FixableMedium, "Artifact 1: Must not be vulnerabilities except one High CVE-001")
	t.EqualValues(0, vulnReport2Stats.Vulnerabilities.FixableLow, "Artifact 1: Must not be vulnerabilities except one High CVE-001")

	// vulnReport3.
	vulnReport3Stats, ok := rMap[t.dbState.vulnReport3.UUID]
	if !ok {
		t.FailNowf("can not find statistics for report ", "%q", t.dbState.vulnReport3.UUID)
	}

	// Checks.
	t.EqualValues(t.dbState.art2.Digest, vulnReport3Stats.ArtifactDigest, "Incorrect Artifact 2 in Vulnerabilities Report 2")
	t.EqualValues(t.dbState.art2.Size, vulnReport3Stats.ArtifactSize, "Incorrect Artifact 2 size")
	t.EqualValues(t.dbState.testPro2.Name, vulnReport3Stats.ProjectName, "Artifact 2 must be in t.dbState.testPro1")
	t.EqualValues(t.dbState.vulnReport3.UUID, vulnReport3Stats.ReportUUID, "Artifact 2: possibly wrong report loaded from DB")
	t.EqualValues(t.dbState.art2.RepositoryName, vulnReport3Stats.RepositoryName, "Artifact 2: must be in repository t.dbState.repo1")

	if !t.NotNil(vulnReport3Stats.Vulnerabilities) {
		t.FailNow("Vulnerabilities statistics is nil - not computed")
	}

	t.EqualValues(1, vulnReport3Stats.Vulnerabilities.High, "Artifact 2: Must be one High CVE-001")
	t.EqualValues(1, vulnReport3Stats.Vulnerabilities.Medium, "Artifact 2: Must be one Medium CVE-002")
	t.EqualValues(1, vulnReport3Stats.Vulnerabilities.Low, "Artifact 2: Must be one Low CVE-003")
	t.EqualValues(1, vulnReport3Stats.Vulnerabilities.FixableHigh, "Artifact 2: High CVE-001 must be 'fixable' - has 'Fix' version")
	t.EqualValues(1, vulnReport3Stats.Vulnerabilities.FixableMedium, "Artifact 2: Medium CVE-002 must be 'fixable' - has 'Fix' version")
	t.EqualValues(0, vulnReport3Stats.Vulnerabilities.FixableLow, "Artifact 2: Low CVE-003 must NOT be 'fixable' - has no 'Fix' version")

}

func TestArtifactCollector(t *testing.T) {
	suite.Run(t, &TestArtifactCollectorSuite{})
}
