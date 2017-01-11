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

package models

// LdapConf holds information about repository that accessed most
type LdapConf struct {
	LdapURL            string `json:"ldap_url"`
	LdapSearchDn       string `json:"ldap_searchdn"`
	LdapSearchPwd      string `json:"ldap_search_pwd"`
	LdapBaseDn         string `json:"ldap_basedn"`
	LdapFilter         string `json:"ldap_filter"`
	LdapUID            string `json:"ldap_uid"`
	LdapScope          int    `json:"ldap_scope"`
	LdapConnectTimeout int    `json:"ldap_connect_timeout"`
}
