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
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor/quota"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/opencontainers/go-digest"
)

var (
	defaultBuilders = []interceptor.Builder{
		&deleteManifestBuilder{},
		&putManifestBuilder{},
	}
)

type deleteManifestBuilder struct {
}

func (*deleteManifestBuilder) Build(req *http.Request) interceptor.Interceptor {
	if req.Method != http.MethodDelete {
		return nil
	}

	match, name, reference := util.MatchManifestURL(req)
	if !match {
		return nil
	}

	dgt, err := digest.Parse(reference)
	if err != nil {
		// Delete manifest only accept digest as reference
		return nil
	}

	projectName := strings.Split(name, "/")[0]
	project, err := dao.GetProjectByName(projectName)
	if err != nil {
		log.Errorf("Failed to get project %s, error: %v", projectName, err)
		return nil
	}
	if project == nil {
		log.Warningf("Project %s not found", projectName)
		return nil
	}

	info := &util.MfInfo{
		ProjectID:  project.ProjectID,
		Repository: name,
		Digest:     dgt.String(),
	}

	// Manifest info will be used by computeQuotaForUpload
	*req = *req.WithContext(util.NewManifestInfoContext(req.Context(), info))

	opts := []quota.Option{
		quota.WithManager("project", strconv.FormatInt(project.ProjectID, 10)),
		quota.WithAction(quota.SubtractAction),
		quota.StatusCode(http.StatusAccepted),
		quota.MutexKeys(mutexKey(info)),
		quota.OnResources(computeQuotaForDelete),
		quota.OnFulfilled(func(http.ResponseWriter, *http.Request) error {
			return dao.DeleteArtifactByDigest(info.ProjectID, info.Repository, info.Digest)
		}),
	}

	return quota.New(opts...)
}

type putManifestBuilder struct {
}

func (b *putManifestBuilder) Build(req *http.Request) interceptor.Interceptor {
	if req.Method != http.MethodPut {
		return nil
	}

	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		// assert that manifest info will be set by others
		return nil
	}

	opts := []quota.Option{
		quota.WithManager("project", strconv.FormatInt(info.ProjectID, 10)),
		quota.WithAction(quota.AddAction),
		quota.StatusCode(http.StatusCreated),
		quota.MutexKeys(mutexKey(info)),
		quota.OnResources(computeQuotaForPut),
		quota.OnFulfilled(func(http.ResponseWriter, *http.Request) error {
			newManifest, overwriteTag := !info.Exist, info.DigestChanged

			if newManifest {
				if err := b.doNewManifest(info); err != nil {
					log.Errorf("Failed to handle response for new manifest, error: %v", err)
				}
			} else if overwriteTag {
				if err := b.doOverwriteTag(info); err != nil {
					log.Errorf("Failed to handle response for overwrite tag, error: %v", err)
				}
			}

			return nil
		}),
	}

	return quota.New(opts...)
}

func (b *putManifestBuilder) doNewManifest(info *util.MfInfo) error {
	artifact := &models.Artifact{
		PID:      info.ProjectID,
		Repo:     info.Repository,
		Tag:      info.Tag,
		Digest:   info.Digest,
		PushTime: time.Now(),
		Kind:     "Docker-Image",
	}

	if _, err := dao.AddArtifact(artifact); err != nil {
		return fmt.Errorf("error to add artifact, %v", err)
	}

	return b.attachBlobsToArtifact(info)
}

func (b *putManifestBuilder) doOverwriteTag(info *util.MfInfo) error {
	artifact := &models.Artifact{
		ID:       info.ArtifactID,
		PID:      info.ProjectID,
		Repo:     info.Repository,
		Tag:      info.Tag,
		Digest:   info.Digest,
		PushTime: time.Now(),
		Kind:     "Docker-Image",
	}

	if err := dao.UpdateArtifactDigest(artifact); err != nil {
		return fmt.Errorf("error to update artifact, %v", err)
	}

	return b.attachBlobsToArtifact(info)
}

func (b *putManifestBuilder) attachBlobsToArtifact(info *util.MfInfo) error {
	self := &models.ArtifactAndBlob{
		DigestAF:   info.Digest,
		DigestBlob: info.Digest,
	}

	artifactBlobs := append([]*models.ArtifactAndBlob{}, self)

	for _, d := range info.Refrerence {
		artifactBlob := &models.ArtifactAndBlob{
			DigestAF:   info.Digest,
			DigestBlob: d.Digest.String(),
		}

		artifactBlobs = append(artifactBlobs, artifactBlob)
	}

	if err := dao.AddArtifactNBlobs(artifactBlobs); err != nil {
		if strings.Contains(err.Error(), dao.ErrDupRows.Error()) {
			log.Warning("the artifact and blobs have already in the DB, it maybe an existing image with different tag")
			return nil
		}

		return fmt.Errorf("error to add artifact and blobs in proxy response handler, %v", err)
	}

	return nil
}
