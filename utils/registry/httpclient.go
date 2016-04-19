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

package registry

import (
	"net/http"

	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry/auth"
)

// NewClient returns a http.Client according to the handlers provided
func NewClient(handlers []auth.Handler) *http.Client {
	transport := NewAuthTransport(http.DefaultTransport, handlers)

	return &http.Client{
		Transport: transport,
	}
}

// NewClientStandardAuthHandlerEmbeded return a http.Client which will authorize the request
// according to the credential provided and send it again when encounters a 401 error
func NewClientStandardAuthHandlerEmbeded(credential auth.Credential) *http.Client {
	handlers := []auth.Handler{}

	tokenHandler := auth.NewStandardTokenHandler(credential)

	handlers = append(handlers, tokenHandler)

	return NewClient(handlers)
}

// NewClientUsernameAuthHandlerEmbeded return a http.Client which will authorize the request
// according to the user's privileges and send it again when encounters a 401 error
func NewClientUsernameAuthHandlerEmbeded(username string) *http.Client {
	handlers := []auth.Handler{}

	tokenHandler := auth.NewUsernameTokenHandler(username)

	handlers = append(handlers, tokenHandler)

	return NewClient(handlers)
}

type authTransport struct {
	transport http.RoundTripper
	handlers  []auth.Handler
}

// NewAuthTransport wraps the AuthHandlers to be http.RounTripper
func NewAuthTransport(transport http.RoundTripper, handlers []auth.Handler) http.RoundTripper {
	return &authTransport{
		transport: transport,
		handlers:  handlers,
	}
}

// RoundTrip ...
func (a *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	originResp, originErr := a.transport.RoundTrip(req)

	if originErr != nil {
		return originResp, originErr
	}

	log.Debugf("%d | %s %s", originResp.StatusCode, req.Method, req.URL)

	if originResp.StatusCode != http.StatusUnauthorized {
		return originResp, nil
	}

	challenges := auth.ParseChallengeFromResponse(originResp)

	reqChanged := false
	for _, challenge := range challenges {

		scheme := challenge.Scheme

		for _, handler := range a.handlers {
			if scheme != handler.Schema() {
				log.Debugf("scheme not match: %s %s, skip", scheme, handler.Schema())
				continue
			}

			if err := handler.AuthorizeRequest(req, challenge.Parameters); err != nil {
				return nil, err
			}
			reqChanged = true
		}
	}

	if !reqChanged {
		log.Warning("no handler match scheme")
		return originResp, nil
	}

	resp, err := a.transport.RoundTrip(req)
	if err == nil {
		log.Debugf("%d | %s %s", resp.StatusCode, req.Method, req.URL)
	}

	return resp, err
}
