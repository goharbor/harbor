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
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/stretchr/testify/assert"
)

func TestMergeNativeReportSummary(t *testing.T) {
	assert := assert.New(t)
	errorStatus := job.ErrorStatus.String()
	runningStatus := job.RunningStatus.String()

	v1 := VulnerabilitySummary{
		Total:   1,
		Fixable: 1,
		Summary: map[Severity]int{Low: 1},
	}

	l := &VulnerabilityItemList{}
	l.Add(&VulnerabilityItem{
		ID:         "cve-id",
		Package:    "openssl-libs",
		Version:    "1:1.1.1g-11.el8",
		Severity:   Low,
		FixVersion: "1:1.1.1g-12.el8_3",
	})

	{
		n1 := NativeReportSummary{
			ScanStatus:            runningStatus,
			Severity:              Low,
			TotalCount:            1,
			Summary:               &v1,
			VulnerabilityItemList: l,
		}

		r := n1.Merge(&NativeReportSummary{
			ScanStatus: errorStatus,
			Severity:   Severity(""),
			TotalCount: 1,
		})

		assert.Equal(runningStatus, r.ScanStatus)
		assert.Equal(Low, r.Severity)
		assert.Equal(v1, *r.Summary)
	}

	{
		n1 := NativeReportSummary{
			ScanStatus: runningStatus,
			Severity:   Severity(""),
			TotalCount: 1,
		}

		r := n1.Merge(&NativeReportSummary{
			ScanStatus: errorStatus,
			Severity:   Severity(""),
			TotalCount: 1,
		})

		assert.Equal(runningStatus, r.ScanStatus)
		assert.Equal(Severity(""), r.Severity)
		assert.Nil(r.Summary)
	}

	{
		n1 := &NativeReportSummary{
			ScanStatus: errorStatus,
			Severity:   Severity(""),
			TotalCount: 1,
		}

		r := n1.Merge(&NativeReportSummary{
			ScanStatus:            runningStatus,
			Severity:              Low,
			TotalCount:            1,
			Summary:               &v1,
			VulnerabilityItemList: l,
		})

		assert.Equal(runningStatus, r.ScanStatus)
		assert.Equal(Low, r.Severity)
		assert.Equal(v1, *r.Summary)
	}

	{
		n1 := &NativeReportSummary{
			ScanStatus:            runningStatus,
			Severity:              Low,
			TotalCount:            1,
			Summary:               &v1,
			VulnerabilityItemList: l,
		}

		r := n1.Merge(&NativeReportSummary{
			ScanStatus:            runningStatus,
			Severity:              Low,
			TotalCount:            1,
			Summary:               &v1,
			VulnerabilityItemList: l,
		})

		assert.Equal(runningStatus, r.ScanStatus)
		assert.Equal(Low, r.Severity)
		assert.Equal(v1, *r.Summary)
	}
}
