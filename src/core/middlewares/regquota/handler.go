package regquota

import (
	"bytes"
	"fmt"
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"io/ioutil"
	"net/http"
)

type regQuotaHandler struct {
	next http.Handler
}

func New(next http.Handler) http.Handler {
	return &regQuotaHandler{
		next: next,
	}
}

//PATCH manifest ...
func (rqh regQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, _, _ := util.MatchManifestURL(req)
	if match {
		var mfSize int64
		var mfDigest string
		mediaType := req.Header.Get("Content-Type")
		if req.Method == http.MethodPut && mediaType == "application/vnd.docker.distribution.manifest.v2+json" {
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				log.Warningf("Error occured when to copy manifest body %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occured when to decode manifest body %v", err)), http.StatusInternalServerError)
				return
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

			_, desc, err := distribution.UnmarshalManifest(mediaType, data)
			if err != nil {
				log.Warningf("Error occured when to Unmarshal Manifest %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occured when to Unmarshal Manifest %v", err)), http.StatusInternalServerError)
				return
			}
			mfDigest = desc.Digest.String()
			mfSize = desc.Size
			log.Infof("manifest digest... %s", mfDigest)
			log.Infof("manifest size... %v", mfSize)
		}
	}

	rqh.next.ServeHTTP(rw, req)
}
