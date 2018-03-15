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

package config

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/vmware/harbor/src/adminserver/client"
	"github.com/vmware/harbor/src/common"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/secret"
	"github.com/vmware/harbor/src/common/utils/log"
	jobservice_client "github.com/vmware/harbor/src/jobservice/client"
	"github.com/vmware/harbor/src/ui/promgr"
	"github.com/vmware/harbor/src/ui/promgr/pmsdriver"
	"github.com/vmware/harbor/src/ui/promgr/pmsdriver/admiral"
	"github.com/vmware/harbor/src/ui/promgr/pmsdriver/local"
)

const (
	defaultKeyPath       string = "/etc/ui/key"
	defaultTokenFilePath string = "/etc/ui/token/tokens.properties"
	secretCookieName     string = "secret"
)

var (
	// SecretStore manages secrets
	SecretStore *secret.Store
	// AdminserverClient is a client for adminserver
	AdminserverClient client.Client
	// GlobalProjectMgr is initialized based on the deploy mode
	GlobalProjectMgr promgr.ProjectManager
	mg               *comcfg.Manager
	keyProvider      comcfg.KeyProvider
	// AdmiralClient is initialized only under integration deploy mode
	// and can be passed to project manager as a parameter
	AdmiralClient *http.Client
	// TokenReader is used in integration mode to read token
	TokenReader admiral.TokenReader
	// GlobalJobserviceClient is a global client for jobservice
	GlobalJobserviceClient jobservice_client.Client
)

// Init configurations
func Init() error {
	//init key provider
	initKeyProvider()
	adminServerURL := os.Getenv("ADMINSERVER_URL")
	if len(adminServerURL) == 0 {
		adminServerURL = "http://adminserver"
	}

	return InitByURL(adminServerURL)

}

// InitByURL Init configurations with given url
func InitByURL(adminServerURL string) error {
	log.Infof("initializing client for adminserver %s ...", adminServerURL)
	cfg := &client.Config{
		Secret: UISecret(),
	}
	AdminserverClient = client.NewClient(adminServerURL, cfg)
	if err := AdminserverClient.Ping(); err != nil {
		return fmt.Errorf("failed to ping adminserver: %v", err)
	}

	mg = comcfg.NewManager(AdminserverClient, true)

	if err := Load(); err != nil {
		return err
	}

	// init secret store
	initSecretStore()

	// init project manager based on deploy mode
	initProjectManager()

	GlobalJobserviceClient = jobservice_client.NewDefaultClient(InternalJobServiceURL(),
		&jobservice_client.Config{
			Secret: UISecret(),
		})

	return nil
}

func initKeyProvider() {
	path := os.Getenv("KEY_PATH")
	if len(path) == 0 {
		path = defaultKeyPath
	}
	log.Infof("key path: %s", path)

	keyProvider = comcfg.NewFileKeyProvider(path)
}

func initSecretStore() {
	m := map[string]string{}
	m[JobserviceSecret()] = secret.JobserviceUser
	SecretStore = secret.NewStore(m)
}

func initProjectManager() {
	var driver pmsdriver.PMSDriver
	if WithAdmiral() {
		// integration with admiral
		log.Info("initializing the project manager based on PMS...")
		// TODO read ca/cert file and pass it to the TLS config
		AdmiralClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}

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
	return cfg[common.AUTHMode].(string), nil
}

// LDAPConf returns the setting of ldap server
func LDAPConf() (*models.LdapConf, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}
	ldapConf := &models.LdapConf{}
	ldapConf.LdapURL = cfg[common.LDAPURL].(string)
	ldapConf.LdapSearchDn = cfg[common.LDAPSearchDN].(string)
	ldapConf.LdapSearchPassword = cfg[common.LDAPSearchPwd].(string)
	ldapConf.LdapBaseDn = cfg[common.LDAPBaseDN].(string)
	ldapConf.LdapUID = cfg[common.LDAPUID].(string)
	ldapConf.LdapFilter = cfg[common.LDAPFilter].(string)
	ldapConf.LdapScope = int(cfg[common.LDAPScope].(float64))
	ldapConf.LdapConnectionTimeout = int(cfg[common.LDAPTimeout].(float64))
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
		ldapGroupConf.LdapGroupBaseDN = cfg[common.LDAPGroupBaseDN].(string)
	}
	if _, ok := cfg[common.LDAPGroupSearchFilter]; ok {
		ldapGroupConf.LdapGroupFilter = cfg[common.LDAPGroupSearchFilter].(string)
	}
	if _, ok := cfg[common.LDAPGroupAttributeName]; ok {
		ldapGroupConf.LdapGroupNameAttribute = cfg[common.LDAPGroupAttributeName].(string)
	}
	if _, ok := cfg[common.LDAPGroupSearchScope]; ok {
		if scopeStr, ok := cfg[common.LDAPGroupSearchScope].(string); ok {
			ldapGroupConf.LdapGroupSearchScope, err = strconv.Atoi(scopeStr)
		}
		if scopeFloat, ok := cfg[common.LDAPGroupSearchScope].(float64); ok {
			ldapGroupConf.LdapGroupSearchScope = int(scopeFloat)
		}
	}
	return ldapGroupConf, nil
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration() (int, error) {
	cfg, err := mg.Get()
	if err != nil {
		return 0, err
	}

	return int(cfg[common.TokenExpiration].(float64)), nil
}

// ExtEndpoint returns the external URL of Harbor: protocol://host:port
func ExtEndpoint() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[common.ExtEndpoint].(string), nil
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
	return cfg[common.SelfRegistration].(bool), nil
}

// RegistryURL ...
func RegistryURL() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[common.RegistryURL].(string), nil
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	cfg, err := mg.Get()
	if err != nil {
		log.Warningf("Failed to Get job service URL from backend, error: %v, will return default value.")

		return "http://jobservice"
	}

	if cfg[common.JobServiceURL] == nil {
		return "http://jobservice"
	}
	return strings.TrimSuffix(cfg[common.JobServiceURL].(string), "/")
}

// InternalTokenServiceEndpoint returns token service endpoint for internal communication between Harbor containers
func InternalTokenServiceEndpoint() string {
	uiURL := "http://ui"
	cfg, err := mg.Get()
	if err != nil {
		log.Warningf("Failed to Get job service UI URL from backend, error: %v, will use default value.")

	} else {
		uiURL = cfg[common.UIURL].(string)
	}
	return strings.TrimSuffix(uiURL, "/") + "/service/token"
}

// InternalNotaryEndpoint returns notary server endpoint for internal communication between Harbor containers
// This is currently a conventional value and can be unaccessible when Harbor is not deployed with Notary.
func InternalNotaryEndpoint() string {
	return "http://notary-server:4443"
}

// InitialAdminPassword returns the initial password for administrator
func InitialAdminPassword() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[common.AdminInitialPassword].(string), nil
}

// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject() (bool, error) {
	cfg, err := mg.Get()
	if err != nil {
		return true, err
	}
	return cfg[common.ProjectCreationRestriction].(string) == common.ProCrtRestrAdmOnly, nil
}

// Email returns email server settings
func Email() (*models.Email, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}

	email := &models.Email{}
	email.Host = cfg[common.EmailHost].(string)
	email.Port = int(cfg[common.EmailPort].(float64))
	email.Username = cfg[common.EmailUsername].(string)
	email.Password = cfg[common.EmailPassword].(string)
	email.SSL = cfg[common.EmailSSL].(bool)
	email.From = cfg[common.EmailFrom].(string)
	email.Identity = cfg[common.EmailIdentity].(string)
	email.Insecure = cfg[common.EmailInsecure].(bool)

	return email, nil
}

// Database returns database settings
func Database() (*models.Database, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}
	database := &models.Database{}
	database.Type = cfg[common.DatabaseType].(string)
	mysql := &models.MySQL{}
	mysql.Host = cfg[common.MySQLHost].(string)
	mysql.Port = int(cfg[common.MySQLPort].(float64))
	mysql.Username = cfg[common.MySQLUsername].(string)
	mysql.Password = cfg[common.MySQLPassword].(string)
	mysql.Database = cfg[common.MySQLDatabase].(string)
	database.MySQL = mysql
	sqlite := &models.SQLite{}
	sqlite.File = cfg[common.SQLiteFile].(string)
	database.SQLite = sqlite

	return database, nil
}

// UISecret returns a secret to mark UI when communicate with
// other component
func UISecret() string {
	return os.Getenv("UI_SECRET")
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
		log.Errorf("Failed to get configuration, will return WithNotary == false")
		return false
	}
	return cfg[common.WithNotary].(bool)
}

// WithClair returns a bool value to indicate if Harbor's deployed with Clair
func WithClair() bool {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration, will return WithClair == false")
		return false
	}
	return cfg[common.WithClair].(bool)
}

// ClairEndpoint returns the end point of clair instance, by default it's the one deployed within Harbor.
func ClairEndpoint() string {
	return common.DefaultClairEndpoint
}

// ClairDB return Clair db info
func ClairDB() (*models.PostGreSQL, error) {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration of Clair DB, Error detail %v", err)
		return nil, err
	}
	clairDB := &models.PostGreSQL{}
	clairDB.Host = cfg[common.ClairDBHost].(string)
	clairDB.Port = int(cfg[common.ClairDBPort].(float64))
	clairDB.Username = cfg[common.ClairDBUsername].(string)
	clairDB.Password = cfg[common.ClairDBPassword].(string)
	clairDB.Database = cfg[common.ClairDB].(string)
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
	return cfg[common.AdmiralEndpoint].(string)
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

//UAASettings returns the UAASettings to access UAA service.
func UAASettings() (*models.UAASettings, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}
	us := &models.UAASettings{
		Endpoint:     cfg[common.UAAEndpoint].(string),
		ClientID:     cfg[common.UAAClientID].(string),
		ClientSecret: cfg[common.UAAClientSecret].(string),
		VerifyCert:   cfg[common.UAAVerifyCert].(bool),
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
	return cfg[common.ReadOnly].(bool)
}
