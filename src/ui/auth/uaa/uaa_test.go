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

package uaa

import (
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common/dao"
	utilstest "github.com/vmware/harbor/src/common/utils/test"
	uaatest "github.com/vmware/harbor/src/common/utils/uaa/test"
	"github.com/vmware/harbor/src/ui/config"

	"os"
	"testing"
)

func TestGetClient(t *testing.T) {
	assert := assert.New(t)
	server, err := utilstest.NewAdminserver(nil)
	if err != nil {
		t.Fatalf("failed to create a mock admin server: %v", err)
	}
	defer server.Close()

	if err := os.Setenv("ADMINSERVER_URL", server.URL); err != nil {
		t.Fatalf("failed to set env %s: %v", "ADMINSERVER_URL", err)
	}
	err = config.Init()
	if err != nil {
		t.Fatalf("failed to init config: %v", err)
	}
	c, err := GetClient()
	assert.Nil(err)
	assert.NotNil(c)
}

func TestDoAuth(t *testing.T) {
	assert := assert.New(t)
	client := &uaatest.FakeClient{
		Username: "user1",
		Password: "password1",
	}
	dao.PrepareTestForMySQL()
	u1, err1 := doAuth("user1", "password1", client)
	assert.Nil(err1)
	assert.True(u1.UserID > 0)
	u2, err2 := doAuth("wrong", "wrong", client)
	assert.NotNil(err2)
	assert.Nil(u2)
}
