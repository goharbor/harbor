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

package scan

import (
	"context"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/pkg/errors"
)

// HandleCheckIn handles the check in data of the scan all job
func HandleCheckIn(ctx context.Context, checkIn string) {
	if len(checkIn) == 0 {
		// Nothing to handle, directly return
		return
	}

	batchSize := 50
	for repo := range fetchRepositories(ctx, batchSize) {
		for artifact := range fetchArtifacts(ctx, repo.RepositoryID, batchSize) {
			if err := DefaultController.Scan(ctx, artifact, WithRequester(checkIn)); err != nil {
				// Just logged
				log.Error(errors.Wrap(err, "handle check in"))
			}
		}
	}
}

func fetchArtifacts(ctx context.Context, repositoryID int64, chunkSize int) <-chan *artifact.Artifact {
	ch := make(chan *artifact.Artifact, chunkSize)
	go func() {
		defer close(ch)

		query := &q.Query{
			Keywords: map[string]interface{}{
				"repository_id": repositoryID,
			},
			PageSize:   int64(chunkSize),
			PageNumber: 1,
		}

		for {
			artifacts, err := artifact.Ctl.List(ctx, query, nil)
			if err != nil {
				log.Errorf("[scan all]: list artifacts failed, error: %v", err)
				return
			}

			for _, artifact := range artifacts {
				ch <- artifact
			}

			if len(artifacts) < chunkSize {
				break
			}

			query.PageNumber++
		}

	}()

	return ch
}

func fetchRepositories(ctx context.Context, chunkSize int) <-chan *models.RepoRecord {
	ch := make(chan *models.RepoRecord, chunkSize)
	go func() {
		defer close(ch)

		query := &q.Query{
			PageSize:   int64(chunkSize),
			PageNumber: 1,
		}

		for {
			repositories, err := repository.Ctl.List(ctx, query)
			if err != nil {
				log.Warningf("[scan all]: list repositories failed, error: %v", err)
				break
			}

			for _, repo := range repositories {
				ch <- repo
			}

			if len(repositories) < chunkSize {
				break
			}

			query.PageNumber++
		}
	}()
	return ch
}
