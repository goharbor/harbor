package exporter

import (
	"strconv"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	proctl "github.com/goharbor/harbor/src/controller/project"
	quotactl "github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/member"
	memberModels "github.com/goharbor/harbor/src/pkg/member/models"
	qtypes "github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/pkg/user"
)

var (
	alice    = models.User{Username: "alice", Password: "password", Email: "alice@test.com"}
	bob      = models.User{Username: "bob", Password: "password", Email: "bob@test.com"}
	eve      = models.User{Username: "eve", Password: "password", Email: "eve@test.com"}
	testPro1 = proModels.Project{OwnerID: 1, Name: "test1", Metadata: map[string]string{"public": "true"}}
	testPro2 = proModels.Project{OwnerID: 1, Name: "test2", Metadata: map[string]string{"public": "false"}}
	rs1      = qtypes.ResourceList{qtypes.ResourceStorage: 100}
	rs2      = qtypes.ResourceList{qtypes.ResourceStorage: 200}
	repo1    = model.RepoRecord{Name: "repo1"}
	repo2    = model.RepoRecord{Name: "repo2"}
	pmIDs    = []int{}
	art1     = artifact.Artifact{RepositoryName: repo1.Name, Type: "IMAGE", Digest: "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180"}
	art2     = artifact.Artifact{RepositoryName: repo1.Name, Type: "IMAGE", Digest: "sha256:3198b18471892718923712837192831287312893712893712897312db1a3bc73"}
)

func setupTest(t *testing.T) {
	test.InitDatabaseFromEnv()
	ctx := orm.Context()

	// register projAdmin and assign project admin role
	aliceID, err := user.Mgr.Create(ctx, &alice)
	if err != nil {
		t.Errorf("register user error %v", err)
	}
	bobID, err := user.Mgr.Create(ctx, &bob)
	if err != nil {
		t.Errorf("register user error %v", err)
	}
	eveID, err := user.Mgr.Create(ctx, &eve)
	if err != nil {
		t.Errorf("register user error %v", err)
	}

	// Create Project
	proID1, err := proctl.Ctl.Create(ctx, &testPro1)
	if err != nil {
		t.Errorf("project creating %v", err)
	}
	proID2, err := proctl.Ctl.Create(ctx, &testPro2)
	if err != nil {
		t.Errorf("project creating %v", err)
	}
	testPro1.ProjectID = proID1
	testPro2.ProjectID = proID2

	// Create quota for project
	quotactl.Ctl.Create(ctx, "project", strconv.Itoa(int(testPro1.ProjectID)), rs1)
	quotactl.Ctl.Create(ctx, "project", strconv.Itoa(int(testPro2.ProjectID)), rs2)
	if err != nil {
		t.Errorf("add project error %v", err)
	}

	// Add repo to project
	repo1.ProjectID = testPro1.ProjectID
	repo1ID, err := pkg.RepositoryMgr.Create(ctx, &repo1)
	if err != nil {
		t.Errorf("add repo error %v", err)
	}
	repo1.RepositoryID = repo1ID
	repo2.ProjectID = testPro2.ProjectID
	repo2ID, err := pkg.RepositoryMgr.Create(ctx, &repo2)
	repo2.RepositoryID = repo2ID
	if err != nil {
		t.Errorf("add repo error %v", err)
	}
	// Add artifacts
	art1.ProjectID = testPro1.ProjectID
	art1.RepositoryID = repo1ID
	art1.PushTime = time.Now()
	_, err = pkg.ArtifactMgr.Create(ctx, &art1)
	if err != nil {
		t.Errorf("add repo error %v", err)
	}

	art2.ProjectID = testPro2.ProjectID
	art2.RepositoryID = repo2ID
	art2.PushTime = time.Now()
	_, err = pkg.ArtifactMgr.Create(ctx, &art2)
	if err != nil {
		t.Errorf("add repo error %v", err)
	}
	// Add member to project
	pmIDs = make([]int, 0)
	alice.UserID, bob.UserID, eve.UserID = int(aliceID), int(bobID), int(eveID)

	p1m1ID, err := member.Mgr.AddProjectMember(ctx, memberModels.Member{ProjectID: proID1, Role: common.RoleDeveloper, EntityID: int(aliceID), EntityType: common.UserMember})
	if err != nil {
		t.Errorf("add project member error %v", err)
	}
	p2m1ID, err := member.Mgr.AddProjectMember(ctx, memberModels.Member{ProjectID: proID2, Role: common.RoleMaintainer, EntityID: int(bobID), EntityType: common.UserMember})
	if err != nil {
		t.Errorf("add project member error %v", err)
	}
	p2m2ID, err := member.Mgr.AddProjectMember(ctx, memberModels.Member{ProjectID: proID2, Role: common.RoleMaintainer, EntityID: int(eveID), EntityType: common.UserMember})

	if err != nil {
		t.Errorf("add project member error %v", err)
	}
	pmIDs = append(pmIDs, p1m1ID, p2m1ID, p2m2ID)
}

func tearDownTest(t *testing.T) {
	dao.GetOrmer().Raw("delete from project_member where project_id in (?, ?)", []int64{testPro1.ProjectID, testPro2.ProjectID}).Exec()
	dao.GetOrmer().Raw("delete from project_metadata where project_id in (?, ?)", []int64{testPro1.ProjectID, testPro2.ProjectID}).Exec()
	dao.GetOrmer().Raw("delete from quota where reference=\"project\" and reference_id in (?, ?)", []int64{testPro1.ProjectID, testPro2.ProjectID}).Exec()
	dao.GetOrmer().Raw("delete from quota_usage where reference=\"project\" and reference_id in (?, ?)", []int64{testPro1.ProjectID, testPro2.ProjectID}).Exec()
	dao.GetOrmer().Raw("delete from project where project_id in (?, ?)", []int64{testPro1.ProjectID, testPro2.ProjectID}).Exec()
	dao.GetOrmer().Raw("delete from artifact where project_id in (?, ?)", []int64{testPro1.ProjectID, testPro2.ProjectID}).Exec()
	dao.GetOrmer().Raw("delete from repository where project_id in (?, ?)", []int64{testPro1.ProjectID, testPro2.ProjectID}).Exec()
	dao.GetOrmer().Raw("delete from cve_allowlist where project_id in (?, ?)", []int64{testPro1.ProjectID, testPro2.ProjectID}).Exec()
	dao.GetOrmer().Raw("delete from harbor_user where user_id in (?, ?, ?)", []int{alice.UserID, bob.UserID, eve.UserID}).Exec()
}

type PorjectCollectorTestSuite struct {
	suite.Suite
}

func (c *PorjectCollectorTestSuite) TestProjectCollector() {
	pMap := make(map[int64]*projectInfo)
	updateProjectBasicInfo(pMap)
	updateProjectMemberInfo(pMap)
	updateProjectRepoInfo(pMap)
	updateProjectArtifactInfo(pMap)

	c.Equalf(testPro1.ProjectID, pMap[testPro1.ProjectID].ProjectID, "pMap %v", pMap)
	c.Equalf(pMap[testPro1.ProjectID].ProjectID, testPro1.ProjectID, "pMap %v", pMap)
	c.Equalf(pMap[testPro1.ProjectID].Name, testPro1.Name, "pMap %v", pMap)
	c.Equalf(strconv.FormatBool(pMap[testPro1.ProjectID].Public), testPro1.Metadata["public"], "pMap %v", pMap)
	c.Equalf(pMap[testPro1.ProjectID].Quota, "{\"storage\": 100}", "pMap %v", pMap)
	c.Equalf(pMap[testPro1.ProjectID].Usage, "{\"storage\": 0}", "pMap %v", pMap)
	c.Equalf(pMap[testPro1.ProjectID].MemberTotal, float64(2), "pMap %v", pMap)
	c.Equalf(pMap[testPro1.ProjectID].PullTotal, float64(0), "pMap %v", pMap)
	c.Equalf(pMap[testPro1.ProjectID].Artifact["IMAGE"].ArtifactTotal, float64(1), "pMap %v", pMap)
	c.Equalf(pMap[testPro1.ProjectID].Artifact["IMAGE"].ArtifactType, "IMAGE", "pMap %v", pMap)

	c.Equalf(pMap[testPro2.ProjectID].ProjectID, testPro2.ProjectID, "pMap %v", pMap)
	c.Equalf(pMap[testPro2.ProjectID].Name, testPro2.Name, "pMap %v", pMap)
	c.Equalf(strconv.FormatBool(pMap[testPro2.ProjectID].Public), testPro2.Metadata["public"], "pMap %v", pMap)
	c.Equalf(pMap[testPro2.ProjectID].Quota, "{\"storage\": 200}", "pMap %v", pMap)
	c.Equalf(pMap[testPro2.ProjectID].Usage, "{\"storage\": 0}", "pMap %v", pMap)
	c.Equalf(pMap[testPro2.ProjectID].MemberTotal, float64(3), "pMap %v", pMap)
	c.Equalf(pMap[testPro2.ProjectID].PullTotal, float64(0), "pMap %v", pMap)
	c.Equalf(pMap[testPro2.ProjectID].Artifact["IMAGE"].ArtifactTotal, float64(1), "pMap %v", pMap)

}

func TestPorjectCollectorTestSuite(t *testing.T) {
	setupTest(t)
	defer tearDownTest(t)
	suite.Run(t, new(PorjectCollectorTestSuite))
}
