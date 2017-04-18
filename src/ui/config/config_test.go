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

// test functions under package ui/config
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

	if err := Load(); err != nil {
		t.Fatalf("failed to load configurations: %v", err)
	}

	if err := Upload(map[string]interface{}{}); err != nil {
		t.Fatalf("failed to upload configurations: %v", err)
	}

	if _, err := GetSystemCfg(); err != nil {
		t.Fatalf("failed to get system configurations: %v", err)
	}

	mode, err := AuthMode()
	if err != nil {
		t.Fatalf("failed to get auth mode: %v", err)
	}
	if mode != "db_auth" {
		t.Errorf("unexpected mode: %s != %s", mode, "db_auth")
	}

	if _, err := LDAP(); err != nil {
		t.Fatalf("failed to get ldap settings: %v", err)
	}

	if _, err := TokenExpiration(); err != nil {
		t.Fatalf("failed to get token expiration: %v", err)
	}

	if _, err := ExtEndpoint(); err != nil {
		t.Fatalf("failed to get domain name: %v", err)
	}

	if _, err := SecretKey(); err != nil {
		t.Fatalf("failed to get secret key: %v", err)
	}

	if _, err := SelfRegistration(); err != nil {
		t.Fatalf("failed to get self registration: %v", err)
	}

	if _, err := RegistryURL(); err != nil {
		t.Fatalf("failed to get registry URL: %v", err)
	}

	if len(InternalJobServiceURL()) == 0 {
		t.Error("the internal job service url is null")
	}

	if len(InternalTokenServiceEndpoint()) == 0 {
		t.Error("the internal token service endpoint is null")
	}

	if _, err := InitialAdminPassword(); err != nil {
		t.Fatalf("failed to get initial admin password: %v", err)
	}

	if _, err := OnlyAdminCreateProject(); err != nil {
		t.Fatalf("failed to get onldy admin create project: %v", err)
	}

	if _, err := VerifyRemoteCert(); err != nil {
		t.Fatalf("failed to get verify remote cert: %v", err)
	}

	if _, err := Email(); err != nil {
		t.Fatalf("failed to get email settings: %v", err)
	}

	if _, err := Database(); err != nil {
		t.Fatalf("failed to get database: %v", err)
	}
	if InternalNotaryEndpoint() != "http://notary-server:4443" {
		t.Errorf("Unexpected notary endpoint: %s", InternalNotaryEndpoint())
	}
	if WithNotary() {
		t.Errorf("Withnotary should be false")
	}
	if !WithAdmiral() {
		t.Errorf("WithAdmiral should be true")
	}
	if AdmiralEndpoint() != "http://www.vmware.com" {
		t.Errorf("Unexpected admiral endpoint: %s", AdmiralEndpoint())
	}

	extURL, err := ExtURL()
	if err != nil {
		t.Errorf("Unexpected error getting external URL: %v", err)
	}
	if extURL != "host01.com" {
		t.Errorf(`extURL should be "host01.com".`)
	}

	// reset configurations
	if err = Reset(); err != nil {
		t.Errorf("failed to reset configurations: %v", err)
		return
	}
	mode, err = AuthMode()
	if err != nil {
		t.Fatalf("failed to get auth mode: %v", err)
	}
	if mode != "db_auth" {
		t.Errorf("unexpected mode: %s != %s", mode, "db_auth")
	}
}
