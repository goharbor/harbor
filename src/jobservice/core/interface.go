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

// Package core provides the main job operation interface and components.
package core

import (
	"github.com/goharbor/harbor/src/jobservice/common/query"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/worker"
)

// Interface defines the related main methods of job operation.
type Interface interface {
	// LaunchJob is used to handle the job submission request.
	//
	// req	*job.Request    : Job request contains related required information of queuing job.
	//
	// Returns:
	//	job.Stats : Job status info with ID and self link returned if job is successfully launched.
	//  error     : Error returned if failed to launch the specified job.
	LaunchJob(req *job.Request) (*job.Stats, error)

	// GetJob is used to handle the job stats query request.
	//
	// jobID	string: ID of job.
	//
	// Returns:
	//	*job.Stats : Job status info if job exists.
	//  error      : Error returned if failed to get the specified job.
	GetJob(jobID string) (*job.Stats, error)

	// StopJob is used to handle the job stopping request.
	//
	// jobID	string: ID of job.
	//
	// Return:
	//  error   : Error returned if failed to stop the specified job.
	StopJob(jobID string) error

	// RetryJob is used to handle the job retrying request.
	//
	// jobID	string        : ID of job.
	//
	// Return:
	//  error   : Error returned if failed to retry the specified job.
	RetryJob(jobID string) error

	// CheckStatus is used to handle the job service healthy status checking request.
	CheckStatus() (*worker.Stats, error)

	// GetJobLogData is used to return the log text data for the specified job if exists
	GetJobLogData(jobID string) ([]byte, error)

	// Get the periodic executions for the specified periodic job.
	// Pagination by query is supported.
	// The total number is also returned.
	GetPeriodicExecutions(periodicJobID string, query *query.Parameter) ([]*job.Stats, int64, error)

	// Get the scheduled jobs by page
	// The page number in the query will be ignored, default 20 is used. This is the limitation of backend lib.
	// The total number is also returned.
	ScheduledJobs(query *query.Parameter) ([]*job.Stats, int64, error)
}
