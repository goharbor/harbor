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

package config

import (
	"os"
	"testing"

	"github.com/vmware/harbor/src/common/utils/test"
)

// test functions under package jobservice/config
func TestConfig(t *testing.T) {
	server, err := test.NewAdminserver(nil)
	if err != nil {
		t.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

	if err := os.Setenv("ADMIN_SERVER_URL", server.URL); err != nil {
		t.Fatalf("failed to set env %s: %v", "ADMIN_SERVER_URL", err)
	}

	secretKeyPath := "/tmp/secretkey"
	_, err = test.GenerateKey(secretKeyPath)
	if err != nil {
		t.Errorf("failed to generate secret key: %v", err)
		return
	}
	defer os.Remove(secretKeyPath)

	if err := os.Setenv("KEY_PATH", secretKeyPath); err != nil {
		t.Fatalf("failed to set env %s: %v", "KEY_PATH", err)
	}

	if err := Init(); err != nil {
		t.Fatalf("failed to initialize configurations: %v", err)
	}

	if _, err := VerifyRemoteCert(); err != nil {
		t.Fatalf("failed to get verify remote cert: %v", err)
	}

	if _, err := Database(); err != nil {
		t.Fatalf("failed to get database settings: %v", err)
	}

	if _, err := MaxJobWorkers(); err != nil {
		t.Fatalf("failed to get max job workers: %v", err)
	}

	if _, err := LocalRegURL(); err != nil {
		t.Fatalf("failed to get registry URL: %v", err)
	}

	if dir := LogDir(); dir != "/var/log/jobs" {
		t.Errorf("unexpected log directory: %s != %s", dir, "/var/log/jobs")
	}

	if _, err := SecretKey(); err != nil {
		t.Fatalf("failed to get secret key: %v", err)
	}

	if len(InternalTokenServiceEndpoint()) == 0 {
		t.Error("the internal token service endpoint is null")
	}

	if _, err := ExtEndpoint(); err != nil {
		t.Fatalf("failed to get ext endpoint: %v", err)
	}
}
