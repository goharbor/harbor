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
	"fmt"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
	"strings"
)

// MountBlobInterceptor ...
type MountBlobInterceptor struct {
	blobInfo *util.BlobInfo
}

// NewMountBlobInterceptor ...
func NewMountBlobInterceptor(blobInfo *util.BlobInfo) *MountBlobInterceptor {
	return &MountBlobInterceptor{
		blobInfo: blobInfo,
	}
}

// HandleRequest ...
func (mbi *MountBlobInterceptor) HandleRequest(req *http.Request) error {
	tProjectID, err := util.GetProjectID(strings.Split(mbi.blobInfo.Repository, "/")[0])
	if err != nil {
		return fmt.Errorf("error occurred when to get target project: %d, %v", tProjectID, err)
	}
	blob, err := dao.GetBlob(mbi.blobInfo.Digest)
	if err != nil {
		return err
	}
	if blob == nil {
		return fmt.Errorf("the blob in the mount request with digest: %s doesn't exist", mbi.blobInfo.Digest)
	}
	mbi.blobInfo.Size = blob.Size
	con, err := util.GetRegRedisCon()
	if err != nil {
		return err
	}
	if err := requireQuota(con, mbi.blobInfo); err != nil {
		return err
	}
	*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, mbi.blobInfo)))
	return nil
}

// HandleResponse ...
func (mbi *MountBlobInterceptor) HandleResponse(rw util.CustomResponseWriter, req *http.Request) {
	if err := HandleBlobCommon(rw, req); err != nil {
		log.Error(err)
	}
}
