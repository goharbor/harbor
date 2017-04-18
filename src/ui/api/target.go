// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"

	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	registry_error "github.com/vmware/harbor/src/common/utils/registry/error"
	"github.com/vmware/harbor/src/ui/config"
)

// TargetAPI handles request to /api/targets/ping /api/targets/{}
type TargetAPI struct {
	api.BaseAPI
	secretKey string
}

// Prepare validates the user
func (t *TargetAPI) Prepare() {
	var err error
	t.secretKey, err = config.SecretKey()
	if err != nil {
		log.Errorf("failed to get secret key: %v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

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

func (t *TargetAPI) ping(endpoint, username, password string) {
	verify, err := config.VerifyRemoteCert()
	if err != nil {
		log.Errorf("failed to check whether insecure or not: %v", err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	registry, err := newRegistryClient(endpoint, !verify, username, password,
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

// PingByID ping target by ID
func (t *TargetAPI) PingByID() {
	id := t.GetIDFromURL()

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	if target == nil {
		t.CustomAbort(http.StatusNotFound, fmt.Sprintf("target %d not found", id))
	}

	endpoint := target.URL
	username := target.Username
	password := target.Password
	if len(password) != 0 {
		password, err = utils.ReversibleDecrypt(password, t.secretKey)
		if err != nil {
			log.Errorf("failed to decrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}
	t.ping(endpoint, username, password)
}

// Ping validates whether the target is reachable and whether the credential is valid
func (t *TargetAPI) Ping() {
	req := struct {
		Endpoint string `json:"endpoint"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{}
	t.DecodeJSONReq(&req)

	if len(req.Endpoint) == 0 {
		t.CustomAbort(http.StatusBadRequest, "endpoint is required")
	}

	t.ping(req.Endpoint, req.Username, req.Password)
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

	target.Password = ""

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
		target.Password = ""
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
	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleDecrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to decrypt password: %v", err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
	}

	req := struct {
		Name     *string `json:"name"`
		Endpoint *string `json:"endpoint"`
		Username *string `json:"username"`
		Password *string `json:"password"`
	}{}
	t.DecodeJSONReq(&req)

	originalName := target.Name
	originalURL := target.URL

	if req.Name != nil {
		target.Name = *req.Name
	}
	if req.Endpoint != nil {
		target.URL = *req.Endpoint
	}
	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Password != nil {
		target.Password = *req.Password
	}

	t.Validate(target)

	if target.Name != originalName {
		ta, err := dao.GetRepTargetByName(target.Name)
		if err != nil {
			log.Errorf("failed to get target %s: %v", target.Name, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if ta != nil {
			t.CustomAbort(http.StatusConflict, "name is already used")
		}
	}

	if target.URL != originalURL {
		ta, err := dao.GetRepTargetByEndpoint(target.URL)
		if err != nil {
			log.Errorf("failed to get target [ %s ]: %v", target.URL, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if ta != nil {
			t.CustomAbort(http.StatusConflict, fmt.Sprintf("the target whose endpoint is %s already exists", target.URL))
		}
	}

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
		t.CustomAbort(http.StatusPreconditionFailed, "the target is used by policies, can not be deleted")
	}

	if err = dao.DeleteRepTarget(id); err != nil {
		log.Errorf("failed to delete target %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func newRegistryClient(endpoint string, insecure bool, username, password, scopeType, scopeName string,
	scopeActions ...string) (*registry.Registry, error) {
	credential := auth.NewBasicAuthCredential(username, password)

	authorizer := auth.NewStandardTokenAuthorizer(credential, insecure,
		"", scopeType, scopeName, scopeActions...)

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
