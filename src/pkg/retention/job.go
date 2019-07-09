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

package retention

import (
	"github.com/goharbor/harbor/src/jobservice/job"
)

// Job of running retention process
type Job struct{}

// MaxFails of the job
func (pj *Job) MaxFails() uint {
	return 3
}

// ShouldRetry indicates job can be retried if failed
func (pj *Job) ShouldRetry() bool {
	return true
}

// Validate the parameters
func (pj *Job) Validate(params Parameters) error {
	return nil
}

// Run the job
func (pj *Job) Run(ctx job.Context, params Parameters) error {
	return nil
}
