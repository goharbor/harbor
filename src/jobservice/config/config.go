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
	"os"

	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/models"
)

var mg *comcfg.Manager

// Init configurations
func Init() error {
	adminServerURL := os.Getenv("ADMIN_SERVER_URL")
	if len(adminServerURL) == 0 {
		adminServerURL = "http://adminserver"
	}
	mg = comcfg.NewManager(adminServerURL, UISecret(), true)

	if err := mg.Init(); err != nil {
		return err
	}

	if _, err := mg.Load(); err != nil {
		return err
	}

	return nil
}

// VerifyRemoteCert returns bool value.
func VerifyRemoteCert() (bool, error) {
	cfg, err := mg.Get()
	if err != nil {
		return true, err
	}
	return cfg[comcfg.VerifyRemoteCert].(bool), nil
}

// Database ...
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

// MaxJobWorkers ...
func MaxJobWorkers() (int, error) {
	cfg, err := mg.Get()
	if err != nil {
		return 0, err
	}
	return int(cfg[comcfg.MaxJobWorkers].(float64)), nil
}

// LocalUIURL returns the local ui url, job service will use this URL to call API hosted on ui process
func LocalUIURL() string {
	return "http://ui"
}

// LocalRegURL returns the local registry url, job service will use this URL to pull image from the registry
func LocalRegURL() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[comcfg.RegistryURL].(string), nil
}

// LogDir returns the absolute path to which the log file will be written
func LogDir() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[comcfg.JobLogDir].(string), nil
}

// SecretKey will return the secret key for encryption/decryption password in target.
func SecretKey() (string, error) {
	cfg, err := mg.Get()
	if err != nil {
		return "", err
	}
	return cfg[comcfg.SecretKey].(string), nil
}

// UISecret returns the value of UI secret cookie, used for communication between UI and JobService
// TODO
func UISecret() string {
	return os.Getenv("UI_SECRET")
}
