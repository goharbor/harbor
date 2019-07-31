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

		{Name: common.AdminInitialPassword, Scope: SystemScope, Group: BasicGroup, EnvKey: "HARBOR_ADMIN_PASSWORD", DefaultValue: "", ItemType: &PasswordType{}, Editable: true},
		{Name: common.AdmiralEndpoint, Scope: SystemScope, Group: BasicGroup, EnvKey: "ADMIRAL_URL", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.AUTHMode, Scope: UserScope, Group: BasicGroup, EnvKey: "AUTH_MODE", DefaultValue: "db_auth", ItemType: &AuthModeType{}, Editable: false},
		{Name: common.ChartRepoURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "CHART_REPOSITORY_URL", DefaultValue: "http://chartmuseum:9999", ItemType: &StringType{}, Editable: false},

		{Name: common.ClairDB, Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB", DefaultValue: "postgres", ItemType: &StringType{}, Editable: false},
		{Name: common.ClairDBHost, Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_HOST", DefaultValue: "postgresql", ItemType: &StringType{}, Editable: false},
		{Name: common.ClairDBPassword, Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_PASSWORD", DefaultValue: "root123", ItemType: &PasswordType{}, Editable: false},
		{Name: common.ClairDBPort, Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_PORT", DefaultValue: "5432", ItemType: &PortType{}, Editable: false},
		{Name: common.ClairDBSSLMode, Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_SSLMODE", DefaultValue: "disable", ItemType: &StringType{}, Editable: false},
		{Name: common.ClairDBUsername, Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_DB_USERNAME", DefaultValue: "postgres", ItemType: &StringType{}, Editable: false},
		{Name: common.ClairURL, Scope: SystemScope, Group: ClairGroup, EnvKey: "CLAIR_URL", DefaultValue: "http://clair:6060", ItemType: &StringType{}, Editable: false},

		{Name: common.CoreURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "CORE_URL", DefaultValue: "http://core:8080", ItemType: &StringType{}, Editable: false},
		{Name: common.DatabaseType, Scope: SystemScope, Group: BasicGroup, EnvKey: "DATABASE_TYPE", DefaultValue: "postgresql", ItemType: &StringType{}, Editable: false},

		{Name: common.EmailFrom, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_FROM", DefaultValue: "admin <sample_admin@mydomain.com>", ItemType: &StringType{}, Editable: false},
		{Name: common.EmailHost, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_HOST", DefaultValue: "smtp.mydomain.com", ItemType: &StringType{}, Editable: false},
		{Name: common.EmailIdentity, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_IDENTITY", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.EmailInsecure, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_INSECURE", DefaultValue: "false", ItemType: &BoolType{}, Editable: false},
		{Name: common.EmailPassword, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_PWD", DefaultValue: "", ItemType: &PasswordType{}, Editable: false},
		{Name: common.EmailPort, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_PORT", DefaultValue: "25", ItemType: &PortType{}, Editable: false},
		{Name: common.EmailSSL, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_SSL", DefaultValue: "false", ItemType: &BoolType{}, Editable: false},
		{Name: common.EmailUsername, Scope: UserScope, Group: EmailGroup, EnvKey: "EMAIL_USR", DefaultValue: "sample_admin@mydomain.com", ItemType: &StringType{}, Editable: false},

		{Name: common.ExtEndpoint, Scope: SystemScope, Group: BasicGroup, EnvKey: "EXT_ENDPOINT", DefaultValue: "https://host01.com", ItemType: &StringType{}, Editable: false},
		{Name: common.JobServiceURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "JOBSERVICE_URL", DefaultValue: "http://jobservice:8080", ItemType: &StringType{}, Editable: false},

		{Name: common.LDAPBaseDN, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_BASE_DN", DefaultValue: "", ItemType: &NonEmptyStringType{}, Editable: false},
		{Name: common.LDAPFilter, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_FILTER", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.LDAPGroupBaseDN, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_BASE_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.LdapGroupAdminDn, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_ADMIN_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.LDAPGroupAttributeName, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_GID", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.LDAPGroupSearchFilter, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_FILTER", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.LDAPGroupSearchScope, Scope: UserScope, Group: LdapGroupGroup, EnvKey: "LDAP_GROUP_SCOPE", DefaultValue: "2", ItemType: &LdapScopeType{}, Editable: false},
		{Name: common.LDAPScope, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SCOPE", DefaultValue: "2", ItemType: &LdapScopeType{}, Editable: false},
		{Name: common.LDAPSearchDN, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SEARCH_DN", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.LDAPSearchPwd, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_SEARCH_PWD", DefaultValue: "", ItemType: &PasswordType{}, Editable: false},
		{Name: common.LDAPTimeout, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_TIMEOUT", DefaultValue: "5", ItemType: &IntType{}, Editable: false},
		{Name: common.LDAPUID, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_UID", DefaultValue: "cn", ItemType: &NonEmptyStringType{}, Editable: false},
		{Name: common.LDAPURL, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_URL", DefaultValue: "", ItemType: &NonEmptyStringType{}, Editable: false},
		{Name: common.LDAPVerifyCert, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_VERIFY_CERT", DefaultValue: "true", ItemType: &BoolType{}, Editable: false},
		{Name: common.LDAPGroupMembershipAttribute, Scope: UserScope, Group: LdapBasicGroup, EnvKey: "LDAP_GROUP_MEMBERSHIP_ATTRIBUTE", DefaultValue: "memberof", ItemType: &StringType{}, Editable: true},

		{Name: common.MaxJobWorkers, Scope: SystemScope, Group: BasicGroup, EnvKey: "MAX_JOB_WORKERS", DefaultValue: "10", ItemType: &IntType{}, Editable: false},
		{Name: common.NotaryURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "NOTARY_URL", DefaultValue: "http://notary-server:4443", ItemType: &StringType{}, Editable: false},
		{Name: common.ScanAllPolicy, Scope: UserScope, Group: BasicGroup, EnvKey: "", DefaultValue: "", ItemType: &MapType{}, Editable: false},

		{Name: common.PostGreSQLDatabase, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_DATABASE", DefaultValue: "registry", ItemType: &StringType{}, Editable: false},
		{Name: common.PostGreSQLHOST, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_HOST", DefaultValue: "postgresql", ItemType: &StringType{}, Editable: false},
		{Name: common.PostGreSQLPassword, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_PASSWORD", DefaultValue: "root123", ItemType: &PasswordType{}, Editable: false},
		{Name: common.PostGreSQLPort, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_PORT", DefaultValue: "5432", ItemType: &PortType{}, Editable: false},
		{Name: common.PostGreSQLSSLMode, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_SSLMODE", DefaultValue: "disable", ItemType: &StringType{}, Editable: false},
		{Name: common.PostGreSQLUsername, Scope: SystemScope, Group: DatabaseGroup, EnvKey: "POSTGRESQL_USERNAME", DefaultValue: "postgres", ItemType: &StringType{}, Editable: false},

		{Name: common.ProjectCreationRestriction, Scope: UserScope, Group: BasicGroup, EnvKey: "PROJECT_CREATION_RESTRICTION", DefaultValue: common.ProCrtRestrEveryone, ItemType: &ProjectCreationRestrictionType{}, Editable: false},
		{Name: common.ReadOnly, Scope: UserScope, Group: BasicGroup, EnvKey: "READ_ONLY", DefaultValue: "false", ItemType: &BoolType{}, Editable: false},

		{Name: common.RegistryStorageProviderName, Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_STORAGE_PROVIDER_NAME", DefaultValue: "filesystem", ItemType: &StringType{}, Editable: false},
		{Name: common.RegistryURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_URL", DefaultValue: "http://registry:5000", ItemType: &StringType{}, Editable: false},
		{Name: common.RegistryControllerURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "REGISTRY_CONTROLLER_URL", DefaultValue: "http://registryctl:8080", ItemType: &StringType{}, Editable: false},
		{Name: common.SelfRegistration, Scope: UserScope, Group: BasicGroup, EnvKey: "SELF_REGISTRATION", DefaultValue: "true", ItemType: &BoolType{}, Editable: false},
		{Name: common.TokenExpiration, Scope: UserScope, Group: BasicGroup, EnvKey: "TOKEN_EXPIRATION", DefaultValue: "30", ItemType: &IntType{}, Editable: false},
		{Name: common.TokenServiceURL, Scope: SystemScope, Group: BasicGroup, EnvKey: "TOKEN_SERVICE_URL", DefaultValue: "http://core:8080/service/token", ItemType: &StringType{}, Editable: false},

		{Name: common.UAAClientID, Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_CLIENTID", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.UAAClientSecret, Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_CLIENTSECRET", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.UAAEndpoint, Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_ENDPOINT", DefaultValue: "", ItemType: &StringType{}, Editable: false},
		{Name: common.UAAVerifyCert, Scope: UserScope, Group: UAAGroup, EnvKey: "UAA_VERIFY_CERT", DefaultValue: "false", ItemType: &BoolType{}, Editable: false},

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

		{Name: common.WithChartMuseum, Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_CHARTMUSEUM", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		{Name: common.WithClair, Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_CLAIR", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		{Name: common.WithNotary, Scope: SystemScope, Group: BasicGroup, EnvKey: "WITH_NOTARY", DefaultValue: "false", ItemType: &BoolType{}, Editable: true},
		// the unit of expiration is minute, 43200 minutes = 30 days
		{Name: common.RobotTokenDuration, Scope: UserScope, Group: BasicGroup, EnvKey: "ROBOT_TOKEN_DURATION", DefaultValue: "43200", ItemType: &IntType{}, Editable: true},
		{Name: common.NotificationEnable, Scope: UserScope, Group: BasicGroup, EnvKey: "NOTIFICATION_ENABLE", DefaultValue: "true", ItemType: &BoolType{}, Editable: true},
	}
)
