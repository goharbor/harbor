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

// Package config provides methods to get configurations required by code in src/ui
package config

import (
	"strconv"
	"strings"

	commonConfig "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/utils/log"
)

// LDAPSetting wraps the setting of an LDAP server
type LDAPSetting struct {
	URL       string
	BaseDn    string
	SearchDn  string
	SearchPwd string
	UID       string
	Filter    string
	Scope     string
}

type uiParser struct{}

// Parse parses the auth settings url settings and other configuration consumed by code under src/ui
func (up *uiParser) Parse(raw map[string]string, config map[string]interface{}) error {
	mode := raw["AUTH_MODE"]
	if mode == "ldap_auth" {
		setting := LDAPSetting{
			URL:       raw["LDAP_URL"],
			BaseDn:    raw["LDAP_BASE_DN"],
			SearchDn:  raw["LDAP_SEARCH_DN"],
			SearchPwd: raw["LDAP_SEARCH_PWD"],
			UID:       raw["LDAP_UID"],
			Filter:    raw["LDAP_FILTER"],
			Scope:     raw["LDAP_SCOPE"],
		}
		config["ldap"] = setting
	}
	config["auth_mode"] = mode
	var tokenExpiration = 30 //minutes
	if len(raw["TOKEN_EXPIRATION"]) > 0 {
		i, err := strconv.Atoi(raw["TOKEN_EXPIRATION"])
		if err != nil {
			log.Warningf("failed to parse token expiration: %v, using default value %d", err, tokenExpiration)
		} else if i <= 0 {
			log.Warningf("invalid token expiration, using default value: %d minutes", tokenExpiration)
		} else {
			tokenExpiration = i
		}
	}
	config["token_exp"] = tokenExpiration
	config["admin_password"] = raw["HARBOR_ADMIN_PASSWORD"]
	config["ext_reg_url"] = raw["EXT_REG_URL"]
	config["ui_secret"] = raw["UI_SECRET"]
	config["secret_key"] = raw["SECRET_KEY"]
	config["self_registration"] = raw["SELF_REGISTRATION"] != "off"
	config["admin_create_project"] = strings.ToLower(raw["PROJECT_CREATION_RESTRICTION"]) == "adminonly"
	registryURL := raw["REGISTRY_URL"]
	registryURL = strings.TrimRight(registryURL, "/")
	config["internal_registry_url"] = registryURL
	jobserviceURL := raw["JOB_SERVICE_URL"]
	jobserviceURL = strings.TrimRight(jobserviceURL, "/")
	config["internal_jobservice_url"] = jobserviceURL
	return nil
}

var uiConfig *commonConfig.Config

func init() {
	uiKeys := []string{"AUTH_MODE", "LDAP_URL", "LDAP_BASE_DN", "LDAP_SEARCH_DN", "LDAP_SEARCH_PWD", "LDAP_UID", "LDAP_FILTER", "LDAP_SCOPE", "TOKEN_EXPIRATION", "HARBOR_ADMIN_PASSWORD", "EXT_REG_URL", "UI_SECRET", "SECRET_KEY", "SELF_REGISTRATION", "PROJECT_CREATION_RESTRICTION", "REGISTRY_URL", "JOB_SERVICE_URL"}
	uiConfig = &commonConfig.Config{
		Config: make(map[string]interface{}),
		Loader: &commonConfig.EnvConfigLoader{Keys: uiKeys},
		Parser: &uiParser{},
	}
	if err := uiConfig.Load(); err != nil {
		panic(err)
	}
}

// Reload ...
func Reload() error {
	return uiConfig.Load()
}

// AuthMode ...
func AuthMode() string {
	return uiConfig.Config["auth_mode"].(string)
}

// LDAP returns the setting of ldap server
func LDAP() LDAPSetting {
	return uiConfig.Config["ldap"].(LDAPSetting)
}

// TokenExpiration returns the token expiration time (in minute)
func TokenExpiration() int {
	return uiConfig.Config["token_exp"].(int)
}

// ExtRegistryURL returns the registry URL to exposed to external client
func ExtRegistryURL() string {
	return uiConfig.Config["ext_reg_url"].(string)
}

// UISecret returns the value of UI secret cookie, used for communication between UI and JobService
func UISecret() string {
	return uiConfig.Config["ui_secret"].(string)
}

// SecretKey returns the secret key to encrypt the password of target
func SecretKey() string {
	return uiConfig.Config["secret_key"].(string)
}

// SelfRegistration returns the enablement of self registration
func SelfRegistration() bool {
	return uiConfig.Config["self_registration"].(bool)
}

// InternalRegistryURL returns registry URL for internal communication between Harbor containers
func InternalRegistryURL() string {
	return uiConfig.Config["internal_registry_url"].(string)
}

// InternalJobServiceURL returns jobservice URL for internal communication between Harbor containers
func InternalJobServiceURL() string {
	return uiConfig.Config["internal_jobservice_url"].(string)
}

// InitialAdminPassword returns the initial password for administrator
func InitialAdminPassword() string {
	return uiConfig.Config["admin_password"].(string)
}

// OnlyAdminCreateProject returns the flag to restrict that only sys admin can create project
func OnlyAdminCreateProject() bool {
	return uiConfig.Config["admin_create_project"].(bool)
}
