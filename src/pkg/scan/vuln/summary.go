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
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	models2 "github.com/goharbor/harbor/src/pkg/allowlist/models"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// NativeReportSummary is the default supported scan report summary model.
// Generated based on the report with v1.MimeTypeNativeReport or the v1.MimeTypeGenericVulnerabilityReport mime type.
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

	TotalCount            int                    `json:"-"`
	CompleteCount         int                    `json:"-"`
	VulnerabilityItemList *VulnerabilityItemList `json:"-"`
	CVESet                models2.CVESet         `json:"-"`
}

// UpdateSeveritySummaryAndByPassed update the Severity, Summary and CVEBypassed of the sum from l and s
func (sum *NativeReportSummary) UpdateSeveritySummaryAndByPassed(l *VulnerabilityItemList, s models2.CVESet) {
	sum.VulnerabilityItemList = l
	sum.CVESet = s

	if l == nil {
		return
	}

	var severity Severity
	severity, sum.Summary, sum.CVEBypassed = l.GetSeveritySummaryAndByPassed(s)

	if len(s) > 0 {
		// Override the overall severity of the filtered list if needed.
		sum.Severity = severity
	}
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
	// choose the scanner from the newer summary
	// because the endtime of the summary is from the newer summary
	if sum.StartTime.After(another.StartTime) {
		r.Scanner = sum.Scanner
	} else {
		r.Scanner = another.Scanner
	}
	r.TotalCount = sum.TotalCount + another.TotalCount
	r.CompleteCount = sum.CompleteCount + another.CompleteCount
	r.CompletePercent = r.CompleteCount * 100 / r.TotalCount
	r.ReportID = mergeReportID(sum.ReportID, another.ReportID)
	r.Severity = mergeSeverity(sum.Severity, another.Severity)
	r.ScanStatus = mergeScanStatus(sum.ScanStatus, another.ScanStatus)

	r.UpdateSeveritySummaryAndByPassed(
		NewVulnerabilityItemList(sum.VulnerabilityItemList, another.VulnerabilityItemList),
		models2.NewCVESet(sum.CVESet, another.CVESet),
	)

	return r
}

// VulnerabilitySummary contains the total number of the found vulnerabilities number
// and numbers of each severity level.
type VulnerabilitySummary struct {
	Total   int             `json:"total"`
	Fixable int             `json:"fixable"`
	Summary SeveritySummary `json:"summary"`
}

// SeveritySummary ...
type SeveritySummary map[Severity]int
