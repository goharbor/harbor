package session

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

import (
	"os"
	"strconv"

	"github.com/volcengine/volcengine-go-sdk/volcengine"
	"github.com/volcengine/volcengine-go-sdk/volcengine/credentials"
	"github.com/volcengine/volcengine-go-sdk/volcengine/defaults"
)

// EnvProviderName provides a name of the provider when config is loaded from environment.
const EnvProviderName = "EnvConfigCredentials"

// envConfig is a collection of environment values the SDK will read
// setup config from. All environment values are optional. But some values
// such as credentials require multiple values to be complete or the values
// will be ignored.
type envConfig struct {
	// Environment configuration values. If set both Access Key ID and Secret Access
	// Key must be provided. Session Token and optionally also be provided, but is
	// not required.
	//
	//	# Access Key ID
	//	VOLCSTACK_ACCESS_KEY_ID=AKID
	//	VOLCSTACK_ACCESS_KEY=AKID # only read if VOLCSTACK_ACCESS_KEY_ID is not set.
	//
	//	# Secret Access Key
	//	VOLCSTACK_SECRET_ACCESS_KEY=SECRET
	//	VOLCSTACK_SECRET_KEY=SECRET=SECRET # only read if VOLCSTACK_SECRET_ACCESS_KEY is not set.
	//
	//	# Session Token
	//	VOLCSTACK_SESSION_TOKEN=TOKEN
	Creds credentials.Value

	// Region value will instruct the SDK where to make service API requests to. If is
	// not provided in the environment the region must be provided before a service
	// client request is made.
	//
	//	VOLCSTACK_REGION=us-east-1
	//
	//	# VOLCSTACK_DEFAULT_REGION is only read if VOLCSTACK_SDK_LOAD_CONFIG is also set,
	//	# and VOLCSTACK_REGION is not also set.
	//	VOLCSTACK_DEFAULT_REGION=us-east-1
	Region string

	// Profile name the SDK should load use when loading shared configuration from the
	// shared configuration files. If not provided "default" will be used as the
	// profile name.
	//
	//	VOLCSTACK_PROFILE=my_profile
	//
	//	# VOLCSTACK_DEFAULT_PROFILE is only read if VOLCSTACK_SDK_LOAD_CONFIG is also set,
	//	# and VOLCSTACK_PROFILE is not also set.
	//	VOLCSTACK_DEFAULT_PROFILE=my_profile
	Profile string

	// SDK load config instructs the SDK to load the shared config in addition to
	// shared credentials. This also expands the configuration loaded from the shared
	// credentials to have parity with the shared config file. This also enables
	// Region and Profile support for the VOLCSTACK_DEFAULT_REGION and VOLCSTACK_DEFAULT_PROFILE
	// env values as well.
	//
	//	VOLCSTACK_SDK_LOAD_CONFIG=1
	EnableSharedConfig bool

	// Shared credentials file path can be set to instruct the SDK to use an alternate
	// file for the shared credentials. If not set the file will be loaded from
	// $HOME/.volcengine/credentials on Linux/Unix based systems, and
	// %USERPROFILE%\.volcengine\credentials on Windows.
	//
	//	VOLCSTACK_SHARED_CREDENTIALS_FILE=$HOME/my_shared_credentials
	SharedCredentialsFile string

	// Shared config file path can be set to instruct the SDK to use an alternate
	// file for the shared config. If not set the file will be loaded from
	// $HOME/.volcengine/config on Linux/Unix based systems, and
	// %USERPROFILE%\.volcengine\config on Windows.
	//
	//	VOLCSTACK_CONFIG_FILE=$HOME/my_shared_config
	SharedConfigFile string

	// Sets the path to a custom Credentials Authority (CA) Bundle PEM file
	// that the SDK will use instead of the system's root CA bundle.
	// Only use this if you want to configure the SDK to use a custom set
	// of CAs.
	//
	// Enabling this option will attempt to merge the Transport
	// into the SDK's HTTP client. If the client's Transport is
	// not a http.Transport an error will be returned. If the
	// Transport's TLS config is set this option will cause the
	// SDK to overwrite the Transport's TLS config's  RootCAs value.
	//
	// Setting a custom HTTPClient in the volcengine.Config options will override this setting.
	// To use this option and custom HTTP client, the HTTP client needs to be provided
	// when creating the session. Not the service client.
	//
	//  VOLCSTACK_CA_BUNDLE=$HOME/my_custom_ca_bundle
	CustomCABundle string

	csmEnabled  string
	CSMEnabled  *bool
	CSMPort     string
	CSMHost     string
	CSMClientID string

	// Enables endpoint discovery via environment variables.
	//
	//	VOLCSTACK_ENABLE_ENDPOINT_DISCOVERY=true
	EnableEndpointDiscovery *bool
	enableEndpointDiscovery string

	// Specifies the WebIdentity token the SDK should use to assume a role
	// with.
	//
	//  VOLCSTACK_WEB_IDENTITY_TOKEN_FILE=file_path
	WebIdentityTokenFilePath string

	// Specifies the IAM role arn to use when assuming an role.
	//
	//  VOLCSTACK_ROLE_ARN=role_arn
	RoleARN string

	// Specifies the IAM role session name to use when assuming a role.
	//
	//  VOLCSTACK_ROLE_SESSION_NAME=session_name
	RoleSessionName string
	// Specifies the endpoint config file path
	//
	//  VOLCSTACK_ENABLE_ENDPOINT_CONFIG_STATE=session_name
	EndpointConfigState *bool
	// Specifies the endpoint config file path
	//
	//  VOLCSTACK_ENDPOINT_CONFIG_FILE=session_name
	EndpointConfigPath string
}

var (
	csmEnabledEnvKey = []string{
		"VOLCSTACK_CSM_ENABLED",
	}
	csmHostEnvKey = []string{
		"VOLCSTACK_CSM_HOST",
	}
	csmPortEnvKey = []string{
		"VOLCSTACK_CSM_PORT",
	}
	csmClientIDEnvKey = []string{
		"VOLCSTACK_CSM_CLIENT_ID",
	}
	credAccessEnvKey = []string{
		"VOLCSTACK_ACCESS_KEY_ID",
		"VOLCSTACK_ACCESS_KEY",
	}
	credSecretEnvKey = []string{
		"VOLCSTACK_SECRET_ACCESS_KEY",
		"VOLCSTACK_SECRET_KEY",
	}
	credSessionEnvKey = []string{
		"VOLCSTACK_SESSION_TOKEN",
	}

	enableEndpointDiscoveryEnvKey = []string{
		"VOLCSTACK_ENABLE_ENDPOINT_DISCOVERY",
	}
	enableEndpointConfigStateEnvKey = []string{
		"VOLCSTACK_ENABLE_ENDPOINT_CONFIG_STATE",
	}

	regionEnvKeys = []string{
		"VOLCSTACK_REGION",
		"VOLCSTACK_DEFAULT_REGION", // Only read if VOLCSTACK_SDK_LOAD_CONFIG is also set
	}
	profileEnvKeys = []string{
		"VOLCSTACK_PROFILE",
		"VOLCSTACK_DEFAULT_PROFILE", // Only read if VOLCSTACK_SDK_LOAD_CONFIG is also set
	}
	sharedCredsFileEnvKey = []string{
		"VOLCSTACK_SHARED_CREDENTIALS_FILE",
	}
	sharedConfigFileEnvKey = []string{
		"VOLCSTACK_CONFIG_FILE",
	}

	EndpointConfigPathEnvKey = []string{
		"VOLCSTACK_ENDPOINT_CONFIG_PATH",
	}
	webIdentityTokenFilePathEnvKey = []string{
		"VOLCSTACK_WEB_IDENTITY_TOKEN_FILE",
	}
	roleARNEnvKey = []string{
		"VOLCSTACK_ROLE_ARN",
	}
	roleSessionNameEnvKey = []string{
		"VOLCSTACK_ROLE_SESSION_NAME",
	}
)

// loadEnvConfig retrieves the SDK's environment configuration.
// See `envConfig` for the values that will be retrieved.
//
// If the environment variable `VOLCSTACK_SDK_LOAD_CONFIG` is set to a truthy value
// the shared SDK config will be loaded in addition to the SDK's specific
// configuration values.
func loadEnvConfig() envConfig {
	enableSharedConfig, _ := strconv.ParseBool(os.Getenv("VOLCSTACK_SDK_LOAD_CONFIG"))
	return envConfigLoad(enableSharedConfig)
}

// loadEnvSharedConfig retrieves the SDK's environment configuration, and the
// SDK shared config. See `envConfig` for the values that will be retrieved.
//
// Loads the shared configuration in addition to the SDK's specific configuration.
// This will load the same values as `loadEnvConfig` if the `VOLCSTACK_SDK_LOAD_CONFIG`
// environment variable is set.
func loadSharedEnvConfig() envConfig {
	return envConfigLoad(true)
}

func envConfigLoad(enableSharedConfig bool) envConfig {
	cfg := envConfig{}

	cfg.EnableSharedConfig = enableSharedConfig

	// Static environment credentials
	var creds credentials.Value
	setFromEnvVal(&creds.AccessKeyID, credAccessEnvKey)
	setFromEnvVal(&creds.SecretAccessKey, credSecretEnvKey)
	setFromEnvVal(&creds.SessionToken, credSessionEnvKey)
	if creds.HasKeys() {
		// Require logical grouping of credentials
		creds.ProviderName = EnvProviderName
		cfg.Creds = creds
	}

	// Role Metadata
	setFromEnvVal(&cfg.RoleARN, roleARNEnvKey)
	setFromEnvVal(&cfg.RoleSessionName, roleSessionNameEnvKey)

	// Web identity environment variables
	setFromEnvVal(&cfg.WebIdentityTokenFilePath, webIdentityTokenFilePathEnvKey)

	// CSM environment variables
	setFromEnvVal(&cfg.csmEnabled, csmEnabledEnvKey)
	setFromEnvVal(&cfg.CSMHost, csmHostEnvKey)
	setFromEnvVal(&cfg.CSMPort, csmPortEnvKey)
	setFromEnvVal(&cfg.CSMClientID, csmClientIDEnvKey)

	if len(cfg.csmEnabled) != 0 {
		v, _ := strconv.ParseBool(cfg.csmEnabled)
		cfg.CSMEnabled = &v
	}

	regionKeys := regionEnvKeys
	profileKeys := profileEnvKeys
	if !cfg.EnableSharedConfig {
		regionKeys = regionKeys[:1]
		profileKeys = profileKeys[:1]
	}

	setFromEnvVal(&cfg.Region, regionKeys)
	setFromEnvVal(&cfg.Profile, profileKeys)

	// endpoint discovery is in reference to it being enabled.
	setFromEnvVal(&cfg.enableEndpointDiscovery, enableEndpointDiscoveryEnvKey)
	if len(cfg.enableEndpointDiscovery) > 0 {
		cfg.EnableEndpointDiscovery = volcengine.Bool(cfg.enableEndpointDiscovery != "false")
	}
	var endpointConfigState string
	setFromEnvVal(&endpointConfigState, enableEndpointConfigStateEnvKey)
	if len(cfg.enableEndpointDiscovery) > 0 {
		cfg.EndpointConfigState = volcengine.Bool(endpointConfigState != "false")
	}
	setFromEnvVal(&cfg.SharedCredentialsFile, sharedCredsFileEnvKey)
	setFromEnvVal(&cfg.SharedConfigFile, sharedConfigFileEnvKey)
	setFromEnvVal(&cfg.EndpointConfigPath, EndpointConfigPathEnvKey)

	if len(cfg.SharedCredentialsFile) == 0 {
		cfg.SharedCredentialsFile = defaults.SharedCredentialsFilename()
	}
	if len(cfg.SharedConfigFile) == 0 {
		cfg.SharedConfigFile = defaults.SharedConfigFilename()
	}

	cfg.CustomCABundle = os.Getenv("VOLCSTACK_CA_BUNDLE")

	return cfg
}

func setFromEnvVal(dst *string, keys []string) {
	for _, k := range keys {
		if v := os.Getenv(k); len(v) > 0 {
			*dst = v
			break
		}
	}
}
