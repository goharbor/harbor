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

func TestResetConfig(t *testing.T) {
	fmt.Println("Testing resetting configurations")
	assert := assert.New(t)
	apiTest := newHarborAPI()

	code, err := apiTest.ResetConfig(*admin)
	if err != nil {
		t.Errorf("failed to get configurations: %v", err)
		return
	}

	if !assert.Equal(200, code, "unexpected response code") {
		return
	}

	code, cfgs, err := apiTest.GetConfig(*admin)
	if err != nil {
		t.Errorf("failed to get configurations: %v", err)
		return
	}

	if !assert.Equal(200, code, "unexpected response code") {
		return
	}

	value, ok := cfgs[common.TokenExpiration]
	if !ok {
		t.Errorf("%s not found", common.TokenExpiration)
		return
	}

	assert.Equal(int(value.Value.(float64)), 30, "unexpected 30")

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

	// length is 512，expected code: 200
	cfg := map[string]interface{}{
		common.LDAPGroupSearchFilter: "OSWvgTrQJuhiPRZt7eCReNku29vrtMBBD2cZt6jl7LQN4OZQcirqEhS2vCnkW8X1OAHMJxiO1LyEY26j" +
			"YhBEiUFliPKDUt8Q9endowT3H60nJibEnCkSRVjix7QujXKRzlmvxcOK76v1oZAoWeHSwwtv7tZrOk16Jj5LTGYdLOnZd2LIgBniTKmceL" +
			"VY5WOgcpmgQCfI5HLbzWsmAqmFfbsDbadirrEDiXYYfZQ0LnF8s6sD4H13eImgenAumXEsBRH43FT37AbNXNxzlaSs8IQYEdPLaMyKoXFb" +
			"rfa0LPipwXnU7bl54IlWOTXwCwum0JGS4qBiMl6LwKUBle34ObZ9fTLh5dFOVE1GdzrGE0kQ7qUmYjMZafQbSXzV80zTc22aZt3RQa9Gxt" +
			"Dn2VqtgcoKAiZHkEySStiwOJtZpwuplyy1jcM3DcN0R9b8IidYAWOsriqetUBThqb75XIZTXAaRWhHLw4ayROYiaw8dPuLRjeVKhdyznqq" +
			"AKxQGyvm",
	}
	code200, _ := apiTest.PutConfig(*admin, cfg)
	assert.Equal(200, code200, "the status code of modifying configurations with admin user should be 200")

	// length is 1059，expected code: 500
	cfg = map[string]interface{}{
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
	code500, _ := apiTest.PutConfig(*admin, cfg)
	assert.Equal(500, code500, "the status code of modifying configurations with admin user should be 500")
}
