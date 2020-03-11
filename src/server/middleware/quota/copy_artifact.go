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

package quota

import (
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/types"
)

// CopyArtifactMiddleware middleware to request count and storage resources for copy artifact API
func CopyArtifactMiddleware() func(http.Handler) http.Handler {
	return RequestMiddleware(RequestConfig{
		ReferenceObject: projectReferenceObject,
		Resources:       copyArtifactResources,
	})
}

func parseRepositoryName(p string) string {
	parts := strings.Split(strings.TrimSuffix(path.Clean(p), "/"), "/repositories/")
	if len(parts) != 2 {
		return ""
	}

	return strings.TrimSuffix(parts[1], "/artifacts")
}

func copyArtifactResources(r *http.Request, reference, referenceID string) (types.ResourceList, error) {
	logPrefix := fmt.Sprintf("[middleware][%s][quota]", r.URL.Path)

	query := r.URL.Query()

	from := query.Get("from")
	if from == "" {
		// miss the from parameter, skip to request the resources
		return nil, nil
	}

	repository, reference, err := distribution.ParseRef(from)
	if err != nil {
		// bad from parameter, skip to request the resources
		log.Errorf("%s: parse from parameter failed, error: %v", logPrefix, err)
		return nil, nil
	}

	ctx := r.Context()

	art, err := artifactController.GetByReference(ctx, repository, reference, nil)
	if ierror.IsNotFoundErr(err) {
		// artifact not found, discontinue the API request
		return nil, ierror.BadRequestError(nil).WithMessage("artifact %s not found", from)
	} else if err != nil {
		log.Errorf("%s: get artifact %s failed, error: %v", logPrefix, from, err)
		return nil, err
	}

	projectID, _ := strconv.ParseInt(referenceID, 10, 64)
	repositoryName := parseRepositoryName(r.URL.EscapedPath())

	if art.ProjectID == projectID && art.RepositoryName == repositoryName {
		return nil, nil
	}

	var artifactDigests []string
	err = artifactController.Walk(ctx, art, func(a *artifact.Artifact) error {
		artifactDigests = append(artifactDigests, a.Digest)
		return nil
	}, nil)
	if err != nil {
		log.Errorf("%s: walk the artifact %s failed, error: %v", logPrefix, art.Digest, err)
		return nil, err
	}

	// HACK: base=* in KeyWords to filter all artifacts
	kw := q.KeyWords{"project_id": projectID, "digest__in": artifactDigests, "repository_name": repositoryName, "base": "*"}
	count, err := artifactController.Count(ctx, q.New(kw))
	if err != nil {
		return nil, err
	}

	copyCount := int64(len(artifactDigests)) - count

	if copyCount == 0 {
		// artifacts  already exist in the repository of the project
		return nil, nil
	}

	allBlobs, err := blobController.List(ctx, blob.ListParams{ArtifactDigests: artifactDigests})
	if err != nil {
		log.Errorf("%s: get blobs for artifacts %s failed, error: %v", logPrefix, strings.Join(artifactDigests, ", "), err)
		return nil, err
	}

	blobs, err := blobController.FindMissingAssociationsForProject(ctx, projectID, allBlobs)
	if err != nil {
		log.Errorf("%s: find missing blobs for project %d failed, error: %v", logPrefix, projectID, err)
		return nil, err
	}

	var size int64
	for _, blob := range blobs {
		if !blob.IsForeignLayer() {
			size += blob.Size
		}
	}

	return types.ResourceList{types.ResourceCount: copyCount, types.ResourceStorage: size}, nil
}
