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

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/goharbor/harbor/src/pkg/types"
)

func mutexKey(info *util.MfInfo) string {
	if info.Tag != "" {
		return "Quota::manifest-lock::" + info.Repository + ":" + info.Tag
	}

	return "Quota::manifest-lock::" + info.Repository + ":" + info.Digest
}

func computeQuotaForDelete(req *http.Request) (types.ResourceList, error) {
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

func computeQuotaForPut(req *http.Request) (types.ResourceList, error) {
	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		return nil, errors.New("manifest info missing")
	}

	artifact, err := getArtifact(info)
	if err != nil {
		return nil, fmt.Errorf("error occurred when to check Manifest existence %v", err)
	}

	if artifact != nil {
		info.ArtifactID = artifact.ID
		info.DigestChanged = artifact.Digest != info.Digest
		info.Exist = true

		return nil, nil
	}

	return quota.ResourceList{quota.ResourceCount: 1}, nil
}

// get artifact by manifest info
func getArtifact(info *util.MfInfo) (*models.Artifact, error) {
	query := &models.ArtifactQuery{
		PID:  info.ProjectID,
		Repo: info.Repository,
		Tag:  info.Tag,
	}

	artifacts, err := dao.ListArtifacts(query)
	if err != nil || len(artifacts) == 0 {
		return nil, err
	}

	return artifacts[0], nil
}
