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
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/api/blob"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/types"
)

// PostInitiateBlobUploadMiddleware middleware to add blob to project after mount blob success
func PostInitiateBlobUploadMiddleware() func(http.Handler) http.Handler {
	return RequestMiddleware(RequestConfig{
		ReferenceObject: projectReferenceObject,
		Resources:       postInitiateBlobUploadResources,
	})
}

func postInitiateBlobUploadResources(r *http.Request, reference, referenceID string) (types.ResourceList, error) {
	logPrefix := fmt.Sprintf("[middleware][%s][quota]", r.URL.Path)

	query := r.URL.Query()

	mount := query.Get("mount")
	if mount == "" {
		// it is not mount blob http request, skip to request the resources
		return nil, nil
	}

	ctx := r.Context()

	blb, err := blobController.Get(ctx, mount)
	if ierror.IsNotFoundErr(err) {
		// mount blob not found, skip to request the resources
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	projectID, _ := strconv.ParseInt(referenceID, 10, 64)

	exist, err := blobController.Exist(ctx, blb.Digest, blob.IsAssociatedWithProject(projectID))
	if err != nil {
		log.Errorf("%s: checking blob %s is associated with project %d failed, error: %v", logPrefix, blb.Digest, projectID, err)
		return nil, err
	}

	if exist {
		return nil, nil
	}

	return types.ResourceList{types.ResourceStorage: blb.Size}, nil
}
