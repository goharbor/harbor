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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	notarytest "github.com/goharbor/harbor/src/common/utils/notary/test"
	testutils "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var endpoint = "10.117.4.142"
var notaryServer *httptest.Server

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

func makeManifest(configSize int64, layerSizes []int64) schema2.Manifest {
	manifest := schema2.Manifest{
		Versioned: manifest.Versioned{SchemaVersion: 2, MediaType: schema2.MediaTypeManifest},
		Config: distribution.Descriptor{
			MediaType: schema2.MediaTypeImageConfig,
			Size:      configSize,
			Digest:    digest.FromString(utils.GenerateRandomString()),
		},
	}

	for _, size := range layerSizes {
		manifest.Layers = append(manifest.Layers, distribution.Descriptor{
			MediaType: schema2.MediaTypeLayer,
			Size:      size,
			Digest:    digest.FromString(utils.GenerateRandomString()),
		})
	}

	return manifest
}

func getDescriptor(manifest schema2.Manifest) distribution.Descriptor {
	buf, _ := json.Marshal(manifest)
	_, desc, _ := distribution.UnmarshalManifest(manifest.Versioned.MediaType, buf)
	return desc
}

func TestParseManifestInfo(t *testing.T) {
	manifest := makeManifest(1, []int64{2, 3, 4})

	tests := []struct {
		name    string
		req     func() *http.Request
		want    *ManifestInfo
		wantErr bool
	}{
		{
			"ok",
			func() *http.Request {
				buf, _ := json.Marshal(manifest)
				req, _ := http.NewRequest(http.MethodPut, "/v2/library/photon/manifests/latest", bytes.NewReader(buf))
				req.Header.Add("Content-Type", manifest.MediaType)

				return req
			},
			&ManifestInfo{
				ProjectID:  1,
				Repository: "library/photon",
				Tag:        "latest",
				Digest:     getDescriptor(manifest).Digest.String(),
				References: manifest.References(),
				Descriptor: getDescriptor(manifest),
			},
			false,
		},
		{
			"bad content type",
			func() *http.Request {
				buf, _ := json.Marshal(manifest)
				req, _ := http.NewRequest(http.MethodPut, "/v2/notfound/photon/manifests/latest", bytes.NewReader(buf))
				req.Header.Add("Content-Type", "application/json")

				return req
			},
			nil,
			true,
		},
		{
			"bad manifest",
			func() *http.Request {
				req, _ := http.NewRequest(http.MethodPut, "/v2/notfound/photon/manifests/latest", bytes.NewReader([]byte("")))
				req.Header.Add("Content-Type", schema2.MediaTypeManifest)

				return req
			},
			nil,
			true,
		},
		{
			"body missing",
			func() *http.Request {
				req, _ := http.NewRequest(http.MethodPut, "/v2/notfound/photon/manifests/latest", nil)
				req.Header.Add("Content-Type", schema2.MediaTypeManifest)

				return req
			},
			nil,
			true,
		},
		{
			"project not found",
			func() *http.Request {

				buf, _ := json.Marshal(manifest)

				req, _ := http.NewRequest(http.MethodPut, "/v2/notfound/photon/manifests/latest", bytes.NewReader(buf))
				req.Header.Add("Content-Type", manifest.MediaType)

				return req
			},
			nil,
			true,
		},
		{
			"url not match",
			func() *http.Request {
				buf, _ := json.Marshal(manifest)
				req, _ := http.NewRequest(http.MethodPut, "/v2/library/photon/manifest/latest", bytes.NewReader(buf))
				req.Header.Add("Content-Type", manifest.MediaType)

				return req
			},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseManifestInfoFromReq(tt.req())
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseManifestInfoFromReq() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseManifestInfoFromReq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseManifestInfoFromPath(t *testing.T) {
	mustRequest := func(method, url string) *http.Request {
		req, _ := http.NewRequest(method, url, nil)
		return req
	}

	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *ManifestInfo
		wantErr bool
	}{
		{
			"ok for digest",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifests/sha256:3e17b60ab9d92d953fb8ebefa25624c0d23fb95f78dde5572285d10158044059")},
			&ManifestInfo{
				ProjectID:  1,
				Repository: "library/photon",
				Digest:     "sha256:3e17b60ab9d92d953fb8ebefa25624c0d23fb95f78dde5572285d10158044059",
			},
			false,
		},
		{
			"ok for tag",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifests/latest")},
			&ManifestInfo{
				ProjectID:  1,
				Repository: "library/photon",
				Tag:        "latest",
			},
			false,
		},
		{
			"project not found",
			args{mustRequest(http.MethodDelete, "/v2/notfound/photon/manifests/latest")},
			nil,
			true,
		},
		{
			"url not match",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifest/latest")},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseManifestInfoFromPath(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseManifestInfoFromPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseManifestInfoFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
