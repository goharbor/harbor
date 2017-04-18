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

package common

// const variables
const (
	DBAuth              = "db_auth"
	LDAPAuth            = "ldap_auth"
	ProCrtRestrEveryone = "everyone"
	ProCrtRestrAdmOnly  = "adminonly"
	LDAPScopeBase       = "1"
	LDAPScopeOnelevel   = "2"
	LDAPScopeSubtree    = "3"

	ExtEndpoint                = "ext_endpoint"
	AUTHMode                   = "auth_mode"
	DatabaseType               = "database_type"
	MySQLHost                  = "mysql_host"
	MySQLPort                  = "mysql_port"
	MySQLUsername              = "mysql_username"
	MySQLPassword              = "mysql_password"
	MySQLDatabase              = "mysql_database"
	SQLiteFile                 = "sqlite_file"
	SelfRegistration           = "self_registration"
	LDAPURL                    = "ldap_url"
	LDAPSearchDN               = "ldap_search_dn"
	LDAPSearchPwd              = "ldap_search_password"
	LDAPBaseDN                 = "ldap_base_dn"
	LDAPUID                    = "ldap_uid"
	LDAPFilter                 = "ldap_filter"
	LDAPScope                  = "ldap_scope"
	LDAPTimeout                = "ldap_timeout"
	TokenServiceURL            = "token_service_url"
	RegistryURL                = "registry_url"
	EmailHost                  = "email_host"
	EmailPort                  = "email_port"
	EmailUsername              = "email_username"
	EmailPassword              = "email_password"
	EmailFrom                  = "email_from"
	EmailSSL                   = "email_ssl"
	EmailIdentity              = "email_identity"
	ProjectCreationRestriction = "project_creation_restriction"
	VerifyRemoteCert           = "verify_remote_cert"
	MaxJobWorkers              = "max_job_workers"
	TokenExpiration            = "token_expiration"
	CfgExpiration              = "cfg_expiration"
	JobLogDir                  = "job_log_dir"
	AdminInitialPassword       = "admin_initial_password"
	AdmiralEndpoint            = "admiral_url"
	WithNotary                 = "with_notary"
)
