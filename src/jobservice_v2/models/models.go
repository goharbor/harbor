// Copyright 2018 The Harbor Authors. All rights reserved.

package models

//JobRequest is the request of launching a job.
type JobRequest struct{}

//JobStats keeps the result of job launching.
type JobStats struct {
	JobID string `json:"job_id"`
}

//JobServiceStats represent the healthy and status of the job service.
type JobServiceStats struct{}
