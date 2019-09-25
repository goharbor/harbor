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
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor/quota"
	"github.com/goharbor/harbor/src/core/middlewares/util"
)

var (
	defaultBuilders = []interceptor.Builder{
		&blobStreamUploadBuilder{},
		&blobStorageQuotaBuilder{},
		&manifestCreationBuilder{},
		&manifestDeletionBuilder{},
	}
)

// blobStreamUploadBuilder interceptor builder for PATCH /v2/<name>/blobs/uploads/<uuid>
type blobStreamUploadBuilder struct{}

func (*blobStreamUploadBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	if !match(req, http.MethodPatch, blobUploadURLRe) {
		return nil, nil
	}

	s := blobUploadURLRe.FindStringSubmatch(req.URL.Path)
	uuid := s[2]

	onResponse := func(w http.ResponseWriter, req *http.Request) {
		size, err := parseUploadedBlobSize(w)
		if err != nil {
			log.Errorf("failed to parse uploaded blob size for upload %s, error: %v", uuid, err)
			return
		}

		ok, err := setUploadedBlobSize(uuid, size)
		if err != nil {
			log.Errorf("failed to update blob update size for upload %s, error: %v", uuid, err)
			return
		}

		if !ok {
			// ToDo discuss what to do here.
			log.Errorf("fail to set bunk: %s size: %d in redis, it causes unable to set correct quota for the artifact", uuid, size)
		}
	}

	return interceptor.ResponseInterceptorFunc(onResponse), nil
}

// blobStorageQuotaBuilder interceptor builder for these requests
// PUT  /v2/<name>/blobs/uploads/<uuid>?digest=<digest>
// POST /v2/<name>/blobs/uploads/?mount=<digest>&from=<repository name>
type blobStorageQuotaBuilder struct{}

func (*blobStorageQuotaBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	parseBlobInfo := getBlobInfoParser(req)
	if parseBlobInfo == nil {
		return nil, nil
	}

	info, err := parseBlobInfo(req)
	if err != nil {
		return nil, err
	}

	// replace req with blob info context
	*req = *(req.WithContext(util.NewBlobInfoContext(req.Context(), info)))

	opts := []quota.Option{
		quota.EnforceResources(config.QuotaPerProjectEnable()),
		quota.WithManager("project", strconv.FormatInt(info.ProjectID, 10)),
		quota.WithAction(quota.AddAction),
		quota.StatusCode(http.StatusCreated), // NOTICE: mount blob and blob upload complete both return 201 when success
		quota.OnResources(computeResourcesForBlob),
		quota.MutexKeys(info.MutexKey()),
		quota.OnFulfilled(func(http.ResponseWriter, *http.Request) error {
			return syncBlobInfoToProject(info)
		}),
	}

	return quota.New(opts...), nil
}

// manifestCreationBuilder interceptor builder for the request PUT /v2/<name>/manifests/<reference>
type manifestCreationBuilder struct{}

func (*manifestCreationBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	if match, _, _ := util.MatchPushManifest(req); !match {
		return nil, nil
	}

	info, err := util.ParseManifestInfoFromReq(req)
	if err != nil {
		return nil, err
	}

	// Replace request with manifests info context
	*req = *req.WithContext(util.NewManifestInfoContext(req.Context(), info))

	// Sync manifest layers to blobs for foreign layers not pushed and they are not in blob table
	if err := info.SyncBlobs(); err != nil {
		log.Warningf("Failed to sync blobs, error: %v", err)
	}

	opts := []quota.Option{
		quota.EnforceResources(config.QuotaPerProjectEnable()),
		quota.WithManager("project", strconv.FormatInt(info.ProjectID, 10)),
		quota.WithAction(quota.AddAction),
		quota.StatusCode(http.StatusCreated),
		quota.OnResources(computeResourcesForManifestCreation),
		quota.MutexKeys(info.MutexKey("size")),
		quota.OnFulfilled(func(http.ResponseWriter, *http.Request) error {
			// manifest created, sync manifest itself as blob to blob and project_blob table
			blobInfo, err := parseBlobInfoFromManifest(req)
			if err != nil {
				return err
			}

			if err := syncBlobInfoToProject(blobInfo); err != nil {
				return err
			}

			// sync blobs from manifest which are not in project to project_blob table
			blobs, err := info.GetBlobsNotInProject()
			if err != nil {
				return err
			}

			_, err = dao.AddBlobsToProject(info.ProjectID, blobs...)

			return err
		}),
	}

	return quota.New(opts...), nil
}

// deleteManifestBuilder interceptor builder for the request DELETE /v2/<name>/manifests/<reference>
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

	blobs, err := dao.GetBlobsByArtifact(info.Digest)
	if err != nil {
		return nil, fmt.Errorf("failed to query blobs of %s, error: %v", info.Digest, err)
	}

	mutexKeys := []string{info.MutexKey("size")}
	for _, blob := range blobs {
		mutexKeys = append(mutexKeys, info.BlobMutexKey(blob))
	}

	opts := []quota.Option{
		quota.EnforceResources(config.QuotaPerProjectEnable()),
		quota.WithManager("project", strconv.FormatInt(info.ProjectID, 10)),
		quota.WithAction(quota.SubtractAction),
		quota.StatusCode(http.StatusAccepted),
		quota.OnResources(computeResourcesForManifestDeletion),
		quota.MutexKeys(mutexKeys...),
		quota.OnFulfilled(func(http.ResponseWriter, *http.Request) error {
			blobs := info.ExclusiveBlobs
			return dao.RemoveBlobsFromProject(info.ProjectID, blobs...)
		}),
	}

	return quota.New(opts...), nil
}
