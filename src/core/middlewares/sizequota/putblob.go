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
	"context"
	"errors"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"
	"net/http"
)

// PutBlobInterceptor ...
type PutBlobInterceptor struct {
	blobInfo *util.BlobInfo
}

// NewPutBlobInterceptor ...
func NewPutBlobInterceptor(blobInfo *util.BlobInfo) *PutBlobInterceptor {
	return &PutBlobInterceptor{
		blobInfo: blobInfo,
	}
}

// HandleRequest ...
func (pbi *PutBlobInterceptor) HandleRequest(req *http.Request) error {
	// the redis connection will be closed in the put response.
	con, err := util.GetRegRedisCon()
	if err != nil {
		return err
	}

	defer func() {
		if pbi.blobInfo.UUID != "" {
			_, err := rmBlobUploadUUID(con, pbi.blobInfo.UUID)
			if err != nil {
				log.Warningf("error occurred when remove UUID for blob, %v", err)
			}
		}
	}()

	dgstStr := req.FormValue("digest")
	if dgstStr == "" {
		return errors.New("blob digest missing")
	}
	dgst, err := digest.Parse(dgstStr)
	if err != nil {
		return errors.New("blob digest parsing failed")
	}

	pbi.blobInfo.Digest = dgst.String()
	pbi.blobInfo.UUID = getUUID(req.URL.Path)
	size, err := util.GetBlobSize(con, pbi.blobInfo.UUID)
	if err != nil {
		return err
	}
	pbi.blobInfo.Size = size
	if err := requireQuota(con, pbi.blobInfo); err != nil {
		return err
	}
	*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, pbi.blobInfo)))
	return nil
}

// HandleResponse ...
func (pbi *PutBlobInterceptor) HandleResponse(rw util.CustomResponseWriter, req *http.Request) {
	if err := HandleBlobCommon(rw, req); err != nil {
		log.Error(err)
	}
}
