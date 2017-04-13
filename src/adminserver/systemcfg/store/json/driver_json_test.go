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
	"os"
	"testing"
)

func TestReadWrite(t *testing.T) {
	path := "/tmp/config.json"
	store, err := NewCfgStore(path)
	if err != nil {
		t.Fatalf("failed to create json cfg store: %v", err)
	}
	defer func() {
		if err := os.Remove(path); err != nil {
			t.Fatalf("failed to remove the json file %s: %v", path, err)
		}
	}()

	if store.Name() != "JSON" {
		t.Errorf("unexpected name: %s != %s", store.Name(), "JSON")
		return
	}

	config := map[string]interface{}{
		"key": "value",
	}
	if err := store.Write(config); err != nil {
		t.Errorf("failed to write configurations to json file: %v", err)
		return
	}

	if _, err = store.Read(); err != nil {
		t.Errorf("failed to read configurations from json file: %v", err)
		return
	}
}
