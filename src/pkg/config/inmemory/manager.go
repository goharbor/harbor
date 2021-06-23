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

package inmemory

import (
	"context"
	"github.com/goharbor/harbor/src/common"
	libCfg "github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/config"
	"github.com/goharbor/harbor/src/pkg/config/store"
	"sync"
)

func init() {
	libCfg.Register(common.InMemoryCfgManager, NewInMemoryManager())
}

// Driver driver for unit testing
type Driver struct {
	sync.Mutex
	cfgMap map[string]interface{}
}

// Load load data from driver, for example load from database,
// it should be invoked before get any user scope config
// for system scope config, because it is immutable, no need to call this method
func (d *Driver) Load(context.Context) (map[string]interface{}, error) {
	d.Lock()
	defer d.Unlock()
	res := make(map[string]interface{})
	for k, v := range d.cfgMap {
		res[k] = v
	}
	return res, nil
}

// Save only save user config setting to driver, for example: database, REST
func (d *Driver) Save(ctx context.Context, cfg map[string]interface{}) error {
	d.Lock()
	defer d.Unlock()
	for k, v := range cfg {
		d.cfgMap[k] = v
	}
	return nil
}

// NewInMemoryManager create a manager for unit testing, doesn't involve database or REST
func NewInMemoryManager() *config.CfgManager {
	manager := &config.CfgManager{Store: store.NewConfigStore(&Driver{cfgMap: map[string]interface{}{}})}
	// load default Value
	manager.LoadDefault()
	// load system config from env
	manager.LoadSystemConfigFromEnv()
	return manager
}
