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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	fmt.Println("Testing getting configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: get configurations without admin role
	code, _, err := apiTest.GetConfig(*testUser)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	assert.Equal(401, code, "the status code of getting configurations with non-admin user should be 401")

	// case 2: get configurations with admin role
	code, cfg, err := apiTest.GetConfig(*admin)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of getting configurations with admin user should be 200") {
		return
	}
	t.Logf("cfg: %+v", cfg)
	mode := cfg[common.AUTHMode].Value.(string)
	assert.Equal(common.DBAuth, mode, fmt.Sprintf("the auth mode should be %s", common.DBAuth))
	ccc, err := config.GetSystemCfg()
	if err != nil {
		t.Logf("failed to get system configurations: %v", err)
	}
	t.Logf("%v", ccc)
}

func TestInternalConfig(t *testing.T) {
	fmt.Println("Testing internal configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	// case 1: get configurations without admin role
	code, _, err := apiTest.GetInternalConfig(*testUser)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	assert.Equal(401, code, "the status code of getting configurations with non-admin user should be 401")

	// case 2: get configurations with admin role
	code, _, err = apiTest.GetInternalConfig(*admin)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of getting configurations with admin user should be 200") {
		return
	}
}

func TestPutConfig(t *testing.T) {
	fmt.Println("Testing modifying configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	cfg := map[string]interface{}{
		common.TokenExpiration: 60,
	}

	code, err := apiTest.PutConfig(*admin, cfg)
	if err != nil {
		t.Fatalf("failed to get configurations: %v", err)
	}

	if !assert.Equal(200, code, "the status code of modifying configurations with admin user should be 200") {
		return
	}
	ccc, err := config.GetSystemCfg()
	if err != nil {
		t.Logf("failed to get system configurations: %v", err)
	}
	t.Logf("%v", ccc)
}

func TestPutConfigMaxLength(t *testing.T) {
	fmt.Println("Testing modifying configurations with max length.")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	// length is 1059，expected code: 200
	cfg := map[string]interface{}{
		common.LDAPGroupSearchFilter: "YU2YcM13JtSx5jtBiftTjfaEM9KZFQ0XA5fKQHU02E9Xe0aLYaSy7YBokrTA8oHFjSkWFSgWZJ6FEmTS" +
			"Vy5Ovsy5to2kWnFtbVNX3pzbeQpZeAqK3mEGnXdMkMSQu9WTq74s99GpwjEdA628pcZqLx6wCR0IvwryqIcNoRtqPlUcuRGODWA8ZXaC0d" +
			"Qs7cRUYSe8onHsM2c9JWuUS8Jv4E7KggfytrxeKAT0WGP5DBZsB7rHZKxoAppE3C0NueEeC4yV791PUOODJt9rc0RrcD6ORUIO5RriCwym" +
			"IinJZa03MtTk3vGFTmL9wM0wEYZP3fEBmoiB0iF8o4wkHGyMpNJoDyPuo7huuCbipAXClEcX1R7xD4aijTF9iOMKymvsObMZ4qqI7flco5" +
			"yLFf7W8cpSisk3YJSvxDWfrl91WT4IFE5KHK976DgLQJhTZ8msGOImnFiUGtuIUNQpOgFFtlXJV41OltSsjW5jwAzxcko0MFkOIc7XuPjB" +
			"XMrdjC9poYldrxNFrGOPFSyh19iS2UWKayKrtnhvDYAWrNCqOmRs01awEXBlwHp17VcLuze6XGCx7ZoPQX1Nu4uF1InAGpSm1B3pKtteeR" +
			"WNNeLZjmNGNuiorHyxLTx1bQTfkG2UzZTTR0e2XatiXt5nCDxSqP2OkOxH7dew36fm9LpkFbmgtlxWxjHX8buYzSJCAjTqqwW3rHCEfQjv" +
			"B4T7CTJrAgehCG9zL82P59DQbGXXWqRHbw5g9QszREQys1m56SHLosNptVPUwy7vD70rRf5s8knohW5npEZS9f3RGel64mj5g7bQBBkopx" +
			"f6uac3MlJAe9d6C0B7fexZJABln2kCtXXYzITflICISwxuZ0YXHJmT2sMSIpn9VwMnMidV4JsM2BD8ykExZ5QyeVyOCXHDxvRvFwQwjQfR" +
			"kkqQmtFREitKWl5njO8wLJw0XyeIVAej75NsGKKZWVjyaupaM9Bqn6NFrWjELFacLox6OCcRIDSDl3ntNN8tIzGOF7aXVCIqljJl0IL9Pz" +
			"NenmmubaNm48YjfkBk8MqOUSYJYaFkO1qCKbVdMg7yTqKEHgSUqEkoFPoJMH6GAozC",
	}
	sc, _ := apiTest.PutConfig(*admin, cfg)
	assert.Equal(200, sc, "the status code of modifying configurations with admin user should be 200")
}
