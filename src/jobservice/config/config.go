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
)

var mg *comcfg.Manager

// Configuration of Jobservice
type Configuration struct {
	Database         *models.Database `json:"database"`
	Registry         *models.Registry `json:"registry"`
	VerifyRemoteCert bool             `json:"verify_remote_cert"`
	MaxJobWorkers    int              `json:"max_job_workers"`
	JobLogDir        string           `json:"job_log_dir"`
	SecretKey        string           `json:"secret_key"`
	CfgExpiration    int              `json:"cfg_expiration"`
}

type parser struct {
}

func (p *parser) Parse(b []byte) (interface{}, error) {
	c := &Configuration{}
	if err := json.Unmarshal(b, c); err != nil {
		return nil, err
	}
	return c, nil
}

// Init configurations
func Init() error {
	adminServerURL := os.Getenv("ADMIN_SERVER_URL")
	if len(adminServerURL) == 0 {
		adminServerURL = "http://adminserver"
	}
	mg = comcfg.NewManager(adminServerURL, UISecret(), &parser{}, true)

	if err := mg.Init(); err != nil {
		return err
	}

	if _, err := mg.Load(); err != nil {
		return err
	}

	return nil
}

func get() (*Configuration, error) {
	c, err := mg.Get()
	if err != nil {
		return nil, err
	}
	return c.(*Configuration), nil
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

// SecretKey will return the secret key for encryption/decryption password in target.
func SecretKey() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.SecretKey, nil
}

// UISecret returns the value of UI secret cookie, used for communication between UI and JobService
// TODO
func UISecret() string {
	return os.Getenv("UI_SECRET")
}
