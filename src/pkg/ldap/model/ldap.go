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

// LdapConf holds information about ldap configuration
type LdapConf struct {
	URL               string `json:"ldap_url"`
	SearchDn          string `json:"ldap_search_dn"`
	SearchPassword    string `json:"ldap_search_password"`
	BaseDn            string `json:"ldap_base_dn"`
	Filter            string `json:"ldap_filter"`
	UID               string `json:"ldap_uid"`
	Scope             int    `json:"ldap_scope"`
	ConnectionTimeout int    `json:"ldap_connection_timeout"`
	VerifyCert        bool   `json:"ldap_verify_cert"`
}

// GroupConf holds information about ldap group
type GroupConf struct {
	BaseDN              string `json:"ldap_group_base_dn,omitempty"`
	Filter              string `json:"ldap_group_filter,omitempty"`
	NameAttribute       string `json:"ldap_group_name_attribute,omitempty"`
	SearchScope         int    `json:"ldap_group_search_scope"`
	AdminDN             string `json:"ldap_group_admin_dn,omitempty"`
	MembershipAttribute string `json:"ldap_group_membership_attribute,omitempty"`
}

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
