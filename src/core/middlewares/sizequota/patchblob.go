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

package sizequota

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
	"strconv"
	"strings"
)

// PatchBlobInterceptor ...
type PatchBlobInterceptor struct {
}

// NewPatchBlobInterceptor ...
func NewPatchBlobInterceptor() *PatchBlobInterceptor {
	return &PatchBlobInterceptor{}
}

// HandleRequest do nothing for patch blob, just let the request to proxy.
func (pbi *PatchBlobInterceptor) HandleRequest(req *http.Request) error {
	return nil
}

// HandleResponse record the upload process with Range attribute, set it into redis with UUID as the key
func (pbi *PatchBlobInterceptor) HandleResponse(rw util.CustomResponseWriter, req *http.Request) {
	if rw.Status() != http.StatusAccepted {
		return
	}

	con, err := util.GetRegRedisCon()
	if err != nil {
		log.Error(err)
		return
	}
	defer con.Close()

	uuid := rw.Header().Get("Docker-Upload-UUID")
	if uuid == "" {
		log.Errorf("no UUID in the patch blob response, the request path %s ", req.URL.Path)
		return
	}

	// Range: Range indicating the current progress of the upload.
	// https://github.com/opencontainers/distribution-spec/blob/master/spec.md#get-blob-upload
	patchRange := rw.Header().Get("Range")
	if uuid == "" {
		log.Errorf("no Range in the patch blob response, the request path %s ", req.URL.Path)
		return
	}

	endRange := strings.Split(patchRange, "-")[1]
	size, err := strconv.ParseInt(endRange, 10, 64)
	// docker registry did '-1' in the response
	if size > 0 {
		size = size + 1
	}
	if err != nil {
		log.Error(err)
		return
	}
	success, err := util.SetBunkSize(con, uuid, size)
	if err != nil {
		log.Error(err)
		return
	}
	if !success {
		// ToDo discuss what to do here.
		log.Warningf(" T_T: Fail to set bunk: %s size: %d in redis, it causes unable to set correct quota for the artifact.", uuid, size)
	}
	return
}
