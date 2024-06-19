package blob

import (
	"net/http"

	"github.com/gorilla/mux"
)

// get route variable for request
func getPathParams(req *http.Request) map[string]string {
	return mux.Vars(req)
}

// sets URL variables in the request for use in route handlers.
func SetURLVars(req *http.Request, varMap map[string]string) *http.Request {
	return mux.SetURLVars(req, varMap)
}
