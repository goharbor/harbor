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

package database

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/vmware/harbor/src/adminserver/systemcfg/store"
	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
)

const (
	name = "database"
)

var (
	numKeys = map[string]bool{
		common.EmailPort:       true,
		common.LDAPScope:       true,
		common.LDAPTimeout:     true,
		common.TokenExpiration: true,
		common.MaxJobWorkers:   true,
		common.CfgExpiration:   true,
		common.ClairDBPort:     true,
		common.PostGreSQLPort:  true,
	}
	boolKeys = map[string]bool{
		common.WithClair:        true,
		common.WithNotary:       true,
		common.SelfRegistration: true,
		common.EmailSSL:         true,
		common.EmailInsecure:    true,
		common.LDAPVerifyCert:   true,
		common.UAAVerifyCert:    true,
		common.ReadOnly:         true,
	}
	mapKeys = map[string]bool{
		common.ScanAllPolicy: true,
	}
)

type cfgStore struct {
	name string
}

// Name The name of the driver
func (c *cfgStore) Name() string {
	return name
}

// NewCfgStore New a cfg store for database driver
func NewCfgStore() (store.Driver, error) {
	return &cfgStore{
		name: name,
	}, nil
}

// Read configuration from database
func (c *cfgStore) Read() (map[string]interface{}, error) {
	configEntries, error := dao.GetConfigEntries()
	if error != nil {
		return nil, error
	}
	return WrapperConfig(configEntries)
}

// WrapperConfig Wrapper the configuration
func WrapperConfig(configEntries []*models.ConfigEntry) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	for _, entry := range configEntries {
		if numKeys[entry.Key] {
			strvalue, err := strconv.Atoi(entry.Value)
			if err != nil {
				return nil, err
			}
			config[entry.Key] = float64(strvalue)
		} else if boolKeys[entry.Key] {
			strvalue, err := strconv.ParseBool(entry.Value)
			if err != nil {
				return nil, err
			}
			config[entry.Key] = strvalue
		} else if mapKeys[entry.Key] {
			m := map[string]interface{}{}
			if err := json.Unmarshal([]byte(entry.Value), &m); err != nil {
				return nil, err
			}
			config[entry.Key] = m
		} else {
			config[entry.Key] = entry.Value
		}

	}
	return config, nil
}

// Write save configuration to database
func (c *cfgStore) Write(config map[string]interface{}) error {
	configEntries, err := TranslateConfig(config)
	if err != nil {
		return err
	}
	return dao.SaveConfigEntries(configEntries)
}

// TranslateConfig Translate configuration from int, bool, float64 to string
func TranslateConfig(config map[string]interface{}) ([]models.ConfigEntry, error) {
	var configEntries []models.ConfigEntry
	for k, v := range config {
		var entry = new(models.ConfigEntry)
		entry.Key = k
		switch v.(type) {
		case string:
			entry.Value = v.(string)
		case int:
			entry.Value = strconv.Itoa(v.(int))
		case bool:
			entry.Value = strconv.FormatBool(v.(bool))
		case float64:
			entry.Value = strconv.Itoa(int(v.(float64)))
		case map[string]interface{}:
			data, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			entry.Value = string(data)
		default:
			return nil, fmt.Errorf("unknown type %v", v)
		}
		configEntries = append(configEntries, *entry)
	}
	return configEntries, nil
}
