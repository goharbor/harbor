// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package regquota

import (
	"bytes"
	"fmt"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"io/ioutil"
	"net/http"
)

type regQuotaHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &regQuotaHandler{
		next: next,
	}
}

// ServeHTTP PATCH manifest ...
func (rqh regQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, _, _ := util.MatchManifestURL(req)
	if match {
		var mfSize int64
		var mfDigest string
		mediaType := req.Header.Get("Content-Type")
		if req.Method == http.MethodPut {
			if mediaType == schema1.MediaTypeManifest ||
				mediaType == schema1.MediaTypeSignedManifest ||
				mediaType == schema2.MediaTypeManifest {
				data, err := ioutil.ReadAll(req.Body)
				if err != nil {
					log.Warningf("Error occurred when to copy manifest body %v", err)
					http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to decode manifest body %v", err)), http.StatusInternalServerError)
					return
				}
				req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

				_, desc, err := distribution.UnmarshalManifest(mediaType, data)
				if err != nil {
					log.Warningf("Error occurred when to Unmarshal Manifest %v", err)
					http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to Unmarshal Manifest %v", err)), http.StatusInternalServerError)
					return
				}
				mfDigest = desc.Digest.String()
				mfSize = desc.Size
				log.Infof("manifest digest... %s", mfDigest)
				log.Infof("manifest size... %v", mfSize)
			}
		}
	}

	rqh.next.ServeHTTP(rw, req)
}
