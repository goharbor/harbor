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

package export

import (
	"encoding/json"
	"errors"
	"time"
)

// Data models a single row of the exported scan vulnerability data

type Data struct {
	Repository     string `orm:"column(repository_name)" csv:"Repository"`
	ArtifactDigest string `orm:"column(artifact_digest)" csv:"Artifact Digest"`
	CVEId          string `orm:"column(cve_id)" csv:"CVE"`
	Package        string `orm:"column(package)" csv:"Package"`
	Version        string `orm:"column(package_version)" csv:"Current Version"`
	FixVersion     string `orm:"column(fixed_version)" csv:"Fixed in version"`
	Severity       string `orm:"column(severity)" csv:"Severity"`
	CWEIds         string `orm:"column(cwe_ids)" csv:"CWE Ids"`
	AdditionalData string `orm:"column(vendor_attributes)" csv:"Additional Data"`
	ScannerName    string `orm:"column(scanner_name)" csv:"Scanner"`
}

// Request encapsulates the filters to be provided when exporting the data for a scan.
type Request struct {

	// UserID contains the database identity of the user initiating the export request
	UserID int

	// UserName contains the name of the user initiating the export request
	UserName string

	// JobName contains the name of the job as specified by the external client.
	JobName string

	// cve ids
	CVEIds string

	// A list of one or more labels for which to export the scan data, defaults to all if empty
	Labels []int64

	// A list of one or more projects for which to export the scan data, defaults to all if empty
	Projects []int64

	// A list of repositories for which to export the scan data, defaults to all if empty
	Repositories string

	// A list of tags for which to export the scan data, defaults to all if empty
	Tags string
}

// FromJSON parses robot from json data
func (c *Request) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), c)
}

// ToJSON marshals Robot to JSON data
func (c *Request) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Execution provides details about the running status of a scan data export job
type Execution struct {
	// ID of the execution
	ID int64
	// UserID triggering the execution
	UserID int64
	// ProjectIDs contains projects ids
	ProjectIDs []int64
	// Status provides the status of the execution
	Status string
	// StatusMessage contains the human-readable status message for the execution
	StatusMessage string
	// Trigger indicates the mode of trigger for the job execution
	Trigger string
	// StartTime contains the start time instant of the execution
	StartTime time.Time
	// EndTime contains the end time instant of the execution
	EndTime time.Time
	// ExportDataDigest contains the SHA256 hash of the exported scan data artifact
	ExportDataDigest string
	// Name of the job as specified during the export task invocation
	JobName string
	// Name of the user triggering the job
	UserName string
	// FilePresent is true if file artifact is actually present, false otherwise
	FilePresent bool
}

type Task struct {
	// ID of the scan data export task
	ID int64
	// Job Id corresponding to the task
	JobID string
	// Status of the current task execution
	Status string
	// Status message for the current task execution
	StatusMessage string
}

type TriggerParam struct {
	TimeWindowMinutes int
	PageSize          int
}
