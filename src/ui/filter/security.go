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

package filter

import (
	"net/http"
	"strings"

	"github.com/astaxie/beego/context"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/security"
)

const (
	// HarborSecurityContext is the name of security context passed to handlers
	HarborSecurityContext = "harbor_security_context"
)

// SecurityFilter authenticates the request and passes a security context with it
// which can be used to do some authorization
func SecurityFilter(ctx *context.Context) {
	if ctx == nil {
		return
	}

	req := ctx.Request
	if req == nil {
		return
	}

	if !strings.HasPrefix(req.RequestURI, "/api/") &&
		!strings.HasPrefix(req.RequestURI, "/service/") {
		return
	}

	securityCtx, err := createSecurityContext(req)
	if err != nil {
		log.Warningf("failed to create security context: %v", err)
		return
	}

	ctx.Input.SetData(HarborSecurityContext, securityCtx)
}

func createSecurityContext(req *http.Request) (security.Context, error) {
	return nil, nil
}
