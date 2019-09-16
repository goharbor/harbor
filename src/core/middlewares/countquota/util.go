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
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
)

// computeResourcesForManifestCreation returns count resource required for manifest
// no count required if the tag of the repository exists in the project
func computeResourcesForManifestCreation(req *http.Request) (types.ResourceList, error) {
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		return nil, errors.New("manifest info missing")
	}

	// only count quota required when push new tag
	if info.IsNewTag() {
		return quota.ResourceList{quota.ResourceCount: 1}, nil
	}

	return nil, nil
}

// computeResourcesForManifestDeletion returns count resource will be released when manifest deleted
// then result will be the sum of manifest count of the same repository in the project
func computeResourcesForManifestDeletion(req *http.Request) (types.ResourceList, error) {
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		return nil, errors.New("manifest info missing")
	}

	total, err := dao.GetTotalOfArtifacts(&models.ArtifactQuery{
		PID:    info.ProjectID,
		Repo:   info.Repository,
		Digest: info.Digest,
	})

	if err != nil {
		return nil, fmt.Errorf("error occurred when get artifacts %v ", err)
	}

	return types.ResourceList{types.ResourceCount: total}, nil
}

// afterManifestCreated the handler after manifest created success
// it will create or update the artifact info in db, and then attach blobs to artifact
func afterManifestCreated(w http.ResponseWriter, req *http.Request) error {
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		return errors.New("manifest info missing")
	}

	artifact := info.Artifact()
	if artifact.ID == 0 {
		if _, err := dao.AddArtifact(artifact); err != nil {
			return fmt.Errorf("error to add artifact, %v", err)
		}
	} else {
		if err := dao.UpdateArtifact(artifact); err != nil {
			return fmt.Errorf("error to update artifact, %v", err)
		}
	}

	return attachBlobsToArtifact(info)
}

// attachBlobsToArtifact attach the blobs which from manifest to artifact
func attachBlobsToArtifact(info *util.ManifestInfo) error {
	temp := make(map[string]interface{})
	artifactBlobs := []*models.ArtifactAndBlob{}

	temp[info.Digest] = nil
	// self
	artifactBlobs = append(artifactBlobs, &models.ArtifactAndBlob{
		DigestAF:   info.Digest,
		DigestBlob: info.Digest,
	})

	// avoid the duplicate layers.
	for _, reference := range info.References {
		_, exist := temp[reference.Digest.String()]
		if !exist {
			temp[reference.Digest.String()] = nil
			artifactBlobs = append(artifactBlobs, &models.ArtifactAndBlob{
				DigestAF:   info.Digest,
				DigestBlob: reference.Digest.String(),
			})
		}
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
