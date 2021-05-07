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
	"encoding/json"
	"reflect"
	"testing"

	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/stretchr/testify/assert"
)

func TestReport_Merge(t *testing.T) {
	a := []*VulnerabilityItem{
		{ID: "CVE-2017-8283"},
		{ID: "CVE-2017-8284"},
	}
	b := []*VulnerabilityItem{
		{ID: "CVE-2017-8285"},
	}
	type fields struct {
		GeneratedAt     string
		Scanner         *v1.Scanner
		Severity        Severity
		Vulnerabilities []*VulnerabilityItem
	}
	type args struct {
		another *Report
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Report
	}{
		{"GeneratedAt", fields{GeneratedAt: "2020-04-06T18:38:34.791086859Z"}, args{&Report{GeneratedAt: "2020-04-06T18:38:34.791086860Z"}}, &Report{GeneratedAt: "2020-04-06T18:38:34.791086860Z"}},
		{"Vulnerabilities nil & a", fields{Vulnerabilities: nil}, args{&Report{Vulnerabilities: a}}, &Report{Vulnerabilities: a}},
		{"Vulnerabilities a & nil", fields{Vulnerabilities: a}, args{&Report{Vulnerabilities: nil}}, &Report{Vulnerabilities: a}},
		{"Vulnerabilities a & b", fields{Vulnerabilities: a}, args{&Report{Vulnerabilities: b}}, &Report{Vulnerabilities: append(a, b...)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := &Report{
				GeneratedAt:     tt.fields.GeneratedAt,
				Scanner:         tt.fields.Scanner,
				Severity:        tt.fields.Severity,
				Vulnerabilities: tt.fields.Vulnerabilities,
			}
			got := report.Merge(tt.args.another)
			got.vulnerabilityItemList = nil
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Report.Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReportMarshalJSON(t *testing.T) {
	assert := assert.New(t)

	report := &Report{
		GeneratedAt: "GeneratedAt",
	}

	b, _ := json.Marshal(report)
	assert.Contains(string(b), "vulnerabilities")
}

func TestGetSummarySeverity(t *testing.T) {
	assert := assert.New(t)

	vul1 := &VulnerabilityItem{
		ID:         "cve1",
		Severity:   Low,
		FixVersion: "1.3",
	}

	vul2 := &VulnerabilityItem{
		ID:       "cve2",
		Severity: Low,
	}

	vul3 := &VulnerabilityItem{
		ID:       "cve3",
		Severity: Medium,
	}

	l := VulnerabilityItemList{}
	l.Add(vul1, vul2, vul3)

	s := SeveritySummary{
		Low:    2,
		Medium: 1,
	}

	severity, sum := l.GetSeveritySummary()
	assert.Equal(Medium, severity)
	assert.Equal(3, sum.Total)
	assert.Equal(1, sum.Fixable)
	assert.Equal(s, sum.Summary)
}
