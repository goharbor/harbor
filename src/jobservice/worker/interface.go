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

package worker

import (
	"github.com/goharbor/harbor/src/jobservice/common/query"
	"github.com/goharbor/harbor/src/jobservice/job"
)

// Interface for worker.
// More like a driver to transparent the lower queue.
type Interface interface {
	// Start to serve
	Start() error

	// Register multiple jobs.
	//
	// jobs	map[string]interface{}: job map, key is job name and value is job handler.
	//
	// Return:
	//  error if failed to register
	RegisterJobs(jobs map[string]interface{}) error

	// Enqueue job
	//
	// jobName string        : the name of enqueuing job
	// params job.Parameters : parameters of enqueuing job
	// isUnique bool         : specify if duplicated job will be discarded
	// webHook string        : the server URL to receive hook events
	//
	// Returns:
	//  *job.Stats : the stats of enqueuing job if succeed
	//  error      : if failed to enqueue
	Enqueue(jobName string, params job.Parameters, isUnique bool, webHook string) (*job.Stats, error)

	// Schedule job to run after the specified interval (seconds).
	//
	// jobName string         : the name of enqueuing job
	// runAfterSeconds uint64 : the waiting interval with seconds
	// params job.Parameters  : parameters of enqueuing job
	// isUnique bool          : specify if duplicated job will be discarded
	// webHook string        : the server URL to receive hook events
	//
	// Returns:
	//  *job.Stats: the stats of enqueuing job if succeed
	//  error          : if failed to enqueue
	Schedule(jobName string, params job.Parameters, runAfterSeconds uint64, isUnique bool, webHook string) (*job.Stats, error)

	// Schedule the job periodically running.
	//
	// jobName string        : the name of enqueuing job
	// params job.Parameters : parameters of enqueuing job
	// cronSetting string    : the periodic duration with cron style like '0 * * * * *'
	// isUnique bool         : specify if duplicated job will be discarded
	// webHook string        : the server URL to receive hook events
	//
	// Returns:
	//  models.JobStats: the stats of enqueuing job if succeed
	//  error          : if failed to enqueue
	PeriodicallyEnqueue(jobName string, params job.Parameters, cronSetting string, isUnique bool, webHook string) (*job.Stats, error)

	// Return the status info of the worker.
	//
	// Returns:
	//  *Stats : the stats info of all running pools
	//  error  :  failed to check
	Stats() (*Stats, error)

	// Check if the job has been already registered.
	//
	// name string : name of job
	//
	// Returns:
	// interface{} : the job type of the known job if it's existing
	// bool        : if the known job requires parameters
	IsKnownJob(name string) (interface{}, bool)

	// Validate the parameters of the known job
	//
	// jobType interface{}            : type of known job
	// params map[string]interface{} : parameters of known job
	//
	// Return:
	//  error if parameters are not valid

	ValidateJobParameters(jobType interface{}, params job.Parameters) error

	// Stop the job
	//
	// jobID string : ID of the enqueued job
	//
	// Return:
	//  error           : error returned if meet any problems
	StopJob(jobID string) error

	// Retry the job
	//
	// jobID string : ID of the enqueued job
	//
	// Return:
	//  error           : error returned if meet any problems
	RetryJob(jobID string) error

	// Get the scheduled jobs by page
	// The page number in the query will be ignored, default 20 is used. This is the limitation of backend lib.
	// The total number is also returned.
	//
	// query *query.Parameter : query parameters
	//
	// Return:
	//   []*job.Stats : list of scheduled jobs
	//   int          : the total number of scheduled jobs
	//   error        : non nil error if meet any issues
	ScheduledJobs(query *query.Parameter) ([]*job.Stats, int64, error)
}
