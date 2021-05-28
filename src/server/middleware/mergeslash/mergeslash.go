package mergeslash

import (
	"net/http"
	"regexp"

	"github.com/goharbor/harbor/src/server/middleware"
)

var multiSlash = regexp.MustCompile(`(/+)`)

// Middleware creates the middleware to merge slashes in the URL path of the request
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		path := multiSlash.ReplaceAll([]byte(r.URL.Path), []byte("/"))
		r.URL.Path = string(path)
		next.ServeHTTP(w, r)
	}, skippers...)
}
