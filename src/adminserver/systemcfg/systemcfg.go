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

	"github.com/vmware/harbor/src/adminserver/systemcfg/store"
	"github.com/vmware/harbor/src/adminserver/systemcfg/store/json"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/utils/log"
)

var cfgStore store.Driver

// Init system configurations. Read from config store first, if null read from env
func Init() (err error) {
	s := getCfgStore()
	switch s {
	case "json":
		path := os.Getenv("JSON_STORE_PATH")
		cfgStore, err = json.NewCfgStore(path)
		if err != nil {
			return
		}
	default:
		return fmt.Errorf("unsupported configuration store driver %s", s)
	}

	log.Infof("configuration store driver: %s", cfgStore.Name())
	cfg, err := cfgStore.Read()
	if err != nil {
		return err
	}

	if cfg == nil {
		log.Info("configurations read from store driver are null, initializing system from environment variables...")
		cfg, err = initFromEnv()
		if err != nil {
			return err
		}
	} else {
		if err := readFromEnv(cfg); err != nil {
			return err
		}
	}

	//sync configurations into cfg store
	if err = cfgStore.Write(cfg); err != nil {
		return err
	}

	return nil
}

func getCfgStore() string {
	t := os.Getenv("CFG_STORE_TYPE")
	if len(t) == 0 {
		t = "json"
	}
	return t
}

//read the following attrs from env every time boots up
func readFromEnv(cfg map[string]interface{}) error {
	cfg[comcfg.ExtEndpoint] = os.Getenv("EXT_ENDPOINT")

	cfg[comcfg.DatabaseType] = os.Getenv("DATABASE_TYPE")
	cfg[comcfg.MySQLHost] = os.Getenv("MYSQL_HOST")
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	if err != nil {
		return err
	}
	cfg[comcfg.MySQLPort] = port
	cfg[comcfg.MySQLUsername] = os.Getenv("MYSQL_USR")
	cfg[comcfg.MySQLPassword] = os.Getenv("MYSQL_PWD")
	cfg[comcfg.MySQLDatabase] = os.Getenv("MYSQL_DATABASE")
	cfg[comcfg.SQLiteFile] = os.Getenv("SQLITE_FILE")
	cfg[comcfg.TokenServiceURL] = os.Getenv("TOKEN_SERVICE_URL")
	tokenExpi, err := strconv.Atoi(os.Getenv("TOKEN_EXPIRATION"))
	if err != nil {
		return err
	}
	cfg[comcfg.TokenExpiration] = tokenExpi
	cfg[comcfg.RegistryURL] = os.Getenv("REGISTRY_URL")
	//TODO remove
	cfg[comcfg.JobLogDir] = os.Getenv("LOG_DIR")
	//TODO remove
	cfg[comcfg.UseCompressedJS] = os.Getenv("USE_COMPRESSED_JS") == "on"
	cfg[comcfg.SecretKey] = os.Getenv("SECRET_KEY")
	cfgExpi, err := strconv.Atoi(os.Getenv("CFG_EXPIRATION"))
	if err != nil {
		return err
	}
	cfg[comcfg.CfgExpiration] = cfgExpi
	workers, err := strconv.Atoi(os.Getenv("MAX_JOB_WORKERS"))
	if err != nil {
		return err
	}
	cfg[comcfg.MaxJobWorkers] = workers

	return nil
}

func initFromEnv() (map[string]interface{}, error) {
	cfg := map[string]interface{}{}

	if err := readFromEnv(cfg); err != nil {
		return nil, err
	}

	cfg[comcfg.AUTHMode] = os.Getenv("AUTH_MODE")
	cfg[comcfg.SelfRegistration] = os.Getenv("SELF_REGISTRATION") == "on"
	cfg[comcfg.LDAPURL] = os.Getenv("LDAP_URL")
	cfg[comcfg.LDAPSearchDN] = os.Getenv("LDAP_SEARCH_DN")
	cfg[comcfg.LDAPSearchPwd] = os.Getenv("LDAP_SEARCH_PWD")
	cfg[comcfg.LDAPBaseDN] = os.Getenv("LDAP_BASE_DN")
	cfg[comcfg.LDAPFilter] = os.Getenv("LDAP_FILTER")
	cfg[comcfg.LDAPUID] = os.Getenv("LDAP_UID")
	scope, err := strconv.Atoi(os.Getenv("LDAP_SCOPE"))
	if err != nil {
		return nil, err
	}
	cfg[comcfg.LDAPScope] = scope
	timeout, err := strconv.Atoi(os.Getenv("LDAP_TIMEOUT"))
	if err != nil {
		return nil, err
	}
	cfg[comcfg.LDAPTimeout] = timeout
	cfg[comcfg.EmailHost] = os.Getenv("EMAIL_HOST")
	port, err := strconv.Atoi(os.Getenv("EMAIL_PORT"))
	if err != nil {
		return nil, err
	}
	cfg[comcfg.EmailPort] = port
	cfg[comcfg.EmailUsername] = os.Getenv("EMAIL_USR")
	cfg[comcfg.EmailPassword] = os.Getenv("EMAIL_PWD")
	cfg[comcfg.EmailSSL] = os.Getenv("EMAIL_SSL") == "true"
	cfg[comcfg.EmailFrom] = os.Getenv("EMAIL_FROM")
	cfg[comcfg.EmailIdentity] = os.Getenv("EMAIL_IDENTITY")
	cfg[comcfg.VerifyRemoteCert] = os.Getenv("VERIFY_REMOTE_CERT") == "on"
	cfg[comcfg.ProjectCreationRestriction] = os.Getenv("PROJECT_CREATION_RESTRICTION")
	cfg[comcfg.AdminInitialPassword] = os.Getenv("HARBOR_ADMIN_PASSWORD")

	return cfg, nil
}

// GetSystemCfg returns the system configurations
func GetSystemCfg() (map[string]interface{}, error) {
	return cfgStore.Read()
}

// UpdateSystemCfg updates the system configurations
func UpdateSystemCfg(cfg map[string]interface{}) error {
	return cfgStore.Write(cfg)
}
