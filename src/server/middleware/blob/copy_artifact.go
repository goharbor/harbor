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

package blob

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/util"
)

func isSuccess(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusBadRequest
}

// CopyArtifactMiddleware middleware to sync the missing associations for the project
func CopyArtifactMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if !isSuccess(statusCode) {
			return nil
		}

		logPrefix := fmt.Sprintf("[middleware][%s][blob]", r.URL.Path)

		query := r.URL.Query()

		from := query.Get("from")
		repository, reference, _ := distribution.ParseRef(from)

		ctx := r.Context()

		art, err := artifactController.GetByReference(ctx, repository, reference, nil)
		if ierror.IsNotFoundErr(err) {
			// artifact not found, discontinue the API request
			return ierror.BadRequestError(nil).WithMessage("artifact %s not found", from)
		} else if err != nil {
			log.Errorf("%s: get artifact %s failed, error: %v", logPrefix, from, err)
			return err
		}

		projectName := util.ParseProjectName(r)
		project, err := projectController.GetByName(ctx, projectName)
		if err != nil {
			log.Errorf("%s: get project %s failed, error: %v", logPrefix, projectName, err)
			return err
		}

		if art.ProjectID == project.ProjectID {
			// copy artifact in same project
			return nil
		}

		var artifactDigests []string
		err = artifactController.Walk(ctx, art, func(a *artifact.Artifact) error {
			artifactDigests = append(artifactDigests, a.Digest)
			return nil
		}, nil)
		if err != nil {
			log.Errorf("%s: walk the artifact %s failed, error: %v", logPrefix, art.Digest, err)
			return err
		}

		allBlobs, err := blobController.List(ctx, blob.ListParams{ArtifactDigests: artifactDigests})
		if err != nil {
			log.Errorf("%s: get blobs for artifacts %s failed, error: %v", logPrefix, strings.Join(artifactDigests, ", "), err)
			return err
		}

		blobs, err := blobController.FindMissingAssociationsForProject(ctx, project.ProjectID, allBlobs)
		if err != nil {
			log.Errorf("%s: find missing blobs for project %d failed, error: %v", logPrefix, project.ProjectID, err)
			return err
		}

		for _, blob := range blobs {
			if err := blobController.AssociateWithProjectByID(ctx, blob.ID, project.ProjectID); err != nil {
				log.Errorf("%s: associate blob %s with project %d failed, error: %v", logPrefix, blob.Digest, project.ProjectID, err)
				return err
			}
		}

		return nil
	})
}
