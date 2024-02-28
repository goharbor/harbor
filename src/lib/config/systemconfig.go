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

package config

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/lib/log"
)

var (
	// SecretStore manages secrets
	SecretStore *secret.Store
	keyProvider encrypt.KeyProvider
	// Use backgroundCtx to access system scope config
	backgroundCtx = context.Background()
)

// It contains all system settings

// TokenPrivateKeyPath returns the path to the key for signing token for registry
func TokenPrivateKeyPath() string {
	path := os.Getenv("TOKEN_PRIVATE_KEY_PATH")
	if len(path) == 0 {
		path = defaultRegistryTokenPrivateKeyPath
	}
	return path
}

// RegistryURL ...
func RegistryURL() (string, error) {
	url := os.Getenv("REGISTRY_URL")
	if len(url) == 0 {
		url = "http://registry:5000"
	}
	return url, nil
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	return os.Getenv("JOBSERVICE_URL")
}

// GetCoreURL returns the url of core from env
func GetCoreURL() string {
	return os.Getenv("CORE_URL")
}

// CoreSecret returns a secret to mark harbor-core when communicate with
// other component
func CoreSecret() string {
	return os.Getenv("CORE_SECRET")
}

// RegistryCredential returns the username and password the core uses to access registry
func RegistryCredential() (string, string) {
	return os.Getenv("REGISTRY_CREDENTIAL_USERNAME"), os.Getenv("REGISTRY_CREDENTIAL_PASSWORD")
}

// JobserviceSecret returns a secret to mark Jobservice when communicate with
// other component
// TODO replace it with method of SecretStore
func JobserviceSecret() string {
	return os.Getenv("JOBSERVICE_SECRET")
}

// GetPortalURL returns the URL of portal
func GetPortalURL() string {
	url := os.Getenv("PORTAL_URL")
	if len(url) == 0 {
		return common.DefaultPortalURL
	}
	return url
}

// GetRegistryCtlURL returns the URL of registryctl
func GetRegistryCtlURL() string {
	url := os.Getenv("REGISTRY_CONTROLLER_URL")
	if len(url) == 0 {
		return common.DefaultRegistryCtlURL
	}
	return url
}

// GetPermittedRegistryTypesForProxyCache returns the permitted registry types for proxy cache
func GetPermittedRegistryTypesForProxyCache() []string {
	types := os.Getenv("PERMITTED_REGISTRY_TYPES_FOR_PROXY_CACHE")
	if len(types) == 0 {
		return []string{}
	}
	return strings.Split(types, ",")
}

// GetGCTimeWindow returns the reserve time window of blob.
func GetGCTimeWindow() int64 {
	// the env is for testing/debugging. For production, Do NOT set it.
	if env, exist := os.LookupEnv("GC_TIME_WINDOW_HOURS"); exist {
		timeWindow, err := strconv.ParseInt(env, 10, 64)
		if err == nil {
			return timeWindow
		}
	}
	return common.DefaultGCTimeWindowHours
}

// GetExecutionStatusRefreshIntervalSeconds returns the interval seconds for the refresh of execution status.
func GetExecutionStatusRefreshIntervalSeconds() int64 {
	return DefaultMgr().Get(backgroundCtx, common.ExecutionStatusRefreshIntervalSeconds).GetInt64()
}

// GetQuotaUpdateProvider returns the provider for updating quota.
func GetQuotaUpdateProvider() string {
	return DefaultMgr().Get(backgroundCtx, common.QuotaUpdateProvider).GetString()
}

// GetBeegoMaxMemoryBytes returns the max memory bytes of beego config
func GetBeegoMaxMemoryBytes() int64 {
	return DefaultMgr().Get(backgroundCtx, common.BeegoMaxMemoryBytes).GetInt64()
}

// GetBeegoMaxUploadSizeBytes returns the max upload size bytes of beego config
func GetBeegoMaxUploadSizeBytes() int64 {
	return DefaultMgr().Get(backgroundCtx, common.BeegoMaxUploadSizeBytes).GetInt64()
}

// WithTrivy returns a bool value to indicate if Harbor's deployed with Trivy.
func WithTrivy() bool {
	return DefaultMgr().Get(backgroundCtx, common.WithTrivy).GetBool()
}

// ExtEndpoint returns the external URL of Harbor: protocol://host:port
func ExtEndpoint() (string, error) {
	return DefaultMgr().Get(backgroundCtx, common.ExtEndpoint).GetString(), nil
}

// ExtURL returns the external URL: host:port
func ExtURL() (string, error) {
	endpoint, err := ExtEndpoint()
	if err != nil {
		log.Errorf("failed to load config, error %v", err)
	}
	l := strings.Split(endpoint, "://")
	if len(l) > 1 {
		return l[1], nil
	}
	return endpoint, nil
}

// SecretKey returns the secret key to encrypt the password of target
func SecretKey() (string, error) {
	return keyProvider.Get(nil)
}

func initKeyProvider() {
	path := os.Getenv("KEY_PATH")
	if len(path) == 0 {
		path = defaultKeyPath
	}
	log.Infof("key path: %s", path)
	keyProvider = encrypt.NewFileKeyProvider(path)
}

func initSecretStore() {
	m := map[string]string{}
	m[JobserviceSecret()] = secret.JobserviceUser
	SecretStore = secret.NewStore(m)
}

// InternalCoreURL returns the local harbor core url
func InternalCoreURL() string {
	return strings.TrimSuffix(DefaultMgr().Get(backgroundCtx, common.CoreURL).GetString(), "/")
}

// LocalCoreURL returns the local harbor core url
func LocalCoreURL() string {
	return DefaultMgr().Get(backgroundCtx, common.CoreLocalURL).GetString()
}

// InternalTokenServiceEndpoint returns token service endpoint for internal communication between Harbor containers
func InternalTokenServiceEndpoint() string {
	return InternalCoreURL() + "/service/token"
}

// TrivyAdapterURL returns the endpoint URL of a Trivy adapter instance, by default it's the one deployed within Harbor.
func TrivyAdapterURL() string {
	return DefaultMgr().Get(backgroundCtx, common.TrivyAdapterURL).GetString()
}

// Metric returns the overall metric settings
func Metric() *models.Metric {
	return &models.Metric{
		Enabled: DefaultMgr().Get(backgroundCtx, common.MetricEnable).GetBool(),
		Port:    DefaultMgr().Get(backgroundCtx, common.MetricPort).GetInt(),
		Path:    DefaultMgr().Get(backgroundCtx, common.MetricPath).GetString(),
	}
}

// InitialAdminPassword returns the initial password for administrator
func InitialAdminPassword() (string, error) {
	return DefaultMgr().Get(backgroundCtx, common.AdminInitialPassword).GetString(), nil
}

// CacheEnabled returns whether enable cache layer.
func CacheEnabled() bool {
	if DefaultMgr() != nil {
		return DefaultMgr().Get(backgroundCtx, common.CacheEnabled).GetBool()
	}
	// backoff read from env.
	return os.Getenv("CACHE_ENABLED") == "true"
}

// CacheExpireHours returns the cache expire hours for cache layer.
func CacheExpireHours() int {
	if DefaultMgr() != nil {
		return DefaultMgr().Get(backgroundCtx, common.CacheExpireHours).GetInt()
	}
	// backoff read from env.
	hours, err := strconv.Atoi(os.Getenv("CACHE_EXPIRE_HOURS"))
	if err != nil {
		// use default if parse error.
		hours = common.DefaultCacheExpireHours
	}

	return hours
}

// ScannerRobotPrefix returns the scanner of robot account prefix.
func ScannerRobotPrefix(ctx context.Context) string {
	if DefaultMgr() != nil {
		return DefaultMgr().Get(ctx, common.RobotScannerNamePrefix).GetString()
	}
	return os.Getenv("ROBOT_SCANNER_NAME_PREFIX")
}

// Database returns database settings
func Database() (*models.Database, error) {
	database := &models.Database{}
	database.Type = DefaultMgr().Get(backgroundCtx, common.DatabaseType).GetString()
	postgresql := &models.PostGreSQL{
		Host:            DefaultMgr().Get(backgroundCtx, common.PostGreSQLHOST).GetString(),
		Port:            DefaultMgr().Get(backgroundCtx, common.PostGreSQLPort).GetInt(),
		Username:        DefaultMgr().Get(backgroundCtx, common.PostGreSQLUsername).GetString(),
		Password:        DefaultMgr().Get(backgroundCtx, common.PostGreSQLPassword).GetPassword(),
		Database:        DefaultMgr().Get(backgroundCtx, common.PostGreSQLDatabase).GetString(),
		SSLMode:         DefaultMgr().Get(backgroundCtx, common.PostGreSQLSSLMode).GetString(),
		MaxIdleConns:    DefaultMgr().Get(backgroundCtx, common.PostGreSQLMaxIdleConns).GetInt(),
		MaxOpenConns:    DefaultMgr().Get(backgroundCtx, common.PostGreSQLMaxOpenConns).GetInt(),
		ConnMaxLifetime: DefaultMgr().Get(backgroundCtx, common.PostGreSQLConnMaxLifetime).GetDuration(),
		ConnMaxIdleTime: DefaultMgr().Get(backgroundCtx, common.PostGreSQLConnMaxIdleTime).GetDuration(),
	}
	database.PostGreSQL = postgresql

	return database, nil
}
