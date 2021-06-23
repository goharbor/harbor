//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package db

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/lib/orm"
	cfgPkg "github.com/goharbor/harbor/src/pkg/config"
	"github.com/goharbor/harbor/src/pkg/config/db/dao"
	"github.com/goharbor/harbor/src/pkg/config/store"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var TestDBConfig = map[string]interface{}{
	"postgresql_host":     "localhost",
	"postgresql_database": "registry",
	"postgresql_password": "root123",
	"postgresql_username": "postgres",
	"postgresql_sslmode":  "disable",
	"email_host":          "127.0.0.1",
	"scan_all_policy":     `{"parameter":{"daily_time":0},"type":"daily"}`,
}

var configManager *cfgPkg.CfgManager
var testCtx context.Context

func TestMain(m *testing.M) {
	configManager = NewDBCfgManager()
	test.InitDatabaseFromEnv()
	testCtx = orm.Context()
	configManager.UpdateConfig(testCtx, TestDBConfig)
	os.Exit(m.Run())
}

func TestLoadFromDatabase(t *testing.T) {
	configManager.UpdateConfig(testCtx, TestDBConfig)
	configManager.Load(testCtx)
	assert.Equal(t, "127.0.0.1", configManager.Get(testCtx, "email_host").GetString())
	assert.Equal(t, `{"parameter":{"daily_time":0},"type":"daily"}`, configManager.Get(testCtx, "scan_all_policy").GetString())
}

func TestLoadUserCfg(t *testing.T) {
	configMap := configManager.GetUserCfgs(testCtx)
	assert.NotNil(t, configMap["ldap_url"])
	assert.NotNil(t, configMap["ldap_base_dn"])
}

func TestSaveToDatabase(t *testing.T) {
	fmt.Printf("database config %#v\n", configManager.GetDatabaseCfg())
	configManager.Load(testCtx)
	configManager.Set(testCtx, "read_only", "true")
	configManager.Save(testCtx)
	configManager.Load(testCtx)
	assert.Equal(t, true, configManager.Get(testCtx, "read_only").GetBool())
}

func TestUpdateCfg(t *testing.T) {
	testConfig := map[string]interface{}{
		"ldap_url":             "ldaps://ldap.vmware.com",
		"ldap_search_dn":       "cn=admin,dc=example,dc=com",
		"ldap_timeout":         10,
		"ldap_search_password": "admin",
		"ldap_base_dn":         "dc=example,dc=com",
	}
	configManager.Load(testCtx)
	configManager.UpdateConfig(testCtx, testConfig)

	assert.Equal(t, "ldaps://ldap.vmware.com", configManager.Get(testCtx, "ldap_url").GetString())
	assert.Equal(t, 10, configManager.Get(testCtx, "ldap_timeout").GetInt())
	assert.Equal(t, "admin", configManager.Get(testCtx, "ldap_search_password").GetPassword())
	assert.Equal(t, "cn=admin,dc=example,dc=com", configManager.Get(testCtx, "ldap_search_dn").GetString())
	assert.Equal(t, "dc=example,dc=com", configManager.Get(testCtx, "ldap_base_dn").GetString())
}

func TestCfgManager_loadDefaultValues(t *testing.T) {
	configManager.LoadDefault()
	if configManager.Get(testCtx, "ldap_timeout").GetInt() != 5 {
		t.Errorf("Failed to load ldap_timeout")
	}
}

func TestCfgManger_loadSystemValues(t *testing.T) {
	configManager.LoadDefault()
	configManager.LoadSystemConfigFromEnv()
	configManager.UpdateConfig(testCtx, map[string]interface{}{
		"postgresql_host": "127.0.0.1",
	})
	if configManager.Get(testCtx, "postgresql_host").GetString() != "127.0.0.1" {
		t.Errorf("Failed to set system value postgresql_host, expected %v, actual %v", "127.0.0.1", configManager.Get(nil, "postgresql_host").GetString())
	}
}
func TestCfgManager_GetDatabaseCfg(t *testing.T) {
	configManager.UpdateConfig(testCtx, map[string]interface{}{
		"postgresql_host":     "localhost",
		"postgresql_database": "registry",
		"postgresql_password": "root123",
		"postgresql_username": "postgres",
		"postgresql_sslmode":  "disable",
	})
	dbCfg := configManager.GetDatabaseCfg()
	assert.Equal(t, "localhost", dbCfg.PostGreSQL.Host)
	assert.Equal(t, "registry", dbCfg.PostGreSQL.Database)
	assert.Equal(t, "root123", dbCfg.PostGreSQL.Password)
	assert.Equal(t, "postgres", dbCfg.PostGreSQL.Username)
	assert.Equal(t, "disable", dbCfg.PostGreSQL.SSLMode)
}

func TestConfigStore_Save(t *testing.T) {
	cfgStore := store.NewConfigStore(&Database{cfgDAO: dao.New()})
	err := cfgStore.Save(testCtx)
	cfgStore.Set("ldap_verify_cert", metadata.ConfigureValue{Name: "ldap_verify_cert", Value: "true"})
	if err != nil {
		t.Fatal(err)
	}
	cfgValue, err := cfgStore.Get("ldap_verify_cert")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, cfgValue.GetBool())

}

func TestConfigStore_Load(t *testing.T) {
	cfgStore := store.NewConfigStore(&Database{cfgDAO: dao.New()})
	err := cfgStore.Load(testCtx)
	if err != nil {
		t.Fatal(err)
	}
	cfgValue, err := cfgStore.Get("ldap_url")
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "ldaps://ldap.vmware.com", cfgValue.GetString())

}

func TestToString(t *testing.T) {
	cases := []struct {
		name   string
		value  interface{}
		expect string
	}{
		{
			name:   "transform int",
			value:  999,
			expect: "999",
		},
		{
			name:   "transform slice",
			value:  []int{0, 1, 2},
			expect: "[0,1,2]",
		},
		{
			name:   "transform map",
			value:  map[string]string{"k": "v"},
			expect: "{\"k\":\"v\"}",
		},
		{
			name:   "transform bool",
			value:  false,
			expect: "false",
		},
		{
			name:   "transform nil",
			value:  nil,
			expect: "nil",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, err := store.ToString(c.value)
			assert.Nil(t, err)
			assert.Equal(t, c.expect, s)
		})
	}
}
