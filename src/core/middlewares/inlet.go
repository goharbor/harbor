package middlewares

import (
	"errors"
	"github.com/goharbor/harbor/src/core/middlewares/registryproxy"
	"net/http"
)

var head http.Handler

// Init initialize the Proxy instance and handler chain.
func Init() error {
	ph := registryproxy.New()
	if ph == nil {
		return errors.New("get nil when to create proxy")
	}
	handlerChain := New(Middlewares).Create()
	head = handlerChain.Then(ph)
	return nil
}

// Handle handles the request.
func Handle(rw http.ResponseWriter, req *http.Request) {
	head.ServeHTTP(rw, req)
}
