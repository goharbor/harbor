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
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/stretchr/testify/assert"
)

func Test_mergeReportID(t *testing.T) {
	type args struct {
		r1 string
		r2 string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"1|2", args{"1", "2"}, base64.StdEncoding.EncodeToString([]byte("1|2"))},
		{"1|2|3", args{base64.StdEncoding.EncodeToString([]byte("1|2")), "3"}, base64.StdEncoding.EncodeToString([]byte("1|2|3"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeReportID(tt.args.r1, tt.args.r2); got != tt.want {
				t.Errorf("mergeReportID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeSeverity(t *testing.T) {
	type args struct {
		s1 Severity
		s2 Severity
	}
	tests := []struct {
		name string
		args args
		want Severity
	}{
		{"empty string and none", args{Severity(""), None}, None},
		{"none and empty string", args{None, Severity("")}, None},
		{"none and low", args{None, Low}, Low},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeSeverity(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("mergeSeverity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeScanStatus(t *testing.T) {
	errorStatus := job.ErrorStatus.String()
	runningStatus := job.RunningStatus.String()
	successStatus := job.SuccessStatus.String()

	type args struct {
		s1 string
		s2 string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"running and error", args{runningStatus, errorStatus}, runningStatus},
		{"running and success", args{runningStatus, successStatus}, runningStatus},
		{"running and running", args{runningStatus, runningStatus}, runningStatus},
		{"success and error", args{successStatus, errorStatus}, successStatus},
		{"success and success", args{successStatus, successStatus}, successStatus},
		{"error and error", args{errorStatus, errorStatus}, errorStatus},
		{"error and empty string", args{errorStatus, ""}, errorStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeScanStatus(tt.args.s1, tt.args.s2); got != tt.want {
				t.Errorf("mergeScanStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeVulnerabilitySummary(t *testing.T) {
	assert := assert.New(t)
	v1 := VulnerabilitySummary{
		Total:   1,
		Fixable: 1,
		Summary: map[Severity]int{Low: 1},
	}

	r := v1.Merge(&VulnerabilitySummary{
		Total:   1,
		Fixable: 1,
		Summary: map[Severity]int{Low: 1, High: 1},
	})

	assert.Equal(2, r.Total)
	assert.Equal(2, r.Fixable)
	assert.Len(r.Summary, 2)
	assert.Equal(2, r.Summary[Low])
	assert.Equal(1, r.Summary[High])
}

func TestMergeNativeReportSummary(t *testing.T) {
	assert := assert.New(t)
	errorStatus := job.ErrorStatus.String()
	runningStatus := job.RunningStatus.String()

	v1 := VulnerabilitySummary{
		Total:   1,
		Fixable: 1,
		Summary: map[Severity]int{Low: 1},
	}

	n1 := NativeReportSummary{
		ScanStatus: runningStatus,
		Severity:   Low,
		TotalCount: 1,
		Summary:    &v1,
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
