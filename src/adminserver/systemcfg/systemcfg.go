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

package systemcfg

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	enpt "github.com/vmware/harbor/src/adminserver/systemcfg/encrypt"
	"github.com/vmware/harbor/src/adminserver/systemcfg/store"
	"github.com/vmware/harbor/src/adminserver/systemcfg/store/encrypt"
	"github.com/vmware/harbor/src/adminserver/systemcfg/store/json"
	"github.com/vmware/harbor/src/common"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/utils/log"
)

const (
	defaultJSONCfgStorePath string = "/etc/adminserver/config/config.json"
	defaultKeyPath          string = "/etc/adminserver/key"
)

var (
	// CfgStore is a storage driver that configurations
	// can be read from and wrote to
	CfgStore store.Driver

	// attrs need to be encrypted or decrypted
	attrs = []string{
		common.EmailPassword,
		common.LDAPSearchPwd,
		common.MySQLPassword,
		common.AdminInitialPassword,
		common.ClairDBPassword,
		common.UAAClientSecret,
	}

	// all configurations need read from environment variables
	allEnvs = map[string]interface{}{
		common.ExtEndpoint: "EXT_ENDPOINT",
		common.AUTHMode:    "AUTH_MODE",
		common.SelfRegistration: &parser{
			env:   "SELF_REGISTRATION",
			parse: parseStringToBool,
		},
		common.DatabaseType: "DATABASE_TYPE",
		common.MySQLHost:    "MYSQL_HOST",
		common.MySQLPort: &parser{
			env:   "MYSQL_PORT",
			parse: parseStringToInt,
		},
		common.MySQLUsername: "MYSQL_USR",
		common.MySQLPassword: "MYSQL_PWD",
		common.MySQLDatabase: "MYSQL_DATABASE",
		common.SQLiteFile:    "SQLITE_FILE",
		common.LDAPURL:       "LDAP_URL",
		common.LDAPSearchDN:  "LDAP_SEARCH_DN",
		common.LDAPSearchPwd: "LDAP_SEARCH_PWD",
		common.LDAPBaseDN:    "LDAP_BASE_DN",
		common.LDAPFilter:    "LDAP_FILTER",
		common.LDAPUID:       "LDAP_UID",
		common.LDAPScope: &parser{
			env:   "LDAP_SCOPE",
			parse: parseStringToInt,
		},
		common.LDAPTimeout: &parser{
			env:   "LDAP_TIMEOUT",
			parse: parseStringToInt,
		},
		common.EmailHost: "EMAIL_HOST",
		common.EmailPort: &parser{
			env:   "EMAIL_PORT",
			parse: parseStringToInt,
		},
		common.EmailUsername: "EMAIL_USR",
		common.EmailPassword: "EMAIL_PWD",
		common.EmailSSL: &parser{
			env:   "EMAIL_SSL",
			parse: parseStringToBool,
		},
		common.EmailInsecure: &parser{
			env:   "EMAIL_INSECURE",
			parse: parseStringToBool,
		},
		common.EmailFrom:     "EMAIL_FROM",
		common.EmailIdentity: "EMAIL_IDENTITY",
		common.RegistryURL:   "REGISTRY_URL",
		common.TokenExpiration: &parser{
			env:   "TOKEN_EXPIRATION",
			parse: parseStringToInt,
		},
		common.CfgExpiration: &parser{
			env:   "CFG_EXPIRATION",
			parse: parseStringToInt,
		},
		common.MaxJobWorkers: &parser{
			env:   "MAX_JOB_WORKERS",
			parse: parseStringToInt,
		},
		common.ProjectCreationRestriction: "PROJECT_CREATION_RESTRICTION",
		common.AdminInitialPassword:       "HARBOR_ADMIN_PASSWORD",
		common.AdmiralEndpoint:            "ADMIRAL_URL",
		common.WithNotary: &parser{
			env:   "WITH_NOTARY",
			parse: parseStringToBool,
		},
		common.WithClair: &parser{
			env:   "WITH_CLAIR",
			parse: parseStringToBool,
		},
		common.ClairDBPassword: "CLAIR_DB_PASSWORD",
		common.UAAEndpoint:     "UAA_ENDPOINT",
		common.UAAClientID:     "UAA_CLIENTID",
		common.UAAClientSecret: "UAA_CLIENTSECRET",
		common.UIURL:           "UI_URL",
		common.JobServiceURL:   "JOBSERVICE_URL",
	}

	// configurations need read from environment variables
	// every time the system startup
	repeatLoadEnvs = map[string]interface{}{
		common.ExtEndpoint:   "EXT_ENDPOINT",
		common.MySQLPassword: "MYSQL_PWD",
		common.MaxJobWorkers: &parser{
			env:   "MAX_JOB_WORKERS",
			parse: parseStringToInt,
		},
		common.CfgExpiration: &parser{
			env:   "CFG_EXPIRATION",
			parse: parseStringToInt,
		},
		common.AdmiralEndpoint: "ADMIRAL_URL",
		common.WithNotary: &parser{
			env:   "WITH_NOTARY",
			parse: parseStringToBool,
		},
		common.WithClair: &parser{
			env:   "WITH_CLAIR",
			parse: parseStringToBool,
		},
		common.ClairDBPassword: "CLAIR_DB_PASSWORD",
		common.UAAEndpoint:     "UAA_ENDPOINT",
		common.UAAClientID:     "UAA_CLIENTID",
		common.UAAClientSecret: "UAA_CLIENTSECRET",
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
	if len(str) == 0 {
		return 0, nil
	}
	return strconv.Atoi(str)
}

func parseStringToBool(str string) (interface{}, error) {
	return strings.ToLower(str) == "true" ||
		strings.ToLower(str) == "on", nil
}

// Init system configurations. If env RESET is set or configurations
// read from storage driver is null, load all configurations from env
func Init() (err error) {
	if err = initCfgStore(); err != nil {
		return err
	}

	loadAll := false
	cfgs := map[string]interface{}{}

	if os.Getenv("RESET") == "true" {
		log.Info("RESET is set, will load all configurations from environment variables")
		loadAll = true
	}

	if !loadAll {
		cfgs, err = CfgStore.Read()
		if cfgs == nil {
			log.Info("configurations read from storage driver are null, will load them from environment variables")
			loadAll = true
			cfgs = map[string]interface{}{}
		}
	}

	if err = LoadFromEnv(cfgs, loadAll); err != nil {
		return err
	}

	return CfgStore.Write(cfgs)
}

func initCfgStore() (err error) {
	path := os.Getenv("JSON_CFG_STORE_PATH")
	if len(path) == 0 {
		path = defaultJSONCfgStorePath
	}
	log.Infof("the path of json configuration storage: %s", path)

	CfgStore, err = json.NewCfgStore(path)
	if err != nil {
		return
	}

	kp := os.Getenv("KEY_PATH")
	if len(kp) == 0 {
		kp = defaultKeyPath
	}
	log.Infof("the path of key used by key provider: %s", kp)

	encryptor := enpt.NewAESEncryptor(
		comcfg.NewFileKeyProvider(kp), nil)

	CfgStore = encrypt.NewCfgStore(encryptor, attrs, CfgStore)
	return nil
}

// LoadFromEnv loads the configurations from allEnvs, if all is false, it just loads
// the repeatLoadEnvs and the env which is absent in cfgs
func LoadFromEnv(cfgs map[string]interface{}, all bool) error {
	var envs map[string]interface{}

	if all {
		envs = allEnvs
	} else {
		envs = make(map[string]interface{})
		for k, v := range repeatLoadEnvs {
			envs[k] = v
		}
		for k, v := range allEnvs {
			if _, exist := cfgs[k]; !exist {
				envs[k] = v
			}
		}
	}

	for k, v := range envs {
		if str, ok := v.(string); ok {
			cfgs[k] = os.Getenv(str)
			continue
		}

		if parser, ok := v.(*parser); ok {
			i, err := parser.parse(os.Getenv(parser.env))
			if err != nil {
				return err
			}
			cfgs[k] = i
			continue
		}

		return fmt.Errorf("%v is not string or parse type", v)
	}

	return nil
}
