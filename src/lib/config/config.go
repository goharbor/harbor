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

package config

import (
	"context"
	"errors"
	"sync"

	"github.com/goharbor/harbor/src/common"
	comModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
)

const (
	// SessionCookieName is the name of the cookie for session ID
	SessionCookieName = "sid"

	defaultKeyPath                     = "/etc/core/key"
	defaultRegistryTokenPrivateKeyPath = "/etc/core/private_key.pem"
)

var (
	// DefaultCfgManager the default change manager, default is DBCfgManager. If InMemoryConfigManager is used, need to set to InMemoryCfgManager in test code
	DefaultCfgManager = common.DBCfgManager
	managersMU        sync.RWMutex
	managers          = make(map[string]Manager)
)

// Manager defines the operation for config
type Manager interface {
	Load(ctx context.Context) error
	Set(ctx context.Context, key string, value interface{})
	Save(ctx context.Context) error
	Get(ctx context.Context, key string) *metadata.ConfigureValue
	UpdateConfig(ctx context.Context, cfgs map[string]interface{}) error
	GetUserCfgs(ctx context.Context) map[string]interface{}
	ValidateCfg(ctx context.Context, cfgs map[string]interface{}) error
	GetAll(ctx context.Context) map[string]interface{}
	GetDatabaseCfg() *comModels.Database
}

// Register  register the config manager
func Register(name string, mgr Manager) {
	managersMU.Lock()
	defer managersMU.Unlock()
	if mgr == nil {
		log.Error("Register manager is nil")
	}
	managers[name] = mgr
}

// GetManager get the configure manager by name
func GetManager(name string) (Manager, error) {
	mgr, ok := managers[name]
	if !ok {
		return nil, errors.New("config manager is not registered: " + name)
	}
	return mgr, nil
}

// DefaultMgr get default config manager
func DefaultMgr() Manager {
	manager, err := GetManager(DefaultCfgManager)
	if err != nil {
		log.Error("failed to get config manager")
	}
	return manager
}

// Init configurations
// need to import following package before calling it
// _ "github.com/goharbor/harbor/src/pkg/config/db"
func Init() {
	// init key provider
	initKeyProvider()
	log.Info("init secret store")
	// init secret store
	initSecretStore()
	DefaultCfgManager = common.DBCfgManager
}

// InitWithSettings init config with predefined configs, and optionally overwrite the keyprovider
// need to import following package before calling it
// _ "github.com/goharbor/harbor/src/pkg/config/inmemory"
func InitWithSettings(cfgs map[string]interface{}, kp ...encrypt.KeyProvider) {
	Init()
	DefaultCfgManager = common.InMemoryCfgManager
	mgr := DefaultMgr()
	mgr.UpdateConfig(backgroundCtx, cfgs)
	if len(kp) > 0 {
		keyProvider = kp[0]
	}
}

// GetCfgManager return the current config manager
func GetCfgManager(ctx context.Context) Manager {
	return DefaultMgr()
}

// Load configurations
func Load(ctx context.Context) error {
	return DefaultMgr().Load(ctx)
}

// Upload save all configurations, used by testing
func Upload(cfg map[string]interface{}) error {
	return DefaultMgr().UpdateConfig(orm.Context(), cfg)
}
