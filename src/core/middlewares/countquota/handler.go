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

package countquota

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	common_redis "github.com/goharbor/harbor/src/common/utils/redis"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = time.Minute + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
)

// ErrRequireQuota ...
var ErrRequireQuota = errors.New("cannot get quota on project for request")

type countQuotaHandler struct {
	next   http.Handler
	mfInfo *util.MfInfo
}

// New ...
func New(next http.Handler) http.Handler {
	return &countQuotaHandler{
		next: next,
	}
}

// ServeHTTP manifest ...
func (cqh *countQuotaHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	match, repository, tag := util.MatchPushManifest(req)
	if match {
		mfInfo := &util.MfInfo{
			Repository: repository,
			Tag:        tag,
		}
		cqh.mfInfo = mfInfo

		mediaType := req.Header.Get("Content-Type")
		if mediaType == schema1.MediaTypeManifest ||
			mediaType == schema1.MediaTypeSignedManifest ||
			mediaType == schema2.MediaTypeManifest {

			tagLock, err := cqh.tryLockTag()
			if err != nil {
				log.Warningf("Error occurred when to lock tag %s:%s with digest %v", repository, tag, err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to lock tag %s:%s with digest %v", repository, tag, err)), http.StatusInternalServerError)
				return
			}
			cqh.mfInfo.TagLock = tagLock

			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				cqh.tryFreeTag()
				log.Warningf("Error occurred when to copy manifest body %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to decode manifest body %v", err)), http.StatusInternalServerError)
				return
			}
			req.Body = ioutil.NopCloser(bytes.NewBuffer(data))

			manifest, desc, err := distribution.UnmarshalManifest(mediaType, data)
			if err != nil {
				cqh.tryFreeTag()
				log.Warningf("Error occurred when to Unmarshal Manifest %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to Unmarshal Manifest %v", err)), http.StatusInternalServerError)
				return
			}
			cqh.mfInfo.Refrerence = manifest.References()
			cqh.mfInfo.Digest = desc.Digest.String()

			projectID, err := cqh.getProjectID(strings.Split(repository, "/")[0])
			if err != nil {
				log.Warningf("Error occurred when to get project ID %v", err)
				return
			}
			cqh.mfInfo.ProjectID = projectID

			imageExist, af, err := cqh.imageExist()
			if err != nil {
				cqh.tryFreeTag()
				log.Warningf("Error occurred when to check Manifest existence by repo and tag name %v", err)
				http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to check Manifest existence %v", err)), http.StatusInternalServerError)
				return
			}
			cqh.mfInfo.Exist = imageExist
			if imageExist {
				if af.Digest != cqh.mfInfo.Digest {
					cqh.mfInfo.DigestChanged = true
				}
			} else {
				quotaRes := &quota.ResourceList{
					quota.ResourceCount: 1,
				}
				err := cqh.tryRequireQuota(quotaRes)
				if err != nil {
					cqh.tryFreeTag()
					log.Errorf("Cannot get quota for the manifest %v", err)
					if err == ErrRequireQuota {
						http.Error(rw, util.MarshalError("StatusNotAcceptable", fmt.Sprintf("Cannot get quota for the manifest %v", err)), http.StatusNotAcceptable)
						return
					}
					http.Error(rw, util.MarshalError("InternalServerError", fmt.Sprintf("Error occurred when to require quota for the manifest %v", err)), http.StatusInternalServerError)
					return
				}
				cqh.mfInfo.Quota = quotaRes
			}

			*req = *(req.WithContext(context.WithValue(req.Context(), util.MFInfokKey, mfInfo)))
		}

	}

	cqh.next.ServeHTTP(rw, req)
}

// tryLockTag locks tag with redis ...
func (cqh *countQuotaHandler) tryLockTag() (*common_redis.Mutex, error) {
	con, err := redis.DialURL(
		config.GetRedisOfRegURL(),
		redis.DialConnectTimeout(dialConnectionTimeout),
		redis.DialReadTimeout(dialReadTimeout),
		redis.DialWriteTimeout(dialWriteTimeout),
	)
	if err != nil {
		return nil, err
	}
	tagLock := common_redis.New(con, cqh.mfInfo.Repository+":"+cqh.mfInfo.Tag, common_util.GenerateRandomString())
	success, err := tagLock.Require()
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("unable to lock tag: %s ", cqh.mfInfo.Repository+":"+cqh.mfInfo.Tag)
	}
	return tagLock, nil
}

func (cqh *countQuotaHandler) tryFreeTag() {
	_, err := cqh.mfInfo.TagLock.Free()
	if err != nil {
		log.Warningf("Error to unlock tag: %s, with error: %v ", cqh.mfInfo.Tag, err)
	}
}

// check the existence of a artifact, if exist, the method will return the artifact model
func (cqh *countQuotaHandler) imageExist() (exist bool, af *models.Artifact, err error) {
	artifactQuery := &models.ArtifactQuery{
		PID:  cqh.mfInfo.ProjectID,
		Repo: cqh.mfInfo.Repository,
		Tag:  cqh.mfInfo.Tag,
	}
	afs, err := dao.ListArtifacts(artifactQuery)
	if err != nil {
		log.Errorf("Error occurred when to get project ID %v", err)
		return false, nil, err
	}
	if len(afs) > 0 {
		return true, afs[0], nil
	}
	return false, nil, nil
}

func (cqh *countQuotaHandler) tryRequireQuota(quotaRes *quota.ResourceList) error {
	quotaMgr, err := quota.NewManager("project", strconv.FormatInt(cqh.mfInfo.ProjectID, 10))
	if err != nil {
		log.Errorf("Error occurred when to new quota manager %v", err)
		return err
	}
	if err := quotaMgr.AddResources(*quotaRes); err != nil {
		log.Errorf("Cannot get quota for the manifest %v", err)
		return ErrRequireQuota
	}
	return nil
}

func (cqh *countQuotaHandler) getProjectID(name string) (int64, error) {
	project, err := dao.GetProjectByName(name)
	if err != nil {
		return 0, err
	}
	if project != nil {
		return project.ProjectID, nil
	}
	return 0, fmt.Errorf("project %s is not found", name)
}
