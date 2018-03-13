// Copyright 2018 The Harbor Authors. All rights reserved.

package models

import (
	"time"
)

//Parameters for job execution.
type Parameters map[string]interface{}

//JobRequest is the request of launching a job.
type JobRequest struct {
	Job *JobData `json:"job"`
}

//JobData keeps the basic info.
type JobData struct {
	Name       string       `json:"name"`
	Parameters Parameters   `json:"parameters"`
	Metadata   *JobMetadata `json:"metadata"`
}

//JobMetadata stores the metadata of job.
type JobMetadata struct {
	JobKind       string `json:"kind"`
	ScheduleDelay uint64 `json:"schedule_delay,omitempty"`
	Cron          string `json:"cron_spec,omitempty"`
	IsUnique      bool   `json:"unique"`
}

//JobStats keeps the result of job launching.
type JobStats struct {
	Stats *JobStatData `json:"job"`
}

//JobStatData keeps the stats of job
type JobStatData struct {
	JobID       string    `json:"id"`
	Status      string    `json:"status"`
	JobName     string    `json:"name"`
	RefLink     string    `json:"ref_link,omitempty"`
	EnqueueTime time.Time `json:"enqueue_time"`
	UpdateTime  time.Time `json:"update_time"`
	RunAt       time.Time `json:"run_at,omitempty"`
}

//JobPoolStats represent the healthy and status of the job service.
type JobPoolStats struct{}
