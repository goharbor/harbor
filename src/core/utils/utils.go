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

// Package utils contains methods to support security, cache, and webhook functions.
package utils

import (
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/core/config"
	"net/http"
)

// NewRepositoryClientForUI creates a repository client that can only be used to
// access the internal registry
func NewRepositoryClientForUI(username, repository string) (*registry.Repository, error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}
	return newRepositoryClient(endpoint, username, repository)
}

func newRepositoryClient(endpoint, username, repository string) (*registry.Repository, error) {
	uam := &auth.UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}
	authorizer := auth.DefaultBasicAuthorizer()
	transport := registry.NewTransport(http.DefaultTransport, authorizer, uam)
	client := &http.Client{
		Transport: transport,
	}
	return registry.NewRepository(repository, endpoint, client)
}
