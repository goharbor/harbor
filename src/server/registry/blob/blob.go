package blob

import (
	"net/http"
	"net/http/httputil"
)

// NewHandler returns the handler to handler catalog request
func NewHandler(proxy *httputil.ReverseProxy) http.Handler {
	return &handler{
		proxy: proxy,
	}
}

type handler struct {
	proxy *httputil.ReverseProxy
}

// ServeHTTP ...
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.proxy.ServeHTTP(w, req)
}
