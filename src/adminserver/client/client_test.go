// Copyright Project Harbor Authors
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

package client

import (
	"fmt"
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/stretchr/testify/assert"
)

var c Client

func TestMain(m *testing.M) {

	server, err := test.NewAdminserver(nil)
	if err != nil {
		fmt.Printf("failed to create adminserver: %v", err)
		os.Exit(1)
	}

	c = NewClient(server.URL, &Config{})

	os.Exit(m.Run())
}

func TestPing(t *testing.T) {
	err := c.Ping()
	assert.Nil(t, err, "unexpected error")
}

func TestGetCfgs(t *testing.T) {
	cfgs, err := c.GetCfgs()
	if !assert.Nil(t, err, "unexpected error") {
		return
	}

	assert.Equal(t, common.DBAuth, cfgs[common.AUTHMode], "unexpected configuration")
}

func TestUpdateCfgs(t *testing.T) {
	cfgs := map[string]interface{}{
		common.AUTHMode: common.LDAPAuth,
	}
	err := c.UpdateCfgs(cfgs)
	if !assert.Nil(t, err, "unexpected error") {
		return
	}
}

func TestResetCfgs(t *testing.T) {
	err := c.ResetCfgs()
	if !assert.Nil(t, err, "unexpected error") {
		return
	}
}
