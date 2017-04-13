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
	"net/http"
	"os"

	gorilla_handlers "github.com/gorilla/handlers"
	"github.com/vmware/harbor/src/adminserver/auth"
	"github.com/vmware/harbor/src/common/utils/log"
)

// NewHandler returns a gorilla router which is wrapped by  authenticate handler
// and logging handler
func NewHandler() http.Handler {
	h := newRouter()
	secrets := map[string]string{
		"uiSecret":         os.Getenv("UI_SECRET"),
		"jobserviceSecret": os.Getenv("JOBSERVICE_SECRET"),
	}
	h = newAuthHandler(auth.NewSecretAuthenticator(secrets), h)
	h = gorilla_handlers.LoggingHandler(os.Stdout, h)
	return h
}

type authHandler struct {
	authenticator auth.Authenticator
	handler       http.Handler
}

func newAuthHandler(authenticator auth.Authenticator, handler http.Handler) http.Handler {
	return &authHandler{
		authenticator: authenticator,
		handler:       handler,
	}
}

func (a *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a.authenticator == nil {
		if a.handler != nil {
			a.handler.ServeHTTP(w, r)
		}
		return
	}

	valid, err := a.authenticator.Authenticate(r)
	if err != nil {
		log.Errorf("failed to authenticate request: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError)
		return
	}

	if !valid {
		http.Error(w, http.StatusText(http.StatusUnauthorized),
			http.StatusUnauthorized)
		return
	}

	if a.handler != nil {
		a.handler.ServeHTTP(w, r)
	}
	return
}
