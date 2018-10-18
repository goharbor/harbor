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

package pool

import "github.com/goharbor/harbor/src/jobservice/models"

// Interface for worker pool.
// More like a driver to transparent the lower queue.
type Interface interface {
	// Start to serve
	//
	// Return:
	//  error if failed to start
	Start() error

	// Register job to the pool.
	//
	// name	string     : job name for referring
	// job	interface{}: job handler which must implement the job.Interface.
	//
	// Return:
	//  error if failed to register
	RegisterJob(name string, job interface{}) error

	// Register multiple jobs.
	//
	// jobs	map[string]interface{}: job map, key is job name and value is job handler.
	//
	// Return:
	//  error if failed to register
	RegisterJobs(jobs map[string]interface{}) error

	// Enqueue job
	//
	// jobName string           : the name of enqueuing job
	// params models.Parameters : parameters of enqueuing job
	// isUnique bool            : specify if duplicated job will be discarded
	//
	// Returns:
	//  models.JobStats: the stats of enqueuing job if succeed
	//  error          : if failed to enqueue
	Enqueue(jobName string, params models.Parameters, isUnique bool) (models.JobStats, error)

	// Schedule job to run after the specified interval (seconds).
	//
	// jobName string           : the name of enqueuing job
	// runAfterSeconds uint64   : the waiting interval with seconds
	// params models.Parameters : parameters of enqueuing job
	// isUnique bool            : specify if duplicated job will be discarded
	//
	// Returns:
	//  models.JobStats: the stats of enqueuing job if succeed
	//  error          : if failed to enqueue
	Schedule(jobName string, params models.Parameters, runAfterSeconds uint64, isUnique bool) (models.JobStats, error)

	// Schedule the job periodically running.
	//
	// jobName string           : the name of enqueuing job
	// params models.Parameters : parameters of enqueuing job
	// cronSetting string       : the periodic duration with cron style like '0 * * * * *'
	//
	// Returns:
	//  models.JobStats: the stats of enqueuing job if succeed
	//  error          : if failed to enqueue
	PeriodicallyEnqueue(jobName string, params models.Parameters, cronSetting string) (models.JobStats, error)

	// Return the status info of the pool.
	//
	// Returns:
	//  models.JobPoolStats : the stats info of all running pools
	//  error               :  failed to check
	Stats() (models.JobPoolStats, error)

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

	ValidateJobParameters(jobType interface{}, params map[string]interface{}) error

	// Get the stats of the specified job
	//
	// jobID string : ID of the enqueued job
	//
	// Returns:
	//  models.JobStats : job stats data
	//  error           : error returned if meet any problems
	GetJobStats(jobID string) (models.JobStats, error)

	// Stop the job
	//
	// jobID string : ID of the enqueued job
	//
	// Return:
	//  error           : error returned if meet any problems
	StopJob(jobID string) error

	// Cancel the job
	//
	// jobID string : ID of the enqueued job
	//
	// Return:
	//  error           : error returned if meet any problems
	CancelJob(jobID string) error

	// Retry the job
	//
	// jobID string : ID of the enqueued job
	//
	// Return:
	//  error           : error returned if meet any problems
	RetryJob(jobID string) error

	// Register hook
	//
	// jobID string   : ID of job
	// hookURL string : the hook url
	//
	// Return:
	//  error        : error returned if meet any problems
	RegisterHook(jobID string, hookURL string) error
}
