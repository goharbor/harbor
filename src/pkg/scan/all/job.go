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
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
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

// MaxCurrency is implementation of same method in Interface.
func (sa *Job) MaxCurrency() uint {
	return 0
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

	if err := ctx.Checkin(requester); err != nil {
		logger.Error(errors.Wrap(err, "check in data: scan all job"))
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
