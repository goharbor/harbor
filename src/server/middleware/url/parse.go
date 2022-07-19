package url

import (
	"net/http"
	"net/url"

	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/server/middleware"
)

// Middleware middleware which validates the raw query, especially for the invalid semicolon separator.
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if r.URL != nil && r.URL.RawQuery != "" {
			_, err := url.ParseQuery(r.URL.RawQuery)
			if err != nil {
				lib_http.SendError(w, errors.New(err).WithCode(errors.BadRequestCode))
				return
			}
		}
		next.ServeHTTP(w, r)
	}, skippers...)
}
