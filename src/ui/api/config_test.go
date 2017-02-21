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

package api

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common/config"
)

func TestGetConfig(t *testing.T) {
	fmt.Println("Testing getting configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	//case 1: get configurations without admin role
	code, _, err := apiTest.GetConfig(*testUser)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	assert.Equal(401, code, "the status code of getting configurations with non-admin user should be 401")

	//case 2: get configurations with admin role
	code, cfg, err := apiTest.GetConfig(*admin)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of getting configurations with admin user should be 200") {
		return
	}

	mode := cfg[config.AUTHMode].Value.(string)
	assert.Equal(config.DBAuth, mode, fmt.Sprintf("the auth mode should be %s", config.DBAuth))
}

func TestPutConfig(t *testing.T) {
	fmt.Println("Testing modifying configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	cfg := map[string]string{
		config.VerifyRemoteCert: "0",
	}

	code, err := apiTest.PutConfig(*admin, cfg)
	if err != nil {
		t.Fatalf("failed to modify configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of modifying configurations with admin user should be 200") {
		return
	}
}

func TestResetConfig(t *testing.T) {
	fmt.Println("Testing resetting configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	m := map[string]string{
		config.VerifyRemoteCert: "0",
	}

	// modify configurations
	code, err := apiTest.PutConfig(*admin, m)
	if err != nil {
		t.Fatalf("failed to modify configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of modifying configurations with admin user should be 200") {
		return
	}

	//reset configurations
	code, err = apiTest.ResetConfig(*admin)
	if err != nil {
		t.Fatalf("failed to reset configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of resetting configurations with admin user should be 200") {
		return
	}

	// verify result
	code, cfg, err := apiTest.GetConfig(*admin)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of getting configurations with admin user should be 200") {
		return
	}

	verify := cfg[config.VerifyRemoteCert].Value.(bool)
	assert.Equal(true, verify, fmt.Sprint("the verify_remote_cert should be true"))
}
