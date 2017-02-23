/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package config

import (
	"encoding/json"
	"os"

	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

const defaultKeyPath string = "/etc/ui/key"

var (
	mg          *comcfg.Manager
	keyProvider comcfg.KeyProvider
)

// Init configurations
func Init() error {
	//init key provider
	initKeyProvider()

	adminServerURL := os.Getenv("ADMIN_SERVER_URL")
	if len(adminServerURL) == 0 {
		adminServerURL = "http://adminserver"
	}
	log.Debugf("admin server URL: %s", adminServerURL)
	mg = comcfg.NewManager(adminServerURL, UISecret(), true)

	if err := mg.Init(); err != nil {
		return err
	}

	if _, err := mg.Load(); err != nil {
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

// Upload uploads all system configutations to admin server
func Upload(cfg map[string]interface{}) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return mg.Upload(b)
}

// GetSystemCfg returns the system configurations
func GetSystemCfg() (map[string]interface{}, error) {
	raw, err := mg.Loader.Load()
	if err != nil {
		return nil, err
	}

	c, err := mg.Parser.Parse(raw)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// AuthMode ...
func AuthMode() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[comcfg.AUTHMode].(string), nil
}

// LDAP returns the setting of ldap server
func LDAP() (*models.LDAP, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}

	ldap := &models.LDAP{}
	ldap.URL = cfg[comcfg.LDAPURL].(string)
	ldap.SearchDN = cfg[comcfg.LDAPSearchDN].(string)
	ldap.SearchPassword = cfg[comcfg.LDAPSearchPwd].(string)
	ldap.BaseDN = cfg[comcfg.LDAPBaseDN].(string)
	ldap.UID = cfg[comcfg.LDAPUID].(string)
	ldap.Filter = cfg[comcfg.LDAPFilter].(string)
	ldap.Scope = int(cfg[comcfg.LDAPScope].(float64))
	ldap.Timeout = int(cfg[comcfg.LDAPTimeout].(float64))

	return ldap, nil
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration() (int, error) {
	cfg, err := mg.Get()
	if err != nil {
		return 0, err
	}
	return int(cfg[comcfg.TokenExpiration].(float64)), nil
}

// ExtEndpoint returns the external URL of Harbor: protocal://host:port
func ExtEndpoint() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[comcfg.ExtEndpoint].(string), nil
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
	return cfg[comcfg.SelfRegistration].(bool), nil
}

// RegistryURL ...
func RegistryURL() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[comcfg.RegistryURL].(string), nil
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	return "http://jobservice"
}

// InternalTokenServiceEndpoint returns token service endpoint for internal communication between Harbor containers
func InternalTokenServiceEndpoint() string {
	return "http://ui/service/token"
}

// InitialAdminPassword returns the initial password for administrator
func InitialAdminPassword() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[comcfg.AdminInitialPassword].(string), nil
}

// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject() (bool, error) {
	cfg, err := mg.Get()
	if err != nil {
		return true, err
	}
	return cfg[comcfg.ProjectCreationRestriction].(string) == comcfg.ProCrtRestrAdmOnly, nil
}

// VerifyRemoteCert returns bool value.
func VerifyRemoteCert() (bool, error) {
	cfg, err := mg.Get()
	if err != nil {
		return true, err
	}
	return cfg[comcfg.VerifyRemoteCert].(bool), nil
}

// Email returns email server settings
func Email() (*models.Email, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}

	email := &models.Email{}
	email.Host = cfg[comcfg.EmailHost].(string)
	email.Port = int(cfg[comcfg.EmailPort].(float64))
	email.Username = cfg[comcfg.EmailUsername].(string)
	email.Password = cfg[comcfg.EmailPassword].(string)
	email.SSL = cfg[comcfg.EmailSSL].(bool)
	email.From = cfg[comcfg.EmailFrom].(string)
	email.Identity = cfg[comcfg.EmailIdentity].(string)

	return email, nil
}

// Database returns database settings
func Database() (*models.Database, error) {
	cfg, err := mg.Get()
	if err != nil {
		return nil, err
	}
	database := &models.Database{}
	database.Type = cfg[comcfg.DatabaseType].(string)
	mysql := &models.MySQL{}
	mysql.Host = cfg[comcfg.MySQLHost].(string)
	mysql.Port = int(cfg[comcfg.MySQLPort].(float64))
	mysql.Username = cfg[comcfg.MySQLUsername].(string)
	mysql.Password = cfg[comcfg.MySQLPassword].(string)
	mysql.Database = cfg[comcfg.MySQLDatabase].(string)
	database.MySQL = mysql
	sqlite := &models.SQLite{}
	sqlite.File = cfg[comcfg.SQLiteFile].(string)
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
