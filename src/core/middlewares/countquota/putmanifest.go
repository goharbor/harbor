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
	"context"
	"errors"
	"fmt"
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

// PutManifestInterceptor ...
type PutManifestInterceptor struct {
	mfInfo *util.MfInfo
}

// NewPutManifestInterceptor ...
func NewPutManifestInterceptor(mfInfo *util.MfInfo) *PutManifestInterceptor {
	return &PutManifestInterceptor{
		mfInfo: mfInfo,
	}
}

// HandleRequest ...
// The context has already contain mfinfo as it was put by size quota handler.
func (pmi *PutManifestInterceptor) HandleRequest(req *http.Request) error {
	mfInfo := req.Context().Value(util.MFInfokKey)
	mf, ok := mfInfo.(*util.MfInfo)
	if !ok {
		return errors.New("failed to get manifest infor from context")
	}

	tagLock, err := tryLockTag(mf)
	if err != nil {
		return fmt.Errorf("error occurred when to lock tag %s:%s with digest %v", mf.Repository, mf.Tag, err)
	}
	mf.TagLock = tagLock

	imageExist, af, err := imageExist(mf)
	if err != nil {
		tryFreeTag(mf)
		return fmt.Errorf("error occurred when to check Manifest existence %v", err)
	}
	mf.Exist = imageExist
	if imageExist {
		if af.Digest != mf.Digest {
			mf.DigestChanged = true
		}
	} else {
		quotaRes := &quota.ResourceList{
			quota.ResourceCount: 1,
		}
		err := util.TryRequireQuota(mf.ProjectID, quotaRes)
		if err != nil {
			tryFreeTag(mf)
			log.Errorf("Cannot get quota for the manifest %v", err)
			if err == util.ErrRequireQuota {
				return err
			}
			return fmt.Errorf("error occurred when to require quota for the manifest %v", err)
		}
		mf.Quota = quotaRes
	}
	*req = *(req.WithContext(context.WithValue(req.Context(), util.MFInfokKey, mf)))
	return nil
}

// HandleResponse ...
func (pmi *PutManifestInterceptor) HandleResponse(rw util.CustomResponseWriter, req *http.Request) {
	mfInfo := req.Context().Value(util.MFInfokKey)
	mf, ok := mfInfo.(*util.MfInfo)
	if !ok {
		log.Error("failed to convert manifest information context into MfInfo")
		return
	}
	defer func() {
		_, err := mf.TagLock.Free()
		if err != nil {
			log.Errorf("Error to unlock in response handler, %v", err)
		}
		if err := mf.TagLock.Conn.Close(); err != nil {
			log.Errorf("Error to close redis connection in response handler, %v", err)
		}
	}()

	// 201
	if rw.Status() == http.StatusCreated {
		af := &models.Artifact{
			PID:      mf.ProjectID,
			Repo:     mf.Repository,
			Tag:      mf.Tag,
			Digest:   mf.Digest,
			PushTime: time.Now(),
			Kind:     "Docker-Image",
		}

		// insert or update
		if !mf.Exist {
			_, err := dao.AddArtifact(af)
			if err != nil {
				log.Errorf("Error to add artifact, %v", err)
				return
			}
		}
		if mf.DigestChanged {
			err := dao.UpdateArtifactDigest(af)
			if err != nil {
				log.Errorf("Error to add artifact, %v", err)
				return
			}
		}

		if !mf.Exist || mf.DigestChanged {
			afnbs := []*models.ArtifactAndBlob{}
			self := &models.ArtifactAndBlob{
				DigestAF:   mf.Digest,
				DigestBlob: mf.Digest,
			}
			afnbs = append(afnbs, self)
			for _, d := range mf.Refrerence {
				afnb := &models.ArtifactAndBlob{
					DigestAF:   mf.Digest,
					DigestBlob: d.Digest.String(),
				}
				afnbs = append(afnbs, afnb)
			}
			if err := dao.AddArtifactNBlobs(afnbs); err != nil {
				if strings.Contains(err.Error(), dao.ErrDupRows.Error()) {
					log.Warning("the artifact and blobs have already in the DB, it maybe an existing image with different tag")
					return
				}
				log.Errorf("Error to add artifact and blobs in proxy response handler, %v", err)
				return
			}
		}

	} else if rw.Status() >= 300 || rw.Status() <= 511 {
		if !mf.Exist {
			success := util.TryFreeQuota(mf.ProjectID, mf.Quota)
			if !success {
				log.Error("error to release resource booked for the manifest")
				return
			}
		}
	}

	return
}

// tryLockTag locks tag with redis ...
func tryLockTag(mfInfo *util.MfInfo) (*common_redis.Mutex, error) {
	con, err := util.GetRegRedisCon()
	if err != nil {
		return nil, err
	}
	tagLock := common_redis.New(con, "Quota::manifest-lock::"+mfInfo.Repository+":"+mfInfo.Tag, common_util.GenerateRandomString())
	success, err := tagLock.Require()
	if err != nil {
		return nil, err
	}
	if !success {
		return nil, fmt.Errorf("unable to lock tag: %s ", mfInfo.Repository+":"+mfInfo.Tag)
	}
	return tagLock, nil
}

func tryFreeTag(mfInfo *util.MfInfo) {
	_, err := mfInfo.TagLock.Free()
	if err != nil {
		log.Warningf("Error to unlock tag: %s, with error: %v ", mfInfo.Tag, err)
	}
}

// check the existence of a artifact, if exist, the method will return the artifact model
func imageExist(mfInfo *util.MfInfo) (exist bool, af *models.Artifact, err error) {
	artifactQuery := &models.ArtifactQuery{
		PID:  mfInfo.ProjectID,
		Repo: mfInfo.Repository,
		Tag:  mfInfo.Tag,
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
