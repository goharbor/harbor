// Copyright 2018 Project Harbor Authors
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
package api

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVolumeInfo(t *testing.T) {
	fmt.Println("Testing Get Volume Info")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: get volume info without admin role
	CommonAddUser()
	code, _, err := apiTest.VolumeInfoGet(*testUser)
	if err != nil {
		t.Error("Error occurred while get system volume info")
		t.Log(err)
	} else {
		assert.Equal(403, code, "Get system volume info should be 403")
	}
	// case 2: get volume info with admin role
	code, info, err := apiTest.VolumeInfoGet(*admin)
	if err != nil {
		t.Error("Error occurred while get system volume info")
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get system volume info should be 200")
		if info.HarborStorage.Total <= 0 {
			assert.Equal(1, info.HarborStorage.Total, "Total storage of system should be larger than 0")
		}
		if info.HarborStorage.Free <= 0 {
			assert.Equal(1, info.HarborStorage.Free, "Free storage of system should be larger than 0")
		}
	}

}

func TestGetGeneralInfo(t *testing.T) {
	apiTest := newHarborAPI()
	code, body, err := apiTest.GetGeneralInfo()
	assert := assert.New(t)
	assert.Nil(err, fmt.Sprintf("Unexpected Error: %v", err))
	assert.Equal(200, code, fmt.Sprintf("Unexpected status code: %d", code))
	g := &GeneralInfo{}
	err = json.Unmarshal(body, g)
	assert.Nil(err, fmt.Sprintf("Unexpected Error: %v", err))
	assert.Equal(false, g.WithNotary, "with notary should be false")
	assert.Equal(true, g.HasCARoot, "has ca root should be true")
	assert.NotEmpty(g.HarborVersion, "harbor version should not be empty")
	assert.Equal(false, g.ReadOnly, "readonly should be false")
}

func TestGetCert(t *testing.T) {
	fmt.Println("Testing Get Cert")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: get cert without admin role
	code, _, err := apiTest.CertGet(*testUser)
	if err != nil {
		t.Error("Error occurred while get system cert")
		t.Log(err)
	} else {
		assert.Equal(403, code, "Get system cert should be 403")
	}
	// case 2: get cert with admin role
	code, content, err := apiTest.CertGet(*admin)
	if err != nil {
		t.Error("Error occurred while get system cert")
		t.Log(err)
	} else {
		assert.Equal(200, code, "Get system cert should be 200")
		assert.Equal("test for ca.crt.\n", string(content), "Get system cert content should be equal")

	}
	CommonDelUser()
}
func TestPing(t *testing.T) {
	apiTest := newHarborAPI()
	code, _, err := apiTest.Ping()
	assert := assert.New(t)
	assert.Nil(err, fmt.Sprintf("Unexpected Error: %v", err))
	assert.Equal(200, code, fmt.Sprintf("Unexpected status code: %d", code))
}
