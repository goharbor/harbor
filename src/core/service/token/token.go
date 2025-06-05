// Copyright Project Harbor Authors
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
	"html/template"
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"github.com/goharbor/harbor/src/lib/log"
)

// Handler handles request on /service/token, which is the auth provider for registry.
type Handler struct {
	web.Controller
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
		log.Error(errMsg)
		h.CustomAbort(http.StatusBadRequest, template.HTMLEscapeString(errMsg))
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
	if err := h.ServeJSON(); err != nil {
		log.Errorf("failed to serve json on /service/token, %v", err)
		h.CustomAbort(http.StatusInternalServerError, "")
	}
}
