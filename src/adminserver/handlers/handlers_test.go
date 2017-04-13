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

package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/adminserver/auth"
)

type fakeAuthenticator struct {
	authenticated bool
	err           error
}

func (f *fakeAuthenticator) Authenticate(req *http.Request) (bool, error) {
	return f.authenticated, f.err
}

type fakeHandler struct {
	responseCode int
}

func (f *fakeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(f.responseCode)
}

func TestNewAuthHandler(t *testing.T) {
	cases := []struct {
		authenticator auth.Authenticator
		handler       http.Handler
		responseCode  int
	}{

		{nil, nil, http.StatusOK},
		{&fakeAuthenticator{
			authenticated: false,
			err:           nil,
		}, nil, http.StatusUnauthorized},
		{&fakeAuthenticator{
			authenticated: false,
			err:           errors.New("error"),
		}, nil, http.StatusInternalServerError},
		{&fakeAuthenticator{
			authenticated: true,
			err:           nil,
		}, &fakeHandler{http.StatusNotFound}, http.StatusNotFound},
	}

	for _, c := range cases {
		handler := newAuthHandler(c.authenticator, c.handler)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, nil)
		assert.Equal(t, c.responseCode, w.Code, "unexpected response code")
	}
}
