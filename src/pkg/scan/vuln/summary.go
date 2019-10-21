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
)

// NativeReportSummary is the default supported scan report summary model.
// Generated based on the report with v1.MimeTypeNativeReport mime type.
type NativeReportSummary struct {
	ReportID    string                `json:"report_id"`
	ScanStatus  string                `json:"scan_status"`
	Severity    Severity              `json:"severity"`
	Duration    int64                 `json:"duration"`
	Summary     *VulnerabilitySummary `json:"summary"`
	CVEBypassed []string              `json:"-"`
	StartTime   time.Time             `json:"start_time"`
	EndTime     time.Time             `json:"end_time"`
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
