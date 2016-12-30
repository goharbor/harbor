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
	"time"

	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/models"
	//"github.com/vmware/harbor/src/common/utils/log"
)

var mg *comcfg.Manager

// Configuration holds configurations of Jobservice
type Configuration struct {
	Database         *models.Database `json:"database"`
	Registry         *models.Registry `json:"registry"`
	VerifyRemoteCert bool             `json:"verify_remote_cert"`
	MaxJobWorkers    int              `json:"max_job_workers"`
	JobLogDir        string           `json:"job_log_dir"`
	SecretKey        string           `json:"secret_key"`
	CfgExpiration    int              `json:"cfg_expiration"`
}

func Init() error {
	adminServerURL := os.Getenv("ADMIN_SERVER_URL")
	if len(adminServerURL) == 0 {
		adminServerURL = "http://admin_server"
	}
	mg = comcfg.NewManager("cfg", adminServerURL)

	if err := mg.Loader.Init(); err != nil {
		return err
	}

	if err := load(); err != nil {
		return err
	}

	path, err := LogDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0600); err != nil {
		return err
	}

	return nil
}

// get returns configurations of jobservice from cache,
// if cache is null, it loads first
func get() (*Configuration, error) {
	cfg := mg.GetFromCache()
	if cfg != nil {
		return cfg.(*Configuration), nil
	}

	if err := load(); err != nil {
		return nil, err
	}

	return mg.GetFromCache().(*Configuration), nil
}

// load loads configurations of jobservice and puts them into cache
func load() error {
	raw, err := mg.Loader.Load()
	if err != nil {
		return err
	}

	cfg := &Configuration{}
	if err = json.Unmarshal(raw, cfg); err != nil {
		return err
	}

	if err = mg.Cache.Put(mg.Key, cfg,
		time.Duration(cfg.CfgExpiration)*time.Second); err != nil {
		return err
	}

	return nil
}

// VerifyRemoteCert returns bool value.
func VerifyRemoteCert() (bool, error) {
	cfg, err := get()
	if err != nil {
		return true, err
	}
	return cfg.VerifyRemoteCert, nil
}

// Database ...
func Database() (*models.Database, error) {
	cfg, err := get()
	if err != nil {
		return nil, err
	}
	return cfg.Database, nil
}

// MaxJobWorkers ...
func MaxJobWorkers() (int, error) {
	cfg, err := get()
	if err != nil {
		return 0, err
	}
	return cfg.MaxJobWorkers, nil
}

// LocalUIURL returns the local ui url, job service will use this URL to call API hosted on ui process
func LocalUIURL() string {
	return "http://ui"
}

// LocalRegURL returns the local registry url, job service will use this URL to pull image from the registry
func LocalRegURL() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.Registry.URL, nil
}

// LogDir returns the absolute path to which the log file will be written
func LogDir() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.JobLogDir, nil
}

// UISecret will return the value of secret cookie for jobsevice to call UI API.
func UISecret() string {
	return os.Getenv("UI_SECRET")
}

// SecretKey will return the secret key for encryption/decryption password in target.
func SecretKey() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.SecretKey, nil
}
