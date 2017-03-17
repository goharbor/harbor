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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego/cache"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
)

// const variables
const (
	DBAuth              = "db_auth"
	LDAPAuth            = "ldap_auth"
	ProCrtRestrEveryone = "everyone"
	ProCrtRestrAdmOnly  = "adminonly"
	LDAPScopeBase       = "1"
	LDAPScopeOnelevel   = "2"
	LDAPScopeSubtree    = "3"

	ExtEndpoint                = "ext_endpoint"
	AUTHMode                   = "auth_mode"
	DatabaseType               = "database_type"
	MySQLHost                  = "mysql_host"
	MySQLPort                  = "mysql_port"
	MySQLUsername              = "mysql_username"
	MySQLPassword              = "mysql_password"
	MySQLDatabase              = "mysql_database"
	SQLiteFile                 = "sqlite_file"
	SelfRegistration           = "self_registration"
	LDAPURL                    = "ldap_url"
	LDAPSearchDN               = "ldap_search_dn"
	LDAPSearchPwd              = "ldap_search_password"
	LDAPBaseDN                 = "ldap_base_dn"
	LDAPUID                    = "ldap_uid"
	LDAPFilter                 = "ldap_filter"
	LDAPScope                  = "ldap_scope"
	LDAPTimeout                = "ldap_timeout"
	TokenServiceURL            = "token_service_url"
	RegistryURL                = "registry_url"
	EmailHost                  = "email_host"
	EmailPort                  = "email_port"
	EmailUsername              = "email_username"
	EmailPassword              = "email_password"
	EmailFrom                  = "email_from"
	EmailSSL                   = "email_ssl"
	EmailIdentity              = "email_identity"
	ProjectCreationRestriction = "project_creation_restriction"
	VerifyRemoteCert           = "verify_remote_cert"
	MaxJobWorkers              = "max_job_workers"
	TokenExpiration            = "token_expiration"
	CfgExpiration              = "cfg_expiration"
	JobLogDir                  = "job_log_dir"
	UseCompressedJS            = "use_compressed_js"
	AdminInitialPassword       = "admin_initial_password"
	AdmiralEndpoint            = "admiral_url"
	WithNotary                 = "with_notary"
)

// Manager manages configurations
type Manager struct {
	Loader *Loader
	Parser *Parser
	Cache  bool
	cache  cache.Cache
	key    string
}

// NewManager returns an instance of Manager
// url: the url from which loader loads configurations
func NewManager(url, secret string, enableCache bool) *Manager {
	m := &Manager{
		Loader: NewLoader(url, secret),
		Parser: &Parser{},
	}

	if enableCache {
		m.Cache = true
		m.cache = cache.NewMemoryCache()
		m.key = "cfg"
	}

	return m
}

// Init loader
func (m *Manager) Init() error {
	return m.Loader.Init()
}

// Load configurations, if cache is enabled, cache the configurations
func (m *Manager) Load() (map[string]interface{}, error) {
	b, err := m.Loader.Load()
	if err != nil {
		return nil, err
	}

	c, err := m.Parser.Parse(b)
	if err != nil {
		return nil, err
	}

	if m.Cache {
		expi, err := getCfgExpiration(c)
		if err != nil {
			return nil, err
		}
		if err = m.cache.Put(m.key, c,
			time.Duration(expi)*time.Second); err != nil {
			return nil, err
		}
	}

	return c, nil
}

// Reset configurations
func (m *Manager) Reset() error {
	return m.Loader.Reset()
}

func getCfgExpiration(m map[string]interface{}) (int, error) {
	if m == nil {
		return 0, fmt.Errorf("can not get cfg expiration as configurations are null")
	}

	expi, ok := m[CfgExpiration]
	if !ok {
		return 0, fmt.Errorf("cfg expiration is not set")
	}

	return int(expi.(float64)), nil
}

// Get : if cache is enabled, read configurations from cache,
// if cache is null or cache is disabled it loads configurations directly
func (m *Manager) Get() (map[string]interface{}, error) {
	if m.Cache {
		c := m.cache.Get(m.key)
		if c != nil {
			return c.(map[string]interface{}), nil
		}
	}
	return m.Load()
}

// Upload configurations
func (m *Manager) Upload(b []byte) error {
	return m.Loader.Upload(b)
}

// Loader loads and uploads configurations
type Loader struct {
	url    string
	secret string
	client *http.Client
}

// NewLoader ...
func NewLoader(url, secret string) *Loader {
	return &Loader{
		url:    url,
		secret: secret,
		client: &http.Client{},
	}
}

// Init waits remote server to be ready by testing connections with it
func (l *Loader) Init() error {
	addr := l.url
	if strings.Contains(addr, "://") {
		addr = strings.Split(addr, "://")[1]
	}

	if !strings.Contains(addr, ":") {
		addr = addr + ":80"
	}

	return utils.TestTCPConn(addr, 60, 2)
}

// Load configurations from remote server
func (l *Loader) Load() ([]byte, error) {
	log.Debug("loading configurations...")
	req, err := http.NewRequest("GET", l.url+"/api/configurations", nil)
	if err != nil {
		return nil, err
	}

	req.AddCookie(&http.Cookie{
		Name:  "secret",
		Value: l.secret,
	})

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Debug("configurations load completed")
	return b, nil
}

// Upload configurations to remote server
func (l *Loader) Upload(b []byte) error {
	req, err := http.NewRequest("PUT", l.url+"/api/configurations", bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{
		Name:  "secret",
		Value: l.secret,
	})

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status code: %d", resp.StatusCode)
	}

	log.Debug("configurations uploaded")

	return nil
}

// Reset sends configurations resetting command to
// remote server
func (l *Loader) Reset() error {
	req, err := http.NewRequest("POST", l.url+"/api/configurations/reset", nil)
	if err != nil {
		return err
	}

	req.AddCookie(&http.Cookie{
		Name:  "secret",
		Value: l.secret,
	})

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected http status code: %d", resp.StatusCode)
	}

	log.Debug("configurations resetted")

	return nil
}

// Parser parses configurations
type Parser struct {
}

// Parse parses []byte to a map configuration
func (p *Parser) Parse(b []byte) (map[string]interface{}, error) {
	c := map[string]interface{}{}
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	return c, nil
}
