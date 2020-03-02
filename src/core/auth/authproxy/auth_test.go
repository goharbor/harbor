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

package authproxy

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/models"
	cut "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/auth/authproxy/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"net/http/httptest"
	"os"
	"testing"
)

var mockSvr *httptest.Server
var a *Auth
var pwd = "1234567ab"
var cmt = "By Authproxy"

func TestMain(m *testing.M) {
	cut.InitDatabaseFromEnv()
	if err := dao.ClearHTTPAuthProxyUsers(); err != nil {
		panic(err)
	}
	mockSvr = test.NewMockServer(map[string]string{"jt": "pp", "Admin@vsphere.local": "Admin!23"})
	defer mockSvr.Close()
	a = &Auth{
		Endpoint:            mockSvr.URL + "/test/login",
		TokenReviewEndpoint: mockSvr.URL + "/test/tokenreview",
	}
	cfgMap := cut.GetUnitTestConfig()
	conf := map[string]interface{}{
		common.HTTPAuthProxyEndpoint:            a.Endpoint,
		common.HTTPAuthProxyTokenReviewEndpoint: a.TokenReviewEndpoint,
		common.HTTPAuthProxyVerifyCert:          false,
		common.PostGreSQLSSLMode:                cfgMap[common.PostGreSQLSSLMode],
		common.PostGreSQLUsername:               cfgMap[common.PostGreSQLUsername],
		common.PostGreSQLPort:                   cfgMap[common.PostGreSQLPort],
		common.PostGreSQLHOST:                   cfgMap[common.PostGreSQLHOST],
		common.PostGreSQLPassword:               cfgMap[common.PostGreSQLPassword],
		common.PostGreSQLDatabase:               cfgMap[common.PostGreSQLDatabase],
	}

	config.InitWithSettings(conf)
	defer dao.ExecuteBatchSQL([]string{"delete from user_group where group_name='onboardtest'"})
	rc := m.Run()
	if err := dao.ClearHTTPAuthProxyUsers(); err != nil {
		panic(err)
	}
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestAuth_Authenticate(t *testing.T) {
	userGroups := []models.UserGroup{
		{GroupName: "vsphere.local\\users", GroupType: common.HTTPGroupType},
		{GroupName: "vsphere.local\\administrators", GroupType: common.HTTPGroupType},
		{GroupName: "vsphere.local\\caadmins", GroupType: common.HTTPGroupType},
		{GroupName: "vsphere.local\\systemconfiguration.bashshelladministrators", GroupType: common.HTTPGroupType},
		{GroupName: "vsphere.local\\systemconfiguration.administrators", GroupType: common.HTTPGroupType},
		{GroupName: "vsphere.local\\licenseservice.administrators", GroupType: common.HTTPGroupType},
		{GroupName: "vsphere.local\\everyone", GroupType: common.HTTPGroupType},
	}

	groupIDs, err := group.PopulateGroup(userGroups)
	if err != nil {
		t.Fatal("Failed to get groupIDs")
	}
	t.Log("auth endpoint: ", a.Endpoint)
	type output struct {
		user models.User
		err  error
	}
	type tc struct {
		input  models.AuthModel
		expect output
	}
	suite := []tc{
		{
			input: models.AuthModel{
				Principal: "jt", Password: "pp"},
			expect: output{
				user: models.User{
					Username: "jt",
					GroupIDs: groupIDs,
				},
				err: nil,
			},
		},
		{
			input: models.AuthModel{
				Principal: "Admin@vsphere.local",
				Password:  "Admin!23",
			},
			expect: output{
				user: models.User{
					Username: "admin@vsphere.local",
					GroupIDs: groupIDs,
					// Email:    "Admin@placeholder.com",
					// Password: pwd,
					// Comment:  fmt.Sprintf(cmtTmpl, path.Join(mockSvr.URL, "/test/login")),
				},
				err: nil,
			},
		},
		{
			input: models.AuthModel{
				Principal: "jt",
				Password:  "ppp",
			},
			expect: output{
				err: auth.ErrAuth{},
			},
		},
	}
	assert := assert.New(t)
	for _, c := range suite {
		r, e := a.Authenticate(c.input)
		if c.expect.err == nil {
			assert.Nil(e)
			assert.Equal(c.expect.user, *r)
		} else {
			assert.Nil(r)
			assert.NotNil(e)
			if _, ok := e.(auth.ErrAuth); ok {
				assert.IsType(auth.ErrAuth{}, e)
			}
		}
	}
}

func TestAuth_PostAuthenticate(t *testing.T) {
	type tc struct {
		input  *models.User
		expect models.User
	}
	suite := []tc{
		{
			input: &models.User{
				Username: "jt",
			},
			expect: models.User{
				Username: "jt",
				Email:    "",
				Realname: "jt",
				Password: pwd,
				Comment:  userEntryComment,
			},
		},
		{
			input: &models.User{
				Username: "admin@vsphere.local",
			},
			expect: models.User{
				Username: "admin@vsphere.local",
				Email:    "admin@vsphere.local",
				Realname: "admin@vsphere.local",
				Password: pwd,
				Comment:  userEntryComment,
			},
		},
	}
	for _, c := range suite {
		a.PostAuthenticate(c.input)
		assert.Equal(t, c.expect.Username, c.input.Username)
		assert.Equal(t, c.expect.Email, c.input.Email)
		assert.Equal(t, c.expect.Realname, c.input.Realname)
		assert.Equal(t, c.expect.Comment, c.input.Comment)
	}

}

func TestAuth_OnBoardGroup(t *testing.T) {
	input := &models.UserGroup{
		GroupName: "OnBoardTest",
		GroupType: common.HTTPGroupType,
	}
	a.OnBoardGroup(input, "")

	assert.True(t, input.ID > 0, "The OnBoardGroup should have a valid group ID")
	g, er := group.GetUserGroup(input.ID)
	assert.Nil(t, er)
	assert.Equal(t, "OnBoardTest", g.GroupName)

	emptyGroup := &models.UserGroup{}
	err := a.OnBoardGroup(emptyGroup, "")
	if err == nil {
		t.Fatal("Empty user group should failed to OnBoard")
	}
}

func TestGetTLSConfig(t *testing.T) {
	type result struct {
		hasError  bool
		insecure  bool
		nilRootCA bool
	}
	cases := []struct {
		input  *models.HTTPAuthProxy
		expect result
	}{
		{
			input: &models.HTTPAuthProxy{
				Endpoint:            "https://127.0.0.1/login",
				TokenReviewEndpoint: "https://127.0.0.1/tokenreview",
				VerifyCert:          false,
				SkipSearch:          false,
				ServerCertificate:   "",
			},
			expect: result{
				hasError:  false,
				insecure:  true,
				nilRootCA: true,
			},
		},
		{
			input: &models.HTTPAuthProxy{
				Endpoint:            "https://127.0.0.1/login",
				TokenReviewEndpoint: "https://127.0.0.1/tokenreview",
				VerifyCert:          false,
				SkipSearch:          false,
				ServerCertificate:   "This does not look like a cert",
			},
			expect: result{
				hasError:  false,
				insecure:  true,
				nilRootCA: true,
			},
		},
		{
			input: &models.HTTPAuthProxy{
				Endpoint:            "https://127.0.0.1/login",
				TokenReviewEndpoint: "https://127.0.0.1/tokenreview",
				VerifyCert:          true,
				SkipSearch:          false,
				ServerCertificate:   "This does not look like a cert",
			},
			expect: result{
				hasError: true,
			},
		},
		{
			input: &models.HTTPAuthProxy{
				Endpoint:            "https://127.0.0.1/login",
				TokenReviewEndpoint: "https://127.0.0.1/tokenreview",
				VerifyCert:          true,
				SkipSearch:          false,
				ServerCertificate:   "",
			},
			expect: result{
				hasError:  false,
				insecure:  false,
				nilRootCA: true,
			},
		},
		{
			input: &models.HTTPAuthProxy{
				Endpoint:            "https://127.0.0.1/login",
				TokenReviewEndpoint: "https://127.0.0.1/tokenreview",
				VerifyCert:          true,
				SkipSearch:          false,
				ServerCertificate: `-----BEGIN CERTIFICATE-----
MIIFXzCCA0egAwIBAgIUY7f2ECRISPMeb1iVNvV5iQsIErUwDQYJKoZIhvcNAQEL
BQAwUjELMAkGA1UEBhMCQ04xDDAKBgNVBAgMA1BFSzERMA8GA1UEBwwIQmVpIEpp
bmcxDzANBgNVBAoMBlZNd2FyZTERMA8GA1UEAwwISGFyYm9yQ0EwHhcNMTkxMTE2
MjI1NjQ0WhcNMjAxMTE1MjI1NjQ0WjBdMQswCQYDVQQGEwJDTjEMMAoGA1UECAwD
UEVLMREwDwYDVQQHDAhCZWkgSmluZzEPMA0GA1UECgwGVk13YXJlMRwwGgYDVQQD
DBNqdC1kZXYuaGFyYm9yLmxvY2FsMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIIC
CgKCAgEAwQ/TQTHwnHwEC/KyHP8Tyv/v35GwXRGW6s1MYoqVnyQMPud0scLHAA2u
PZv2F5jy7PtnhcR0ZHGf05L/igY1utV7/4F2aFgOq0ExYMKxvzilitdcvsxmfTLI
m2pwS8+kH/1s1xR9/7ZPlPSdHuxcHgjtMqorljJykRyq0RBLvXCG+fmAY91kdLil
XWiuIU73lNpZHuXEDl4m2XUzb9cuhwvaHYs7aT6BhwqAJZUjwURUqMe1PIOo7vkQ
cKUHe3u3Fg/vbxfecEr3AHcKfIqm5fwI9vdzj5BP3lGT9hrxduAwI6SgehxGGWP4
aN/cKGIKt/2kzgFoQi/d5p3RBkLVNP/sEyAt9dLJj12ovkQwJzdKDVOy50t3ws9g
Mf3rUUb/wdZADK26lxolep9EXVe4kuWpOo1RvdI+lJJvWc3QaJIoVbr9LM8QN3e7
Iyk3pYRyaQj9EKZ4k0RgWVbIZfRLy1LkGMqmCcIqunHVdGDDjbO/ri8z0sKocMGl
qrqcBTPYmsau7PEcfzJY8k/HUDYdhZgIv2C1iLBl6eoTVDRbrGFcu8LzleWx2/19
OA1Cx7S8WyzN+9mjygqwEYc6qMtoeutAkOA5J8JkxBp0yqjUEnAB6E7R07xQP8AY
IKq5oVpkbD8WRI3w7l/X0AAkDtnijbgYWTfPVGihXHhRtkr/b9cCAwEAAaMiMCAw
HgYDVR0RBBcwFYITanQtZGV2LmhhcmJvci5sb2NhbDANBgkqhkiG9w0BAQsFAAOC
AgEAqwU10WwhI5W8w62vOpT+PKSXRVjHKhm3ltaIzAN7S772hiGl6L81cP9dXZ5y
FN0tFUtUVK01JRJHJaduXNwx24HlwRPNp7mLa4IPpeeVfG14/QCoOd8vxHtKG+V6
zE7Jx2FBVfUJ7P4SngEv4QfvZPt+lCXGK3V1RRTpkLD2knhBfu85rjPi+VW56Z7b
Jb2IEmVRlfR7Z0oYm8z3Obt2XuLIC1/8NtfxthggKr6DeSwZSJXrdGVbyvdiAmk2
iHQch0+UTkRDinL0je6WWbxBoAPXsWA9Hc69o8kmjcXUa99/i8FrC2QxPDUoxxMn
1zWk0jct2Tsr3VZ5HnaI5e8ifG7RUcE5Vr6w7MI5P44Q88zhboP1ShYQ/s513cu2
heELKvO3+mqv96lERtkUUwe8tm1zoPKzQI6ecGuqaTcMbXAGax+ud5XnUlz4xzTI
cByAsQ9DNhYIcOftnfz349zkHeWmMum4uiQwfp/+OrqX+O8U0eJYhlfu9vqCU05T
3mE8Hw5veNdLaZx+mzUVIDzrOB3fh/O62J9CsaZKtxwgLlGiT2ltuC1xUqn3DL8s
pkgODrJUf0p5dhcnLyA2nZolRV1rtwlgJstnEV4JpG1MwtmAZYZUilLvnfpVxTtA
y1bQusZMygQezfCuEzsewF+OpANFovCTUEs6s5vyoVNP8lk=
-----END CERTIFICATE-----
`,
			},
			expect: result{
				hasError:  false,
				insecure:  false,
				nilRootCA: false,
			},
		},
	}

	for _, c := range cases {
		output, err := getTLSConfig(c.input)
		if c.expect.hasError {
			assert.NotNil(t, err)
			continue
		} else {
			assert.Nil(t, err)
		}
		if output != nil {
			assert.Equal(t, c.expect.insecure, output.InsecureSkipVerify)
			assert.Equal(t, c.expect.nilRootCA, output.RootCAs == nil)
		}
	}

}
