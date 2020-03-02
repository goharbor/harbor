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

package job

import "fmt"

const (
	// PendingStatus   : job status pending
	PendingStatus Status = "Pending"
	// RunningStatus   : job status running
	RunningStatus Status = "Running"
	// StoppedStatus   : job status stopped
	StoppedStatus Status = "Stopped"
	// ErrorStatus     : job status error
	ErrorStatus Status = "Error"
	// SuccessStatus   : job status success
	SuccessStatus Status = "Success"
	// ScheduledStatus : job status scheduled
	ScheduledStatus Status = "Scheduled"
)

// Status of job
type Status string

// Validate the status
// If it's valid, then return nil error
// otherwise an non nil error is returned
func (s Status) Validate() error {
	if s.Code() == -1 {
		return fmt.Errorf("%s is not valid job status", s)
	}

	return nil
}

// Code of job status
func (s Status) Code() int {
	switch s {
	case "Pending":
		return 0
	case "Scheduled":
		return 1
	case "Running":
		return 2
	// All final status share the same code
	// Each job will have only 1 final status
	case "Stopped":
		return 3
	case "Error":
		return 3
	case "Success":
		return 3
	default:
	}

	return -1
}

// Compare the two job status
// if < 0, s before another status
// if == 0, same status
// if > 0, s after another status
func (s Status) Compare(another Status) int {
	return s.Code() - another.Code()
}

// String returns the raw string value of the status
func (s Status) String() string {
	return string(s)
}

// Final returns if the status is final status
// e.g: "Stopped", "Error" or "Success"
func (s Status) Final() bool {
	return s.Code() == 3
}
