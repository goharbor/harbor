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

package config

import (
	"os"
	"testing"

	"github.com/vmware/harbor/src/common/utils/test"
)

// test functions under package jobservice/config
func TestConfig(t *testing.T) {
	server, err := test.NewAdminserver()
	if err != nil {
		t.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

	url := os.Getenv("ADMIN_SERVER_URL")
	defer os.Setenv("ADMIN_SERVER_URL", url)

	if err := os.Setenv("ADMIN_SERVER_URL", server.URL); err != nil {
		t.Fatalf("failed to set env %s: %v", "ADMIN_SERVER_URL", err)
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

	LocalUIURL()

	if _, err := LocalRegURL(); err != nil {
		t.Fatalf("failed to get registry URL: %v", err)
	}

	if _, err := LogDir(); err != nil {
		t.Fatalf("failed to get log directory: %v", err)
	}

	if _, err := SecretKey(); err != nil {
		t.Fatalf("failed to get secret key: %v", err)
	}

	UISecret()
}
