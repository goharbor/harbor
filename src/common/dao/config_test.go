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

package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
)

func TestAuthModeCanBeModified(t *testing.T) {
	c, err := GetOrmer().QueryTable(&models.User{}).Count()
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}

	if c == 2 {
		flag, err := AuthModeCanBeModified()
		if err != nil {
			t.Fatalf("failed to determine whether auth mode can be modified: %v", err)
		}
		if !flag {
			t.Errorf("unexpected result: %t != %t", flag, true)
		}

		user := models.User{
			Username: "user_for_config_test",
			Email:    "user_for_config_test@vmware.com",
			Password: "P@ssword",
			Realname: "user_for_config_test",
		}
		id, err := Register(user)
		if err != nil {
			t.Fatalf("failed to register user: %v", err)
		}
		defer func(id int64) {
			if err := CleanUser(id); err != nil {
				t.Fatalf("failed to delete user %d: %v", id, err)
			}
		}(id)

		flag, err = AuthModeCanBeModified()
		if err != nil {
			t.Fatalf("failed to determine whether auth mode can be modified: %v", err)
		}
		if flag {
			t.Errorf("unexpected result: %t != %t", flag, false)
		}

	} else {
		flag, err := AuthModeCanBeModified()
		if err != nil {
			t.Fatalf("failed to determine whether auth mode can be modified: %v", err)
		}
		if flag {
			t.Errorf("unexpected result: %t != %t", flag, false)
		}
	}
}
