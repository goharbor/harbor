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
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
)

const (
	// reportIDSeparator the separator of the ReportID in the summary when its merged by multi summaries
	reportIDSeparator = "|"
)

// ParseReportIDs returns report ids from s
func ParseReportIDs(s string) []string {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		data = []byte(s)
	}

	return strings.Split(string(data), reportIDSeparator)
}

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
	src = append(src, []byte(reportIDSeparator+r2)...)

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
