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

package util

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	notarytest "github.com/goharbor/harbor/src/common/utils/notary/test"
	testutils "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/quota"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var endpoint = "10.117.4.142"
var notaryServer *httptest.Server

const testingRedisHost = "REDIS_HOST"

var admiralEndpoint = "http://127.0.0.1:8282"
var token = ""

func TestMain(m *testing.M) {
	testutils.InitDatabaseFromEnv()
	notaryServer = notarytest.NewNotaryServer(endpoint)
	defer notaryServer.Close()
	var defaultConfig = map[string]interface{}{
		common.ExtEndpoint:     "https://" + endpoint,
		common.WithNotary:      true,
		common.TokenExpiration: 30,
	}
	config.InitWithSettings(defaultConfig)
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestMatchPullManifest(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("POST", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	res1, _, _ := MatchPullManifest(req1)
	assert.False(res1, "%s %v is not a request to pull manifest", req1.Method, req1.URL)

	req2, _ := http.NewRequest("GET", "http://192.168.0.3:80/v2/library/ubuntu/manifests/14.04", nil)
	res2, repo2, tag2 := MatchPullManifest(req2)
	assert.True(res2, "%s %v is a request to pull manifest", req2.Method, req2.URL)
	assert.Equal("library/ubuntu", repo2)
	assert.Equal("14.04", tag2)

	req3, _ := http.NewRequest("GET", "https://192.168.0.5:443/v1/library/ubuntu/manifests/14.04", nil)
	res3, _, _ := MatchPullManifest(req3)
	assert.False(res3, "%s %v is not a request to pull manifest", req3.Method, req3.URL)

	req4, _ := http.NewRequest("GET", "https://192.168.0.5/v2/library/ubuntu/manifests/14.04", nil)
	res4, repo4, tag4 := MatchPullManifest(req4)
	assert.True(res4, "%s %v is a request to pull manifest", req4.Method, req4.URL)
	assert.Equal("library/ubuntu", repo4)
	assert.Equal("14.04", tag4)

	req5, _ := http.NewRequest("GET", "https://myregistry.com/v2/path1/path2/golang/manifests/1.6.2", nil)
	res5, repo5, tag5 := MatchPullManifest(req5)
	assert.True(res5, "%s %v is a request to pull manifest", req5.Method, req5.URL)
	assert.Equal("path1/path2/golang", repo5)
	assert.Equal("1.6.2", tag5)

	req6, _ := http.NewRequest("GET", "https://myregistry.com/v2/myproject/registry/manifests/sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", nil)
	res6, repo6, tag6 := MatchPullManifest(req6)
	assert.True(res6, "%s %v is a request to pull manifest", req6.Method, req6.URL)
	assert.Equal("myproject/registry", repo6)
	assert.Equal("sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", tag6)

	req7, _ := http.NewRequest("GET", "https://myregistry.com/v2/myproject/manifests/sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", nil)
	res7, repo7, tag7 := MatchPullManifest(req7)
	assert.True(res7, "%s %v is a request to pull manifest", req7.Method, req7.URL)
	assert.Equal("myproject", repo7)
	assert.Equal("sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", tag7)
}

func TestMatchPutBlob(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/blobs/uploads/67bb4d9b-4dab-4bbe-b726-2e39322b8303?_state=7W3kWkgdr3fTW", nil)
	res1, repo1 := MatchPutBlobURL(req1)
	assert.True(res1, "%s %v is not a request to put blob", req1.Method, req1.URL)
	assert.Equal("library/ubuntu", repo1)

	req2, _ := http.NewRequest("PATCH", "http://127.0.0.1:5000/v2/library/blobs/uploads/67bb4d9b-4dab-4bbe-b726-2e39322b8303?_state=7W3kWkgdr3fTW", nil)
	res2, _ := MatchPutBlobURL(req2)
	assert.False(res2, "%s %v is a request to put blob", req2.Method, req2.URL)

	req3, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/manifest/67bb4d9b-4dab-4bbe-b726-2e39322b8303?_state=7W3kWkgdr3fTW", nil)
	res3, _ := MatchPutBlobURL(req3)
	assert.False(res3, "%s %v is not a request to put blob", req3.Method, req3.URL)
}

func TestMatchMountBlobURL(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("POST", "http://127.0.0.1:5000/v2/library/ubuntu/blobs/uploads/?mount=digtest123&from=testrepo", nil)
	res1, repo1, mount, from := MatchMountBlobURL(req1)
	assert.True(res1, "%s %v is not a request to mount blob", req1.Method, req1.URL)
	assert.Equal("library/ubuntu", repo1)
	assert.Equal("digtest123", mount)
	assert.Equal("testrepo", from)

	req2, _ := http.NewRequest("PATCH", "http://127.0.0.1:5000/v2/library/ubuntu/blobs/uploads/?mount=digtest123&from=testrepo", nil)
	res2, _, _, _ := MatchMountBlobURL(req2)
	assert.False(res2, "%s %v is a request to mount blob", req2.Method, req2.URL)

	req3, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/blobs/uploads/?mount=digtest123&from=testrepo", nil)
	res3, _, _, _ := MatchMountBlobURL(req3)
	assert.False(res3, "%s %v is not a request to put blob", req3.Method, req3.URL)
}

func TestPatchBlobURL(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("PATCH", "http://127.0.0.1:5000/v2/library/ubuntu/blobs/uploads/1234-1234-abcd", nil)
	res1, repo1 := MatchPatchBlobURL(req1)
	assert.True(res1, "%s %v is not a request to patch blob", req1.Method, req1.URL)
	assert.Equal("library/ubuntu", repo1)

	req2, _ := http.NewRequest("POST", "http://127.0.0.1:5000/v2/library/ubuntu/blobs/uploads/1234-1234-abcd", nil)
	res2, _ := MatchPatchBlobURL(req2)
	assert.False(res2, "%s %v is a request to patch blob", req2.Method, req2.URL)

	req3, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/blobs/uploads/?mount=digtest123&from=testrepo", nil)
	res3, _ := MatchPatchBlobURL(req3)
	assert.False(res3, "%s %v is not a request to patch blob", req3.Method, req3.URL)
}

func TestMatchPushManifest(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("POST", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	res1, _, _ := MatchPushManifest(req1)
	assert.False(res1, "%s %v is not a request to push manifest", req1.Method, req1.URL)

	req2, _ := http.NewRequest("PUT", "http://192.168.0.3:80/v2/library/ubuntu/manifests/14.04", nil)
	res2, repo2, tag2 := MatchPushManifest(req2)
	assert.True(res2, "%s %v is a request to push manifest", req2.Method, req2.URL)
	assert.Equal("library/ubuntu", repo2)
	assert.Equal("14.04", tag2)

	req3, _ := http.NewRequest("GET", "https://192.168.0.5:443/v1/library/ubuntu/manifests/14.04", nil)
	res3, _, _ := MatchPushManifest(req3)
	assert.False(res3, "%s %v is not a request to push manifest", req3.Method, req3.URL)

	req4, _ := http.NewRequest("PUT", "https://192.168.0.5/v2/library/ubuntu/manifests/14.04", nil)
	res4, repo4, tag4 := MatchPushManifest(req4)
	assert.True(res4, "%s %v is a request to push manifest", req4.Method, req4.URL)
	assert.Equal("library/ubuntu", repo4)
	assert.Equal("14.04", tag4)

	req5, _ := http.NewRequest("PUT", "https://myregistry.com/v2/path1/path2/golang/manifests/1.6.2", nil)
	res5, repo5, tag5 := MatchPushManifest(req5)
	assert.True(res5, "%s %v is a request to push manifest", req5.Method, req5.URL)
	assert.Equal("path1/path2/golang", repo5)
	assert.Equal("1.6.2", tag5)

	req6, _ := http.NewRequest("PUT", "https://myregistry.com/v2/myproject/registry/manifests/sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", nil)
	res6, repo6, tag6 := MatchPushManifest(req6)
	assert.True(res6, "%s %v is a request to push manifest", req6.Method, req6.URL)
	assert.Equal("myproject/registry", repo6)
	assert.Equal("sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", tag6)

	req7, _ := http.NewRequest("PUT", "https://myregistry.com/v2/myproject/manifests/sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", nil)
	res7, repo7, tag7 := MatchPushManifest(req7)
	assert.True(res7, "%s %v is a request to push manifest", req7.Method, req7.URL)
	assert.Equal("myproject", repo7)
	assert.Equal("sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", tag7)

	req8, _ := http.NewRequest("PUT", "http://192.168.0.3:80/v2/library/ubuntu/manifests/14.04", nil)
	res8, repo8, tag8 := MatchPushManifest(req8)
	assert.True(res8, "%s %v is a request to push manifest", req8.Method, req8.URL)
	assert.Equal("library/ubuntu", repo8)
	assert.Equal("14.04", tag8)
}

func TestPMSPolicyChecker(t *testing.T) {
	var defaultConfigAdmiral = map[string]interface{}{
		common.ExtEndpoint:        "https://" + endpoint,
		common.WithNotary:         true,
		common.TokenExpiration:    30,
		common.DatabaseType:       "postgresql",
		common.PostGreSQLHOST:     "127.0.0.1",
		common.PostGreSQLPort:     5432,
		common.PostGreSQLUsername: "postgres",
		common.PostGreSQLPassword: "root123",
		common.PostGreSQLDatabase: "registry",
	}

	if err := config.Init(); err != nil {
		panic(err)
	}

	config.Upload(defaultConfigAdmiral)

	name := "project_for_test_get_sev_low"
	id, err := config.GlobalProjectMgr.Create(&models.Project{
		Name:    name,
		OwnerID: 1,
		Metadata: map[string]string{
			models.ProMetaEnableContentTrust:   "true",
			models.ProMetaPreventVul:           "true",
			models.ProMetaSeverity:             "low",
			models.ProMetaReuseSysCVEWhitelist: "false",
		},
	})
	require.Nil(t, err)
	defer func(id int64) {
		if err := config.GlobalProjectMgr.Delete(id); err != nil {
			t.Logf("failed to delete project %d: %v", id, err)
		}
	}(id)

	contentTrustFlag := GetPolicyChecker().ContentTrustEnabled("project_for_test_get_sev_low")
	assert.True(t, contentTrustFlag)
	projectVulnerableEnabled, projectVulnerableSeverity, wl := GetPolicyChecker().VulnerablePolicy("project_for_test_get_sev_low")
	assert.True(t, projectVulnerableEnabled)
	assert.Equal(t, projectVulnerableSeverity, models.SevLow)
	assert.Empty(t, wl.Items)
}

func TestCopyResp(t *testing.T) {
	assert := assert.New(t)
	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	rec1.Header().Set("X-Test", "mytest")
	rec1.WriteHeader(418)
	CopyResp(rec1, rec2)
	assert.Equal(418, rec2.Result().StatusCode)
	assert.Equal("mytest", rec2.Header().Get("X-Test"))
}

func TestMarshalError(t *testing.T) {
	assert := assert.New(t)
	js1 := MarshalError("PROJECT_POLICY_VIOLATION", "Not Found")
	assert.Equal("{\"errors\":[{\"code\":\"PROJECT_POLICY_VIOLATION\",\"message\":\"Not Found\",\"detail\":\"Not Found\"}]}", js1)
	js2 := MarshalError("DENIED", "The action is denied")
	assert.Equal("{\"errors\":[{\"code\":\"DENIED\",\"message\":\"The action is denied\",\"detail\":\"The action is denied\"}]}", js2)
}

func TestTryRequireQuota(t *testing.T) {
	quotaRes := &quota.ResourceList{
		quota.ResourceStorage: 100,
	}
	err := TryRequireQuota(1, quotaRes)
	assert.Nil(t, err)
}

func TestTryFreeQuota(t *testing.T) {
	quotaRes := &quota.ResourceList{
		quota.ResourceStorage: 1,
	}
	success := TryFreeQuota(1, quotaRes)
	assert.True(t, success)
}

func TestGetBlobSize(t *testing.T) {
	con, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", getRedisHost(), 6379),
		redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(time.Minute+10*time.Second),
		redis.DialWriteTimeout(10*time.Second),
	)
	assert.Nil(t, err)
	defer con.Close()

	size, err := GetBlobSize(con, "test-TestGetBlobSize")
	assert.Nil(t, err)
	assert.Equal(t, size, int64(0))
}

func TestSetBunkSize(t *testing.T) {
	con, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", getRedisHost(), 6379),
		redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(time.Minute+10*time.Second),
		redis.DialWriteTimeout(10*time.Second),
	)
	assert.Nil(t, err)
	defer con.Close()

	size, err := GetBlobSize(con, "TestSetBunkSize")
	assert.Nil(t, err)
	assert.Equal(t, size, int64(0))

	_, err = SetBunkSize(con, "TestSetBunkSize", 123)
	assert.Nil(t, err)

	size1, err := GetBlobSize(con, "TestSetBunkSize")
	assert.Nil(t, err)
	assert.Equal(t, size1, int64(123))
}

func TestGetProjectID(t *testing.T) {
	name := "project_for_TestGetProjectID"
	project := models.Project{
		OwnerID: 1,
		Name:    name,
	}

	id, err := dao.AddProject(project)
	if err != nil {
		t.Fatalf("failed to add project: %v", err)
	}

	idget, err := GetProjectID(name)
	assert.Nil(t, err)
	assert.Equal(t, id, idget)
}

func getRedisHost() string {
	redisHost := os.Getenv(testingRedisHost)
	if redisHost == "" {
		redisHost = "127.0.0.1" // for local test
	}

	return redisHost
}
