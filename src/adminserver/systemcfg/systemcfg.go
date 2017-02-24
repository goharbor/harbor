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

package systemcfg

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/vmware/harbor/src/adminserver/systemcfg/store"
	"github.com/vmware/harbor/src/adminserver/systemcfg/store/json"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	defaultCfgStoreDriver   string = "json"
	defaultJSONCfgStorePath string = "/etc/adminserver/config.json"
	defaultKeyPath          string = "/etc/adminserver/key"
)

var (
	cfgStore    store.Driver
	keyProvider comcfg.KeyProvider

	// attrs need to be encrypted or decrypted
	attrs = []string{
		comcfg.EmailPassword,
		comcfg.LDAPSearchPwd,
		comcfg.MySQLPassword,
		comcfg.AdminInitialPassword,
	}

	// envs are configurations need read from environment variables
	envs = map[string]interface{}{
		comcfg.ExtEndpoint: "EXT_ENDPOINT",
		comcfg.AUTHMode:    "AUTH_MODE",
		comcfg.SelfRegistration: &parser{
			env:   "SELF_REGISTRATION",
			parse: parseStringToBool,
		},
		comcfg.DatabaseType: "DATABASE_TYPE",
		comcfg.MySQLHost:    "MYSQL_HOST",
		comcfg.MySQLPort: &parser{
			env:   "MYSQL_PORT",
			parse: parseStringToInt,
		},
		comcfg.MySQLUsername: "MYSQL_USR",
		comcfg.MySQLPassword: "MYSQL_PWD",
		comcfg.MySQLDatabase: "MYSQL_DATABASE",
		comcfg.SQLiteFile:    "SQLITE_FILE",
		comcfg.LDAPURL:       "LDAP_URL",
		comcfg.LDAPSearchDN:  "LDAP_SEARCH_DN",
		comcfg.LDAPSearchPwd: "LDAP_SEARCH_PWD",
		comcfg.LDAPBaseDN:    "LDAP_BASE_DN",
		comcfg.LDAPFilter:    "LDAP_FILTER",
		comcfg.LDAPUID:       "LDAP_UID",
		comcfg.LDAPScope: &parser{
			env:   "LDAP_SCOPE",
			parse: parseStringToInt,
		},
		comcfg.LDAPTimeout: &parser{
			env:   "LDAP_TIMEOUT",
			parse: parseStringToInt,
		},
		comcfg.EmailHost: "EMAIL_HOST",
		comcfg.EmailPort: &parser{
			env:   "EMAIL_PORT",
			parse: parseStringToInt,
		},
		comcfg.EmailUsername: "EMAIL_USR",
		comcfg.EmailPassword: "EMAIL_PWD",
		comcfg.EmailSSL: &parser{
			env:   "EMAIL_SSL",
			parse: parseStringToBool,
		},
		comcfg.EmailFrom:     "EMAIL_FROM",
		comcfg.EmailIdentity: "EMAIL_IDENTITY",
		comcfg.RegistryURL:   "REGISTRY_URL",
		comcfg.TokenExpiration: &parser{
			env:   "TOKEN_EXPIRATION",
			parse: parseStringToInt,
		},
		comcfg.JobLogDir: "LOG_DIR",
		comcfg.UseCompressedJS: &parser{
			env:   "USE_COMPRESSED_JS",
			parse: parseStringToBool,
		},
		comcfg.CfgExpiration: &parser{
			env:   "CFG_EXPIRATION",
			parse: parseStringToInt,
		},
		comcfg.MaxJobWorkers: &parser{
			env:   "MAX_JOB_WORKERS",
			parse: parseStringToInt,
		},
		comcfg.VerifyRemoteCert: &parser{
			env:   "VERIFY_REMOTE_CERT",
			parse: parseStringToBool,
		},
		comcfg.ProjectCreationRestriction: "PROJECT_CREATION_RESTRICTION",
		comcfg.AdminInitialPassword:       "HARBOR_ADMIN_PASSWORD",
	}
)

type parser struct {
	// the name of env
	env string
	// parse the value of env, e.g. parse string to int or
	// parse string to bool
	parse func(string) (interface{}, error)
}

func parseStringToInt(str string) (interface{}, error) {
	return strconv.Atoi(str)
}

func parseStringToBool(str string) (interface{}, error) {
	return strings.ToLower(str) == "true" ||
		strings.ToLower(str) == "on", nil
}

// Init system configurations. Read from config store first,
// if null read from env
func Init() (err error) {
	//init configuation store
	if err = initCfgStore(); err != nil {
		return err
	}

	//init key provider
	initKeyProvider()

	cfg, err := GetSystemCfg()
	if err != nil {
		return err
	}

	if cfg != nil {
		return nil
	}

	log.Info("configurations read from store driver are null, initializing system from environment variables...")
	cfg, err = loadFromEnv()
	if err != nil {
		return err
	}

	//sync configurations into cfg store
	return UpdateSystemCfg(cfg)
}

func initCfgStore() (err error) {
	t := os.Getenv("CFG_STORE_DRIVER")
	if len(t) == 0 {
		t = defaultCfgStoreDriver
	}
	log.Infof("configuration store driver: %s", t)

	switch t {
	case "json":
		path := os.Getenv("JSON_CFG_STORE_PATH")
		if len(path) == 0 {
			path = defaultJSONCfgStorePath
		}
		log.Infof("json configuration store path: %s", path)

		cfgStore, err = json.NewCfgStore(path)
	default:
		err = fmt.Errorf("unsupported configuration store driver %s", t)
	}

	return err
}

func initKeyProvider() {
	path := os.Getenv("KEY_PATH")
	if len(path) == 0 {
		path = defaultKeyPath
	}
	log.Infof("key path: %s", path)

	keyProvider = comcfg.NewFileKeyProvider(path)
}

//load the configurations from env
func loadFromEnv() (map[string]interface{}, error) {
	cfg := map[string]interface{}{}

	for k, v := range envs {
		if str, ok := v.(string); ok {
			cfg[k] = os.Getenv(str)
			continue
		}

		if parser, ok := v.(*parser); ok {
			i, err := parser.parse(os.Getenv(parser.env))
			if err != nil {
				return nil, err
			}
			cfg[k] = i
			continue
		}

		return nil, fmt.Errorf("%v is not string or parse type", v)
	}

	return cfg, nil
}

// GetSystemCfg returns the system configurations
func GetSystemCfg() (map[string]interface{}, error) {
	m, err := cfgStore.Read()
	if err != nil {
		return nil, err
	}

	key, err := keyProvider.Get(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %v", err)
	}

	if err = decrypt(m, attrs, key); err != nil {
		return nil, err
	}

	return m, nil
}

// UpdateSystemCfg updates the system configurations
func UpdateSystemCfg(cfg map[string]interface{}) error {

	key, err := keyProvider.Get(nil)
	if err != nil {
		return fmt.Errorf("failed to get key: %v", err)
	}

	if err := encrypt(cfg, attrs, key); err != nil {
		return err
	}

	return cfgStore.Write(cfg)
}

func encrypt(m map[string]interface{}, keys []string, secretKey string) error {
	for _, key := range keys {
		v, ok := m[key]
		if !ok {
			continue
		}

		if len(v.(string)) == 0 {
			continue
		}

		cipherText, err := utils.ReversibleEncrypt(v.(string), secretKey)
		if err != nil {
			return err
		}
		m[key] = cipherText
	}
	return nil
}

func decrypt(m map[string]interface{}, keys []string, secretKey string) error {
	for _, key := range keys {
		v, ok := m[key]
		if !ok {
			continue
		}

		if len(v.(string)) == 0 {
			continue
		}

		text, err := utils.ReversibleDecrypt(v.(string), secretKey)
		if err != nil {
			return err
		}
		m[key] = text
	}
	return nil
}
