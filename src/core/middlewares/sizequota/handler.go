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
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
	"strings"
	"time"
)

type sizeQuotaHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &sizeQuotaHandler{
		next: next,
	}
}

// ServeHTTP ...
func (sqh *sizeQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	sizeInteceptor := getInteceptor(req)
	if sizeInteceptor == nil {
		sqh.next.ServeHTTP(rw, req)
		return
	}

	// handler request
	if err := sizeInteceptor.HandleRequest(req); err != nil {
		log.Warningf("Error occurred when to handle request in size quota handler: %v", err)
		http.Error(rw, util.MarshalError("InternalError", fmt.Sprintf("Error occurred when to handle request in size quota handler: %v", err)),
			http.StatusInternalServerError)
		return
	}
	sqh.next.ServeHTTP(rw, req)

	// handler response
	sizeInteceptor.HandleResponse(*rw.(*util.CustomResponseWriter), req)
}

func getInteceptor(req *http.Request) util.RegInterceptor {
	// POST /v2/<name>/blobs/uploads/?mount=<digest>&from=<repository name>
	matchMountBlob, repository, mount, _ := util.MatchMountBlobURL(req)
	if matchMountBlob {
		bb := util.BlobInfo{}
		bb.Repository = repository
		bb.Digest = mount
		return NewMountBlobInterceptor(&bb)
	}

	// PUT /v2/<name>/blobs/uploads/<uuid>?digest=<digest>
	matchPutBlob, repository := util.MatchPutBlobURL(req)
	if matchPutBlob {
		bb := util.BlobInfo{}
		bb.Repository = repository
		return NewPutBlobInterceptor(&bb)
	}

	// PUT /v2/<name>/manifests/<reference>
	matchPushMF, repository, tag := util.MatchPushManifest(req)
	if matchPushMF {
		bb := util.BlobInfo{}
		mfInfo := util.MfInfo{}
		bb.Repository = repository
		mfInfo.Repository = repository
		mfInfo.Tag = tag
		return NewPutManifestInterceptor(&bb, &mfInfo)
	}

	// PATCH /v2/<name>/blobs/uploads/<uuid>
	matchPatchBlob, _ := util.MatchPatchBlobURL(req)
	if matchPatchBlob {
		return NewPatchBlobInterceptor()
	}

	return nil
}

func requireQuota(conn redis.Conn, blobInfo *util.BlobInfo) error {
	projectID, err := util.GetProjectID(strings.Split(blobInfo.Repository, "/")[0])
	if err != nil {
		return err
	}
	blobInfo.ProjectID = projectID

	digestLock, err := tryLockBlob(conn, blobInfo)
	if err != nil {
		log.Infof("failed to lock digest in redis, %v", err)
		return err
	}
	blobInfo.DigestLock = digestLock

	blobExist, err := dao.HasBlobInProject(blobInfo.ProjectID, blobInfo.Digest)
	if err != nil {
		tryFreeBlob(blobInfo)
		return err
	}
	blobInfo.Exist = blobExist
	if blobExist {
		return nil
	}

	// only require quota for non existing blob.
	quotaRes := &quota.ResourceList{
		quota.ResourceStorage: blobInfo.Size,
	}
	err = util.TryRequireQuota(blobInfo.ProjectID, quotaRes)
	if err != nil {
		log.Infof("project id, %d, size %d", blobInfo.ProjectID, blobInfo.Size)
		tryFreeBlob(blobInfo)
		log.Errorf("cannot get quota for the blob %v", err)
		return err
	}
	blobInfo.Quota = quotaRes

	return nil
}

// HandleBlobCommon handles put blob complete request
// 1, add blob into DB if success
// 2, roll back resource if failure.
func HandleBlobCommon(rw util.CustomResponseWriter, req *http.Request) error {
	bbInfo := req.Context().Value(util.BBInfokKey)
	bb, ok := bbInfo.(*util.BlobInfo)
	if !ok {
		return errors.New("failed to convert blob information context into BBInfo")
	}
	defer func() {
		_, err := bb.DigestLock.Free()
		if err != nil {
			log.Errorf("Error to unlock blob digest:%s in response handler, %v", bb.Digest, err)
		}
		if err := bb.DigestLock.Conn.Close(); err != nil {
			log.Errorf("Error to close redis connection in put blob response handler, %v", err)
		}
	}()

	// Do nothing for a existing blob.
	if bb.Exist {
		return nil
	}

	if rw.Status() == http.StatusCreated {
		blob := &models.Blob{
			Digest:       bb.Digest,
			ContentType:  bb.ContentType,
			Size:         bb.Size,
			CreationTime: time.Now(),
		}
		_, err := dao.AddBlob(blob)
		if err != nil {
			return err
		}
	} else if rw.Status() >= 300 && rw.Status() <= 511 {
		success := util.TryFreeQuota(bb.ProjectID, bb.Quota)
		if !success {
			return fmt.Errorf("Error to release resource booked for the blob, %d, digest: %s ", bb.ProjectID, bb.Digest)
		}
	}
	return nil
}

// tryLockBlob locks blob with redis ...
func tryLockBlob(conn redis.Conn, blobInfo *util.BlobInfo) (*common_redis.Mutex, error) {
	// Quota::blob-lock::projectname::digest
	digestLock := common_redis.New(conn, "Quota::blob-lock::"+strings.Split(blobInfo.Repository, "/")[0]+":"+blobInfo.Digest, common_util.GenerateRandomString())
	success, err := digestLock.Require()
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("unable to lock digest: %s, %s ", blobInfo.Repository, blobInfo.Digest)
	}
	return digestLock, nil
}

func tryFreeBlob(blobInfo *util.BlobInfo) {
	_, err := blobInfo.DigestLock.Free()
	if err != nil {
		log.Warningf("Error to unlock digest: %s,%s with error: %v ", blobInfo.Repository, blobInfo.Digest, err)
	}
}

func rmBlobUploadUUID(conn redis.Conn, UUID string) (bool, error) {
	exists, err := redis.Int(conn.Do("EXISTS", UUID))
	if err != nil {
		return false, err
	}
	if exists == 1 {
		res, err := redis.Int(conn.Do("DEL", UUID))
		if err != nil {
			return false, err
		}
		return res == 1, nil
	}
	return true, nil
}

// put blob path: /v2/<name>/blobs/uploads/<uuid>
func getUUID(path string) string {
	if !strings.Contains(path, "/") {
		log.Infof("it's not a valid path string: %s", path)
		return ""
	}
	strs := strings.Split(path, "/")
	return strs[len(strs)-1]
}
