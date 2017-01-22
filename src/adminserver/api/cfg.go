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

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	cfg "github.com/vmware/harbor/src/adminserver/systemcfg"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

func isAuthenticated(r *http.Request) (bool, error) {
	secret := os.Getenv("UI_SECRET")
	c, err := r.Cookie("secret")
	if err != nil {
		if err == http.ErrNoCookie {
			return false, nil
		}
		return false, err
	}
	return c != nil && c.Value == secret, nil
}

// ListCfgs lists configurations
func ListCfgs(w http.ResponseWriter, r *http.Request) {
	authenticated, err := isAuthenticated(r)
	if err != nil {
		log.Errorf("failed to check whether the request is authenticated or not: %v", err)
		handleInternalServerError(w)
		return
	}

	if !authenticated {
		handleUnauthorized(w)
		return
	}

	cfg, err := cfg.GetSystemCfg()
	if err != nil {
		log.Errorf("failed to get system configurations: %v", err)
		handleInternalServerError(w)
		return
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Errorf("failed to marshal configurations: %v", err)
		handleInternalServerError(w)
		return
	}
	if _, err = w.Write(b); err != nil {
		log.Errorf("failed to write response: %v", err)
	}
}

// UpdateCfgs updates configurations
func UpdateCfgs(w http.ResponseWriter, r *http.Request) {
	authenticated, err := isAuthenticated(r)
	if err != nil {
		log.Errorf("failed to check whether the request is authenticated or not: %v", err)
		handleInternalServerError(w)
		return
	}

	if !authenticated {
		handleUnauthorized(w)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Errorf("failed to read request body: %v", err)
		handleInternalServerError(w)
		return
	}

	m := &map[string]string{}
	if err = json.Unmarshal(b, m); err != nil {
		handleBadRequestError(w, err.Error())
		return
	}

	system, err := cfg.GetSystemCfg()
	if err != nil {
		handleInternalServerError(w)
		return
	}

	if err := populate(system, *m); err != nil {
		log.Errorf("failed to populate system configurations: %v", err)
		handleInternalServerError(w)
		return
	}

	if err = cfg.UpdateSystemCfg(system); err != nil {
		log.Errorf("failed to update system configurations: %v", err)
		handleInternalServerError(w)
		return
	}
}

// populate attrs of cfg according to m
func populate(cfg *models.SystemCfg, m map[string]string) error {
	if mode, ok := m[comcfg.AUTHMode]; ok {
		cfg.Authentication.Mode = mode
	}
	if value, ok := m[comcfg.SelfRegistration]; ok {
		cfg.Authentication.SelfRegistration = value == "1"
	}
	if url, ok := m[comcfg.LDAPURL]; ok {
		cfg.Authentication.LDAP.URL = url
	}
	if dn, ok := m[comcfg.LDAPSearchDN]; ok {
		cfg.Authentication.LDAP.SearchDN = dn
	}
	if pwd, ok := m[comcfg.LDAPSearchPwd]; ok {
		cfg.Authentication.LDAP.SearchPwd = pwd
	}
	if dn, ok := m[comcfg.LDAPBaseDN]; ok {
		cfg.Authentication.LDAP.BaseDN = dn
	}
	if uid, ok := m[comcfg.LDAPUID]; ok {
		cfg.Authentication.LDAP.UID = uid
	}
	if filter, ok := m[comcfg.LDAPFilter]; ok {
		cfg.Authentication.LDAP.Filter = filter
	}
	if scope, ok := m[comcfg.LDAPScope]; ok {
		i, err := strconv.Atoi(scope)
		if err != nil {
			return err
		}
		cfg.Authentication.LDAP.Scope = i
	}
	if timeout, ok := m[comcfg.LDAPTimeout]; ok {
		i, err := strconv.Atoi(timeout)
		if err != nil {
			return err
		}
		cfg.Authentication.LDAP.Timeout = i
	}

	if value, ok := m[comcfg.EmailHost]; ok {
		cfg.Email.Host = value
	}
	if value, ok := m[comcfg.EmailPort]; ok {
		cfg.Email.Port = value
	}
	if value, ok := m[comcfg.EmailUsername]; ok {
		cfg.Email.Username = value
	}
	if value, ok := m[comcfg.EmailPassword]; ok {
		cfg.Email.Password = value
	}
	if value, ok := m[comcfg.EmailSSL]; ok {
		cfg.Email.SSL = value == "1"
	}
	if value, ok := m[comcfg.EmailFrom]; ok {
		cfg.Email.From = value
	}
	if value, ok := m[comcfg.EmailIdentity]; ok {
		cfg.Email.Identity = value
	}

	if value, ok := m[comcfg.ProjectCreationRestriction]; ok {
		cfg.ProjectCreationRestriction = value
	}

	if value, ok := m[comcfg.VerifyRemoteCert]; ok {
		cfg.VerifyRemoteCert = value == "1"
	}

	return nil
}
