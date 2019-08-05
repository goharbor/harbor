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
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewMountBlobInterceptor(t *testing.T) {
	blobinfo := &util.BlobInfo{}
	blobinfo.Repository = "TestNewMountBlobInterceptor/latest"

	bi := NewMountBlobInterceptor(blobinfo)
	assert.NotNil(t, bi)
}

func TestMountBlobHandleRequest(t *testing.T) {
	blobInfo := util.BlobInfo{
		Repository:  "TestHandleRequest/test",
		Digest:      "sha256:TestHandleRequest1234",
		ContentType: "ContentType",
		Size:        101,
		Exist:       false,
	}
	req, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	bi := NewMountBlobInterceptor(&blobInfo)
	assert.NotNil(t, bi.HandleRequest(req))
}

func TestMountBlobHandleResponse(t *testing.T) {
	con, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", getRedisHost(), 6379),
		redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(time.Minute+10*time.Second),
		redis.DialWriteTimeout(10*time.Second),
	)
	assert.Nil(t, err)
	defer con.Close()

	req, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	blobInfo := util.BlobInfo{
		Repository:  "TestHandleResponse/test",
		Digest:      "sha256:TestHandleResponseabcdf12345678sdfefeg1246",
		ContentType: "ContentType",
		Size:        101,
		Exist:       false,
	}

	rw := httptest.NewRecorder()
	customResW := util.CustomResponseWriter{ResponseWriter: rw}
	customResW.WriteHeader(201)

	lock, err := tryLockBlob(con, &blobInfo)
	assert.Nil(t, err)
	blobInfo.DigestLock = lock

	*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, &blobInfo)))

	bi := NewMountBlobInterceptor(&blobInfo)
	assert.NotNil(t, bi)

	bi.HandleResponse(customResW, req)

}
