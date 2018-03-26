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
	UAAAuth             = "uaa_auth"
	ProCrtRestrEveryone = "everyone"
	ProCrtRestrAdmOnly  = "adminonly"
	LDAPScopeBase       = 0
	LDAPScopeOnelevel   = 1
	LDAPScopeSubtree    = 2

	RoleProjectAdmin = 1
	RoleDeveloper    = 2
	RoleGuest        = 3

	LabelLevelSystem  = "s"
	LabelLevelUser    = "u"
	LabelScopeGlobal  = "g"
	LabelScopeProject = "p"

	ResourceTypeProject    = "p"
	ResourceTypeRepository = "r"
	ResourceTypeImage      = "i"

	ExtEndpoint                 = "ext_endpoint"
	AUTHMode                    = "auth_mode"
	DatabaseType                = "database_type"
	MySQLHost                   = "mysql_host"
	MySQLPort                   = "mysql_port"
	MySQLUsername               = "mysql_username"
	MySQLPassword               = "mysql_password"
	MySQLDatabase               = "mysql_database"
	SQLiteFile                  = "sqlite_file"
	SelfRegistration            = "self_registration"
	UIURL                       = "ui_url"
	JobServiceURL               = "jobservice_url"
	LDAPURL                     = "ldap_url"
	LDAPSearchDN                = "ldap_search_dn"
	LDAPSearchPwd               = "ldap_search_password"
	LDAPBaseDN                  = "ldap_base_dn"
	LDAPUID                     = "ldap_uid"
	LDAPFilter                  = "ldap_filter"
	LDAPScope                   = "ldap_scope"
	LDAPTimeout                 = "ldap_timeout"
	LDAPVerifyCert              = "ldap_verify_cert"
	LDAPGroupBaseDN             = "ldap_group_base_dn"
	LDAPGroupSearchFilter       = "ldap_group_search_filter"
	LDAPGroupAttributeName      = "ldap_group_attribute_name"
	LDAPGroupSearchScope        = "ldap_group_search_scope"
	TokenServiceURL             = "token_service_url"
	RegistryURL                 = "registry_url"
	EmailHost                   = "email_host"
	EmailPort                   = "email_port"
	EmailUsername               = "email_username"
	EmailPassword               = "email_password"
	EmailFrom                   = "email_from"
	EmailSSL                    = "email_ssl"
	EmailIdentity               = "email_identity"
	EmailInsecure               = "email_insecure"
	ProjectCreationRestriction  = "project_creation_restriction"
	MaxJobWorkers               = "max_job_workers"
	TokenExpiration             = "token_expiration"
	CfgExpiration               = "cfg_expiration"
	JobLogDir                   = "job_log_dir"
	AdminInitialPassword        = "admin_initial_password"
	AdmiralEndpoint             = "admiral_url"
	WithNotary                  = "with_notary"
	WithClair                   = "with_clair"
	ScanAllPolicy               = "scan_all_policy"
	ClairDBPassword             = "clair_db_password"
	ClairDBHost                 = "clair_db_host"
	ClairDBPort                 = "clair_db_port"
	ClairDB                     = "clair_db"
	ClairDBUsername             = "clair_db_username"
	UAAEndpoint                 = "uaa_endpoint"
	UAAClientID                 = "uaa_client_id"
	UAAClientSecret             = "uaa_client_secret"
	UAAVerifyCert               = "uaa_verify_cert"
	DefaultClairEndpoint        = "http://clair:6060"
	CfgDriverDB                 = "db"
	CfgDriverJSON               = "json"
	NewHarborAdminName          = "admin@harbor.local"
	RegistryStorageProviderName = "registry_storage_provider_name"
	UserMember                  = "u"
	GroupMember                 = "g"
	ReadOnly                    = "read_only"
	ClairURL                    = "clair_url"
	NotaryURL                   = "notary_url"
	DefaultAdminserverEndpoint  = "http://adminserver:8080"
	DefaultJobserviceEndpoint   = "http://jobservice:8080"
	DefaultUIEndpoint           = "http://ui:8080"
	DefaultNotaryEndpoint       = "http://notary-server:4443"
	LdapGroupType               = 1
)
