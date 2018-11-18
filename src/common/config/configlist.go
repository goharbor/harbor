package config

import "github.com/goharbor/harbor/src/common"

// Item - Configure item include default value, type, env name
type Item struct {
	// The Scope of this configuration item: eg: SystemScope, UserScope
	Scope string `json:"scope,omitempty"`
	// email, ldapbasic, ldapgroup, uaa settings, used to retieve configure items by group, for example GetLDAPBasicSetting, GetLDAPGroupSetting settings
	Group string `json:"group,omitempty"`
	// environment key to retrieves this value when initialize, for example: POSTGRESQL_HOST, only used for system settings, for user settings no EnvironmentKey
	EnvironmentKey string `json:"environment_key,omitempty"`
	// The default string value for this key
	DefaultValue string `json:"default_value,omitempty"`
	// The key for current configure settings in database and rerest api
	Name string `json:"name,omitempty"`
	// It can be integer, string, bool, password, map
	Type string `json:"type,omitempty"`
	// The validation function for this field.
	Validator ValidateFunc `json:"validator,omitempty"`
	// Is this settign can be modified after configure
	Editable bool `json:"editable,omitempty"`
	// Reloadable - reload config from env after restart, if it is true, the setting is only reload from env
	Reloadable bool `json:"reloadable,omitempty"`
}

var (
	// ConfigList - All configure items used in harbor
	// Steps to onboard a new setting
	// 1. Add configure item in configlist.go
	// 2. Get settings by config.Client
	ConfigList = []Item{
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "HARBOR_ADMIN_PASSWORD", DefaultValue: "", Name: "admin_initial_password", Type: PasswordType, Editable: true, Reloadable: false},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "ADMIRAL_URL", DefaultValue: "", Name: "admiral_url", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "AUTH_MODE", DefaultValue: "db_auth", Name: "auth_mode", Type: StringType, Editable: false},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "CFG_EXPIRATION", DefaultValue: "", Name: "cfg_expiration", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "CHART_REPOSITORY_URL", DefaultValue: "", Name: "chart_repository_url", Type: StringType, Editable: false, Reloadable: true},

		{Scope: SystemScope, Group: ClairGroup, EnvironmentKey: "CLAIR_DB", DefaultValue: "", Name: "clair_db", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: ClairGroup, EnvironmentKey: "CLAIR_DB_HOST", DefaultValue: "", Name: "clair_db_host", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: ClairGroup, EnvironmentKey: "CLAIR_DB_PASSWORD", DefaultValue: "", Name: "clair_db_password", Type: PasswordType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: ClairGroup, EnvironmentKey: "CLAIR_DB_PORT", DefaultValue: "", Name: "clair_db_port", Type: IntType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: ClairGroup, EnvironmentKey: "CLAIR_DB_SSLMODE", DefaultValue: "", Name: "clair_db_sslmode", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: ClairGroup, EnvironmentKey: "CLAIR_DB_USERNAME", DefaultValue: "", Name: "clair_db_username", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: ClairGroup, EnvironmentKey: "CLAIR_URL", DefaultValue: "", Name: "clair_url", Type: StringType, Editable: false, Reloadable: true},

		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "CORE_URL", DefaultValue: "", Name: "core_url", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "DATABASE_TYPE", DefaultValue: "postgresql", Name: "database_type", Type: StringType, Editable: false, Reloadable: true},

		{Scope: UserScope, Group: EmailGroup, EnvironmentKey: "EMAIL_FROM", DefaultValue: "admin <sample_admin@mydomain.com>", Name: "email_from", Type: StringType, Editable: false},
		{Scope: UserScope, Group: EmailGroup, EnvironmentKey: "EMAIL_HOST", DefaultValue: "smtp.mydomain.com", Name: "email_host", Type: StringType, Editable: false},
		{Scope: UserScope, Group: EmailGroup, EnvironmentKey: "EMAIL_IDENTITY", DefaultValue: "", Name: "email_identity", Type: StringType, Editable: false},
		{Scope: UserScope, Group: EmailGroup, EnvironmentKey: "EMAIL_INSECURE", DefaultValue: "false", Name: "email_insecure", Type: BoolType, Editable: false},
		{Scope: UserScope, Group: EmailGroup, EnvironmentKey: "EMAIL_PWD", DefaultValue: "", Name: "email_password", Type: PasswordType, Editable: false},
		{Scope: UserScope, Group: EmailGroup, EnvironmentKey: "EMAIL_PORT", DefaultValue: "25", Name: "email_port", Type: IntType, Editable: false},
		{Scope: UserScope, Group: EmailGroup, EnvironmentKey: "EMAIL_SSL", DefaultValue: "false", Name: "email_ssl", Type: BoolType, Editable: false},
		{Scope: UserScope, Group: EmailGroup, EnvironmentKey: "EMAIL_USR", DefaultValue: "sample_admin@mydomain.com", Name: "email_username", Type: StringType, Editable: false},

		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "EXT_ENDPOINT", DefaultValue: "", Name: "ext_endpoint", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "JOBSERVICE_URL", DefaultValue: "", Name: "jobservice_url", Type: StringType, Editable: false, Reloadable: true},

		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_BASE_DN", DefaultValue: "", Name: "ldap_base_dn", Type: StringType, Editable: false},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_FILTER", DefaultValue: "", Name: "ldap_filter", Type: StringType, Editable: false},
		{Scope: UserScope, Group: LdapGroupGroup, EnvironmentKey: "LDAP_GROUP_ADMIN_DN", DefaultValue: "", Name: "ldap_group_admin_dn", Type: StringType, Editable: false},
		{Scope: UserScope, Group: LdapGroupGroup, EnvironmentKey: "LDAP_GROUP_GID", DefaultValue: "", Name: "ldap_group_attribute_name", Type: StringType, Editable: false},
		{Scope: UserScope, Group: LdapGroupGroup, EnvironmentKey: "LDAP_GROUP_FILTER", DefaultValue: "", Name: "ldap_group_search_filter", Type: StringType, Editable: false},
		{Scope: UserScope, Group: LdapGroupGroup, EnvironmentKey: "LDAP_GROUP_SCOPE", DefaultValue: "2", Name: "ldap_group_search_scope", Type: IntType, Editable: false},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_SCOPE", DefaultValue: "2", Name: "ldap_scope", Type: IntType, Editable: false},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_SEARCH_DN", DefaultValue: "", Name: "ldap_search_dn", Type: StringType, Editable: false},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_SEARCH_PWD", DefaultValue: "", Name: "ldap_search_password", Type: PasswordType, Editable: false},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_TIMEOUT", DefaultValue: "5", Name: "ldap_timeout", Type: IntType, Editable: false},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_UID", DefaultValue: "", Name: "ldap_uid", Type: StringType, Editable: true},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_URL", DefaultValue: "", Name: "ldap_url", Type: StringType, Editable: true},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "LDAP_VERIFY_CERT", DefaultValue: "true", Name: "ldap_verify_cert", Type: BoolType, Editable: false},

		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "MAX_JOB_WORKERS", DefaultValue: "", Name: "max_job_workers", Type: IntType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "NOTARY_URL", DefaultValue: "", Name: "notary_url", Type: StringType, Editable: false, Reloadable: true},

		{Scope: SystemScope, Group: DatabaseGroup, EnvironmentKey: "POSTGRESQL_DATABASE", DefaultValue: "registry", Name: "postgresql_database", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: DatabaseGroup, EnvironmentKey: "POSTGRESQL_HOST", DefaultValue: "", Name: "postgresql_host", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: DatabaseGroup, EnvironmentKey: "POSTGRESQL_PASSWORD", DefaultValue: "root123", Name: "postgresql_password", Type: PasswordType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: DatabaseGroup, EnvironmentKey: "POSTGRESQL_PORT", DefaultValue: "5432", Name: "postgresql_port", Type: IntType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: DatabaseGroup, EnvironmentKey: "POSTGRESQL_SSLMODE", DefaultValue: "disable", Name: "postgresql_sslmode", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: DatabaseGroup, EnvironmentKey: "POSTGRESQL_USERNAME", DefaultValue: "postgres", Name: "postgresql_username", Type: StringType, Editable: false, Reloadable: true},

		{Scope: UserScope, Group: BasicGroup, EnvironmentKey: "PROJECT_CREATION_RESTRICTION", DefaultValue: common.ProCrtRestrEveryone, Name: "project_creation_restriction", Type: StringType, Editable: false},
		{Scope: UserScope, Group: BasicGroup, EnvironmentKey: "READ_ONLY", DefaultValue: "false", Name: "read_only", Type: BoolType, Editable: false},

		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "REGISTRY_STORAGE_PROVIDER_NAME", DefaultValue: "", Name: "registry_storage_provider_name", Type: StringType, Editable: false, Reloadable: true},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "REGISTRY_URL", DefaultValue: "", Name: "registry_url", Type: StringType, Editable: false, Reloadable: true},
		{Scope: UserScope, Group: BasicGroup, EnvironmentKey: "SELF_REGISTRATION", DefaultValue: "true", Name: "self_registration", Type: BoolType, Editable: false},
		{Scope: UserScope, Group: BasicGroup, EnvironmentKey: "TOKEN_EXPIRATION", DefaultValue: "30", Name: "token_expiration", Type: IntType, Editable: false},
		{Scope: SystemScope, Group: BasicGroup, EnvironmentKey: "TOKEN_SERVICE_URL", DefaultValue: "", Name: "token_service_url", Type: StringType, Editable: false, Reloadable: true},

		{Scope: UserScope, Group: UAAGroup, EnvironmentKey: "UAA_CLIENTID", DefaultValue: "", Name: "uaa_client_id", Type: StringType, Editable: false, Reloadable: true},
		{Scope: UserScope, Group: UAAGroup, EnvironmentKey: "UAA_CLIENTSECRET", DefaultValue: "", Name: "uaa_client_secret", Type: StringType, Editable: false, Reloadable: true},
		{Scope: UserScope, Group: UAAGroup, EnvironmentKey: "UAA_ENDPOINT", DefaultValue: "", Name: "uaa_endpoint", Type: StringType, Editable: false, Reloadable: true},
		{Scope: UserScope, Group: UAAGroup, EnvironmentKey: "UAA_VERIFY_CERT", DefaultValue: "true", Name: "uaa_verify_cert", Type: BoolType, Editable: false, Reloadable: true},

		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "WITH_CHARTMUSEUM", DefaultValue: "", Name: "with_chartmuseum", Type: BoolType, Editable: true, Reloadable: true},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "WITH_CLAIR", DefaultValue: "", Name: "with_clair", Type: BoolType, Editable: true, Reloadable: true},
		{Scope: UserScope, Group: LdapBasicGroup, EnvironmentKey: "WITH_NOTARY", DefaultValue: "", Name: "with_notary", Type: BoolType, Editable: true, Reloadable: true},
	}
)
