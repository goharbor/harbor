package models

const (
	// PreheatingImageTypeImage defines the 'image' type of preheating images
	PreheatingImageTypeImage = "image"
	// PreheatingStatusPending means the preheating is waiting for starting
	PreheatingStatusPending = "PENDING"
	// PreheatingStatusRunning means the preheating is ongoing
	PreheatingStatusRunning = "RUNNING"
	// PreheatingStatusSuccess means the preheating is success
	PreheatingStatusSuccess = "SUCCESS"
	// PreheatingStatusFail means the preheating is failed
	PreheatingStatusFail = "FAIL"
)

// Metadata represents the basic info of one working node for the specified provider.
type Metadata struct {
	// Unique ID
	ID int64 `json:"id"`

	// Instance name
	Name string `json:"name"`

	// Description of instance
	Description string `json:"description"`

	// Based on which driver, identified by ID
	Provider string `json:"provider"`

	// The service endpoint of this instance
	Endpoint string `json:"endpoint"`

	// The authentication way supported
	AuthMode string `json:"auth_mode,omitempty"`

	// The auth credential data if exists
	AuthData map[string]string `json:"auth_data,omitempty"`

	// The health status
	Status string `json:"status,omitempty"`

	// Whether the instance is activated or not
	Enabled bool `json:"enabled"`

	// The timestamp of instance setting up
	SetupTimestamp int64 `json:"setup_timestamp,omitempty"`

	// Append more described data if needed
	Extensions map[string]string `json:"extensions,omitempty"`
}

// HistoryRecord represents one record of the image preheating process.
type HistoryRecord struct {
	ID         int64  `json:"id"`
	TaskID     string `json:"task_id"` // mapping to the provider task ID
	Image      string `json:"image"`
	StartTime  string `json:"start_time"`
	FinishTime string `json:"finish_time"`
	Status     string `json:"status"`
	Provider   string `json:"provider"`
	Instance   int64  `json:"instance"`
}
