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

package common

type contextKey string

// const variables
const (
	DBAuth              = "db_auth"
	LDAPAuth            = "ldap_auth"
	UAAAuth             = "uaa_auth"
	HTTPAuth            = "http_auth"
	OIDCAuth            = "oidc_auth"
	ProCrtRestrEveryone = "everyone"
	ProCrtRestrAdmOnly  = "adminonly"
	LDAPScopeBase       = 0
	LDAPScopeOnelevel   = 1
	LDAPScopeSubtree    = 2

	RoleProjectAdmin = 1
	RoleDeveloper    = 2
	RoleGuest        = 3
	RoleMaster       = 4

	LabelLevelSystem  = "s"
	LabelLevelUser    = "u"
	LabelScopeGlobal  = "g"
	LabelScopeProject = "p"

	ResourceTypeProject    = "p"
	ResourceTypeRepository = "r"
	ResourceTypeImage      = "i"
	ResourceTypeChart      = "c"

	ExtEndpoint                      = "ext_endpoint"
	AUTHMode                         = "auth_mode"
	DatabaseType                     = "database_type"
	PostGreSQLHOST                   = "postgresql_host"
	PostGreSQLPort                   = "postgresql_port"
	PostGreSQLUsername               = "postgresql_username"
	PostGreSQLPassword               = "postgresql_password"
	PostGreSQLDatabase               = "postgresql_database"
	PostGreSQLSSLMode                = "postgresql_sslmode"
	SelfRegistration                 = "self_registration"
	CoreURL                          = "core_url"
	JobServiceURL                    = "jobservice_url"
	LDAPURL                          = "ldap_url"
	LDAPSearchDN                     = "ldap_search_dn"
	LDAPSearchPwd                    = "ldap_search_password"
	LDAPBaseDN                       = "ldap_base_dn"
	LDAPUID                          = "ldap_uid"
	LDAPFilter                       = "ldap_filter"
	LDAPScope                        = "ldap_scope"
	LDAPTimeout                      = "ldap_timeout"
	LDAPVerifyCert                   = "ldap_verify_cert"
	LDAPGroupBaseDN                  = "ldap_group_base_dn"
	LDAPGroupSearchFilter            = "ldap_group_search_filter"
	LDAPGroupAttributeName           = "ldap_group_attribute_name"
	LDAPGroupSearchScope             = "ldap_group_search_scope"
	TokenServiceURL                  = "token_service_url"
	RegistryURL                      = "registry_url"
	EmailHost                        = "email_host"
	EmailPort                        = "email_port"
	EmailUsername                    = "email_username"
	EmailPassword                    = "email_password"
	EmailFrom                        = "email_from"
	EmailSSL                         = "email_ssl"
	EmailIdentity                    = "email_identity"
	EmailInsecure                    = "email_insecure"
	ProjectCreationRestriction       = "project_creation_restriction"
	MaxJobWorkers                    = "max_job_workers"
	TokenExpiration                  = "token_expiration"
	CfgExpiration                    = "cfg_expiration"
	AdminInitialPassword             = "admin_initial_password"
	AdmiralEndpoint                  = "admiral_url"
	WithNotary                       = "with_notary"
	WithClair                        = "with_clair"
	ScanAllPolicy                    = "scan_all_policy"
	ClairDBPassword                  = "clair_db_password"
	ClairDBHost                      = "clair_db_host"
	ClairDBPort                      = "clair_db_port"
	ClairDB                          = "clair_db"
	ClairDBUsername                  = "clair_db_username"
	ClairDBSSLMode                   = "clair_db_sslmode"
	UAAEndpoint                      = "uaa_endpoint"
	UAAClientID                      = "uaa_client_id"
	UAAClientSecret                  = "uaa_client_secret"
	UAAVerifyCert                    = "uaa_verify_cert"
	HTTPAuthProxyEndpoint            = "http_authproxy_endpoint"
	HTTPAuthProxyTokenReviewEndpoint = "http_authproxy_tokenreview_endpoint"
	HTTPAuthProxyVerifyCert          = "http_authproxy_verify_cert"
	HTTPAuthProxyAlwaysOnboard       = "http_authproxy_always_onboard"
	OIDCName                         = "oidc_name"
	OIDCEndpoint                     = "oidc_endpoint"
	OIDCCLientID                     = "oidc_client_id"
	OIDCClientSecret                 = "oidc_client_secret"
	OIDCVerifyCert                   = "oidc_verify_cert"
	OIDCScope                        = "oidc_scope"

	DefaultClairEndpoint              = "http://clair:6060"
	CfgDriverDB                       = "db"
	NewHarborAdminName                = "admin@harbor.local"
	RegistryStorageProviderName       = "registry_storage_provider_name"
	RegistryControllerURL             = "registry_controller_url"
	UserMember                        = "u"
	GroupMember                       = "g"
	ReadOnly                          = "read_only"
	ClairURL                          = "clair_url"
	NotaryURL                         = "notary_url"
	DefaultCoreEndpoint               = "http://core:8080"
	DefaultNotaryEndpoint             = "http://notary-server:4443"
	LdapGroupType                     = 1
	LdapGroupAdminDn                  = "ldap_group_admin_dn"
	LDAPGroupMembershipAttribute      = "ldap_group_membership_attribute"
	DefaultRegistryControllerEndpoint = "http://registryctl:8080"
	WithChartMuseum                   = "with_chartmuseum"
	ChartRepoURL                      = "chart_repository_url"
	DefaultChartRepoURL               = "http://chartmuseum:9999"
	DefaultPortalURL                  = "http://portal"
	DefaultRegistryCtlURL             = "http://registryctl:8080"
	DefaultClairHealthCheckServerURL  = "http://clair:6061"
	// Use this prefix to distinguish harbor user, the prefix contains a special character($), so it cannot be registered as a harbor user.
	RobotPrefix = "robot$"
	// Use this prefix to index user who tries to login with web hook token.
	AuthProxyUserNamePrefix = "tokenreview$"
	CoreConfigPath          = "/api/internal/configurations"
	RobotTokenDuration      = "robot_token_duration"

	OIDCCallbackPath = "/c/oidc/callback"

	ChartUploadCtxKey = contextKey("chart_upload_event")
)
