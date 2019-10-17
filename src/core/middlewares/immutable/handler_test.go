package immutable

import (
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"

	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/immutabletag"
	immu_model "github.com/goharbor/harbor/src/pkg/immutabletag/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type HandlerSuite struct {
	suite.Suite
}

func doPutManifestRequest(projectID int64, projectName, name, tag, dgt string, next ...http.HandlerFunc) int {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/manifests/%s", repository, tag)
	req, _ := http.NewRequest("PUT", url, nil)

	mfInfo := &util.ManifestInfo{
		ProjectID:  projectID,
		Repository: repository,
		Tag:        tag,
		Digest:     dgt,
		References: []distribution.Descriptor{
			{Digest: digest.FromString(randomString(15))},
			{Digest: digest.FromString(randomString(15))},
		},
	}
	ctx := util.NewManifestInfoContext(req.Context(), mfInfo)
	rr := httptest.NewRecorder()

	var n http.HandlerFunc
	if len(next) > 0 {
		n = next[0]
	} else {
		n = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}
	}

	h := New(http.HandlerFunc(n))
	h.ServeHTTP(util.NewCustomResponseWriter(rr), req.WithContext(ctx))

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

func (suite *HandlerSuite) addProject(projectName string) int64 {
	projectID, err := dao.AddProject(models.Project{
		Name:    projectName,
		OwnerID: 1,
	})
	suite.Nil(err, fmt.Sprintf("Add project failed for %s", projectName))
	return projectID
}

func (suite *HandlerSuite) addArt(pid int64, repo string, tag string) int64 {
	afid, err := dao.AddArtifact(&models.Artifact{
		PID:    pid,
		Repo:   repo,
		Tag:    tag,
		Digest: digest.FromString(randomString(15)).String(),
		Kind:   "Docker-Image",
	})
	suite.Nil(err, fmt.Sprintf("Add artifact failed for %s", repo))
	return afid
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
	id, err := immutabletag.ImmuCtr.CreateImmutableRule(metadata)
	require.NoError(suite.T(), err, "nil error expected but got %s", err)
	return id
}

func (suite *HandlerSuite) TestPutManifestCreated() {
	projectName := randomString(5)

	projectID := suite.addProject(projectName)
	immuRuleID := suite.addImmutableRule(projectID)
	afID := suite.addArt(projectID, projectName+"/photon", "release-1.10")
	defer func() {
		dao.DeleteProject(projectID)
		dao.DeleteArtifact(afID)
		immutabletag.ImmuCtr.DeleteImmutableRule(immuRuleID)
	}()

	dgt := digest.FromString(randomString(15)).String()
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
