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

package quota

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/api/blob"
	"github.com/goharbor/harbor/src/api/event/metadata"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/internal"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/types"
)

// PutManifestMiddleware middleware to request count and storage resources for the project
func PutManifestMiddleware() func(http.Handler) http.Handler {
	return RequestMiddleware(RequestConfig{
		ReferenceObject:   projectReferenceObject,
		Resources:         putManifestResources,
		ResourcesExceeded: putManifestResourcesExceeded,
	})
}

var (
	unmarshalManifest = func(r *http.Request) (distribution.Manifest, distribution.Descriptor, error) {
		internal.NopCloseRequest(r)

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return nil, distribution.Descriptor{}, err
		}

		contentType := r.Header.Get("Content-Type")
		return distribution.UnmarshalManifest(contentType, body)
	}
)

func putManifestResources(r *http.Request, reference, referenceID string) (types.ResourceList, error) {
	logger := log.G(r.Context()).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

	projectID, _ := strconv.ParseInt(referenceID, 10, 64)

	manifest, descriptor, err := unmarshalManifest(r)
	if err != nil {
		logger.Errorf("unmarshal manifest failed, error: %v", err)
		return nil, err
	}

	exist, err := blobController.Exist(r.Context(), descriptor.Digest.String(), blob.IsAssociatedWithProject(projectID))
	if err != nil {
		logger.Errorf("check manifest %s is associated with project failed, error: %v", descriptor.Digest.String(), err)
		return nil, err
	}

	if exist {
		return nil, nil
	}

	size := descriptor.Size

	var blobs []*models.Blob
	for _, reference := range manifest.References() {
		blobs = append(blobs, &models.Blob{
			Digest:      reference.Digest.String(),
			Size:        reference.Size,
			ContentType: reference.MediaType,
		})
	}

	missing, err := blobController.FindMissingAssociationsForProject(r.Context(), projectID, blobs)
	if err != nil {
		return nil, err
	}

	for _, m := range missing {
		if !m.IsForeignLayer() {
			size += m.Size
		}
	}

	return types.ResourceList{types.ResourceCount: 1, types.ResourceStorage: size}, nil
}

func putManifestResourcesExceeded(r *http.Request, reference, referenceID string, exceededErr error) event.Metadata {
	ctx := r.Context()

	logger := log.G(ctx).WithFields(log.Fields{"middleware": "quota", "action": "request", "url": r.URL.Path})

	_, descriptor, err := unmarshalManifest(r)
	if err != nil {
		logger.Errorf("unmarshal manifest failed, error: %v", err)
		return nil
	}

	projectID, _ := strconv.ParseInt(referenceID, 10, 64)
	project, err := projectController.Get(ctx, projectID)
	if err != nil {
		logger.Errorf("get project %d failed, error: %v", projectID, err)

		return nil
	}

	path := r.URL.EscapedPath()

	var tag string
	if ref := distribution.ParseReference(path); !distribution.IsDigest(ref) {
		tag = ref
	}

	return &metadata.QuotaMetaData{
		Project:  project,
		Tag:      tag,
		Digest:   descriptor.Digest.String(),
		RepoName: distribution.ParseName(path),
		Level:    1,
		Msg:      exceededErr.Error(),
		OccurAt:  time.Now(),
	}
}
