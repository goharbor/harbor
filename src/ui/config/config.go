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
	"fmt"
	"os"
	"strings"

	"github.com/vmware/harbor/src/adminserver/client"
	"github.com/vmware/harbor/src/adminserver/client/auth"
	"github.com/vmware/harbor/src/common"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	defaultKeyPath   string = "/etc/ui/key"
	secretCookieName string = "secret"
)

var (
	// AdminserverClient is a client for adminserver
	AdminserverClient client.Client
	mg                *comcfg.Manager
	keyProvider       comcfg.KeyProvider
)

// Init configurations
func Init() error {
	//init key provider
	initKeyProvider()

	adminServerURL := os.Getenv("ADMIN_SERVER_URL")
	if len(adminServerURL) == 0 {
		adminServerURL = "http://adminserver"
	}

	log.Infof("initializing client for adminserver %s ...", adminServerURL)
	authorizer := auth.NewSecretAuthorizer(secretCookieName, UISecret())
	AdminserverClient = client.NewClient(adminServerURL, authorizer)
	if err := AdminserverClient.Ping(); err != nil {
		return fmt.Errorf("failed to ping adminserver: %v", err)
	}

	mg = comcfg.NewManager(AdminserverClient, true)

	if err := Load(); err != nil {
		return err
	}

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

// Load configurations
func Load() error {
	_, err := mg.Load()
	return err
}

// Reset configurations
func Reset() error {
	return mg.Reset()
}

// Upload uploads all system configutations to admin server
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

// LDAP returns the setting of ldap server
func LDAP() (*models.LDAP, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}

	ldap := &models.LDAP{}
	ldap.URL = cfg[common.LDAPURL].(string)
	ldap.SearchDN = cfg[common.LDAPSearchDN].(string)
	ldap.SearchPassword = cfg[common.LDAPSearchPwd].(string)
	ldap.BaseDN = cfg[common.LDAPBaseDN].(string)
	ldap.UID = cfg[common.LDAPUID].(string)
	ldap.Filter = cfg[common.LDAPFilter].(string)
	ldap.Scope = int(cfg[common.LDAPScope].(float64))
	ldap.Timeout = int(cfg[common.LDAPTimeout].(float64))

	return ldap, nil
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
	return "http://jobservice"
}

// InternalTokenServiceEndpoint returns token service endpoint for internal communication between Harbor containers
func InternalTokenServiceEndpoint() string {
	return "http://ui/service/token"
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

// VerifyRemoteCert returns bool value.
func VerifyRemoteCert() (bool, error) {
	cfg, err := mg.Get()
	if err != nil {
		return true, err
	}
	return cfg[common.VerifyRemoteCert].(bool), nil
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

// AdmiralEndpoint returns the URL of admiral, if Harbor is not deployed with admiral it should return an empty string.
func AdmiralEndpoint() string {
	cfg, err := mg.Get()
	if err != nil {
		log.Errorf("Failed to get configuration, will return empty string as admiral's endpoint")

		return ""
	}
	if e, ok := cfg[common.AdmiralEndpoint].(string); !ok || e == "NA" {
		cfg[common.AdmiralEndpoint] = ""
	}
	return cfg[common.AdmiralEndpoint].(string)
}

// WithAdmiral returns a bool to indicate if Harbor's deployed with admiral.
func WithAdmiral() bool {
	return len(AdmiralEndpoint()) > 0
}
