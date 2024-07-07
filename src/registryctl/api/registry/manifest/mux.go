package manifest

import (
	"net/http"

	"github.com/gorilla/mux"
)

// extract the path variable
func getPathVars(r *http.Request, path string) string {
	return mux.Vars(r)[path]
}
