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

package vuln

import (
	"encoding/base64"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

const (
	// SummaryReportIDSeparator the separator of the ReportID in the summary when its merged by multi summaries
	SummaryReportIDSeparator = "|"
)

// NativeReportSummary is the default supported scan report summary model.
// Generated based on the report with v1.MimeTypeNativeReport mime type.
type NativeReportSummary struct {
	ReportID        string                `json:"report_id"`
	ScanStatus      string                `json:"scan_status"`
	Severity        Severity              `json:"severity"`
	Duration        int64                 `json:"duration"`
	Summary         *VulnerabilitySummary `json:"summary"`
	CVEBypassed     []string              `json:"-"`
	StartTime       time.Time             `json:"start_time"`
	EndTime         time.Time             `json:"end_time"`
	Scanner         *v1.Scanner           `json:"scanner,omitempty"`
	CompletePercent int                   `json:"complete_percent"`

	TotalCount    int `json:"-"`
	CompleteCount int `json:"-"`
}

// IsSuccessStatus returns true when the scan status is success
func (sum *NativeReportSummary) IsSuccessStatus() bool {
	return sum.ScanStatus == job.SuccessStatus.String()
}

// Merge ...
func (sum *NativeReportSummary) Merge(another *NativeReportSummary) *NativeReportSummary {
	r := &NativeReportSummary{}

	r.StartTime = minTime(sum.StartTime, another.StartTime)
	r.EndTime = maxTime(sum.EndTime, another.EndTime)
	r.Duration = r.EndTime.Unix() - r.StartTime.Unix()
	r.Scanner = sum.Scanner
	r.TotalCount = sum.TotalCount + another.TotalCount
	r.CompleteCount = sum.CompleteCount + another.CompleteCount
	r.CompletePercent = r.CompleteCount * 100 / r.TotalCount
	r.ReportID = mergeReportID(sum.ReportID, another.ReportID)
	r.Severity = mergeSeverity(sum.Severity, another.Severity)
	r.ScanStatus = mergeScanStatus(sum.ScanStatus, another.ScanStatus)

	if sum.Summary != nil && another.Summary != nil {
		r.Summary = sum.Summary.Merge(another.Summary)
	} else if sum.Summary != nil {
		r.Summary = sum.Summary
	} else {
		r.Summary = another.Summary
	}

	return r
}

// VulnerabilitySummary contains the total number of the found vulnerabilities number
// and numbers of each severity level.
type VulnerabilitySummary struct {
	Total   int             `json:"total"`
	Fixable int             `json:"fixable"`
	Summary SeveritySummary `json:"summary"`
}

// Merge ...
func (v *VulnerabilitySummary) Merge(a *VulnerabilitySummary) *VulnerabilitySummary {
	r := &VulnerabilitySummary{
		Total:   v.Total + a.Total,
		Fixable: v.Fixable + a.Fixable,
		Summary: SeveritySummary{},
	}

	for k, v := range v.Summary {
		r.Summary[k] = v
	}

	for k, v := range a.Summary {
		if _, ok := r.Summary[k]; ok {
			r.Summary[k] += v
		} else {
			r.Summary[k] = v
		}
	}

	return r
}

// SeveritySummary ...
type SeveritySummary map[Severity]int

func minTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t1
	}

	return t2
}

func maxTime(t1, t2 time.Time) time.Time {
	if t1.Before(t2) {
		return t2
	}

	return t1
}

func mergeReportID(r1, r2 string) string {
	src, err := base64.StdEncoding.DecodeString(r1)
	if err != nil {
		src = []byte(r1)
	}
	src = append(src, []byte(SummaryReportIDSeparator+r2)...)

	return base64.StdEncoding.EncodeToString(src)
}

func mergeSeverity(s1, s2 Severity) Severity {
	severityValue := func(s Severity) int {
		if s.String() == "" {
			return -1
		}

		return s.Code()
	}

	if severityValue(s1) > severityValue(s2) {
		return s1
	}

	return s2
}

func mergeScanStatus(s1, s2 string) string {
	j1, j2 := job.Status(s1), job.Status(s2)

	if j1 == job.RunningStatus || j2 == job.RunningStatus {
		return job.RunningStatus.String()
	} else if j1 == job.SuccessStatus || j2 == job.SuccessStatus {
		return job.SuccessStatus.String()
	}

	if j1.Compare(j2) > 0 {
		return s1
	}

	return s2
}
