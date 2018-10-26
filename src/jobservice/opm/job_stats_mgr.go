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

package opm

import "github.com/goharbor/harbor/src/jobservice/models"

// JobStatsManager defines the methods to handle stats of job.
type JobStatsManager interface {
	// Start to serve
	Start()

	// Shutdown the manager
	Shutdown()

	// Save the job stats
	// Async method to retry and improve performance
	//
	// jobStats models.JobStats : the job stats to be saved
	Save(jobStats models.JobStats)

	// Get the job stats from backend store
	// Sync method as we need the data
	//
	// Returns:
	//  models.JobStats : job stats data
	//  error           : error if meet any problems
	Retrieve(jobID string) (models.JobStats, error)

	// Update the properties of the job stats
	//
	// jobID string                  : ID of the being retried job
	// fieldAndValues ...interface{} : One or more properties being updated
	//
	// Returns:
	//  error if update failed
	Update(jobID string, fieldAndValues ...interface{}) error

	// SetJobStatus will mark the status of job to the specified one
	// Async method to retry
	SetJobStatus(jobID string, status string)

	// Send command fro the specified job
	//
	// jobID string   : ID of the being retried job
	// command string : the command applied to the job like stop/cancel
	//
	// Returns:
	//  error if it was not successfully sent
	SendCommand(jobID string, command string) error

	// CtlCommand checks if control command is fired for the specified job.
	//
	// jobID string : ID of the job
	//
	// Returns:
	//  the command if it was fired
	//  error if it was not fired yet to meet some other problems
	CtlCommand(jobID string) (string, error)

	// CheckIn message for the specified job like detailed progress info.
	//
	// jobID string   : ID of the job
	// message string : The message being checked in
	//
	CheckIn(jobID string, message string)

	// DieAt marks the failed jobs with the time they put into dead queue.
	//
	// jobID string   : ID of the job
	// message string : The message being checked in
	//
	DieAt(jobID string, dieAt int64)

	// RegisterHook is used to save the hook url or cache the url in memory.
	//
	// jobID string   : ID of job
	// hookURL string : the hook url being registered
	// isCached bool  :  to indicate if only cache the hook url
	//
	// Returns:
	//  error if meet any problems
	RegisterHook(jobID string, hookURL string, isCached bool) error

	// Get hook returns the web hook url for the specified job if it is registered
	//
	// jobID string   : ID of job
	//
	// Returns:
	//  the web hook url if existing
	//  non-nil error if meet any problems
	GetHook(jobID string) (string, error)

	// Mark the periodic job stats expired
	//
	// jobID string   : ID of job
	//
	// Returns:
	//  error if meet any problems
	ExpirePeriodicJobStats(jobID string) error

	// Persist the links between upstream job and the executions.
	//
	// upstreamJobID string: ID of the upstream job
	// executions  ...string: IDs of the execution jobs
	//
	// Returns:
	//  error if meet any issues
	AttachExecution(upstreamJobID string, executions ...string) error

	// Get all the executions (IDs) fro the specified upstream Job.
	//
	// upstreamJobID string: ID of the upstream job
	//
	// Returns:
	//  the ID list of the executions if no error occurred
	//  or a non-nil error is returned
	GetExecutions(upstreamJobID string) ([]string, error)
}
