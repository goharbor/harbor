/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package dao

/*
import (
	"testing"

	"github.com/vmware/harbor/src/common/models"
)

func deleteConfigByKey(key string) error {
	if _, err := GetOrmer().Raw("delete from properties where k = ?", key).
		Exec(); err != nil {
		return err
	}
	return nil
}

func TestGetConfigByKey(t *testing.T) {
	cfg := &models.Config{
		Key:   "key",
		Value: "value",
	}

	if err := InsertConfig(cfg); err != nil {
		t.Fatalf("failed to insert configuration into table: %v", err)
	}
	defer func(key string) {
		if err := deleteConfigByKey(key); err != nil {
			t.Fatalf("failed to delete configuration %s: %v", key, err)
		}
	}(cfg.Key)

	config, err := GetConfigByKey(cfg.Key)
	if err != nil {
		t.Fatalf("failed to get configuration by key %s: %v", cfg.Key, err)
	}

	if config == nil {
		t.Fatal("configuration is nil")
	}

	if config.Value != cfg.Value {
		t.Fatalf("unexpected value: %s != %s", config.Value, cfg.Value)
	}
}

func TestListConfigs(t *testing.T) {
	configs, err := ListConfigs()
	if err != nil {
		t.Fatalf("failed to list configurations: %v", err)
	}
	size := len(configs)

	cfg := &models.Config{
		Key:   "key",
		Value: "value",
	}
	if err := InsertConfig(cfg); err != nil {
		t.Fatalf("failed to insert configuration into table: %v", err)
	}
	defer func(key string) {
		if err := deleteConfigByKey(key); err != nil {
			t.Fatalf("failed to delete configuration %s: %v", key, err)
		}
	}(cfg.Key)

	configs, err = ListConfigs()
	if err != nil {
		t.Fatalf("failed to list configurations: %v", err)
	}

	if size+1 != len(configs) {
		t.Fatalf("unexpected length of configurations: %d != %d", len(configs), size+1)
	}
}

func TestInsertConfig(t *testing.T) {
	cfg := &models.Config{
		Key:   "key1",
		Value: "value1",
	}

	if err := InsertConfig(cfg); err != nil {
		t.Fatalf("failed to insert configuration into table: %v", err)
	}
	defer func(key string) {
		if err := deleteConfigByKey(key); err != nil {
			t.Fatalf("failed to delete configuration %s: %v", key, err)
		}
	}(cfg.Key)

	config, err := GetConfigByKey(cfg.Key)
	if err != nil {
		t.Fatalf("failed to get configuration by key %s: %v", cfg.Key, err)
	}
	if config == nil {
		t.Fatal("configuration is nil")
	}

	if config.Value != cfg.Value {
		t.Fatalf("unexpected value: %s != %s", config.Value, cfg.Value)
	}
}

func TestUpdateConfig(t *testing.T) {
	cfg := &models.Config{
		Key:   "key",
		Value: "value",
	}

	if err := InsertConfig(cfg); err != nil {
		t.Fatalf("failed to insert configuration into table: %v", err)
	}
	defer func(key string) {
		if err := deleteConfigByKey(key); err != nil {
			t.Fatalf("failed to delete configuration %s: %v", key, err)
		}
	}(cfg.Key)

	newCfg := &models.Config{
		Key:   "key",
		Value: "new_value",
	}
	if err := UpdateConfig(newCfg); err != nil {
		t.Fatalf("failed to update configuration: %v", err)
	}

	config, err := GetConfigByKey(cfg.Key)
	if err != nil {
		t.Fatalf("failed to get configuration by key %s: %v", cfg.Key, err)
	}

	if config == nil {
		t.Fatal("configuration is nil")
	}

	if config.Value != newCfg.Value {
		t.Fatalf("unexpected value: %s != %s", config.Value, newCfg.Value)
	}
}

func TestInsertOrUpdateConfigs(t *testing.T) {
	cfg1 := &models.Config{
		Key:   "key1",
		Value: "value1",
	}

	if err := InsertConfig(cfg1); err != nil {
		t.Fatalf("failed to insert configuration into table: %v", err)
	}
	defer func(key string) {
		if err := deleteConfigByKey(key); err != nil {
			t.Fatalf("failed to delete configuration %s: %v", key, err)
		}
	}(cfg1.Key)

	cfg2 := &models.Config{
		Key:   "key2",
		Value: "value2",
	}

	if err := InsertOrUpdateConfigs([]*models.Config{cfg1, cfg2}); err != nil {
		t.Fatalf("failed to insert or update configurations: %v", err)
	}
	defer func(key string) {
		if err := deleteConfigByKey(key); err != nil {
			t.Fatalf("failed to delete configuration %s: %v", key, err)
		}
	}(cfg2.Key)
}

func TestAuthModeCanBeModified(t *testing.T) {
	c, err := GetOrmer().QueryTable(&models.User{}).Count()
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}

	if c == 1 {
		flag, err := AuthModeCanBeModified()
		if err != nil {
			t.Fatalf("failed to determine whether auth mode can be modified: %v", err)
		}
		if !flag {
			t.Errorf("unexpected result: %t != %t", flag, true)
		}

		user := models.User{
			Username: "user_for_config_test",
			Email:    "user_for_config_test@vmware.com",
			Password: "P@ssword",
			Realname: "user_for_config_test",
		}
		id, err := Register(user)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}
		defer func(id int64) {
			if err := deleteUser(id); err != nil {
				t.Fatalf("failed to delete user %d: %v", id, err)
			}
		}(id)

		flag, err = AuthModeCanBeModified()
		if err != nil {
			t.Fatalf("failed to determine whether auth mode can be modified: %v", err)
		}
		if flag {
			t.Errorf("unexpected result: %t != %t", flag, false)
		}

	} else {
		flag, err := AuthModeCanBeModified()
		if err != nil {
			t.Fatalf("failed to determine whether auth mode can be modified: %v", err)
		}
		if flag {
			t.Errorf("unexpected result: %t != %t", flag, false)
		}
	}
}
*/
