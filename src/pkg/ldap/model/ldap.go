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

package model

// User ...
type User struct {
	Username    string   `json:"ldap_username"`
	Email       string   `json:"ldap_email"`
	Realname    string   `json:"ldap_realname"`
	DN          string   `json:"-"`
	GroupDNList []string `json:"ldap_groupdn"`
}

// ImportUser ...
type ImportUser struct {
	UIDList []string `json:"ldap_uid_list"`
}

// FailedImportUser ...
type FailedImportUser struct {
	UID   string `json:"uid"`
	Error string `json:"err_msg"`
}

// Group ...
type Group struct {
	Name string `json:"group_name,omitempty"`
	Dn   string `json:"ldap_group_dn,omitempty"`
}
