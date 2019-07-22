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
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/models"
	cut "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/core/auth/authproxy/test"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
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
	defer dao.ExecuteBatchSQL([]string{"delete from user_group where group_name='OnBoardTest'"})
	a = &Auth{
		Endpoint:            mockSvr.URL + "/test/login",
		TokenReviewEndpoint: mockSvr.URL + "/test/tokenreview",
		SkipCertVerify:      true,
		// So it won't require mocking the cfgManager
		settingTimeStamp: time.Now(),
	}
	cfgMap := cut.GetUnitTestConfig()
	conf := map[string]interface{}{
		common.HTTPAuthProxyEndpoint:            a.Endpoint,
		common.HTTPAuthProxyTokenReviewEndpoint: a.TokenReviewEndpoint,
		common.HTTPAuthProxyVerifyCert:          !a.SkipCertVerify,
		common.PostGreSQLSSLMode:                cfgMap[common.PostGreSQLSSLMode],
		common.PostGreSQLUsername:               cfgMap[common.PostGreSQLUsername],
		common.PostGreSQLPort:                   cfgMap[common.PostGreSQLPort],
		common.PostGreSQLHOST:                   cfgMap[common.PostGreSQLHOST],
		common.PostGreSQLPassword:               cfgMap[common.PostGreSQLPassword],
		common.PostGreSQLDatabase:               cfgMap[common.PostGreSQLDatabase],
	}

	config.InitWithSettings(conf)
	rc := m.Run()
	if err := dao.ClearHTTPAuthProxyUsers(); err != nil {
		panic(err)
	}
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestAuth_Authenticate(t *testing.T) {
	groupIDs, err := group.GetGroupIDByGroupName([]string{"vsphere.local\\users", "vsphere.local\\administrators"}, common.HTTPGroupType)
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
					Username: "Admin@vsphere.local",
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
				Email:    "jt@placeholder.com",
				Realname: "jt",
				Password: pwd,
				Comment:  userEntryComment,
			},
		},
		{
			input: &models.User{
				Username: "Admin@vsphere.local",
			},
			expect: models.User{
				Username: "Admin@vsphere.local",
				Email:    "Admin@vsphere.local",
				Realname: "Admin@vsphere.local",
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

	emptyGroup := &models.UserGroup{}
	err := a.OnBoardGroup(emptyGroup, "")
	if err == nil {
		t.Fatal("Empty user group should failed to OnBoard")
	}
}
