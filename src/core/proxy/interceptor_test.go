package proxy

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	notarytest "github.com/goharbor/harbor/src/common/utils/notary/test"
	testutils "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var endpoint = "10.117.4.142"
var notaryServer *httptest.Server

var admiralEndpoint = "http://127.0.0.1:8282"
var token = ""

func TestMain(m *testing.M) {
	notaryServer = notarytest.NewNotaryServer(endpoint)
	defer notaryServer.Close()
	NotaryEndpoint = notaryServer.URL
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

func TestMatchListRepos(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("POST", "http://127.0.0.1:5000/v2/_catalog", nil)
	res1 := MatchListRepos(req1)
	assert.False(res1, "%s %v is not a request to list repos", req1.Method, req1.URL)

	req2, _ := http.NewRequest("GET", "http://127.0.0.1:5000/v2/_catalog", nil)
	res2 := MatchListRepos(req2)
	assert.True(res2, "%s %v is a request to list repos", req2.Method, req2.URL)

	req3, _ := http.NewRequest("GET", "https://192.168.0.5:443/v1/_catalog", nil)
	res3 := MatchListRepos(req3)
	assert.False(res3, "%s %v is not a request to pull manifest", req3.Method, req3.URL)

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
	testutils.InitDatabaseFromEnv()

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

	contentTrustFlag := getPolicyChecker().contentTrustEnabled("project_for_test_get_sev_low")
	assert.True(t, contentTrustFlag)
	projectVulnerableEnabled, projectVulnerableSeverity, wl := getPolicyChecker().vulnerablePolicy("project_for_test_get_sev_low")
	assert.True(t, projectVulnerableEnabled)
	assert.Equal(t, projectVulnerableSeverity, models.SevLow)
	assert.Empty(t, wl.Items)
}

func TestMatchNotaryDigest(t *testing.T) {
	assert := assert.New(t)
	// The data from common/utils/notary/helper_test.go
	img1 := imageInfo{"notary-demo/busybox", "1.0", "notary-demo", "sha256:1359608115b94599e5641638bac5aef1ddfaa79bb96057ebf41ebc8d33acf8a7"}
	img2 := imageInfo{"notary-demo/busybox", "2.0", "notary-demo", "sha256:12345678"}

	res1, err := matchNotaryDigest(img1)
	assert.Nil(err, "Unexpected error: %v, image: %#v", err, img1)
	assert.True(res1)

	res2, err := matchNotaryDigest(img2)
	assert.Nil(err, "Unexpected error: %v, image: %#v, take 2", err, img2)
	assert.False(res2)
}

func TestCopyResp(t *testing.T) {
	assert := assert.New(t)
	rec1 := httptest.NewRecorder()
	rec2 := httptest.NewRecorder()
	rec1.Header().Set("X-Test", "mytest")
	rec1.WriteHeader(418)
	copyResp(rec1, rec2)
	assert.Equal(418, rec2.Result().StatusCode)
	assert.Equal("mytest", rec2.Header().Get("X-Test"))
}

func TestMarshalError(t *testing.T) {
	assert := assert.New(t)
	js1 := marshalError("PROJECT_POLICY_VIOLATION", "Not Found")
	assert.Equal("{\"errors\":[{\"code\":\"PROJECT_POLICY_VIOLATION\",\"message\":\"Not Found\",\"detail\":\"Not Found\"}]}", js1)
	js2 := marshalError("DENIED", "The action is denied")
	assert.Equal("{\"errors\":[{\"code\":\"DENIED\",\"message\":\"The action is denied\",\"detail\":\"The action is denied\"}]}", js2)
}

func TestIsDigest(t *testing.T) {
	assert := assert.New(t)
	assert.False(isDigest("latest"))
	assert.True(isDigest("sha256:1359608115b94599e5641638bac5aef1ddfaa79bb96057ebf41ebc8d33acf8a7"))
}
