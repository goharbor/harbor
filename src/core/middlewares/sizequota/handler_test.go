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
	utilstest "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

const testingRedisHost = "REDIS_HOST"

func TestMain(m *testing.M) {
	utilstest.InitDatabaseFromEnv()
	rc := m.Run()
	if rc != 0 {
		os.Exit(rc)
	}
}

func TestGetInteceptor(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	res1 := getInteceptor(req1)

	_, ok := res1.(*PutManifestInterceptor)
	assert.True(ok)

	req2, _ := http.NewRequest("POST", "http://127.0.0.1:5000/v2/library/ubuntu/TestGetInteceptor/14.04", nil)
	res2 := getInteceptor(req2)
	assert.Nil(res2)

}

func TestRequireQuota(t *testing.T) {
	con, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", getRedisHost(), 6379),
		redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(time.Minute+10*time.Second),
		redis.DialWriteTimeout(10*time.Second),
	)
	assert.Nil(t, err)
	defer con.Close()

	assert := assert.New(t)
	blobInfo := &util.BlobInfo{
		Repository: "library/test",
		Digest:     "sha256:abcdf123sdfefeg1246",
	}

	err = requireQuota(con, blobInfo)
	assert.Nil(err)

}

func TestGetUUID(t *testing.T) {
	str1 := "test/1/2/uuid-1"
	uuid1 := getUUID(str1)
	assert.Equal(t, uuid1, "uuid-1")

	// not a valid path, just return empty
	str2 := "test-1-2-uuid-2"
	uuid2 := getUUID(str2)
	assert.Equal(t, uuid2, "")
}

func TestAddRmUUID(t *testing.T) {
	con, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", getRedisHost(), 6379),
		redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(time.Minute+10*time.Second),
		redis.DialWriteTimeout(10*time.Second),
	)
	assert.Nil(t, err)
	defer con.Close()

	rmfail, err := rmBlobUploadUUID(con, "test-rm-uuid")
	assert.Nil(t, err)
	assert.True(t, rmfail)

	success, err := util.SetBunkSize(con, "test-rm-uuid", 1000)
	assert.Nil(t, err)
	assert.True(t, success)

	rmSuccess, err := rmBlobUploadUUID(con, "test-rm-uuid")
	assert.Nil(t, err)
	assert.True(t, rmSuccess)

}

func TestTryFreeLockBlob(t *testing.T) {
	con, err := redis.Dial(
		"tcp",
		fmt.Sprintf("%s:%d", getRedisHost(), 6379),
		redis.DialConnectTimeout(30*time.Second),
		redis.DialReadTimeout(time.Minute+10*time.Second),
		redis.DialWriteTimeout(10*time.Second),
	)
	assert.Nil(t, err)
	defer con.Close()

	blobInfo := util.BlobInfo{
		Repository: "lock/test",
		Digest:     "sha256:abcdf123sdfefeg1246",
	}

	lock, err := tryLockBlob(con, &blobInfo)
	assert.Nil(t, err)
	blobInfo.DigestLock = lock
	tryFreeBlob(&blobInfo)
}

func TestBlobCommon(t *testing.T) {
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
		Repository:  "TestBlobCommon/test",
		Digest:      "sha256:abcdf12345678sdfefeg1246",
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

	err = HandleBlobCommon(customResW, req)
	assert.Nil(t, err)

}

func getRedisHost() string {
	redisHost := os.Getenv(testingRedisHost)
	if redisHost == "" {
		redisHost = "127.0.0.1" // for local test
	}

	return redisHost
}
