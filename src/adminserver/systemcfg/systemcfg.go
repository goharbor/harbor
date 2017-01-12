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
	"github.com/vmware/harbor/src/common/models"
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
func readFromEnv(cfg *models.SystemCfg) error {
	cfg.DomainName = os.Getenv("EXT_ENDPOINT")

	cfg.Database = &models.Database{
		Type: os.Getenv("DATABASE_TYPE"),
		MySQL: &models.MySQL{
			Host:     os.Getenv("MYSQL_HOST"),
			Username: os.Getenv("MYSQL_USR"),
			Password: os.Getenv("MYSQL_PWD"),
			Database: os.Getenv("MYSQL_DATABASE"),
		},
		SQLite: &models.SQLite{
			File: os.Getenv("SQLITE_FILE"),
		},
	}
	port, err := strconv.Atoi(os.Getenv("MYSQL_PORT"))
	if err != nil {
		return err
	}
	cfg.Database.MySQL.Port = port

	cfg.TokenService = &models.TokenService{
		URL: os.Getenv("TOKEN_SERVICE_URL"),
	}
	cfg.Registry = &models.Registry{
		URL: os.Getenv("REGISTRY_URL"),
	}

	//TODO remove
	cfg.JobLogDir = os.Getenv("LOG_DIR")
	//TODO remove
	cfg.CompressJS = os.Getenv("USE_COMPRESSED_JS") == "on"
	exp, err := strconv.Atoi(os.Getenv("TOKEN_EXPIRATION"))
	if err != nil {
		return err
	}
	cfg.TokenExpiration = exp
	cfg.SecretKey = os.Getenv("SECRET_KEY")

	cfgExp, err := strconv.Atoi(os.Getenv("CFG_EXPIRATION"))
	if err != nil {
		return err
	}
	cfg.CfgExpiration = cfgExp

	workers, err := strconv.Atoi(os.Getenv("MAX_JOB_WORKERS"))
	if err != nil {
		return err
	}
	cfg.MaxJobWorkers = workers

	return nil
}

func initFromEnv() (*models.SystemCfg, error) {
	cfg := &models.SystemCfg{}

	if err := readFromEnv(cfg); err != nil {
		return nil, err
	}

	cfg.Authentication = &models.Authentication{
		Mode:             os.Getenv("AUTH_MODE"),
		SelfRegistration: os.Getenv("SELF_REGISTRATION") == "on",
		LDAP: &models.LDAP{
			URL:       os.Getenv("LDAP_URL"),
			SearchDN:  os.Getenv("LDAP_SEARCH_DN"),
			SearchPwd: os.Getenv("LDAP_SEARCH_PWD"),
			BaseDN:    os.Getenv("LDAP_BASE_DN"),
			Filter:    os.Getenv("LDAP_FILTER"),
			UID:       os.Getenv("LDAP_UID"),
		},
	}
	scope, err := strconv.Atoi(os.Getenv("LDAP_SCOPE"))
	if err != nil {
		return nil, err
	}
	cfg.Authentication.LDAP.Scope = scope
	timeout, err := strconv.Atoi(os.Getenv("LDAP_TIMEOUT"))
	if err != nil {
		return nil, err
	}
	cfg.Authentication.LDAP.Timeout = timeout

	cfg.Email = &models.Email{
		Host:     os.Getenv("EMAIL_HOST"),
		Port:     os.Getenv("EMAIL_PORT"),
		Username: os.Getenv("EMAIL_USR"),
		Password: os.Getenv("EMAIL_PWD"),
		SSL:      os.Getenv("EMAIL_SSL") == "true",
		From:     os.Getenv("EMAIL_FROM"),
		Identity: os.Getenv("EMAIL_IDENTITY"),
	}
	cfg.VerifyRemoteCert = os.Getenv("VERIFY_REMOTE_CERT") == "on"
	cfg.ProjectCreationRestriction = os.Getenv("PROJECT_CREATION_RESTRICTION")

	cfg.InitialAdminPwd = os.Getenv("HARBOR_ADMIN_PASSWORD")
	return cfg, nil
}

// GetSystemCfg returns the system configurations
func GetSystemCfg() (*models.SystemCfg, error) {
	return cfgStore.Read()
}

// UpdateSystemCfg updates the system configurations
func UpdateSystemCfg(cfg *models.SystemCfg) error {
	return cfgStore.Write(cfg)
}
