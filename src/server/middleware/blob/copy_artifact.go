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

		ctx := r.Context()

		logger := log.G(ctx).WithFields(log.Fields{"middleware": "blob"})

		query := r.URL.Query()
		from := query.Get("from")
		repository, reference, _ := distribution.ParseRef(from)

		art, err := artifactController.GetByReference(ctx, repository, reference, nil)
		if ierror.IsNotFoundErr(err) {
			// artifact not found, discontinue the API request
			return ierror.BadRequestError(nil).WithMessage("artifact %s not found", from)
		} else if err != nil {
			logger.Errorf("get artifact %s failed, error: %v", from, err)
			return err
		}

		projectName := util.ParseProjectName(r)
		project, err := projectController.GetByName(ctx, projectName)
		if err != nil {
			logger.Errorf("get project %s failed, error: %v", projectName, err)
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
			logger.Errorf("walk the artifact %s failed, error: %v", art.Digest, err)
			return err
		}

		allBlobs, err := blobController.List(ctx, blob.ListParams{ArtifactDigests: artifactDigests})
		if err != nil {
			logger.Errorf("get blobs for artifacts %s failed, error: %v", strings.Join(artifactDigests, ", "), err)
			return err
		}

		blobs, err := blobController.FindMissingAssociationsForProject(ctx, project.ProjectID, allBlobs)
		if err != nil {
			logger.Errorf("find missing blobs for project %d failed, error: %v", project.ProjectID, err)
			return err
		}

		for _, blob := range blobs {
			if err := blobController.AssociateWithProjectByID(ctx, blob.ID, project.ProjectID); err != nil {
				logger.Errorf("associate blob %s with project %d failed, error: %v", blob.Digest, project.ProjectID, err)
				return err
			}
		}

		return nil
	})
}
