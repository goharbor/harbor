// Copyright 2018 Project Harbor Authors
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
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/core/config"
)

// TargetAPI handles request to /api/targets/ping /api/targets/{}
type TargetAPI struct {
	BaseController
	secretKey string
}

// Prepare validates the user
func (t *TargetAPI) Prepare() {
	t.BaseController.Prepare()
	if !t.SecurityCtx.IsAuthenticated() {
		t.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	if !t.SecurityCtx.IsSysAdmin() {
		t.SendForbiddenError(errors.New(t.SecurityCtx.GetUsername()))
		return
	}

	var err error
	t.secretKey, err = config.SecretKey()
	if err != nil {
		log.Errorf("failed to get secret key: %v", err)
		t.SendInternalServerError(errors.New(""))
		return
	}
}

func (t *TargetAPI) ping(endpoint, username, password string, insecure bool) {
	registry, err := newRegistryClient(endpoint, insecure, username, password)
	if err == nil {
		err = registry.Ping()
	}

	if err != nil {
		log.Errorf("failed to ping target: %v", err)
		// do not return any detail information of the error, or may cause SSRF security issue #3755
		t.SendConflictError(errors.New("failed to ping target"))
		return
	}
}

// Ping validates whether the target is reachable and whether the credential is valid
func (t *TargetAPI) Ping() {
	req := struct {
		ID       *int64  `json:"id"`
		Endpoint *string `json:"endpoint"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Insecure *bool   `json:"insecure"`
	}{}
	if err := t.DecodeJSONReq(&req); err != nil {
		t.SendBadRequestError(err)
		return
	}

	target := &models.RepTarget{}
	if req.ID != nil {
		var err error
		target, err = dao.GetRepTarget(*req.ID)
		if err != nil {
			t.SendInternalServerError(fmt.Errorf("failed to get target %d: %v", *req.ID, err))
			return
		}
		if target == nil {
			t.SendNotFoundError(fmt.Errorf("target %d not found", *req.ID))
			return
		}
		if len(target.Password) != 0 {
			target.Password, err = utils.ReversibleDecrypt(target.Password, t.secretKey)
			if err != nil {
				t.SendInternalServerError(fmt.Errorf("failed to decrypt password: %v", err))
				return
			}
		}
	}

	if req.Endpoint != nil {
		url, err := utils.ParseEndpoint(*req.Endpoint)
		if err != nil {
			t.SendBadRequestError(err)
			return
		}

		// Prevent SSRF security issue #3755
		target.URL = url.Scheme + "://" + url.Host + url.Path
	}
	if req.Username != nil {
		target.Username = *req.Username
	}
	if req.Password != nil {
		target.Password = *req.Password
	}
	if req.Insecure != nil {
		target.Insecure = *req.Insecure
	}

	t.ping(target.URL, target.Username, target.Password, target.Insecure)
}

// Get ...
func (t *TargetAPI) Get() {
	id, err := t.GetIDFromURL()
	if err != nil {
		t.SendBadRequestError(err)
		return
	}

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	if target == nil {
		t.SendNotFoundError(fmt.Errorf("target %d not found", id))
		return
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
		t.SendInternalServerError(errors.New(""))
		return
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
	isValid, err := t.DecodeJSONReqAndValidate(target)
	if !isValid {
		t.SendBadRequestError(err)
		return
	}

	ta, err := dao.GetRepTargetByName(target.Name)
	if err != nil {
		log.Errorf("failed to get target %s: %v", target.Name, err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	if ta != nil {
		t.SendConflictError(errors.New("name is already used"))
		return
	}

	ta, err = dao.GetRepTargetByEndpoint(target.URL)
	if err != nil {
		log.Errorf("failed to get target [ %s ]: %v", target.URL, err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	if ta != nil {
		t.SendConflictError(fmt.Errorf("the target whose endpoint is %s already exists", target.URL))
		return
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleEncrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			t.SendInternalServerError(errors.New(""))
			return
		}
	}

	id, err := dao.AddRepTarget(*target)
	if err != nil {
		log.Errorf("failed to add target: %v", err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	t.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put ...
func (t *TargetAPI) Put() {
	id, err := t.GetIDFromURL()
	if err != nil {
		t.SendBadRequestError(err)
		return
	}

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	if target == nil {
		t.SendNotFoundError(fmt.Errorf("target %d not found", id))
		return
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleDecrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to decrypt password: %v", err)
			t.SendInternalServerError(errors.New(""))
			return
		}
	}

	req := struct {
		Name     *string `json:"name"`
		Endpoint *string `json:"endpoint"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Insecure *bool   `json:"insecure"`
	}{}
	if err := t.DecodeJSONReq(&req); err != nil {
		t.SendBadRequestError(err)
		return
	}

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
	if req.Insecure != nil {
		target.Insecure = *req.Insecure
	}

	t.Validate(target)

	if target.Name != originalName {
		ta, err := dao.GetRepTargetByName(target.Name)
		if err != nil {
			log.Errorf("failed to get target %s: %v", target.Name, err)
			t.SendInternalServerError(errors.New(""))
			return
		}

		if ta != nil {
			t.SendConflictError(errors.New("name is already used"))
			return
		}
	}

	if target.URL != originalURL {
		ta, err := dao.GetRepTargetByEndpoint(target.URL)
		if err != nil {
			log.Errorf("failed to get target [ %s ]: %v", target.URL, err)
			t.SendInternalServerError(errors.New(""))
			return
		}

		if ta != nil {
			t.SendConflictError(fmt.Errorf("the target whose endpoint is %s already exists", target.URL))
			return
		}
	}

	if len(target.Password) != 0 {
		target.Password, err = utils.ReversibleEncrypt(target.Password, t.secretKey)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			t.SendInternalServerError(errors.New(""))
			return
		}
	}

	if err := dao.UpdateRepTarget(*target); err != nil {
		log.Errorf("failed to update target %d: %v", id, err)
		t.SendInternalServerError(errors.New(""))
		return
	}
}

// Delete ...
func (t *TargetAPI) Delete() {
	id, err := t.GetIDFromURL()
	if err != nil {
		t.SendBadRequestError(err)
		return
	}

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	if target == nil {
		t.SendNotFoundError(fmt.Errorf("target %d not found", id))
		return
	}

	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according target %d: %v", id, err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	if len(policies) > 0 {
		log.Error("the target is used by policies, can not be deleted")
		t.SendPreconditionFailedError(errors.New("the target is used by policies, can not be deleted"))
		return
	}

	if err = dao.DeleteRepTarget(id); err != nil {
		log.Errorf("failed to delete target %d: %v", id, err)
		t.SendInternalServerError(errors.New(""))
		return
	}
}

func newRegistryClient(endpoint string, insecure bool, username, password string) (*registry.Registry, error) {
	transport := registry.GetHTTPTransport(insecure)
	credential := auth.NewBasicAuthCredential(username, password)
	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential)
	return registry.NewRegistry(endpoint, &http.Client{
		Transport: registry.NewTransport(transport, authorizer),
	})
}

// ListPolicies ...
func (t *TargetAPI) ListPolicies() {
	id, err := t.GetIDFromURL()
	if err != nil {
		t.SendBadRequestError(err)
		return
	}

	target, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get target %d: %v", id, err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	if target == nil {
		t.SendNotFoundError(fmt.Errorf("target %d not found", id))
		return
	}

	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("failed to get policies according target %d: %v", id, err)
		t.SendInternalServerError(errors.New(""))
		return
	}

	t.Data["json"] = policies
	t.ServeJSON()
}
