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
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/cache"
	libCfg "github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/config"
	"github.com/goharbor/harbor/src/pkg/config/db/dao"
	"github.com/goharbor/harbor/src/pkg/config/store"
)

func init() {
	libCfg.Register(common.DBCfgManager, NewDBCfgManager())
}

// NewDBCfgManager - create DB config manager
func NewDBCfgManager() *config.CfgManager {
	cfgDriver := (store.Driver)(&Database{cfgDAO: dao.New()})

	if cache.Default() != nil {
		log.Debug("create DB config manager with cache enabled")
		cfgDriver = NewCacheDriver(cache.Default(), cfgDriver)
	}

	manager := &config.CfgManager{Store: store.NewConfigStore(cfgDriver)}
	// load default value
	manager.LoadDefault()
	// load system config from env
	manager.LoadSystemConfigFromEnv()
	return manager
}

// EnableConfigCache ...
func EnableConfigCache() {
	if cache.Default() == nil {
		log.Error("failed to enable config cache, cache is not ready.")
		return
	}
	libCfg.Register(common.DBCfgManager, NewDBCfgManager())
}
