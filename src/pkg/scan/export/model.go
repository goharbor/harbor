package export

import (
	"encoding/json"
	"errors"
	"time"
)

// Data models a single row of the exported scan vulnerability data

type Data struct {
	ID             int64  `orm:"column(result_row_id)" csv:"RowId"`
	ProjectName    string `orm:"column(project_name)" csv:"Project"`
	ProjectOwner   string `orm:"column(project_owner)" csv:"Owner"`
	ScannerName    string `orm:"column(scanner_name)" csv:"Scanner"`
	Repository     string `orm:"column(repository_name)" csv:"Repository"`
	ArtifactDigest string `orm:"column(artifact_digest)" csv:"Artifact Digest"`
	CVEId          string `orm:"column(cve_id)" csv:"CVE"`
	Package        string `orm:"column(package)" csv:"Package"`
	Severity       string `orm:"column(severity)" csv:"Severity"`
	CVSSScoreV3    string `orm:"column(cvss_score_v3)" csv:"CVSS V3 Score"`
	CVSSScoreV2    string `orm:"column(cvss_score_v2)" csv:"CVSS V2 Score"`
	CVSSVectorV3   string `orm:"column(cvss_vector_v3)" csv:"CVSS V3 Vector"`
	CVSSVectorV2   string `orm:"column(cvss_vector_v2)" csv:"CVSS V2 Vector"`
	CWEIds         string `orm:"column(cwe_ids)" csv:"CWE Ids"`
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
