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

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor/quota"
	"github.com/goharbor/harbor/src/core/middlewares/util"
)

var (
	defaultBuilders = []interceptor.Builder{
		&manifestDeletionBuilder{},
		&manifestCreationBuilder{},
	}
)

type manifestDeletionBuilder struct{}

func (*manifestDeletionBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	if match, _, _ := util.MatchDeleteManifest(req); !match {
		return nil, nil
	}

	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		var err error
		info, err = util.ParseManifestInfoFromPath(req)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest, error %v", err)
		}

		// Manifest info will be used by computeResourcesForDeleteManifest
		*req = *(req.WithContext(util.NewManifestInfoContext(req.Context(), info)))
	}

	opts := []quota.Option{
		quota.EnforceResources(config.QuotaPerProjectEnable()),
		quota.WithManager("project", strconv.FormatInt(info.ProjectID, 10)),
		quota.WithAction(quota.SubtractAction),
		quota.StatusCode(http.StatusAccepted),
		quota.MutexKeys(info.MutexKey("count")),
		quota.OnResources(computeResourcesForManifestDeletion),
		quota.OnFulfilled(func(http.ResponseWriter, *http.Request) error {
			return dao.DeleteArtifactByDigest(info.ProjectID, info.Repository, info.Digest)
		}),
	}

	return quota.New(opts...), nil
}

type manifestCreationBuilder struct{}

func (*manifestCreationBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	if match, _, _ := util.MatchPushManifest(req); !match {
		return nil, nil
	}

	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		var err error
		info, err = util.ParseManifestInfo(req)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest, error %v", err)
		}

		// Manifest info will be used by computeResourcesForCreateManifest
		*req = *(req.WithContext(util.NewManifestInfoContext(req.Context(), info)))
	}

	opts := []quota.Option{
		quota.EnforceResources(config.QuotaPerProjectEnable()),
		quota.WithManager("project", strconv.FormatInt(info.ProjectID, 10)),
		quota.WithAction(quota.AddAction),
		quota.StatusCode(http.StatusCreated),
		quota.MutexKeys(info.MutexKey("count")),
		quota.OnResources(computeResourcesForManifestCreation),
		quota.OnFulfilled(afterManifestCreated),
	}

	return quota.New(opts...), nil
}
