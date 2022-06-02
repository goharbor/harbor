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

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/config/db/dao"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/redis" // redis cache
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/stretchr/testify/assert"
)

func TestCacheLoadAndSave(t *testing.T) {
	ctx := orm.Context()
	cache, _ := cache.New("redis")
	driver := NewCacheDriver(cache, &Database{cfgDAO: dao.New()})

	cfgs := map[string]interface{}{
		common.AUTHMode: "db_auth",
		common.LDAPURL:  "ldap://ldap.vmware.com",
	}
	driver.Save(orm.Context(), cfgs)

	cf, err := driver.Load(ctx)
	if err != nil {
		fmt.Printf("load failed %v", err)
	}

	assert.Contains(t, cf, common.AUTHMode)
	assert.Contains(t, cf, common.LDAPURL)
}

func BenchmarkCacheLoad(b *testing.B) {
	ctx := orm.Context()
	cfgs := map[string]interface{}{}
	for _, item := range metadata.Instance().GetAll() {
		cfgs[item.Name] = item.DefaultValue
	}

	driver := Database{}
	driver.Save(ctx, cfgs)

	cache, _ := cache.New("redis")
	c := Cache{cache: cache, driver: &driver}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if _, err := c.Load(ctx); err != nil {
				fmt.Printf("load failed, %v", err)
			}
		}
	})

	if err := cache.Delete(ctx, cacheKey); err != nil {
		fmt.Printf("Delete cache failed, %v\n", err)
	}
}
