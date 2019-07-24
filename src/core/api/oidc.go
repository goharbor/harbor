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

package api

import (
	"errors"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/oidc"
)

// OIDCAPI handles the requests to /api/system/oidc/xxx
type OIDCAPI struct {
	BaseController
}

// Prepare validates the request initially
func (oa *OIDCAPI) Prepare() {
	oa.BaseController.Prepare()
	if !oa.SecurityCtx.IsAuthenticated() {
		oa.SendUnAuthorizedError(errors.New("unauthorized"))
		return
	}
	if !oa.SecurityCtx.IsSysAdmin() {
		msg := "only system admin has permission to access this API"
		log.Errorf(msg)
		oa.SendForbiddenError(errors.New(msg))
		return
	}
}

// Ping will handles the request to test connection to OIDC endpoint
func (oa *OIDCAPI) Ping() {
	var c oidc.Conn
	if err := oa.DecodeJSONReq(&c); err != nil {
		log.Error("Failed to decode JSON request.")
		oa.SendBadRequestError(err)
		return
	}
	if err := oidc.TestEndpoint(c); err != nil {
		log.Errorf("Failed to verify connection: %+v, err: %v", c, err)
		oa.SendBadRequestError(err)
		return
	}
}
