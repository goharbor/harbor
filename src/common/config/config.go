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

// Package config provide methods to get the configurations reqruied by code in src/common
package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/astaxie/beego/cache"
	"github.com/vmware/harbor/src/common/utils"
)

const (
	//auth mode
	DB_AUTH   = "db_auth"
	LDAP_AUTH = "ldap_auth"

	//project_creation_restriction
	PRO_CRT_RESTR_EVERYONE = "everyone"
	PRO_CRT_RESTR_ADM_ONLY = "adminonly"

	LDAP_SCOPE_BASE     = "1"
	LDAP_SCOPE_ONELEVEL = "2"
	LDAP_SCOPE_SUBTREE  = "3"

	AUTH_MODE                    = "auth_mode"
	SELF_REGISTRATION            = "self_registration"
	LDAP_URL                     = "ldap_url"
	LDAP_SEARCH_DN               = "ldap_search_dn"
	LDAP_SEARCH_PWD              = "ldap_search_pwd"
	LDAP_BASE_DN                 = "ldap_base_dn"
	LDAP_UID                     = "ldap_uid"
	LDAP_FILTER                  = "ldap_filter"
	LDAP_SCOPE                   = "ldap_scope"
	EMAIL_SERVER                 = "email_server"
	EMAIL_SERVER_PORT            = "email_server_port"
	EMAIL_USERNAME               = "email_server_username"
	EMAIL_PWD                    = "email_server_pwd"
	EMAIL_FROM                   = "email_from"
	EMAIL_SSL                    = "email_ssl"
	EMAIL_IDENTITY               = "email_identity"
	PROJECT_CREATION_RESTRICTION = "project_creation_restriction"
	VERIFY_REMOTE_CERT           = "verify_remote_cert"
	MAX_JOB_WORKERS              = "max_job_workers"
	CFG_EXPIRATION               = "cfg_expiration"
)

type Manager struct {
	Key    string
	Cache  cache.Cache
	Loader *Loader
}

func NewManager(key, url string) *Manager {
	return &Manager{
		Key:    key,
		Cache:  cache.NewMemoryCache(),
		Loader: NewLoader(url),
	}
}

func (m *Manager) GetFromCache() interface{} {
	value := m.Cache.Get(m.Key)
	if value != nil {
		return value
	}
	return nil
}

type Loader struct {
	url    string
	client *http.Client
}

func NewLoader(url string) *Loader {
	return &Loader{
		url:    url,
		client: &http.Client{},
	}
}

func (l *Loader) Init() error {
	addr := l.url
	if strings.Contains(addr, "://") {
		addr = strings.Split(addr, "://")[1]
	}
	return utils.TestTCPConn(addr, 60, 2)
}

func (l *Loader) Load() ([]byte, error) {
	resp, err := l.client.Get(l.url + "/api/configurations")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (l *Loader) Upload(b []byte) error {
	req, err := http.NewRequest("PUT", l.url+"/api/configurations", bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status code: %d", resp.StatusCode)
	}

	return nil
}
