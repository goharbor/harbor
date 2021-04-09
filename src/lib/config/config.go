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
	"errors"
	"github.com/goharbor/harbor/src/lib/log"
	"sync"
)

var (
	managersMU sync.RWMutex
	managers   = make(map[string]Manager)
)

// Register  register the config manager
func Register(name string, mgr Manager) {
	managersMU.Lock()
	defer managersMU.Unlock()
	if mgr == nil {
		log.Error("Register manager is nil")
	}
	if _, dup := managers[name]; dup {
		log.Errorf("Register called twice for manager " + name)
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
