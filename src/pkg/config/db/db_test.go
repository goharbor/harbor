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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/pkg/config/db/dao"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_Load(t *testing.T) {

	cfgs := map[string]interface{}{
		common.AUTHMode: "db_auth",
		common.LDAPURL:  "ldap://ldap.vmware.com",
	}
	driver := Database{cfgDAO: dao.New()}
	driver.Save(testCtx, cfgs)
	cfgMap, err := driver.Load(testCtx)
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
	driver := Database{cfgDAO: dao.New()}
	prevCfg, err := driver.Load(testCtx)
	if err != nil {
		t.Errorf("failed to load config %v", err)
	}
	cfgMap := map[string]interface{}{"ldap_url": ldapURL}
	driver.Save(testCtx, cfgMap)
	updatedMap, err := driver.Load(testCtx)
	if err != nil {
		t.Errorf("failed to load config %v", err)
	}
	assert.Equal(t, updatedMap["ldap_url"], ldapURL)
	driver.Save(testCtx, prevCfg)

}

func BenchmarkDatabaseLoad(b *testing.B) {
	cfgs := map[string]interface{}{}
	for _, item := range metadata.Instance().GetAll() {
		cfgs[item.Name] = item.DefaultValue
	}

	driver := Database{cfgDAO: dao.New()}
	driver.Save(testCtx, cfgs)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := driver.Load(testCtx); err != nil {
				fmt.Printf("load failed %v", err)
			}
		}
	})
}
