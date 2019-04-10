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
	"github.com/goharbor/harbor/src/common"
	config2 "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	conf := map[string]interface{}{
		common.OIDCName:           "test",
		common.OIDCEndpoint:       "https://accounts.google.com",
		common.OIDCSkipCertVerify: "false",
		common.OIDCScope:          "openid, profile, offline_access",
		common.OIDCCLientID:       "client",
		common.OIDCClientSecret:   "secret",
		common.ExtEndpoint:        "https://harbor.test",
	}
	kp := &config2.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}

	config.InitWithSettings(conf, kp)

	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}
func TestHelperLoadConf(t *testing.T) {
	testP := &providerHelper{}
	assert.Nil(t, testP.setting.Load())
	err := testP.reload()
	assert.Nil(t, err)
	assert.Equal(t, "test", testP.setting.Load().(models.OIDCSetting).Name)
	assert.Equal(t, endpoint{}, testP.ep)
}

func TestHelperCreate(t *testing.T) {
	testP := &providerHelper{}
	err := testP.reload()
	assert.Nil(t, err)
	assert.Nil(t, testP.instance.Load())
	err = testP.create()
	assert.Nil(t, err)
	assert.EqualValues(t, "https://accounts.google.com", testP.ep.url)
	assert.NotNil(t, testP.instance.Load())
}

func TestHelperGet(t *testing.T) {
	testP := &providerHelper{}
	p, err := testP.get()
	assert.Nil(t, err)
	assert.Equal(t, "https://oauth2.googleapis.com/token", p.Endpoint().TokenURL)

	update := map[string]interface{}{
		common.OIDCName:           "test",
		common.OIDCEndpoint:       "https://accounts.google.com",
		common.OIDCSkipCertVerify: "false",
		common.OIDCScope:          "openid, profile, offline_access",
		common.OIDCCLientID:       "client",
		common.OIDCClientSecret:   "new-secret",
		common.ExtEndpoint:        "https://harbor.test",
	}
	config.GetCfgManager().UpdateConfig(update)

	t.Log("Sleep for 5 seconds")
	time.Sleep(5 * time.Second)
	assert.Equal(t, "new-secret", testP.setting.Load().(models.OIDCSetting).ClientSecret)
}

func TestAuthCodeURL(t *testing.T) {
	res, err := AuthCodeURL("random")
	assert.Nil(t, err)
	u, err := url.ParseRequestURI(res)
	assert.Nil(t, err)
	q, err := url.ParseQuery(u.RawQuery)
	assert.Nil(t, err)
	assert.Equal(t, "offline", q.Get("access_type"))
	assert.False(t, strings.Contains(q.Get("scope"), "offline_access"))
}
