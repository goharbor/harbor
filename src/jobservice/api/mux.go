package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

// BaseRouter provides the basic routes for the job service based on the golang http server mux.
type BaseRouter struct {
	// Use mux to keep the routes mapping.
	router *mux.Router

	// Handler used to handle the requests
	handler Handler

	// Do auth
	authenticator Authenticator
}

// get route variable for request
func getPathParams(req *http.Request) map[string]string {
	return mux.Vars(req)
}

func NewMuxRouter() *mux.Router {
	return mux.NewRouter()
}

func addTracingMiddleware(br *BaseRouter) {
	br.router.Use(otelmux.Middleware("serve-http"))
}
