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
	"reflect"
	"testing"

	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

func TestReport_Merge(t *testing.T) {
	emptyVulnerabilities := []*VulnerabilityItem{}
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
		{"Vulnerabilities nil & nil", fields{Vulnerabilities: nil}, args{&Report{Vulnerabilities: nil}}, &Report{Vulnerabilities: nil}},
		{"Vulnerabilities nil & not nil", fields{Vulnerabilities: nil}, args{&Report{Vulnerabilities: emptyVulnerabilities}}, &Report{Vulnerabilities: emptyVulnerabilities}},
		{"Vulnerabilities not nil & nil", fields{Vulnerabilities: emptyVulnerabilities}, args{&Report{Vulnerabilities: nil}}, &Report{Vulnerabilities: emptyVulnerabilities}},
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
			if got := report.Merge(tt.args.another); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Report.Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}
