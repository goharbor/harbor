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

package server

import (
	"net/http"

	"github.com/beego/beego/v2/server/web"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/controllers"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/server/handler"
	"github.com/goharbor/harbor/src/server/router"
)

func ignoreNotification(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func registerRoutes() {
	// API version
	router.NewRoute().Method(http.MethodGet).Path("/api/version").HandlerFunc(GetAPIVersion)

	// Controller API:
	web.Router("/c/login", &controllers.CommonController{}, "post:Login")
	web.Router("/c/log_out", &controllers.CommonController{}, "get:LogOut")
	web.Router("/c/userExists", &controllers.CommonController{}, "post:UserExists")
	web.Router(common.OIDCLoginPath, &controllers.OIDCController{}, "get:RedirectLogin")
	web.Router("/c/oidc/onboard", &controllers.OIDCController{}, "post:Onboard")
	web.Router(common.OIDCCallbackPath, &controllers.OIDCController{}, "get:Callback")
	web.Router(common.AuthProxyRedirectPath, &controllers.AuthProxyController{}, "get:HandleRedirect")

	web.Router("/api/internal/renameadmin", &api.InternalAPI{}, "post:RenameAdmin")
	web.Router("/api/internal/syncquota", &api.InternalAPI{}, "post:SyncQuota")

	router.NewRoute().Method(http.MethodPost).Path("/service/notifications/jobs/adminjob/:id([0-9]+)").Handler(handler.NewJobStatusHandler())         // legacy job status hook endpoint for adminjob
	router.NewRoute().Method(http.MethodPost).Path("/service/notifications/jobs/scan/:uuid").HandlerFunc(ignoreNotification)                          // ignore legacy scan job notifaction
	router.NewRoute().Method(http.MethodPost).Path("/service/notifications/schedules/:id([0-9]+)").Handler(handler.NewJobStatusHandler())             // legacy job status hook endpoint for scheduler
	router.NewRoute().Method(http.MethodPost).Path("/service/notifications/jobs/replication/:id([0-9]+)").Handler(handler.NewJobStatusHandler())      // legacy job status hook endpoint for replication scheduler
	router.NewRoute().Method(http.MethodPost).Path("/service/notifications/jobs/replication/task/:id([0-9]+)").Handler(handler.NewJobStatusHandler()) // legacy job status hook endpoint for replication task
	router.NewRoute().Method(http.MethodPost).Path("/service/notifications/jobs/retention/task/:id([0-9]+)").Handler(handler.NewJobStatusHandler())
	router.NewRoute().Method(http.MethodPost).Path("/service/notifications/tasks/:id").Handler(handler.NewJobStatusHandler())

	web.Router("/service/token", &token.Handler{})

	// Error pages
	web.ErrorController(&controllers.ErrorController{})
}
