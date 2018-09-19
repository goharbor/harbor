// Copyright Project Harbor Authors. All rights reserved.

package models

// Parameters for job execution.
type Parameters map[string]interface{}

// JobRequest is the request of launching a job.
type JobRequest struct {
	Job *JobData `json:"job"`
}

// JobData keeps the basic info.
type JobData struct {
	Name       string       `json:"name"`
	Parameters Parameters   `json:"parameters"`
	Metadata   *JobMetadata `json:"metadata"`
	StatusHook string       `json:"status_hook"`
}

// JobMetadata stores the metadata of job.
type JobMetadata struct {
	JobKind       string `json:"kind"`
	ScheduleDelay uint64 `json:"schedule_delay,omitempty"`
	Cron          string `json:"cron_spec,omitempty"`
	IsUnique      bool   `json:"unique"`
}

// JobStats keeps the result of job launching.
type JobStats struct {
	Stats *JobStatData `json:"job"`
}

// JobStatData keeps the stats of job
type JobStatData struct {
	JobID       string `json:"id"`
	Status      string `json:"status"`
	JobName     string `json:"name"`
	JobKind     string `json:"kind"`
	IsUnique    bool   `json:"unique"`
	RefLink     string `json:"ref_link,omitempty"`
	CronSpec    string `json:"cron_spec,omitempty"`
	EnqueueTime int64  `json:"enqueue_time"`
	UpdateTime  int64  `json:"update_time"`
	RunAt       int64  `json:"run_at,omitempty"`
	CheckIn     string `json:"check_in,omitempty"`
	CheckInAt   int64  `json:"check_in_at,omitempty"`
	DieAt       int64  `json:"die_at,omitempty"`
	HookStatus  string `json:"hook_status,omitempty"`
}

// JobPoolStats represents the healthy and status of all the running worker pools.
type JobPoolStats struct {
	Pools []*JobPoolStatsData `json:"worker_pools"`
}

// JobPoolStatsData represent the healthy and status of the worker pool.
type JobPoolStatsData struct {
	WorkerPoolID string   `json:"worker_pool_id"`
	StartedAt    int64    `json:"started_at"`
	HeartbeatAt  int64    `json:"heartbeat_at"`
	JobNames     []string `json:"job_names"`
	Concurrency  uint     `json:"concurrency"`
	Status       string   `json:"status"`
}

// JobActionRequest defines for triggering job action like stop/cancel.
type JobActionRequest struct {
	Action string `json:"action"`
}

// JobStatusChange is designed for reporting the status change via hook.
type JobStatusChange struct {
	JobID    string       `json:"job_id"`
	Status   string       `json:"status"`
	CheckIn  string       `json:"check_in,omitempty"`
	Metadata *JobStatData `json:"metadata,omitempty"`
}

// Message is designed for sub/pub messages
type Message struct {
	Event string
	Data  interface{} // generic format
}
