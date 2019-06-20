package multiplmanifest

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
	"strings"
)

type MultipleManifestHandler struct {
	next http.Handler
}

func New(next http.Handler) http.Handler {
	return &MultipleManifestHandler{
		next: next,
	}
}

// The handler is responsible for blocking request to upload manifest list by docker client, which is not supported so far by Harbor.
func (mh MultipleManifestHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, _, _ := util.MatchManifestURL(req)
	if match {
		contentType := req.Header.Get("Content-type")
		// application/vnd.docker.distribution.manifest.list.v2+json
		if strings.Contains(contentType, "manifest.list.v2") {
			log.Debugf("Content-type: %s is not supported, failing the response.", contentType)
			http.Error(rw, util.MarshalError("UNSUPPORTED_MEDIA_TYPE", "Manifest.list is not supported."), http.StatusUnsupportedMediaType)
			return
		}
	}
	mh.next.ServeHTTP(rw, req)
}
