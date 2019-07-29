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
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares"

	"errors"
	"github.com/goharbor/harbor/src/core/middlewares/registryproxy"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/core/service/token"
	"net/http/httptest"
)

// NewRepositoryClientForUI creates a repository client that can only be used to
// access the internal registry
func NewRepositoryClientForUI(username, repository string) (*registry.Repository, error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}

	uam := &auth.UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}
	authorizer := auth.NewRawTokenAuthorizer(username, token.Registry)
	transport := registry.NewTransport(http.DefaultTransport, authorizer, uam)
	client := &http.Client{
		Transport: transport,
	}
	return registry.NewRepository(repository, endpoint, client)
}

// NewRepositoryClientForUIWithMiddleware creates a repository client that can only be used to
// access the internal registry with quota middle
func NewRepositoryClientForUIWithMiddleware(username, repository string) (*registry.Repository, error) {
	endpoint, err := config.RegistryURL()
	if err != nil {
		return nil, err
	}

	uam := &auth.UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}
	authorizer := auth.NewRawTokenAuthorizer(username, token.Registry)
	transport := NewTransportWithMiddleware(authorizer, uam)
	client := &http.Client{
		Transport: transport,
	}
	return registry.NewRepository(repository, endpoint, client)
}

// TransportWithMiddleware holds information about base transport and modifiers
type TransportWithMiddleware struct {
	modifiers []modifier.Modifier
}

// NewTransportWithMiddleware ...
func NewTransportWithMiddleware(modifiers ...modifier.Modifier) *TransportWithMiddleware {
	return &TransportWithMiddleware{
		modifiers: modifiers,
	}
}

// RoundTrip ...
func (t *TransportWithMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	for _, modifier := range t.modifiers {
		if err := modifier.Modify(req); err != nil {
			return nil, err
		}
	}
	ph := registryproxy.New()
	if ph == nil {
		return nil, errors.New("get nil when to create proxy")
	}
	rw := httptest.NewRecorder()
	customResW := util.NewCustomResponseWriter(rw)
	handlerChain := middlewares.New(middlewares.MiddlewaresInternal).Create()
	head := handlerChain.Then(ph)
	head.ServeHTTP(customResW, req)

	log.Infof("%d | %s %s", rw.Result().StatusCode, req.Method, req.URL.String())
	return rw.Result(), nil
}

// WaitForManifestReady implements exponential sleeep to wait until manifest is ready in registry.
// This is a workaround for https://github.com/docker/distribution/issues/2625
func WaitForManifestReady(repository string, tag string, maxRetry int) bool {
	// The initial wait interval, hard-coded to 50ms
	interval := 50 * time.Millisecond
	repoClient, err := NewRepositoryClientForUI("harbor-core", repository)
	if err != nil {
		log.Errorf("Failed to create repo client.")
		return false
	}
	for i := 0; i < maxRetry; i++ {
		_, exist, err := repoClient.ManifestExist(tag)
		if err != nil {
			log.Errorf("Unexpected error when checking manifest existence, image:  %s:%s, error: %v", repository, tag, err)
			continue
		}
		if exist {
			return true
		}
		log.Warningf("manifest for image %s:%s is not ready, retry after %v", repository, tag, interval)
		time.Sleep(interval)
		interval = interval * 2
	}
	return false
}
