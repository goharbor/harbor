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

package dao

import (
	"context"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/orm"
	"os"
	"testing"
)

var testCtx context.Context

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	testCtx = orm.Context()
	os.Exit(m.Run())
}

func TestSaveConfigEntries(t *testing.T) {
	dao := New()
	configEntries := []models.ConfigEntry{
		{
			Key:   "teststringkey",
			Value: "192.168.111.211",
		},
		{
			Key:   "testboolkey",
			Value: "true",
		},
		{
			Key:   "testnumberkey",
			Value: "5",
		},
		{
			Key:   common.CfgDriverDB,
			Value: "db",
		},
	}
	err := dao.SaveConfigEntries(testCtx, configEntries)
	if err != nil {
		t.Fatalf("failed to save configuration to database %v", err)
	}
	readEntries, err := GetConfigEntries(testCtx)
	if err != nil {
		t.Fatalf("Failed to get configuration from database %v", err)
	}
	findItem := 0
	for _, entry := range readEntries {
		switch entry.Key {
		case "teststringkey":
			if "192.168.111.211" == entry.Value {
				findItem++
			}
		case "testnumberkey":
			if "5" == entry.Value {
				findItem++
			}
		case "testboolkey":
			if "true" == entry.Value {
				findItem++
			}
		default:
		}
	}
	if findItem != 3 {
		t.Fatalf("Should update 3 configuration but only update %d", findItem)
	}

	configEntries = []models.ConfigEntry{
		{
			Key:   "teststringkey",
			Value: "192.168.111.215",
		},
		{
			Key:   "testboolkey",
			Value: "false",
		},
		{
			Key:   "testnumberkey",
			Value: "7",
		},
		{
			Key:   common.CfgDriverDB,
			Value: "db",
		},
	}
	err = dao.SaveConfigEntries(testCtx, configEntries)
	if err != nil {
		t.Fatalf("failed to save configuration to database %v", err)
	}
	readEntries, err = GetConfigEntries(testCtx)
	if err != nil {
		t.Fatalf("Failed to get configuration from database %v", err)
	}
	findItem = 0
	for _, entry := range readEntries {
		switch entry.Key {
		case "teststringkey":
			if "192.168.111.215" == entry.Value {
				findItem++
			}
		case "testnumberkey":
			if "7" == entry.Value {
				findItem++
			}
		case "testboolkey":
			if "false" == entry.Value {
				findItem++
			}
		default:
		}
	}
	if findItem != 3 {
		t.Fatalf("Should update 3 configuration but only update %d", findItem)
	}
}
