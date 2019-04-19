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
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/gorilla/mux"
)

const (
	baseRoute  = "/api"
	apiVersion = "v1"
)

// Router defines the related routes for the job service and directs the request
// to the right handler method.
type Router interface {
	// ServeHTTP used to handle the http requests
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

// BaseRouter provides the basic routes for the job service based on the golang http server mux.
type BaseRouter struct {
	// Use mux to keep the routes mapping.
	router *mux.Router

	// Handler used to handle the requests
	handler Handler

	// Do auth
	authenticator Authenticator
}

// NewBaseRouter is the constructor of BaseRouter.
func NewBaseRouter(handler Handler, authenticator Authenticator) Router {
	br := &BaseRouter{
		router:        mux.NewRouter(),
		handler:       handler,
		authenticator: authenticator,
	}

	// Register routes here
	br.registerRoutes()

	return br
}

// ServeHTTP is the implementation of Router interface.
func (br *BaseRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// No auth required for /stats as it is a health check endpoint
	// Do auth for other services
	if req.URL.String() != fmt.Sprintf("%s/%s/stats", baseRoute, apiVersion) {
		if err := br.authenticator.DoAuth(req); err != nil {
			authErr := errs.UnauthorizedError(err)
			logger.Errorf("Serve http request '%s %s' failed with error: %s", req.Method, req.URL.String(), authErr.Error())
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(authErr.Error()))
			return
		}
	}

	// Directly pass requests to the server mux.
	br.router.ServeHTTP(w, req)
}

// registerRoutes adds routes to the server mux.
func (br *BaseRouter) registerRoutes() {
	subRouter := br.router.PathPrefix(fmt.Sprintf("%s/%s", baseRoute, apiVersion)).Subrouter()

	subRouter.HandleFunc("/jobs", br.handler.HandleLaunchJobReq).Methods(http.MethodPost)
	subRouter.HandleFunc("/jobs/scheduled", br.handler.HandleScheduledJobs).Methods(http.MethodGet)
	subRouter.HandleFunc("/jobs/{job_id}", br.handler.HandleGetJobReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/jobs/{job_id}", br.handler.HandleJobActionReq).Methods(http.MethodPost)
	subRouter.HandleFunc("/jobs/{job_id}/log", br.handler.HandleJobLogReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/stats", br.handler.HandleCheckStatusReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/jobs/{job_id}/executions", br.handler.HandlePeriodicExecutions).Methods(http.MethodGet)
}
