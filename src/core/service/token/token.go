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

package token

import (
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// Handler handles request on /service/token, which is the auth provider for registry.
type Handler struct {
	beego.Controller
}

// Get handles GET request, it checks the http header for user credentials
// and parse service and scope based on docker registry v2 standard,
// checks the permission against local DB and generates jwt token.
func (h *Handler) Get() {
	request := h.Ctx.Request
	log.Debugf("URL for token request: %s", request.URL.String())
	service := h.GetString("service")
	tokenCreator, ok := creatorMap[service]
	if !ok {
		errMsg := fmt.Sprintf("Unable to handle service: %s", service)
		log.Errorf(errMsg)
		h.CustomAbort(http.StatusBadRequest, errMsg)
	}
	token, err := tokenCreator.Create(request)
	if err != nil {
		if _, ok := err.(*unauthorizedError); ok {
			h.CustomAbort(http.StatusUnauthorized, "")
		}
		log.Errorf("Unexpected error when creating the token, error: %v", err)
		h.CustomAbort(http.StatusInternalServerError, "")
	}
	h.Data["json"] = token
	h.ServeJSON()

}
