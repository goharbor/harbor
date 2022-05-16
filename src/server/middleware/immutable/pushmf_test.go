package immutable

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/pkg"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib"
	internal_orm "github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/artifact"
	immu_model "github.com/goharbor/harbor/src/pkg/immutable/model"
	"github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/pkg/tag"
	tag_model "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type HandlerSuite struct {
	suite.Suite
}

func doPutManifestRequest(projectID int64, projectName, name, tag, dgt string, next ...http.HandlerFunc) int {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/manifests/%s", repository, tag)
	req, _ := http.NewRequest("PUT", url, nil)

	afInfo := lib.ArtifactInfo{
		ProjectName: projectName,
		Repository:  repository,
		Tag:         tag,
		Digest:      dgt,
	}
	rr := httptest.NewRecorder()

	var n http.HandlerFunc
	if len(next) > 0 {
		n = next[0]
	} else {
		n = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}
	}
	*req = *(req.WithContext(internal_orm.NewContext(context.TODO(), dao.GetOrmer())))
	*req = *(req.WithContext(lib.WithArtifactInfo(req.Context(), afInfo)))
	h := Middleware()(n)
	h.ServeHTTP(rr, req)

	return rr.Code
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func (suite *HandlerSuite) addProject(ctx context.Context, projectName string) int64 {
	projectID, err := pkg.ProjectMgr.Create(ctx, &proModels.Project{
		Name:    projectName,
		OwnerID: 1,
	})
	suite.Nil(err, fmt.Sprintf("Add project failed for %s", projectName))
	return projectID
}

func (suite *HandlerSuite) addArt(ctx context.Context, pid, repositoryID int64, repositoryName, dgt string) int64 {
	af := &artifact.Artifact{
		Type:           "Docker-Image",
		ProjectID:      pid,
		RepositoryID:   repositoryID,
		RepositoryName: repositoryName,
		Digest:         dgt,
		Size:           1024,
		PushTime:       time.Now(),
		PullTime:       time.Now(),
	}
	afid, err := pkg.ArtifactMgr.Create(ctx, af)
	suite.Nil(err, fmt.Sprintf("Add artifact failed for %d", repositoryID))
	return afid
}

func (suite *HandlerSuite) addRepo(ctx context.Context, pid int64, repo string) int64 {
	repoRec := &model.RepoRecord{
		Name:      repo,
		ProjectID: pid,
	}
	repoid, err := pkg.RepositoryMgr.Create(ctx, repoRec)
	suite.Nil(err, fmt.Sprintf("Add repository failed for %s", repo))
	return repoid
}

func (suite *HandlerSuite) addTags(ctx context.Context, repoid int64, afid int64, name string) int64 {
	t := &tag_model.Tag{
		RepositoryID: repoid,
		ArtifactID:   afid,
		Name:         name,
		PushTime:     time.Time{},
		PullTime:     time.Time{},
	}
	tid, err := tag.Mgr.Create(ctx, t)
	suite.Nil(err, fmt.Sprintf("Add artifact failed for %s", name))
	return tid
}

func (suite *HandlerSuite) addImmutableRule(pid int64) int64 {
	metadata := &immu_model.Metadata{
		ProjectID: pid,
		Priority:  1,
		Action:    "immutable",
		Template:  "immutable_template",
		TagSelectors: []*immu_model.Selector{
			{
				Kind:       "doublestar",
				Decoration: "matches",
				Pattern:    "release-**",
			},
		},
		ScopeSelectors: map[string][]*immu_model.Selector{
			"repository": {
				{
					Kind:       "doublestar",
					Decoration: "repoMatches",
					Pattern:    "**",
				},
			},
		},
	}
	id, err := immutable.Ctr.CreateImmutableRule(internal_orm.Context(), metadata)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)
	return id
}

func (suite *HandlerSuite) TestPutDeleteManifestCreated() {
	projectName := randomString(5)
	repoName := projectName + "/photon"
	dgt := digest.FromString(randomString(15)).String()
	ctx := internal_orm.NewContext(context.TODO(), dao.GetOrmer())

	projectID := suite.addProject(ctx, projectName)
	immuRuleID := suite.addImmutableRule(projectID)
	repoID := suite.addRepo(ctx, projectID, repoName)
	afID := suite.addArt(ctx, projectID, repoID, repoName, dgt)
	tagID := suite.addTags(ctx, repoID, afID, "release-1.10")

	defer func() {
		pkg.ProjectMgr.Delete(ctx, projectID)
		pkg.ArtifactMgr.Delete(ctx, afID)
		pkg.RepositoryMgr.Delete(ctx, repoID)
		tag.Mgr.Delete(ctx, tagID)
		immutable.Ctr.DeleteImmutableRule(internal_orm.Context(), immuRuleID)
	}()

	code1 := doPutManifestRequest(projectID, projectName, "photon", "release-1.10", dgt)
	suite.Equal(http.StatusPreconditionFailed, code1)

	code2 := doPutManifestRequest(projectID, projectName, "photon", "latest", dgt)
	suite.Equal(http.StatusCreated, code2)

}

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()

	if result := m.Run(); result != 0 {
		os.Exit(result)
	}
}

func TestRunHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}
