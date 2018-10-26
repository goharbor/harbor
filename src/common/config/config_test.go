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
package config

import "testing"

// the functions in common/config/config.go have been tested
// by cases in UI and Jobservice

func TestInitMetaData(t *testing.T) {

	MetaData.InitMetaData()
	if item, err := MetaData.GetConfigMetaData("ldap_search_base_dn"); err != nil {
		t.Error("Failed to find key ldap_search_base_dn after initial")
	} else {
		if item.Type != StringType {
			t.Error("Wrong Type for this item!")
		}
	}
	if item, err := MetaData.GetConfigMetaData("ldap_search_scope"); err != nil {
		t.Error("Failed to find key ldap_search_scope after initial")
	} else {
		if item.Type != IntType {
			t.Error("Wrong Type for this item!")
		}
	}
	if item, err := MetaData.GetConfigMetaData("ldap_search_password"); err != nil {
		t.Error("Failed to find key ldap_search_password after initial")
	} else {
		if item.Type != PasswordType {
			t.Error("Wrong Type for this item!")
		}
	}
	if item, err := MetaData.GetConfigMetaData("ldap_verify_cert"); err != nil {
		t.Error("Failed to find key ldap_verify_cert after initial")
	} else {
		if item.Type != BoolType {
			t.Error("Wrong Type for this item!")
		}
	}

}
