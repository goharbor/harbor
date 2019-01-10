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
package driver

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestDatabase_Load(t *testing.T) {

	cfgs := map[string]interface{}{
		common.AUTHMode: "db_auth",
		common.LDAPURL:  "ldap://ldap.vmware.com",
	}
	driver := Database{}
	driver.Save(cfgs)
	cfgMap, err := driver.Load()
	if err != nil {
		t.Errorf("failed to load, error %v", err)
	}
	assert.True(t, len(cfgMap) >= 1)

	if _, ok := cfgMap["ldap_url"]; !ok {
		t.Error("Can not find ldap_url")
	}
}

func TestDatabase_Save(t *testing.T) {
	ldapURL := "ldap://ldap.vmware.com"
	driver := Database{}
	prevCfg, err := driver.Load()
	if err != nil {
		t.Errorf("failed to load config %v", err)
	}
	cfgMap := map[string]interface{}{"ldap_url": ldapURL}
	driver.Save(cfgMap)
	updatedMap, err := driver.Load()
	if err != nil {
		t.Errorf("failed to load config %v", err)
	}
	assert.Equal(t, updatedMap["ldap_url"], ldapURL)
	driver.Save(prevCfg)

}
