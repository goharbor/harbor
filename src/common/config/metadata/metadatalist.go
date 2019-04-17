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

package metadata

import "github.com/goharbor/harbor/src/common"

// Item - Configure item include default value, type, env name
type Item struct {
	// The Scope of this configuration item: eg: SystemScope, UserScope
	Scope string `json:"scope,omitempty"`
	// email, ldapbasic, ldapgroup, uaa settings, used to retieve configure items by group
	Group string `json:"group,omitempty"`
	// environment key to retrieves this value when initialize, for example: POSTGRESQL_HOST, only used for system settings, for user settings no EnvKey
	EnvKey string `json:"environment_key,omitempty"`
	// The default string value for this key
	DefaultValue string `json:"default_value,omitempty"`
	// The key for current configure settings in database or rest api
	Name string `json:"name,omitempty"`
	// It can be &IntType{}, &StringType{}, &BoolType{}, &PasswordType{}, &MapType{} etc, any type interface implementation
	ItemType Type
	// TODO: Clarify the usage of this attribute
	Editable bool `json:"editable,omitempty"`
}

// Constant for configure item
const (
	// Scope
	UserScope   = "user"
	SystemScope = "system"
	// Group
	LdapBasicGroup = "ldapbasic"
	LdapGroupGroup = "ldapgroup"
	EmailGroup     = "email"
	UAAGroup       = "uaa"
	HTTPAuthGroup  = "http_auth"
	OIDCGroup      = "oidc"
	DatabaseGroup  = "database"
	// Put all config items do not belong a existing group into basic
	BasicGroup = "basic"
	ClairGroup = "clair"
)

var (
	// ConfigList - All configure items used in harbor
	// Steps to onboard a new setting
	// 1. Add configure item in metadatalist.go
	// 2. Get/Set config settings by CfgManager
	// 3. CfgManager.Load()/CfgManager.Save() to load/save from configure storage.
	ConfigList = []Item{
		// TODO: All these Name should be reference to const, see #7040
		{Name: "admin_initial_password", Scope: SystemScope, Group: BasicGroup, EnvKey: "HARBOR_ADMIN_PASSWORD", DefaultValue: "", ItemType: &PasswordType{}, Editable: true},
		{Name: "admiral_url", Scope: SystemScope, Group: BasicGroup, EnvKey: "ADMIRAL_URL", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "auth_mode", Scope: UserScope, Group: BasicGroup, EnvKey: "AUTH_MODE", DefaultValue: "db_auth", ItemType: &AuthModeType{}, Editable: false},
		{Name: "cfg_expiration", Scope: SystemScope, Group: BasicGroup, EnvKey: "CFG_EXPIRATION", DefaultValue: "5", ItemType: &IntType{}, Editable: false},
		{Name: "chart_repository_url", Scope: SystemScope, Group: BasicGroup, EnvKey: "CHART_REPOSITORY_URL", DefaultValue: "http://chartmuseum:9999", ItemType: &StringType{}, Editable: false},

		{Name: "clair_db", Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB", DefaultValue: "postgres", ItemType: &StringType{}, Editable: false},
		{Name: "clair_db_host", Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_HOST", DefaultValue: "postgresql", ItemType: &StringType{}, Editable: false},
		{Name: "clair_db_password", Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_PASSWORD", DefaultValue: "root123", ItemType: &PasswordType{}, Editable: false},
		{Name: "clair_db_port", Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_PORT", DefaultValue: "5432", ItemType: &PortType{}, Editable: false},
		{Name: "clair_db_sslmode", Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_SSLMODE", DefaultValue: "disable", ItemType: &StringType{}, Editable: false},
		{Name: "clair_db_username", Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_USERNAME", DefaultValue: "postgres", ItemType: &StringType{}, Editable: false},
		{Name: "clair_url", Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_URL", DefaultValue: "http://clair:6060", ItemType: &StringType{}, Editable: false},

		{Name: "core_url", Scope: SystemScope, Group: BasicGroup, EnvKey: "CORE_URL", DefaultValue: "http://core:8080", ItemType: &StringType{}, Editable: false},
		{Name: "database_type", Scope: SystemScope, Group: BasicGroup, EnvKey: "DATABASE_TYPE", DefaultValue: "postgresql", ItemType: &StringType{}, Editable: false},

		{Name: "email_from", Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_FROM", DefaultValue: "admin <sample_admin@mydomain.com>", ItemType: &StringType{}, Editable: false},
		{Name: "email_host", Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_HOST", DefaultValue: "smtp.mydomain.com", ItemType: &StringType{}, Editable: false},
		{Name: "email_identity", Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_IDENTITY", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "email_insecure", Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_INSECURE", DefaultValue: "false", ItemType: &BoolType{}, Editable: false},
		{Name: "email_password", Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_PWD", DefaultValue: "", ItemType: &PasswordType{}, Editable: false},
		{Name: "email_port", Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_PORT", DefaultValue: "25", ItemType: &PortType{}, Editable: false},
		{Name: "email_ssl", Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_SSL", DefaultValue: "false", ItemType: &BoolType{}, Editable: false},
		{Name: "email_username", Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_USR", DefaultValue: "sample_admin@mydomain.com", ItemType: &StringType{}, Editable: false},

		{Name: "ext_endpoint", Scope: SystemScope, Group: BasicGroup, EnvKey: "EXT_ENDPOINT", DefaultValue: "https://host01.com", ItemType: &StringType{}, Editable: false},
		{Name: "jobservice_url", Scope: SystemScope, Group: BasicGroup, EnvKey: "JOBSERVICE_URL", DefaultValue: "http://jobservice:8080", ItemType: &StringType{}, Editable: false},

		{Name: "ldap_base_dn", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_BASE_DN", DefaultValue: "", ItemType: &NonEmptyStringType{}, Editable: false},
		{Name: "ldap_filter", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_FILTER", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "ldap_group_base_dn", Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_BASE_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "ldap_group_admin_dn", Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_ADMIN_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "ldap_group_attribute_name", Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_GID", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "ldap_group_search_filter", Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_FILTER", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "ldap_group_search_scope", Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_SCOPE", DefaultValue: "2", ItemType: &LdapScopeType{}, Editable: false},
		{Name: "ldap_scope", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SCOPE", DefaultValue: "2", ItemType: &LdapScopeType{}, Editable: false},
		{Name: "ldap_search_dn", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SEARCH_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "ldap_search_password", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SEARCH_PWD", DefaultValue: "", ItemType: &PasswordType{}, Editable: false},
		{Name: "ldap_timeout", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_TIMEOUT", DefaultValue: "5", ItemType: &IntType{}, Editable: false},
		{Name: "ldap_uid", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_UID", DefaultValue: "cn", ItemType: &NonEmptyStringType{}, Editable: false},
		{Name: "ldap_url", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_URL", DefaultValue: "", ItemType: &NonEmptyStringType{}, Editable: false},
		{Name: "ldap_verify_cert", Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_VERIFY_CERT", DefaultValue: "true", ItemType: &BoolType{}, Editable: false},
		{Name: common.LDAPGroupMembershipAttribute, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_GROUP_MEMBERSHIP_ATTRIBUTE", DefaultValue: "memberof", ItemType: &StringType{}, Editable: true},

		{Name: "max_job_workers", Scope: SystemScope, Group: BasicGroup, EnvKey: "MAX_JOB_WORKERS", DefaultValue: "10", ItemType: &IntType{}, Editable: false},
		{Name: "notary_url", Scope: SystemScope, Group: BasicGroup, EnvKey: "NOTARY_URL", DefaultValue: "http://notary-server:4443", ItemType: &StringType{}, Editable: false},
		{Name: "scan_all_policy", Scope: UserScope, Group: BasicGroup, EnvKey: "", DefaultValue: "", ItemType: &MapType{}, Editable: false},

		{Name: "postgresql_database", Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_DATABASE", DefaultValue: "registry", ItemType: &StringType{}, Editable: false},
		{Name: "postgresql_host", Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_HOST", DefaultValue: "postgresql", ItemType: &StringType{}, Editable: false},
		{Name: "postgresql_password", Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_PASSWORD", DefaultValue: "root123", ItemType: &PasswordType{}, Editable: false},
		{Name: "postgresql_port", Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_PORT", DefaultValue: "5432", ItemType: &PortType{}, Editable: false},
		{Name: "postgresql_sslmode", Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_SSLMODE", DefaultValue: "disable", ItemType: &StringType{}, Editable: false},
		{Name: "postgresql_username", Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_USERNAME", DefaultValue: "postgres", ItemType: &StringType{}, Editable: false},

		{Name: "project_creation_restriction", Scope: UserScope, Group: BasicGroup, EnvKey: "PROJECT_CREATION_RESTRICTION", DefaultValue: common.ProCrtRestrEveryone, ItemType: &ProjectCreationRestrictionType{}, Editable: false},
		{Name: "read_only", Scope: UserScope, Group: BasicGroup, EnvKey: "READ_ONLY", DefaultValue: "false", ItemType: &BoolType{}, Editable: false},

		{Name: "registry_storage_provider_name", Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_STORAGE_PROVIDER_NAME", DefaultValue: "filesystem", ItemType: &StringType{}, Editable: false},
		{Name: "registry_url", Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_URL", DefaultValue: "http://registry:5000", ItemType: &StringType{}, Editable: false},
		{Name: "registry_controller_url", Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_CONTROLLER_URL", DefaultValue: "http://registryctl:8080", ItemType: &StringType{}, Editable: false},
		{Name: "self_registration", Scope: UserScope, Group: BasicGroup, EnvKey: "SELF_REGISTRATION", DefaultValue: "true", ItemType: &BoolType{}, Editable: false},
		{Name: "token_expiration", Scope: UserScope, Group: BasicGroup, EnvKey: "TOKEN_EXPIRATION", DefaultValue: "30", ItemType: &IntType{}, Editable: false},
		{Name: "token_service_url", Scope: SystemScope, Group: BasicGroup, EnvKey: "TOKEN_SERVICE_URL", DefaultValue: "http://core:8080/service/token", ItemType: &StringType{}, Editable: false},

		{Name: "uaa_client_id", Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_CLIENTID", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "uaa_client_secret", Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_CLIENTSECRET", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "uaa_endpoint", Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_ENDPOINT", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: "uaa_verify_cert", Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_VERIFY_CERT", DefaultValue: "false", ItemType: &BoolType{}, Editable: false},

		{Name: common.HTTPAuthProxyEndpoint, Scope: UserScope, Group: HTTPAuthGroup, ItemType: &StringType{}},
		{Name: common.HTTPAuthProxyTokenReviewEndpoint, Scope: UserScope, Group: HTTPAuthGroup, ItemType: &StringType{}},
		{Name: common.HTTPAuthProxyVerifyCert, Scope: UserScope, Group: HTTPAuthGroup, DefaultValue: "true", ItemType: &BoolType{}},
		{Name: common.HTTPAuthProxyAlwaysOnboard, Scope: UserScope, Group: HTTPAuthGroup, DefaultValue: "false", ItemType: &BoolType{}},

		{Name: common.OIDCName, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}},
		{Name: common.OIDCEndpoint, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}},
		{Name: common.OIDCCLientID, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}},
		{Name: common.OIDCClientSecret, Scope: UserScope, Group: OIDCGroup, ItemType: &PasswordType{}},
		{Name: common.OIDCScope, Scope: UserScope, Group: OIDCGroup, ItemType: &StringType{}},
		{Name: common.OIDCVerifyCert, Scope: UserScope, Group: OIDCGroup, DefaultValue: "true", ItemType: &BoolType{}},

		{Name: "with_chartmuseum", Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_CHARTMUSEUM", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		{Name: "with_clair", Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_CLAIR", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		{Name: "with_notary", Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_NOTARY", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		// the unit of expiration is minute, 43200 minutes = 30 days
		{Name: "robot_token_duration", Scope: UserScope, Group: BasicGroup, EnvKey: "ROBOT_TOKEN_DURATION", DefaultValue: "43200", ItemType: &IntType{}, Editable: true},
	}
)
