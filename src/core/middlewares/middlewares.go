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

package middlewares

import (
	"github.com/goharbor/harbor/src/server/middleware/readonly"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/astaxie/beego"
	"github.com/docker/distribution/reference"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/orm"
	"github.com/goharbor/harbor/src/server/middleware/requestid"
	"github.com/goharbor/harbor/src/server/middleware/transaction"
)

var (
	match         = regexp.MustCompile
	numericRegexp = match(`[0-9]+`)

	blobURLRe = match("^/v2/(" + reference.NameRegexp.String() + ")/blobs/" + reference.DigestRegexp.String())

	// fetchBlobAPISkipper skip transaction middleware for fetch blob API
	// because transaction use the ResponseBuffer for the response which will degrade the performance for fetch blob
	fetchBlobAPISkipper = middleware.MethodAndPathSkipper(http.MethodGet, blobURLRe)

	// readonlySkippers skip the post request when harbor sets to readonly.
	readonlySkippers = []middleware.Skipper{
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/c/login")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/c/userExists")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/c/oidc/onboard")),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/adminjob/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/replication/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/replication/task/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/webhook/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/retention/task/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/schedules/"+numericRegexp.String())),
		middleware.MethodAndPathSkipper(http.MethodPost, match("^/service/notifications/jobs/webhook/"+numericRegexp.String())),
	}
)

// legacyAPISkipper skip middleware for legacy APIs
func legacyAPISkipper(r *http.Request) bool {
	path := path.Clean(r.URL.EscapedPath())
	for _, prefix := range []string{"/v2/", "/api/v2.0/"} {
		if strings.HasPrefix(path, prefix) {
			return false
		}
	}

	return true
}

// MiddleWares returns global middlewares
func MiddleWares() []beego.MiddleWare {
	return []beego.MiddleWare{
		requestid.Middleware(),
		readonly.Middleware(readonlySkippers...),
		orm.Middleware(legacyAPISkipper),
		transaction.Middleware(legacyAPISkipper, fetchBlobAPISkipper),
	}
}
