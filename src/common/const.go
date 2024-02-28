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

import "time"

type contextKey string

// const variables
const (
	DBAuth              = "db_auth"
	LDAPAuth            = "ldap_auth"
	UAAAuth             = "uaa_auth"
	HTTPAuth            = "http_auth"
	OIDCAuth            = "oidc_auth"
	DBCfgManager        = "db_cfg_manager"
	InMemoryCfgManager  = "in_memory_manager"
	RestCfgManager      = "rest_config_manager"
	ProCrtRestrEveryone = "everyone"
	ProCrtRestrAdmOnly  = "adminonly"
	LDAPScopeBase       = 0
	LDAPScopeOnelevel   = 1
	LDAPScopeSubtree    = 2

	RoleProjectAdmin = 1
	RoleDeveloper    = 2
	RoleGuest        = 3
	RoleMaintainer   = 4
	RoleLimitedGuest = 5

	LabelLevelSystem  = "s"
	LabelLevelUser    = "u"
	LabelScopeGlobal  = "g"
	LabelScopeProject = "p"

	ResourceTypeProject    = "p"
	ResourceTypeRepository = "r"
	ResourceTypeImage      = "i"

	ExtEndpoint                      = "ext_endpoint"
	AUTHMode                         = "auth_mode"
	PrimaryAuthMode                  = "primary_auth_mode"
	DatabaseType                     = "database_type"
	PostGreSQLHOST                   = "postgresql_host"
	PostGreSQLPort                   = "postgresql_port"
	PostGreSQLUsername               = "postgresql_username"
	PostGreSQLPassword               = "postgresql_password"
	PostGreSQLDatabase               = "postgresql_database"
	PostGreSQLSSLMode                = "postgresql_sslmode"
	PostGreSQLMaxIdleConns           = "postgresql_max_idle_conns"
	PostGreSQLMaxOpenConns           = "postgresql_max_open_conns"
	PostGreSQLConnMaxLifetime        = "postgresql_conn_max_lifetime"
	PostGreSQLConnMaxIdleTime        = "postgresql_conn_max_idle_time"
	SelfRegistration                 = "self_registration"
	CoreURL                          = "core_url"
	CoreLocalURL                     = "core_local_url"
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
	AdminInitialPassword             = "admin_initial_password"
	WithTrivy                        = "with_trivy"
	ScanAllPolicy                    = "scan_all_policy"
	UAAEndpoint                      = "uaa_endpoint"
	UAAClientID                      = "uaa_client_id"
	UAAClientSecret                  = "uaa_client_secret"
	UAAVerifyCert                    = "uaa_verify_cert"
	HTTPAuthProxyEndpoint            = "http_authproxy_endpoint"
	HTTPAuthProxyTokenReviewEndpoint = "http_authproxy_tokenreview_endpoint"
	HTTPAuthProxyAdminGroups         = "http_authproxy_admin_groups"
	HTTPAuthProxyAdminUsernames      = "http_authproxy_admin_usernames"
	HTTPAuthProxyVerifyCert          = "http_authproxy_verify_cert"
	HTTPAuthProxySkipSearch          = "http_authproxy_skip_search"
	HTTPAuthProxyServerCertificate   = "http_authproxy_server_certificate"
	OIDCName                         = "oidc_name"
	OIDCEndpoint                     = "oidc_endpoint"
	OIDCCLientID                     = "oidc_client_id"
	OIDCClientSecret                 = "oidc_client_secret"
	OIDCVerifyCert                   = "oidc_verify_cert"
	OIDCAdminGroup                   = "oidc_admin_group"
	OIDCGroupsClaim                  = "oidc_groups_claim"
	OIDCGroupFilter                  = "oidc_group_filter"
	OIDCAutoOnboard                  = "oidc_auto_onboard"
	OIDCExtraRedirectParms           = "oidc_extra_redirect_parms"
	OIDCScope                        = "oidc_scope"
	OIDCUserClaim                    = "oidc_user_claim"

	CfgDriverDB                       = "db"
	NewHarborAdminName                = "admin@harbor.local"
	RegistryStorageProviderName       = "registry_storage_provider_name"
	RegistryControllerURL             = "registry_controller_url"
	UserMember                        = "u"
	GroupMember                       = "g"
	ReadOnly                          = "read_only"
	TrivyAdapterURL                   = "trivy_adapter_url"
	DefaultCoreEndpoint               = "http://core:8080"
	LDAPGroupType                     = 1
	HTTPGroupType                     = 2
	OIDCGroupType                     = 3
	LDAPGroupAdminDn                  = "ldap_group_admin_dn"
	LDAPGroupMembershipAttribute      = "ldap_group_membership_attribute"
	DefaultRegistryControllerEndpoint = "http://registryctl:8080"
	DefaultPortalURL                  = "http://portal:8080"
	DefaultRegistryCtlURL             = "http://registryctl:8080"
	// Use this prefix to distinguish harbor user, the prefix contains a special character($), so it cannot be registered as a harbor user.
	RobotPrefix = "robot$"
	// System admin defined the robot name prefix.
	RobotNamePrefix = "robot_name_prefix"
	// Scanner robot name prefix
	RobotScannerNamePrefix = "robot_scanner_name_prefix"
	// Use this prefix to index user who tries to login with web hook token.
	AuthProxyUserNamePrefix = "tokenreview$"
	CoreConfigPath          = "/api/v2.0/internalconfig"
	RobotTokenDuration      = "robot_token_duration"

	OIDCCallbackPath = "/c/oidc/callback"
	OIDCLoginPath    = "/c/oidc/login"

	AuthProxyRediretPath = "/c/authproxy/redirect"

	// Global notification enable configuration
	NotificationEnable = "notification_enable"

	// Quota setting items for project
	QuotaPerProjectEnable = "quota_per_project_enable"
	StoragePerProject     = "storage_per_project"

	// DefaultGCTimeWindowHours is the reserve blob time window used by GC, default is 2 hours
	DefaultGCTimeWindowHours = int64(2)

	// Metric setting items
	MetricEnable = "metric_enable"
	MetricPort   = "metric_port"
	MetricPath   = "metric_path"

	// Trace setting items
	TraceEnabled         = "trace_enabled"
	TraceServiceName     = "trace_service_name"
	TraceSampleRate      = "trace_sample_rate"
	TraceNamespace       = "trace_namespace"
	TraceAttributes      = "trace_attribute"
	TraceJaegerEndpoint  = "trace_jaeger_endpoint"
	TraceJaegerUsername  = "trace_jaeger_username"
	TraceJaegerPassword  = "trace_jaeger_password"
	TraceJaegerAgentHost = "trace_jaeger_agent_host"
	TraceJaegerAgentPort = "trace_jaeger_agent_port"
	TraceOtelEndpoint    = "trace_otel_endpoint"
	TraceOtelURLPath     = "trace_otel_url_path"
	TraceOtelCompression = "trace_otel_compression"
	TraceOtelInsecure    = "trace_otel_insecure"
	TraceOtelTimeout     = "trace_otel_timeout"

	GDPRDeleteUser = "gdpr_delete_user"
	GDPRAuditLogs  = "gdpr_audit_logs"

	//  These variables are temporary solution for issue: https://github.com/goharbor/harbor/issues/16039
	//  When user disable the pull count/time/audit log, it will decrease the database access, especially in large concurrency pull scenarios.
	// TODO: Once we have a complete solution, delete these variables.
	// PullCountUpdateDisable indicate if pull count is disable for pull request.
	PullCountUpdateDisable = "pull_count_update_disable"
	// PullTimeUpdateDisable indicate if pull time is disable for pull request.
	PullTimeUpdateDisable = "pull_time_update_disable"
	// PullAuditLogDisable indicate if pull audit log is disable for pull request.
	PullAuditLogDisable = "pull_audit_log_disable"

	// Cache layer settings
	// CacheEnabled indicate whether enable cache layer.
	CacheEnabled = "cache_enabled"
	// CacheExpireHours is the cache expiration time, unit is hour.
	CacheExpireHours = "cache_expire_hours"
	// DefaultCacheExpireHours is the default cache expire hours, default is
	// 24h.
	DefaultCacheExpireHours = 24

	PurgeAuditIncludeOperations = "include_operations"
	PurgeAuditDryRun            = "dry_run"
	PurgeAuditRetentionHour     = "audit_retention_hour"
	// AuditLogForwardEndpoint indicate to forward the audit log to an endpoint
	AuditLogForwardEndpoint = "audit_log_forward_endpoint"
	// SkipAuditLogDatabase skip to log audit log in database
	SkipAuditLogDatabase = "skip_audit_log_database"
	// MaxAuditRetentionHour allowed in audit log purge
	MaxAuditRetentionHour = 240000
	// ScannerSkipUpdatePullTime
	ScannerSkipUpdatePullTime = "scanner_skip_update_pulltime"

	// SessionTimeout defines the web session timeout
	SessionTimeout = "session_timeout"

	// Customized banner message
	BannerMessage = "banner_message"

	// UIMaxLengthLimitedOfNumber is the max length that UI limited for type number
	UIMaxLengthLimitedOfNumber = 10
	// ExecutionStatusRefreshIntervalSeconds is the interval seconds for refreshing execution status
	ExecutionStatusRefreshIntervalSeconds = "execution_status_refresh_interval_seconds"
	// QuotaUpdateProvider is the provider for updating quota, currently support Redis and DB
	QuotaUpdateProvider = "quota_update_provider"
	// IllegalCharsInUsername is the illegal chars in username
	IllegalCharsInUsername = `,"~#%$`

	// Beego web config
	// BeegoMaxMemoryBytes is the max memory(bytes) of the beego web config
	BeegoMaxMemoryBytes = "beego_max_memory_bytes"
	// DefaultBeegoMaxMemoryBytes sets default max memory to 128GB
	DefaultBeegoMaxMemoryBytes = 1 << 37
	// BeegoMaxUploadSizeBytes is the max upload size(bytes) of the beego web config
	BeegoMaxUploadSizeBytes = "beego_max_upload_size_bytes"
	// DefaultBeegoMaxUploadSizeBytes sets default max upload size to 128GB
	DefaultBeegoMaxUploadSizeBytes = 1 << 37

	// Global Leeway used for token validation
	JwtLeeway = 60 * time.Second
)
