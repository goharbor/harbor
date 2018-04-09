// Copyright 2018 The Harbor Authors. All rights reserved.

package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vmware/harbor/src/jobservice/errs"
)

const (
	baseRoute  = "/api"
	apiVersion = "v1"
)

//Router defines the related routes for the job service and directs the request
//to the right handler method.
type Router interface {
	//ServeHTTP used to handle the http requests
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

//BaseRouter provides the basic routes for the job service based on the golang http server mux.
type BaseRouter struct {
	//Use mux to keep the routes mapping.
	router *mux.Router

	//Handler used to handle the requests
	handler Handler

	//Do auth
	authenticator Authenticator
}

//NewBaseRouter is the constructor of BaseRouter.
func NewBaseRouter(handler Handler, authenticator Authenticator) Router {
	br := &BaseRouter{
		router:        mux.NewRouter(),
		handler:       handler,
		authenticator: authenticator,
	}

	//Register routes here
	br.registerRoutes()

	return br
}

//ServeHTTP is the implementation of Router interface.
func (br *BaseRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	//Do auth
	if err := br.authenticator.DoAuth(req); err != nil {
		authErr := errs.UnauthorizedError(err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(authErr.Error()))
		return
	}

	//Directly pass requests to the server mux.
	br.router.ServeHTTP(w, req)
}

//registerRoutes adds routes to the server mux.
func (br *BaseRouter) registerRoutes() {
	subRouter := br.router.PathPrefix(fmt.Sprintf("%s/%s", baseRoute, apiVersion)).Subrouter()

	subRouter.HandleFunc("/jobs", br.handler.HandleLaunchJobReq).Methods(http.MethodPost)
	subRouter.HandleFunc("/jobs/{job_id}", br.handler.HandleGetJobReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/jobs/{job_id}", br.handler.HandleJobActionReq).Methods(http.MethodPost)
	subRouter.HandleFunc("/jobs/{job_id}/log", br.handler.HandleJobLogReq).Methods(http.MethodGet)
	subRouter.HandleFunc("/stats", br.handler.HandleCheckStatusReq).Methods(http.MethodGet)
}
