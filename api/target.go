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
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils"
	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry"
	"github.com/vmware/harbor/utils/registry/auth"
	registry_error "github.com/vmware/harbor/utils/registry/error"
)

// TargetAPI handles request to /api/targets/ping /api/targets/{}
type TargetAPI struct {
	BaseAPI
	secretKey string
}

// Prepare validates the user
func (t *TargetAPI) Prepare() {
	//TODO:move to config
	t.secretKey = os.Getenv("SECRET_KEY")

	userID := t.ValidateUser()
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("error occurred in IsAdminRole: %v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if !isSysAdmin {
		t.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
}

// Ping validates whether the target is reachable and whether the credential is valid
func (t *TargetAPI) Ping() {
	var endpoint, username, password string

	idStr := t.GetString("id")
	if len(idStr) != 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			t.CustomAbort(http.StatusBadRequest, fmt.Sprintf("id %s is invalid", idStr))
		}

		target, err := dao.GetRepTarget(id)
		if err != nil {
			log.Errorf("failed to get target %d: %v", id, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if target == nil {
			t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
		}

		endpoint = target.URL
		username = target.Username
		password = target.Password

		if len(password) != 0 {
			password, err = utils.ReversibleDecrypt(password, t.secretKey)
			if err != nil {
				log.Errorf("failed to decrypt password: %v", err)
				t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}
		}
	} else {
		endpoint = t.GetString("endpoint")
		if len(endpoint) == 0 {
			t.CustomAbort(http.StatusBadRequest, "id or endpoint is needed")
		}

		username = t.GetString("username")
		password = t.GetString("password")
	}

	registry, err := newRegistryClient(endpoint, getIsInsecure(), username, password,
		"", "", "")
	if err != nil {
		// timeout, dns resolve error, connection refused, etc.
		if urlErr, ok := err.(*url.Error); ok {
			if netErr, ok := urlErr.Err.(net.Error); ok {
				t.CustomAbort(http.StatusBadRequest, netErr.Error())
			}

			t.CustomAbort(http.StatusBadRequest, urlErr.Error())
		}

		log.Errorf("failed to create registry client: %#v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if err = registry.Ping(); err != nil {
		if regErr, ok := err.(*registry_error.Error); ok {
			t.CustomAbort(regErr.StatusCode, regErr.Detail)
		}

		log.Errorf("failed to ping registry %s: %v", registry.Endpoint.String(), err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Get ...
func (t *TargetAPI) Get() {
	id := t.GetIDFromURL()

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	// The reason why the password is returned is that when user just wants to
	// modify other fields of target he does not need to input the password again.
	// The security issue can be fixed by enable https.
	if len(target.Password) != 0 {
		pwd, err := utils.ReversibleDecrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to decrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		target.Password = pwd
	}

	t.Data["json"] = target
	t.ServeJSON()
}

// List ...
func (t *TargetAPI) List() {
	name := t.GetString("name")
	targets, err := dao.FilterRepTargets(name)
	if err != nil {
		log.Errorf("failed to filter targets %s: %v", name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	for _, target := range targets {
		if len(target.Password) == 0 {
			continue
		}

		str, err := utils.ReversibleDecrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to decrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		target.Password = str
	}

	t.Data["json"] = targets
	t.ServeJSON()
	return
}

// Post ...
func (t *TargetAPI) Post() {
	target := &models.RepTarget{}
	t.DecodeJSONReqAndValidate(target)

	ta, err := dao.GetRepTargetByName(target.Name)
	if err != nil {
		log.Errorf("failed to get target %s: %v", target.Name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if ta != nil {
		t.CustomAbort(http.StatusConflict, "name is already used")
	}

	ta, err = dao.GetRepTargetByEndpoint(target.URL)
	if err != nil {
		log.Errorf("failed to get target [ %s ]: %v", target.URL, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if ta != nil {
		t.CustomAbort(http.StatusConflict, fmt.Sprintf("the target whose endpoint is %s already exists", target.URL))
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleEncrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	id, err := dao.AddRepTarget(*target)
	if err != nil {
		log.Errorf("failed to add target: %v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	t.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put ...
func (t *TargetAPI) Put() {
	id := t.GetIDFromURL()

	originalTarget, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if originalTarget == nil {
		t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	hasEnabledPolicy := false
	for _, policy := range policies {
		if policy.Enabled == 1 {
			hasEnabledPolicy = true
			break
		}
	}

	if hasEnabledPolicy {
		t.CustomAbort(http.StatusBadRequest, "the target is associated with policy which is enabled")
	}

	target := &models.RepTarget{}
	t.DecodeJSONReqAndValidate(target)

	if target.Name != originalTarget.Name {
		ta, err := dao.GetRepTargetByName(target.Name)
		if err != nil {
			log.Errorf("failed to get target %s: %v", target.Name, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if ta != nil {
			t.CustomAbort(http.StatusConflict, "name is already used")
		}
	}

	if target.URL != originalTarget.URL {
		ta, err := dao.GetRepTargetByEndpoint(target.URL)
		if err != nil {
			log.Errorf("failed to get target [ %s ]: %v", target.URL, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if ta != nil {
			t.CustomAbort(http.StatusConflict, fmt.Sprintf("the target whose endpoint is %s already exists", target.URL))
		}
	}

	target.ID = id

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleEncrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	if err := dao.UpdateRepTarget(*target); err != nil {
		log.Errorf("failed to update target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Delete ...
func (t *TargetAPI) Delete() {
	id := t.GetIDFromURL()

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if len(policies) > 0 {
		t.CustomAbort(http.StatusBadRequest, "the target is used by policies, can not be deleted")
	}

	if err = dao.DeleteRepTarget(id); err != nil {
		log.Errorf("failed to delete target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func newRegistryClient(endpoint string, insecure bool, username, password, scopeType, scopeName string,
	scopeActions ...string) (*registry.Registry, error) {
	credential := auth.NewBasicAuthCredential(username, password)
	authorizer := auth.NewStandardTokenAuthorizer(credential, insecure, scopeType, scopeName, scopeActions...)

	store, err := auth.NewAuthorizerStore(endpoint, insecure, authorizer)
	if err != nil {
		return nil, err
	}

	client, err := registry.NewRegistryWithModifiers(endpoint, insecure, store)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// ListPolicies ...
func (t *TargetAPI) ListPolicies() {
	id := t.GetIDFromURL()

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		t.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	t.Data["json"] = policies
	t.ServeJSON()
}
