package provider

import (
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/provider"
)

const (
	// DriverStatusHealthy represents the healthy status
	DriverStatusHealthy = "Healthy"

	// DriverStatusUnHealthy represents the unhealthy status
	DriverStatusUnHealthy = "Unhealthy"
)

// Driver defines the capabilities one distribution provider should have.
// Includes:
//   Self descriptor
//   Health checking
//   Preheat related : Preheat means transfer the preheating image to the network of distribution provider in advance.
type Driver interface {
	// Self returns the metadata of the driver.
	// The metadata includes: name, icon(optional), maintainers(optional), version and source repo.
	Self() *Metadata

	// Try to get the health status of the driver.
	// If succeed, a non nil status object will be returned;
	// otherwise, a non nil error will be set.
	GetHealth() (*DriverStatus, error)

	// Preheat the specified image
	// If succeed, a non nil result object with preheating task id will be returned;
	// otherwise, a non nil error will be set.
	Preheat(preheatingImage *PreheatImage) (*PreheatingStatus, error)

	// Check the progress of the preheating process.
	// If succeed, a non nil status object with preheating status will be returned;
	// otherwise, a non nil error will be set.
	CheckProgress(taskID string) (*PreheatingStatus, error)
}

// Metadata contains the basic information of the provider.
type Metadata struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Icon        string   `json:"icon,omitempty"`
	Maintainers []string `json:"maintainers,omitempty"`
	Version     string   `json:"version"`
	Source      string   `json:"source,omitempty"`
}

// DriverStatus keeps the health status of driver.
type DriverStatus struct {
	Status string `json:"status"`
}

// PreheatingStatus contains the related results/status of the preheating operation
// from the provider.
type PreheatingStatus struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	StartTime  string `json:"start_time"`
	FinishTime string `json:"finish_time"`
}

// String format of PreheatingStatus
func (ps *PreheatingStatus) String() string {
	t := fmt.Sprintf("Task [%s] is %s", ps.TaskID, strings.ToLower(ps.Status))
	switch ps.Status {
	case provider.PreheatingStatusFail:
		t = fmt.Sprintf("%s with error: %s", t, ps.Error)
	case provider.PreheatingStatusSuccess:
		if len(ps.StartTime) > 0 && len(ps.FinishTime) > 0 {
			if st, err := time.Parse(time.RFC3339, ps.StartTime); err == nil {
				if ft, err := time.Parse(time.RFC3339, ps.FinishTime); err == nil {
					d := ft.Sub(st)
					t = fmt.Sprintf("%s with duration: %s", t, d)
				}
			}
		}
	default:
		t = fmt.Sprintf("%s, start time=%s", t, ps.StartTime)

	}

	return t
}
