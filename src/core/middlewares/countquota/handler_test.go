// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package countquota

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getProjectCountUsage(projectID int64) (int64, error) {
	usage := models.QuotaUsage{Reference: "project", ReferenceID: fmt.Sprintf("%d", projectID)}
	err := dao.GetOrmer().Read(&usage, "reference", "reference_id")
	if err != nil {
		return 0, err
	}
	used, err := types.NewResourceList(usage.Used)
	if err != nil {
		return 0, err
	}

	return used[types.ResourceCount], nil
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyz"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}

	return string(b)
}

func doDeleteManifestRequest(projectID int64, projectName, name, dgt string, next ...http.HandlerFunc) int {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/manifests/%s", repository, dgt)
	req, _ := http.NewRequest("DELETE", url, nil)

	ctx := util.NewManifestInfoContext(req.Context(), &util.ManifestInfo{
		ProjectID:  projectID,
		Repository: repository,
		Digest:     dgt,
	})

	rr := httptest.NewRecorder()

	var n http.HandlerFunc
	if len(next) > 0 {
		n = next[0]
	} else {
		n = func(w http.ResponseWriter, req *http.Request) {
			w.WriteHeader(http.StatusAccepted)
		}
	}

	h := New(http.HandlerFunc(n))
	h.ServeHTTP(util.NewCustomResponseWriter(rr), req.WithContext(ctx))

	return rr.Code
}

func doPutManifestRequest(projectID int64, projectName, name, tag, dgt string, next ...http.HandlerFunc) int {
	repository := fmt.Sprintf("%s/%s", projectName, name)

	url := fmt.Sprintf("/v2/%s/manifests/%s", repository, tag)
	req, _ := http.NewRequest("PUT", url, nil)

	ctx := util.NewManifestInfoContext(req.Context(), &util.ManifestInfo{
		ProjectID:  projectID,
		Repository: repository,
		Tag:        tag,
		Digest:     dgt,
		References: []distribution.Descriptor{
			{Digest: digest.FromString(randomString(15))},
			{Digest: digest.FromString(randomString(15))},
		},
	})

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

type HandlerSuite struct {
	suite.Suite
}

func (suite *HandlerSuite) addProject(projectName string) int64 {
	projectID, err := dao.AddProject(models.Project{
		Name:    projectName,
		OwnerID: 1,
	})

	suite.Nil(err, fmt.Sprintf("Add project failed for %s", projectName))

	return projectID
}

func (suite *HandlerSuite) checkCountUsage(expected, projectID int64) {
	count, err := getProjectCountUsage(projectID)
	suite.Nil(err, fmt.Sprintf("Failed to get count usage of project %d, error: %v", projectID, err))
	suite.Equal(expected, count, "Failed to check count usage for project %d", projectID)
}

func (suite *HandlerSuite) TearDownTest() {
	for _, table := range []string{
		"artifact", "blob",
		"artifact_blob", "project_blob",
		"quota", "quota_usage",
	} {
		dao.ClearTable(table)
	}
}

func (suite *HandlerSuite) TestPutManifestCreated() {
	projectName := randomString(5)

	projectID := suite.addProject(projectName)
	defer func() {
		dao.DeleteProject(projectID)
	}()

	dgt := digest.FromString(randomString(15)).String()
	code := doPutManifestRequest(projectID, projectName, "photon", "latest", dgt)
	suite.Equal(http.StatusCreated, code)
	suite.checkCountUsage(1, projectID)

	total, err := dao.GetTotalOfArtifacts(&models.ArtifactQuery{Digest: dgt})
	suite.Nil(err)
	suite.Equal(int64(1), total, "Artifact should be created")

	// Push the photon:latest with photon:dev
	code = doPutManifestRequest(projectID, projectName, "photon", "dev", dgt)
	suite.Equal(http.StatusCreated, code)
	suite.checkCountUsage(2, projectID)

	total, err = dao.GetTotalOfArtifacts(&models.ArtifactQuery{Digest: dgt})
	suite.Nil(err)
	suite.Equal(int64(2), total, "Artifact should be created")

	// Push the photon:latest with new image
	newDgt := digest.FromString(randomString(15)).String()
	code = doPutManifestRequest(projectID, projectName, "photon", "latest", newDgt)
	suite.Equal(http.StatusCreated, code)
	suite.checkCountUsage(2, projectID)

	total, err = dao.GetTotalOfArtifacts(&models.ArtifactQuery{Digest: newDgt})
	suite.Nil(err)
	suite.Equal(int64(1), total, "Artifact should be updated")
}

func (suite *HandlerSuite) TestPutManifestFailed() {
	projectName := randomString(5)

	projectID := suite.addProject(projectName)
	defer func() {
		dao.DeleteProject(projectID)
	}()

	next := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	dgt := digest.FromString(randomString(15)).String()
	code := doPutManifestRequest(projectID, projectName, "photon", "latest", dgt, next)
	suite.Equal(http.StatusInternalServerError, code)
	suite.checkCountUsage(0, projectID)

	total, err := dao.GetTotalOfArtifacts(&models.ArtifactQuery{Digest: dgt})
	suite.Nil(err)
	suite.Equal(int64(0), total, "Artifact should not be created")
}

func (suite *HandlerSuite) TestDeleteManifestAccepted() {
	projectName := randomString(5)

	projectID := suite.addProject(projectName)
	defer func() {
		dao.DeleteProject(projectID)
	}()

	dgt := digest.FromString(randomString(15)).String()
	code := doPutManifestRequest(projectID, projectName, "photon", "latest", dgt)
	suite.Equal(http.StatusCreated, code)
	suite.checkCountUsage(1, projectID)

	code = doDeleteManifestRequest(projectID, projectName, "photon", dgt)
	suite.Equal(http.StatusAccepted, code)
	suite.checkCountUsage(0, projectID)
}

func (suite *HandlerSuite) TestDeleteManifestFailed() {
	projectName := randomString(5)

	projectID := suite.addProject(projectName)
	defer func() {
		dao.DeleteProject(projectID)
	}()

	dgt := digest.FromString(randomString(15)).String()
	code := doPutManifestRequest(projectID, projectName, "photon", "latest", dgt)
	suite.Equal(http.StatusCreated, code)
	suite.checkCountUsage(1, projectID)

	next := func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	code = doDeleteManifestRequest(projectID, projectName, "photon", dgt, next)
	suite.Equal(http.StatusInternalServerError, code)
	suite.checkCountUsage(1, projectID)
}

func (suite *HandlerSuite) TestDeleteManifestInMultiProjects() {
	projectName := randomString(5)

	projectID := suite.addProject(projectName)
	defer func() {
		dao.DeleteProject(projectID)
	}()

	dgt := digest.FromString(randomString(15)).String()
	code := doPutManifestRequest(projectID, projectName, "photon", "latest", dgt)
	suite.Equal(http.StatusCreated, code)
	suite.checkCountUsage(1, projectID)

	{
		projectName := randomString(5)

		projectID := suite.addProject(projectName)
		defer func() {
			dao.DeleteProject(projectID)
		}()

		code := doPutManifestRequest(projectID, projectName, "photon", "latest", dgt)
		suite.Equal(http.StatusCreated, code)
		suite.checkCountUsage(1, projectID)

		code = doDeleteManifestRequest(projectID, projectName, "photon", dgt)
		suite.Equal(http.StatusAccepted, code)
		suite.checkCountUsage(0, projectID)
	}

	code = doDeleteManifestRequest(projectID, projectName, "photon", dgt)
	suite.Equal(http.StatusAccepted, code)
	suite.checkCountUsage(0, projectID)
}

func TestMain(m *testing.M) {
	config.Init()
	dao.PrepareTestForPostgresSQL()

	if result := m.Run(); result != 0 {
		os.Exit(result)
	}
}

func TestRunHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}
