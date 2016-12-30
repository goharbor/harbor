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
)

var mg *comcfg.Manager

type Configuration struct {
	DomainName                 string                 `json:"domain_name"` // Harbor external URL: protocal://host:port
	Authentication             *models.Authentication `json:"authentication"`
	Database                   *models.Database       `json:"database"`
	TokenService               *models.TokenService   `json:"token_service"`
	Registry                   *models.Registry       `json:"registry"`
	Email                      *models.Email          `json:"email"`
	VerifyRemoteCert           bool                   `json:"verify_remote_cert"`
	ProjectCreationRestriction string                 `json:"project_creation_restriction"`
	InitialAdminPwd            string                 `json:"initial_admin_pwd"`
	//TODO remove
	CompressJS      bool   `json:"compress_js"`
	TokenExpiration int    `json:"token_expiration"`
	SecretKey       string `json:"secret_key"`
	CfgExpiration   int    `json:"cfg_expiration`
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

	if err := Load(); err != nil {
		return err
	}

	return nil
}

// Get returns configurations of UI, if cache is null, it loads first
func get() (*Configuration, error) {
	cfg := mg.GetFromCache()
	if cfg != nil {
		return cfg.(*Configuration), nil
	}

	if err := Load(); err != nil {
		return nil, err
	}

	return mg.GetFromCache().(*Configuration), nil
}

// Load loads configurations of UI and puts them into cache
func Load() error {
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

// Upload uploads all system configutations to admin server
func Upload(cfg map[string]string) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return mg.Loader.Upload(b)
}

// GetSystemCfg returns the system configurations
func GetSystemCfg() (*models.SystemCfg, error) {
	raw, err := mg.Loader.Load()
	if err != nil {
		return nil, err
	}

	cfg := &models.SystemCfg{}
	if err = json.Unmarshal(raw, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// AuthMode ...
func AuthMode() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.Authentication.Mode, nil
}

// LDAP returns the setting of ldap server
func LDAP() (*models.LDAP, error) {
	cfg, err := get()
	if err != nil {
		return nil, err
	}
	return cfg.Authentication.LDAP, nil
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration() (int, error) {
	cfg, err := get()
	if err != nil {
		return 0, err
	}
	return cfg.TokenExpiration, nil
}

// DomainName returns the external URL of Harbor: protocal://host:port
func DomainName() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.DomainName, nil
}

// SecretKey returns the secret key to encrypt the password of target
func SecretKey() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.SecretKey, nil
}

// SelfRegistration returns the enablement of self registration
func SelfRegistration() (bool, error) {
	cfg, err := get()
	if err != nil {
		return false, err
	}
	return cfg.Authentication.SelfRegistration, nil
}

// RegistryURL ...
func RegistryURL() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.Registry.URL, nil
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	return "http://jobservice"
}

// InitialAdminPassword returns the initial password for administrator
func InitialAdminPassword() (string, error) {
	cfg, err := get()
	if err != nil {
		return "", err
	}
	return cfg.InitialAdminPwd, nil
}

// TODO
// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject() (bool, error) {
	cfg, err := get()
	if err != nil {
		return true, err
	}
	return cfg.ProjectCreationRestriction == comcfg.PRO_CRT_RESTR_ADM_ONLY, nil
}

// VerifyRemoteCert returns bool value.
func VerifyRemoteCert() (bool, error) {
	cfg, err := get()
	if err != nil {
		return true, err
	}
	return cfg.VerifyRemoteCert, nil
}

func Email() (*models.Email, error) {
	cfg, err := get()
	if err != nil {
		return nil, err
	}
	return cfg.Email, nil
}

func Database() (*models.Database, error) {
	cfg, err := get()
	if err != nil {
		return nil, err
	}
	return cfg.Database, nil
}

// TODO
// UISecret returns the value of UI secret cookie, used for communication between UI and JobService
func UISecret() string {
	return os.Getenv("UI_SECRET")
}
