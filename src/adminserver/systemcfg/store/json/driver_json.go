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

package json

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/vmware/harbor/src/adminserver/systemcfg/store"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	// the default path of configuration file
	defaultPath = "/etc/harbor/config.json"
)

type cfgStore struct {
	path string // the path of cfg file
	sync.RWMutex
}

// NewCfgStore returns an instance of cfgStore that stores the configurations
// in a json file. The file will be created if it does not exist.
func NewCfgStore(path ...string) (store.Driver, error) {
	p := defaultPath
	if len(path) > 0 && len(path[0]) > 0 {
		p = path[0]
	}

	log.Debugf("path of configuration file: %s", p)

	if _, err := os.Stat(p); os.IsNotExist(err) {
		log.Infof("the configuration file %s does not exist, creating it...", p)
		if err = os.MkdirAll(filepath.Dir(p), 0600); err != nil {
			return nil, err
		}
		if err = ioutil.WriteFile(p, []byte{}, 0600); err != nil {
			return nil, err
		}
	}

	return &cfgStore{
		path: p,
	}, nil
}

// Name ...
func (c *cfgStore) Name() string {
	return "JSON"
}

// Read ...
func (c *cfgStore) Read() (map[string]interface{}, error) {
	c.RLock()
	defer c.RUnlock()

	return read(c.path)
}

func read(path string) (map[string]interface{}, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// empty file
	if len(b) == 0 {
		return nil, nil
	}

	config := map[string]interface{}{}
	if err = json.Unmarshal(b, &config); err != nil {
		return nil, err
	}

	return config, nil
}

// Write ...
func (c *cfgStore) Write(config map[string]interface{}) error {
	c.Lock()
	defer c.Unlock()

	cfg, err := read(c.path)
	if err != nil {
		return err
	}

	if cfg == nil {
		cfg = config
	} else {
		for k, v := range config {
			cfg[k] = v
		}
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(c.path, b, 0600); err != nil {
		return err
	}

	return nil
}
