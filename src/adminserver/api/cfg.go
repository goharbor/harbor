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
	"strconv"

	cfg "github.com/vmware/harbor/src/adminserver/systemcfg"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// ListCfgs lists configurations
func ListCfgs(w http.ResponseWriter, r *http.Request) {
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

	log.Info(m)

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

	log.Info(system.Authentication.SelfRegistration)

	if err = cfg.UpdateSystemCfg(system); err != nil {
		log.Errorf("failed to update system configurations: %v", err)
		handleInternalServerError(w)
		return
	}
}

// populate attrs of cfg according to m
func populate(cfg *models.SystemCfg, m map[string]string) error {
	if mode, ok := m[comcfg.AUTH_MODE]; ok {
		cfg.Authentication.Mode = mode
	}
	if value, ok := m[comcfg.SELF_REGISTRATION]; ok {
		cfg.Authentication.SelfRegistration = value == "true"
	}
	if url, ok := m[comcfg.LDAP_URL]; ok {
		cfg.Authentication.LDAP.URL = url
	}
	if dn, ok := m[comcfg.LDAP_SEARCH_DN]; ok {
		cfg.Authentication.LDAP.SearchDN = dn
	}
	if pwd, ok := m[comcfg.LDAP_SEARCH_PWD]; ok {
		cfg.Authentication.LDAP.SearchPwd = pwd
	}
	if dn, ok := m[comcfg.LDAP_BASE_DN]; ok {
		cfg.Authentication.LDAP.BaseDN = dn
	}
	if uid, ok := m[comcfg.LDAP_UID]; ok {
		cfg.Authentication.LDAP.UID = uid
	}
	if filter, ok := m[comcfg.LDAP_FILTER]; ok {
		cfg.Authentication.LDAP.Filter = filter
	}
	if scope, ok := m[comcfg.LDAP_SCOPE]; ok {
		i, err := strconv.Atoi(scope)
		if err != nil {
			return err
		}
		cfg.Authentication.LDAP.Scope = i
	}

	if value, ok := m[comcfg.EMAIL_SERVER]; ok {
		cfg.Email.Host = value
	}
	if value, ok := m[comcfg.EMAIL_SERVER_PORT]; ok {
		cfg.Email.Port = value
	}
	if value, ok := m[comcfg.EMAIL_USERNAME]; ok {
		cfg.Email.Username = value
	}
	if value, ok := m[comcfg.EMAIL_PWD]; ok {
		cfg.Email.Host = value
	}
	if value, ok := m[comcfg.EMAIL_SSL]; ok {
		cfg.Email.Password = value
	}
	if value, ok := m[comcfg.EMAIL_FROM]; ok {
		cfg.Email.From = value
	}
	if value, ok := m[comcfg.EMAIL_IDENTITY]; ok {
		cfg.Email.Identity = value
	}

	if value, ok := m[comcfg.PROJECT_CREATION_RESTRICTION]; ok {
		cfg.ProjectCreationRestriction = value
	}

	if value, ok := m[comcfg.VERIFY_REMOTE_CERT]; ok {
		cfg.VerifyRemoteCert = value == "true"
	}

	if value, ok := m[comcfg.MAX_JOB_WORKERS]; ok {
		if i, err := strconv.Atoi(value); err != nil {
			return err
		} else {
			cfg.MaxJobWorkers = i
		}
	}

	return nil
}
