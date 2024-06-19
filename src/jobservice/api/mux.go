package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type RouterWrapper struct {
	*mux.Router
}

// get route variable for request
func getPathParams(req *http.Request) map[string]string {
	return mux.Vars(req)
}

func newMuxRouter() *mux.Router {
	return mux.NewRouter()
}
