// Copyright 2018 Project Harbor Authors
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

package oidc

import (
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	cfgModels "github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/orm"
	_ "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	conf := map[string]any{
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://accounts.google.com",
		common.OIDCVerifyCert:   "true",
		common.OIDCScope:        "openid, profile, offline_access",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "secret",
		common.ExtEndpoint:      "https://harbor.test",
	}
	kp := &encrypt.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}

	config.InitWithSettings(conf, kp)

	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestHelperCreate(t *testing.T) {
	testP := &providerHelper{}
	assert.Nil(t, testP.instance.Load())
	err := testP.create(orm.Context())
	assert.Nil(t, err)
	assert.NotNil(t, testP.instance.Load())
	assert.True(t, time.Now().Sub(testP.creationTime) < 2*time.Second)
}

func TestHelperGet(t *testing.T) {
	testP := &providerHelper{}
	ctx := orm.Context()
	p, err := testP.get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "https://oauth2.googleapis.com/token", p.Endpoint().TokenURL)

	update := map[string]any{
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://accounts.google.com",
		common.OIDCVerifyCert:   "true",
		common.OIDCScope:        "openid, profile, offline_access",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "new-secret",
		common.ExtEndpoint:      "https://harbor.test",
	}
	config.GetCfgManager(ctx).UpdateConfig(ctx, update)

	t.Log("Sleep for 5 seconds")
	time.Sleep(5 * time.Second)
	oidcSetting, err := config.OIDCSetting(ctx)
	assert.Nil(t, err)
	assert.Equal(t, "new-secret", oidcSetting.ClientSecret)
}

func TestAuthCodeURL(t *testing.T) {
	conf := map[string]any{
		common.OIDCName:               "test",
		common.OIDCEndpoint:           "https://accounts.google.com",
		common.OIDCVerifyCert:         "true",
		common.OIDCScope:              "openid, profile, offline_access",
		common.OIDCCLientID:           "client",
		common.OIDCClientSecret:       "secret",
		common.ExtEndpoint:            "https://harbor.test",
		common.OIDCExtraRedirectParms: `{"test_key":"test_value"}`,
	}
	ctx := orm.Context()
	config.GetCfgManager(ctx).UpdateConfig(ctx, conf)
	res, err := AuthCodeURL(ctx, "random", "this-is-a-pkce-code")
	assert.Nil(t, err)
	u, err := url.ParseRequestURI(res)
	assert.Nil(t, err)
	q, err := url.ParseQuery(u.RawQuery)
	assert.Nil(t, err)
	assert.Equal(t, "test_value", q.Get("test_key"))
	assert.Equal(t, "offline", q.Get("access_type"))
	assert.True(t, len(q.Get("code_challenge")) > 0)
	assert.Equal(t, "S256", q.Get("code_challenge_method"))
	assert.False(t, strings.Contains(q.Get("scope"), "offline_access"))
}

func TestTestEndpoint(t *testing.T) {
	c1 := Conn{
		URL:        googleEndpoint,
		VerifyCert: true,
	}
	c2 := Conn{
		URL:        "https://www.baidu.com",
		VerifyCert: false,
	}
	assert.Nil(t, TestEndpoint(c1))
	assert.NotNil(t, TestEndpoint(c2))
}

type fakeClaims struct {
	claims map[string]any
}

func (fc *fakeClaims) Claims(n any) error {
	b, err := json.Marshal(fc.claims)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, n)
}

func TestGroupsFromClaim(t *testing.T) {
	in := map[string]any{
		"user":         "user1",
		"groups":       []any{"group1", "group2"},
		"groups_2":     []any{"group1", "group2", 2},
		"single_group": "onlygroup",
		"empty_string": "",
	}

	m := []struct {
		input  map[string]any
		key    string
		expect []string
		ok     bool
	}{
		{
			// Test single string value (supported for OIDC providers that return string for single group)
			in,
			"user",
			[]string{"user1"},
			true,
		},
		{
			in,
			"prg",
			[]string{},
			false,
		},
		{
			in,
			"groups",
			[]string{"group1", "group2"},
			true,
		},
		{
			in,
			"groups_2",
			[]string{"group1", "group2"},
			true,
		},
		{
			// Test single string group
			in,
			"single_group",
			[]string{"onlygroup"},
			true,
		},
		{
			// Test empty string group
			in,
			"empty_string",
			[]string{""},
			true,
		},
	}

	for _, tc := range m {

		r, ok := groupsFromClaims(&fakeClaims{tc.input}, tc.key)
		assert.Equal(t, tc.expect, r)
		assert.Equal(t, tc.ok, ok)
	}
}

func TestUserInfoFromClaims(t *testing.T) {
	s := []struct {
		input   map[string]any
		setting cfgModels.OIDCSetting
		expect  *UserInfo
	}{
		{
			input: map[string]any{
				"name":   "Daniel",
				"email":  "daniel@gmail.com",
				"groups": []any{"g1", "g2"},
			},
			setting: cfgModels.OIDCSetting{
				Name:        "t1",
				GroupsClaim: "grouplist",
				UserClaim:   "",
				AdminGroup:  "g1",
			},
			expect: &UserInfo{
				Issuer:              "",
				Subject:             "",
				autoOnboardUsername: "",
				Username:            "Daniel",
				Email:               "daniel@gmail.com",
				Groups:              []string{},
				hasGroupClaim:       false,
			},
		},
		{
			input: map[string]any{
				"name":   "Daniel",
				"email":  "daniel@gmail.com",
				"groups": []any{"g1", "g2"},
			},
			setting: cfgModels.OIDCSetting{
				Name:        "t2",
				GroupsClaim: "groups",
				UserClaim:   "",
				AdminGroup:  "g1",
			},
			expect: &UserInfo{
				Issuer:              "",
				Subject:             "",
				autoOnboardUsername: "",
				Username:            "Daniel",
				Email:               "daniel@gmail.com",
				Groups:              []string{"g1", "g2"},
				AdminGroupMember:    true,
				hasGroupClaim:       true,
			},
		},
		{
			input: map[string]any{
				"iss":        "issuer",
				"sub":        "subject000",
				"name":       "jack",
				"email":      "jack@gmail.com",
				"groupclaim": []any{},
			},
			setting: cfgModels.OIDCSetting{
				Name:        "t3",
				GroupsClaim: "groupclaim",
				UserClaim:   "",
				AdminGroup:  "g1",
			},
			expect: &UserInfo{
				Issuer:              "issuer",
				Subject:             "subject000",
				autoOnboardUsername: "",
				Username:            "jack",
				Email:               "jack@gmail.com",
				Groups:              []string{},
				hasGroupClaim:       true,
				AdminGroupMember:    false,
			},
		},
		{
			input: map[string]any{
				"name":   "Alvaro",
				"email":  "airadier@gmail.com",
				"groups": []any{"g1", "g2"},
			},
			setting: cfgModels.OIDCSetting{
				Name:        "t4",
				GroupsClaim: "grouplist",
				UserClaim:   "email",
				AdminGroup:  "g1",
			},
			expect: &UserInfo{
				Issuer:              "",
				Subject:             "",
				autoOnboardUsername: "airadier@gmail.com",
				Username:            "airadier@gmail.com", // Set Username based on configured UserClaim
				Email:               "airadier@gmail.com",
				Groups:              []string{},
				hasGroupClaim:       false,
				AdminGroupMember:    false,
			},
		},
	}
	for _, tc := range s {
		out, err := userInfoFromClaims(&fakeClaims{tc.input}, tc.setting)
		assert.Nil(t, err)
		assert.Equal(t, *tc.expect, *out)
	}
}

func TestMergeUserInfo(t *testing.T) {
	s := []struct {
		fromInfo    *UserInfo
		fromIDToken *UserInfo
		expected    *UserInfo
	}{
		{
			fromInfo: &UserInfo{
				Issuer:              "",
				Subject:             "",
				autoOnboardUsername: "",
				Username:            "daniel",
				Email:               "daniel@gmail.com",
				Groups:              []string{},
				hasGroupClaim:       false,
			},
			fromIDToken: &UserInfo{
				Issuer:              "issuer-google",
				Subject:             "subject-daniel",
				autoOnboardUsername: "",
				Username:            "daniel",
				Email:               "daniel@yahoo.com",
				Groups:              []string{"developers", "everyone"},
				hasGroupClaim:       true,
			},
			expected: &UserInfo{
				Issuer:        "issuer-google",
				Subject:       "subject-daniel",
				Username:      "daniel",
				Email:         "daniel@gmail.com",
				Groups:        []string{"developers", "everyone"},
				hasGroupClaim: true,
			},
		},
		{
			fromInfo: &UserInfo{
				Issuer:              "",
				Subject:             "",
				autoOnboardUsername: "",
				Username:            "tom",
				Email:               "tom@gmail.com",
				Groups:              nil,
				hasGroupClaim:       false,
			},
			fromIDToken: &UserInfo{
				Issuer:              "issuer-okta",
				Subject:             "subject-jiangtan",
				autoOnboardUsername: "",
				Username:            "tom",
				Email:               "tom@okta.com",
				Groups:              []string{"nouse"},
				hasGroupClaim:       false,
			},
			expected: &UserInfo{
				Issuer:        "issuer-okta",
				Subject:       "subject-jiangtan",
				Username:      "tom",
				Email:         "tom@gmail.com",
				Groups:        []string{},
				hasGroupClaim: false,
			},
		},
		{
			fromInfo: &UserInfo{
				Issuer:              "",
				Subject:             "",
				autoOnboardUsername: "",
				Username:            "jim",
				Email:               "jim@gmail.com",
				Groups:              []string{},
				hasGroupClaim:       true,
			},
			fromIDToken: &UserInfo{
				Issuer:              "issuer-yahoo",
				Subject:             "subject-jim",
				autoOnboardUsername: "",
				Username:            "jim",
				Email:               "jim@yaoo.com",
				Groups:              []string{"g1", "g2"},
				hasGroupClaim:       true,
			},
			expected: &UserInfo{
				Issuer:        "issuer-yahoo",
				Subject:       "subject-jim",
				Username:      "jim",
				Email:         "jim@gmail.com",
				Groups:        []string{},
				hasGroupClaim: true,
			},
		},
		{
			fromInfo: &UserInfo{
				Issuer:              "",
				Subject:             "",
				autoOnboardUsername: "",
				Username:            "",
				Email:               "kevin@whatever.com",
				Groups:              []string{},
				hasGroupClaim:       false,
			},
			fromIDToken: &UserInfo{
				Issuer:              "issuer-whatever",
				Subject:             "subject-kevin",
				autoOnboardUsername: "",
				Username:            "kevin",
				Email:               "kevin@whatever.com",
				Groups:              []string{"g1", "g2"},
				hasGroupClaim:       true,
			},
			expected: &UserInfo{
				Issuer:        "issuer-whatever",
				Subject:       "subject-kevin",
				Username:      "kevin",
				Email:         "kevin@whatever.com",
				Groups:        []string{"g1", "g2"},
				hasGroupClaim: true,
			},
		},
		{
			fromInfo: &UserInfo{
				Issuer:  "",
				Subject: "",
				// only the auto onboard username from token will be used
				autoOnboardUsername: "info-jt",
				Username:            "",
				Email:               "jt@whatever.com",
				Groups:              []string{},
				hasGroupClaim:       false,
			},
			fromIDToken: &UserInfo{
				Issuer:              "issuer-whatever",
				Subject:             "subject-jt",
				autoOnboardUsername: "token-jt",
				Username:            "jt",
				Email:               "jt@whatever.com",
				Groups:              []string{"g1", "g2"},
				hasGroupClaim:       true,
			},
			expected: &UserInfo{
				Issuer:        "issuer-whatever",
				Subject:       "subject-jt",
				Username:      "token-jt",
				Email:         "jt@whatever.com",
				Groups:        []string{"g1", "g2"},
				hasGroupClaim: true,
			},
		},
	}

	for _, tc := range s {
		m := mergeUserInfo(tc.fromInfo, tc.fromIDToken)
		assert.Equal(t, *tc.expected, *m)
	}
}

func TestInjectGroupsToUser(t *testing.T) {
	cases := []struct {
		userInfo *UserInfo
		old      *models.User
		new      *models.User
	}{
		{
			userInfo: &UserInfo{
				Issuer:           "issuer-yahoo",
				Subject:          "subject-jim",
				Username:         "jim",
				Email:            "jim@gmail.com",
				Groups:           []string{},
				hasGroupClaim:    true,
				AdminGroupMember: false,
			},
			old: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: false,
			},
			new: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: false,
			},
		},
		{
			userInfo: &UserInfo{
				Issuer:           "issuer-yahoo",
				Subject:          "subject-jim",
				Username:         "jim",
				Email:            "jim@gmail.com",
				Groups:           []string{"1", "abc"},
				hasGroupClaim:    true,
				AdminGroupMember: true,
			},
			old: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: false,
			},
			new: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: true,
			},
		},
		{
			userInfo: &UserInfo{
				Issuer:           "issuer-yahoo",
				Subject:          "subject-jim",
				Username:         "jim",
				Email:            "jim@gmail.com",
				Groups:           []string{"1", "2"},
				hasGroupClaim:    true,
				AdminGroupMember: true,
			},
			old: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{},
				AdminRoleInAuth: false,
			},
			new: &models.User{
				Username:        "jim",
				Email:           "jim@gmail.com",
				GroupIDs:        []int{1, 2},
				AdminRoleInAuth: true,
			},
		},
	}
	for _, c := range cases {
		u := c.old
		InjectGroupsToUser(c.userInfo, u, mockPopulateGroups)
		assert.Equal(t, *c.new, *u)
	}
}

func Test_filterGroup(t *testing.T) {
	type args struct {
		groupNames []string
		filter     string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{"normal", args{[]string{"admin_user"}, "^admin.*"}, []string{"admin_user"}},
		{"multiple ", args{[]string{"admin_user", "harbor_admin"}, "^admin.*"}, []string{"admin_user"}},
		{"no match", args{[]string{"harbor_admin", "harbor_user", "sample_admin", "myadmin"}, "^admin.*"}, []string{}},
		{"empty filter", args{[]string{"harbor_admin", "harbor_user", "sample_admin", "myadmin"}, ""}, []string{"harbor_admin", "harbor_user", "sample_admin", "myadmin"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, filterGroup(tt.args.groupNames, tt.args.filter), "filterGroup(%v, %v)", tt.args.groupNames, tt.args.filter)
		})
	}
}
