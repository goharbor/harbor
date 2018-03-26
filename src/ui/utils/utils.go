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

// Package utils contains methods to support security, cache, and webhook functions.
package utils

import (
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/ui/config"
	"github.com/vmware/harbor/src/ui/service/token"

	"io"
	"net/http"
)

// RequestAsUI is a shortcut to make a request attach UI secret and send the request.
// Do not use this when you want to handle the response
func RequestAsUI(method, url string, body io.Reader, h ResponseHandler) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	AddUISecret(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	return h.Handle(resp)
}

//AddUISecret add secret cookie to a request
func AddUISecret(req *http.Request) {
	if req != nil {
		req.AddCookie(&http.Cookie{
			Name:  models.UISecretCookie,
			Value: config.UISecret(),
		})
	}
}

// NewRepositoryClientForUI creates a repository client that can only be used to
// access the internal registry
func NewRepositoryClientForUI(username, repository string) (*registry.Repository, error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}

	authorizer := auth.NewRawTokenAuthorizer(username, token.Registry)
	transport := registry.NewTransport(http.DefaultTransport, authorizer)
	client := &http.Client{
		Transport: transport,
	}
	return registry.NewRepository(repository, endpoint, client)
}
