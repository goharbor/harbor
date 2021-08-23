package trace

import (
	"net/http"

	tracelib "github.com/goharbor/harbor/src/lib/trace"
)

func traceHandler(next http.Handler) http.Handler {
	if tracelib.Enabled() {
		return tracelib.NewHandler(next, "handle-http-request")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// Middleware returns a middleware for handling requests
func Middleware() func(http.Handler) http.Handler {
	return traceHandler
}
