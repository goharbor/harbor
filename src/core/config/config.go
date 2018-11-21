// Copyright 2018 Project Harbor Authors
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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/adminserver/client"
	"github.com/goharbor/harbor/src/common"
	comcfg "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/config/client/db"
	"github.com/goharbor/harbor/src/common/config/encrypt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/core/promgr/pmsdriver"
	"github.com/goharbor/harbor/src/core/promgr/pmsdriver/admiral"
	"github.com/goharbor/harbor/src/core/promgr/pmsdriver/local"
)

const (
	defaultKeyPath                     = "/etc/core/key"
	defaultTokenFilePath               = "/etc/core/token/tokens.properties"
	defaultRegistryTokenPrivateKeyPath = "/etc/core/private_key.pem"
)

var (
	// SecretStore manages secrets
	SecretStore *secret.Store
	// AdminserverClient is a client for adminserver
	AdminserverClient client.Client
	// GlobalProjectMgr is initialized based on the deploy mode
	GlobalProjectMgr promgr.ProjectManager
	// mg               *comcfg.Manager
	mg          comcfg.ManagerInterface
	keyProvider encrypt.KeyProvider
	// AdmiralClient is initialized only under integration deploy mode
	// and can be passed to project manager as a parameter
	AdmiralClient *http.Client
	// TokenReader is used in integration mode to read token
	TokenReader admiral.TokenReader
	// defined as a var for testing.
	defaultCACertPath = "/etc/core/ca/ca.crt"
)

// Init configurations
func Init() error {
	// init key provider
	initKeyProvider()
	// adminServerURL := os.Getenv("ADMINSERVER_URL")
	// if len(adminServerURL) == 0 {
	// 	adminServerURL = common.DefaultAdminserverEndpoint
	// }
	return InitDBConfigManager()
}

// InitDBConfigManager ...
func InitDBConfigManager() error {
	mg = db.NewCoreConfigManager()

	_, err := mg.Load()
	if err != nil {
		return err
	}
	// init secret store
	initSecretStore()

	// init project manager based on deploy mode
	if err := initProjectManager(); err != nil {
		log.Errorf("Failed to initialise project manager, error: %v", err)
		return err
	}

	return err
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

func initProjectManager() error {
	var driver pmsdriver.PMSDriver
	if WithAdmiral() {
		log.Debugf("Initialising Admiral client with certificate: %s", defaultCACertPath)
		content, err := ioutil.ReadFile(defaultCACertPath)
		if err != nil {
			return err
		}
		pool := x509.NewCertPool()
		if ok := pool.AppendCertsFromPEM(content); !ok {
			return fmt.Errorf("failed to append cert content into cert pool")
		}
		AdmiralClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					RootCAs: pool,
				},
			},
		}

		// integration with admiral
		log.Info("initializing the project manager based on PMS...")
		path := os.Getenv("SERVICE_TOKEN_FILE_PATH")
		if len(path) == 0 {
			path = defaultTokenFilePath
		}
		log.Infof("service token file path: %s", path)
		TokenReader = &admiral.FileTokenReader{
			Path: path,
		}
		driver = admiral.NewDriver(AdmiralClient, AdmiralEndpoint(), TokenReader)
	} else {
		// standalone
		log.Info("initializing the project manager based on local database...")
		driver = local.NewDriver()
	}
	GlobalProjectMgr = promgr.NewDefaultProjectManager(driver, true)
	return nil

}

// Load configurations
func Load() error {
	_, err := mg.Load()
	return err
}

// Reset configurations
func Reset() error {
	return mg.Reset()
}

// Upload uploads all system configurations to admin server
func Upload(cfg map[string]interface{}) error {
	return mg.Upload(cfg)
}

// GetSystemCfg returns the system configurations
func GetSystemCfg() (map[string]interface{}, error) {
	return mg.Load()
}

// AuthMode ...
func AuthMode() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return utils.SafeCastString(cfg[common.AUTHMode]), nil
}

// TokenPrivateKeyPath returns the path to the key for signing token for registry
func TokenPrivateKeyPath() string {
	path := os.Getenv("TOKEN_PRIVATE_KEY_PATH")
	if len(path) == 0 {
		path = defaultRegistryTokenPrivateKeyPath
	}
	return path
}

// LDAPConf returns the setting of ldap server
func LDAPConf() (*models.LdapConf, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}
	ldapConf := &models.LdapConf{}
	ldapConf.LdapURL = utils.SafeCastString(cfg[common.LDAPURL])
	ldapConf.LdapSearchDn = utils.SafeCastString(cfg[common.LDAPSearchDN])
	ldapConf.LdapSearchPassword = utils.SafeCastString(cfg[common.LDAPSearchPwd])
	ldapConf.LdapBaseDn = utils.SafeCastString(cfg[common.LDAPBaseDN])
	ldapConf.LdapUID = utils.SafeCastString(cfg[common.LDAPUID])
	ldapConf.LdapFilter = utils.SafeCastString(cfg[common.LDAPFilter])
	ldapConf.LdapScope = utils.SafeCastInt(cfg[common.LDAPScope])
	ldapConf.LdapConnectionTimeout = utils.SafeCastInt(cfg[common.LDAPTimeout])
	if cfg[common.LDAPVerifyCert] != nil {
		ldapConf.LdapVerifyCert = cfg[common.LDAPVerifyCert].(bool)
	} else {
		ldapConf.LdapVerifyCert = true
	}

	return ldapConf, nil
}

// LDAPGroupConf returns the setting of ldap group search
func LDAPGroupConf() (*models.LdapGroupConf, error) {

	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}

	ldapGroupConf := &models.LdapGroupConf{LdapGroupSearchScope: 2}
	if _, ok := cfg[common.LDAPGroupBaseDN]; ok {
		ldapGroupConf.LdapGroupBaseDN = utils.SafeCastString(cfg[common.LDAPGroupBaseDN])
	}
	if _, ok := cfg[common.LDAPGroupSearchFilter]; ok {
		ldapGroupConf.LdapGroupFilter = utils.SafeCastString(cfg[common.LDAPGroupSearchFilter])
	}
	if _, ok := cfg[common.LDAPGroupAttributeName]; ok {
		ldapGroupConf.LdapGroupNameAttribute = utils.SafeCastString(cfg[common.LDAPGroupAttributeName])
	}
	if _, ok := cfg[common.LDAPGroupSearchScope]; ok {
		if scopeStr, ok := cfg[common.LDAPGroupSearchScope].(string); ok {
			ldapGroupConf.LdapGroupSearchScope, err = strconv.Atoi(scopeStr)
		}
		if scopeFloat, ok := cfg[common.LDAPGroupSearchScope].(float64); ok {
			ldapGroupConf.LdapGroupSearchScope = int(scopeFloat)
		}
	}
	if _, ok := cfg[common.LdapGroupAdminDn]; ok {
		ldapGroupConf.LdapGroupAdminDN = cfg[common.LdapGroupAdminDn].(string)
	}
	return ldapGroupConf, nil
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration() (int, error) {
	cfg, err := mg.Get()
	if err != nil {
		return 0, err
	}

	return int(utils.SafeCastFloat64(cfg[common.TokenExpiration])), nil
}

// ExtEndpoint returns the external URL of Harbor: protocol://host:port
func ExtEndpoint() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return utils.SafeCastString(cfg[common.ExtEndpoint]), nil
}

// ExtURL returns the external URL: host:port
func ExtURL() (string, error) {
	endpoint, err := ExtEndpoint()
	if err != nil {
		return "", err
	}
	l := strings.Split(endpoint, "://")
	if len(l) > 0 {
		return l[1], nil
	}
	return endpoint, nil
}

// SecretKey returns the secret key to encrypt the password of target
func SecretKey() (string, error) {
	return keyProvider.Get(nil)
}

// SelfRegistration returns the enablement of self registration
func SelfRegistration() (bool, error) {
	cfg, err := mg.Get()
	if err != nil {
		return false, err
	}
	return utils.SafeCastBool(cfg[common.SelfRegistration]), nil
}

// RegistryURL ...
func RegistryURL() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return utils.SafeCastString(cfg[common.RegistryURL]), nil
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	cfg, err := mg.Get()
	if err != nil {
		log.Warningf("Failed to Get job service URL from backend, error: %v, will return default value.", err)
		return common.DefaultJobserviceEndpoint
	}

	if cfg[common.JobServiceURL] == nil {
		return common.DefaultJobserviceEndpoint
	}
	return strings.TrimSuffix(utils.SafeCastString(cfg[common.JobServiceURL]), "/")
}

// InternalCoreURL returns the local harbor core url
func InternalCoreURL() string {
	cfg, err := mg.Get()
	if err != nil {
		log.Warningf("Failed to Get job service Core URL from backend, error: %v, will return default value.", err)
		return common.DefaultCoreEndpoint
	}
	return strings.TrimSuffix(utils.SafeCastString(cfg[common.CoreURL]), "/")

}

// InternalTokenServiceEndpoint returns token service endpoint for internal communication between Harbor containers
func InternalTokenServiceEndpoint() string {
	return InternalCoreURL() + "/service/token"
}

// InternalNotaryEndpoint returns notary server endpoint for internal communication between Harbor containers
// This is currently a conventional value and can be unaccessible when Harbor is not deployed with Notary.
func InternalNotaryEndpoint() string {
	cfg, err := mg.Get()
	if err != nil {
		log.Warningf("Failed to get Notary endpoint from backend, error: %v, will use default value.", err)
		return common.DefaultNotaryEndpoint
	}
	if cfg[common.NotaryURL] == nil {
		return common.DefaultNotaryEndpoint
	}
	return utils.SafeCastString(cfg[common.NotaryURL])
}

// InitialAdminPassword returns the initial password for administrator
func InitialAdminPassword() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return utils.SafeCastString(cfg[common.AdminInitialPassword]), nil
}

// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject() (bool, error) {
	cfg, err := mg.Get()
	if err != nil {
		return true, err
	}
	return utils.SafeCastString(cfg[common.ProjectCreationRestriction]) == common.ProCrtRestrAdmOnly, nil
}

// Email returns email server settings
func Email() (*models.Email, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}

	email := &models.Email{}
	email.Host = utils.SafeCastString(cfg[common.EmailHost])
	email.Port = int(utils.SafeCastFloat64(cfg[common.EmailPort]))
	email.Username = utils.SafeCastString(cfg[common.EmailUsername])
	email.Password = utils.SafeCastString(cfg[common.EmailPassword])
	email.SSL = utils.SafeCastBool(cfg[common.EmailSSL])
	email.From = utils.SafeCastString(cfg[common.EmailFrom])
	email.Identity = utils.SafeCastString(cfg[common.EmailIdentity])
	email.Insecure = utils.SafeCastBool(cfg[common.EmailInsecure])

	return email, nil
}

// Database returns database settings
func Database() (*models.Database, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}

	log.Infof("database connections = %+v", cfg)

	database := &models.Database{}
	database.Type = utils.SafeCastString(cfg[common.DatabaseType])

	postgresql := &models.PostGreSQL{}
	postgresql.Host = utils.SafeCastString(cfg[common.PostGreSQLHOST])
	postgresql.Port = utils.SafeCastInt(cfg[common.PostGreSQLPort])
	postgresql.Username = utils.SafeCastString(cfg[common.PostGreSQLUsername])
	postgresql.Password = utils.SafeCastString(cfg[common.PostGreSQLPassword])
	postgresql.Database = utils.SafeCastString(cfg[common.PostGreSQLDatabase])
	postgresql.SSLMode = utils.SafeCastString(cfg[common.PostGreSQLSSLMode])
	database.PostGreSQL = postgresql

	return database, nil
}

// CoreSecret returns a secret to mark harbor-core when communicate with
// other component
func CoreSecret() string {
	return os.Getenv("CORE_SECRET")
}

// JobserviceSecret returns a secret to mark Jobservice when communicate with
// other component
// TODO replace it with method of SecretStore
func JobserviceSecret() string {
	return os.Getenv("JOBSERVICE_SECRET")
}

// WithNotary returns a bool value to indicate if Harbor's deployed with Notary
func WithNotary() bool {
	cfg, err := mg.Get()
	if err != nil {
		log.Warningf("Failed to get configuration, will return WithNotary == false")
		return false
	}
	return utils.SafeCastBool(cfg[common.WithNotary])
}

// WithClair returns a bool value to indicate if Harbor's deployed with Clair
func WithClair() bool {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration, will return WithClair == false")
		return false
	}
	return utils.SafeCastBool(cfg[common.WithClair])
}

// ClairEndpoint returns the end point of clair instance, by default it's the one deployed within Harbor.
func ClairEndpoint() string {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration, use default clair endpoint")
		return common.DefaultClairEndpoint
	}
	return utils.SafeCastString(cfg[common.ClairURL])
}

// ClairDB return Clair db info
func ClairDB() (*models.PostGreSQL, error) {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration of Clair DB, Error detail %v", err)
		return nil, err
	}
	clairDB := &models.PostGreSQL{}
	clairDB.Host = utils.SafeCastString(cfg[common.ClairDBHost])
	clairDB.Port = utils.SafeCastInt(cfg[common.ClairDBPort])
	clairDB.Username = utils.SafeCastString(cfg[common.ClairDBUsername])
	clairDB.Password = utils.SafeCastString(cfg[common.ClairDBPassword])
	clairDB.Database = utils.SafeCastString(cfg[common.ClairDB])
	clairDB.SSLMode = utils.SafeCastString(cfg[common.ClairDBSSLMode])
	return clairDB, nil
}

// AdmiralEndpoint returns the URL of admiral, if Harbor is not deployed with admiral it should return an empty string.
func AdmiralEndpoint() string {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration, will return empty string as admiral's endpoint, error: %v", err)
		return ""
	}

	if e, ok := cfg[common.AdmiralEndpoint].(string); !ok || e == "NA" {
		return ""
	}
	return utils.SafeCastString(cfg[common.AdmiralEndpoint])
}

// ScanAllPolicy returns the policy which controls the scan all.
func ScanAllPolicy() models.ScanAllPolicy {
	var res models.ScanAllPolicy
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration, will return default scan all policy, error: %v", err)
		return models.DefaultScanAllPolicy
	}
	v, ok := cfg[common.ScanAllPolicy]
	if !ok {
		return models.DefaultScanAllPolicy
	}
	b, err := json.Marshal(v)
	if err != nil {
		log.Errorf("Failed to Marshal the value in configuration for Scan All policy, error: %v, returning the default policy", err)
		return models.DefaultScanAllPolicy
	}
	if err := json.Unmarshal(b, &res); err != nil {
		log.Errorf("Failed to unmarshal the value in configuration for Scan All policy, error: %v, returning the default policy", err)
		return models.DefaultScanAllPolicy
	}
	return res
}

// WithAdmiral returns a bool to indicate if Harbor's deployed with admiral.
func WithAdmiral() bool {
	return len(AdmiralEndpoint()) > 0
}

// UAASettings returns the UAASettings to access UAA service.
func UAASettings() (*models.UAASettings, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}
	us := &models.UAASettings{
		Endpoint:     utils.SafeCastString(cfg[common.UAAEndpoint]),
		ClientID:     utils.SafeCastString(cfg[common.UAAClientID]),
		ClientSecret: utils.SafeCastString(cfg[common.UAAClientSecret]),
		VerifyCert:   utils.SafeCastBool(cfg[common.UAAVerifyCert]),
	}
	return us, nil
}

// ReadOnly returns a bool to indicates if Harbor is in read only mode.
func ReadOnly() bool {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration, will return false as read only, error: %v", err)
		return false
	}
	return utils.SafeCastBool(cfg[common.ReadOnly])
}

// WithChartMuseum returns a bool to indicate if chartmuseum is deployed with Harbor.
func WithChartMuseum() bool {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get 'with_chartmuseum' configuration with error: %s; return false as default", err.Error())
		return false
	}

	return utils.SafeCastBool(cfg[common.WithChartMuseum])
}

// GetChartMuseumEndpoint returns the endpoint of the chartmuseum service
// otherwise an non nil error is returned
func GetChartMuseumEndpoint() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get 'chart_repository_url' configuration with error: %s; return false as default", err.Error())
		return "", err
	}

	chartEndpoint := strings.TrimSpace(utils.SafeCastString(cfg[common.ChartRepoURL]))
	if len(chartEndpoint) == 0 {
		return "", errors.New("empty chartmuseum endpoint")
	}

	return chartEndpoint, nil
}
