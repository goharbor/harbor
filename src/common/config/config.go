// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

// Package config provide methods to get the configurations reqruied by code in src/common
package config

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/cache"
	"github.com/vmware/harbor/src/adminserver/client"
	"github.com/vmware/harbor/src/common"
)

// Manager manages configurations
type Manager struct {
	client client.Client
	Cache  bool
	cache  cache.Cache
	key    string
}

// NewManager returns an instance of Manager
func NewManager(client client.Client, enableCache bool) *Manager {
	m := &Manager{
		client: client,
	}

	if enableCache {
		m.Cache = true
		m.cache = cache.NewMemoryCache()
		m.key = "cfg"
	}

	return m
}

// Load configurations, if cache is enabled, cache the configurations
func (m *Manager) Load() (map[string]interface{}, error) {
	c, err := m.client.GetCfgs()
	if err != nil {
		return nil, err
	}

	if m.Cache {
		expi, err := getCfgExpiration(c)
		if err != nil {
			return nil, err
		}

		// copy the configuration map so that later modification to the
		// map does not effect the cached value
		cachedCfgs := map[string]interface{}{}
		for k, v := range c {
			cachedCfgs[k] = v
		}

		if err = m.cache.Put(m.key, cachedCfgs,
			time.Duration(expi)*time.Second); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Reset configurations
func (m *Manager) Reset() error {
	return m.client.ResetCfgs()
}

func getCfgExpiration(m map[string]interface{}) (int, error) {
	if m == nil {
		return 0, fmt.Errorf("can not get cfg expiration as configurations are null")
	}

	expi, ok := m[common.CfgExpiration]
	if !ok {
		return 0, fmt.Errorf("cfg expiration is not set")
	}

	return int(expi.(float64)), nil
}

// Get : if cache is enabled, read configurations from cache,
// if cache is null or cache is disabled it loads configurations directly
func (m *Manager) Get() (map[string]interface{}, error) {
	if m.Cache {
		c := m.cache.Get(m.key)
		if c != nil {
			return c.(map[string]interface{}), nil
		}
	}
	return m.Load()
}

// Upload configurations
func (m *Manager) Upload(cfgs map[string]interface{}) error {
	return m.client.UpdateCfgs(cfgs)
}
