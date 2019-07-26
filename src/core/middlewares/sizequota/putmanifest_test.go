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

func TestNewPutManifestInterceptor(t *testing.T) {
	blobinfo := &util.BlobInfo{}
	blobinfo.Repository = "TestNewPutManifestInterceptor/latest"

	mfinfo := &util.MfInfo{
		Repository: "TestNewPutManifestInterceptor",
	}

	mi := NewPutManifestInterceptor(blobinfo, mfinfo)
	assert.NotNil(t, mi)
}

func TestPutManifestHandleRequest(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	blobinfo := &util.BlobInfo{}
	blobinfo.Repository = "TestPutManifestHandleRequest/latest"

	mfinfo := &util.MfInfo{
		Repository: "TestPutManifestHandleRequest",
	}

	mi := NewPutManifestInterceptor(blobinfo, mfinfo)
	assert.NotNil(t, mi.HandleRequest(req))
}

func TestPutManifestHandleResponse(t *testing.T) {
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
		Repository:  "TestPutManifestandleResponse/test",
		Digest:      "sha256:TestPutManifestandleResponseabcdf12345678sdfefeg1246",
		ContentType: "ContentType",
		Size:        101,
		Exist:       false,
	}

	mfinfo := util.MfInfo{
		Repository: "TestPutManifestandleResponse",
	}

	rw := httptest.NewRecorder()
	customResW := util.CustomResponseWriter{ResponseWriter: rw}
	customResW.WriteHeader(201)

	lock, err := tryLockBlob(con, &blobInfo)
	assert.Nil(t, err)
	blobInfo.DigestLock = lock

	*req = *(req.WithContext(context.WithValue(req.Context(), util.BBInfokKey, &blobInfo)))

	bi := NewPutManifestInterceptor(&blobInfo, &mfinfo)
	assert.NotNil(t, bi)

	bi.HandleResponse(customResW, req)
}
