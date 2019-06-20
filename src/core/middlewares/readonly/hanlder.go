package readonly

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
)

type readonlyHandler struct {
	next http.Handler
}

func New(next http.Handler) http.Handler {
	return &readonlyHandler{
		next: next,
	}
}

func (rh readonlyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if config.ReadOnly() {
		if req.Method == http.MethodDelete || req.Method == http.MethodPost || req.Method == http.MethodPatch || req.Method == http.MethodPut {
			log.Warningf("The request is prohibited in readonly mode, url is: %s", req.URL.Path)
			http.Error(rw, util.MarshalError("DENIED", "The system is in read only mode. Any modification is prohibited."), http.StatusForbidden)
			return
		}
	}
	rh.next.ServeHTTP(rw, req)
}
