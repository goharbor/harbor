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

package all

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/q"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/pkg/errors"
)

const (
	// The max number of the goroutines to retrieve the tags
	maxProcessors = 25
	// Job parameter key for the admin job ID
	jobParamAJID = "admin_job_id"
)

// Job query the DB and Registry for all image and tags,
// then call Harbor's API to scan each of them.
type Job struct{}

// MaxFails implements the interface in job/Interface
func (sa *Job) MaxFails() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (sa *Job) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (sa *Job) Validate(params job.Parameters) error {
	_, err := parseAJID(params)
	if err != nil {
		return errors.Wrap(err, "job validation: scan all job")
	}

	return nil
}

// Run implements the interface in job/Interface
func (sa *Job) Run(ctx job.Context, params job.Parameters) error {
	logger := ctx.GetLogger()
	logger.Info("Scanning all the images in the registry")

	// No need to check error any more as it has been checked in job validation.
	requester, _ := parseAJID(params)

	// List all the repositories of registry
	// TODO: REPLACE DAO WITH CORRESPONDING MANAGER OR CTL
	repos, err := dao.GetRepositories()
	if err != nil {
		err = errors.Wrap(err, "list repositories : scan all job")
		logger.Error(err)
		return err
	}
	logger.Infof("Found %d repositories", len(repos))

	// Initialize tokens
	tokens := make(chan bool, maxProcessors)
	for i := 0; i < maxProcessors; i++ {
		// Assign tokens at first
		tokens <- true
	}

	// Get the tags under the repository
	for _, r := range repos {
		// Get token first
		<-tokens

		go func(repo *models.RepoRecord) {
			defer func() {
				// Return the token when process ending
				tokens <- true
			}()

			logger.Infof("Scan artifacts under repository: %s", repo.Name)

			// Query artifacts under the repository
			query := &q.Query{
				Keywords: make(map[string]interface{}),
			}
			query.Keywords["repo"] = repo.Name

			al, err := art.DefaultController.List(query)
			if err != nil {
				logger.Errorf("Failed to get tags for repo: %s, error: %v", repo.Name, err)
				return
			}

			if len(al) > 0 {
				// Check in the data
				arts := make([]*v1.Artifact, 0)

				for _, a := range al {
					artf := &v1.Artifact{
						NamespaceID: repo.ProjectID,
						Repository:  repo.Name,
						Tag:         a.Tag,
						Digest:      a.Digest,
						MimeType:    v1.MimeTypeDockerArtifact, // default
					}

					arts = append(arts, artf)
				}

				logger.Infof("Found %d artifacts under repository %s", len(arts), repo.Name)

				ck := &CheckInData{
					Artifacts: arts,
					Requester: requester,
				}

				jsn, err := ck.ToJSON()
				if err != nil {
					logger.Error(errors.Wrap(err, "scan all job"))
					return
				}

				if err := ctx.Checkin(jsn); err != nil {
					logger.Error(errors.Wrap(err, "check in data: scan all job"))
				}

				logger.Infof("Check in scanning artifacts for repository: %s", repo.Name)
				// Debug more
				logger.Debugf("Check in: %s\n", jsn)
			} else {
				logger.Infof("No scanning artifacts found under repository: %s", repo.Name)
			}
		}(r)
	}

	return nil
}

func parseAJID(params job.Parameters) (string, error) {
	if len(params) > 0 {
		if v, ok := params[jobParamAJID]; ok {
			if id, y := v.(string); y {
				return id, nil
			}
		}
	}

	return "", errors.Errorf("missing required job parameter: %s", jobParamAJID)
}
