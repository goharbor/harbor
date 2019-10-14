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

package report

import (
	"reflect"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/pkg/errors"
)

// SupportedGenerators declares mappings between mime type and summary generator func.
var SupportedGenerators = map[string]SummaryGenerator{
	v1.MimeTypeNativeReport: GenerateNativeSummary,
}

// GenerateSummary is a helper function to generate report
// summary based on the given report.
func GenerateSummary(r *scan.Report) (interface{}, error) {
	g, ok := SupportedGenerators[r.MimeType]
	if !ok {
		return nil, errors.Errorf("no generator bound with mime type %s", r.MimeType)
	}

	return g(r)
}

// SummaryGenerator is a func template which used to generated report
// summary for relevant mime type.
type SummaryGenerator func(r *scan.Report) (interface{}, error)

// GenerateNativeSummary generates the report summary for the native report.
func GenerateNativeSummary(r *scan.Report) (interface{}, error) {
	sum := &vuln.NativeReportSummary{}
	sum.ReportID = r.UUID
	sum.StartTime = r.StartTime
	sum.EndTime = r.EndTime
	sum.Duration = r.EndTime.Unix() - r.StartTime.Unix()

	sum.ScanStatus = job.ErrorStatus.String()
	if job.Status(r.Status).Code() != -1 {
		sum.ScanStatus = r.Status
	}

	// If the status is not success/stopped, there will not be any report.
	if r.Status != job.SuccessStatus.String() &&
		r.Status != job.StoppedStatus.String() {
		return sum, nil
	}

	// Probably no report data if the job is interrupted
	if len(r.Report) == 0 {
		return nil, errors.Errorf("no report data for %s, status is: %s", r.UUID, sum.ScanStatus)
	}

	raw, err := ResolveData(r.MimeType, []byte(r.Report))
	if err != nil {
		return nil, err
	}

	rp, ok := raw.(*vuln.Report)
	if !ok {
		return nil, errors.Errorf("type mismatch: expect *vuln.Report but got %s", reflect.TypeOf(raw).String())
	}

	sum.Severity = rp.Severity
	vsum := &vuln.VulnerabilitySummary{
		Total:   len(rp.Vulnerabilities),
		Summary: make(vuln.SeveritySummary),
	}

	for _, v := range rp.Vulnerabilities {
		if num, ok := vsum.Summary[v.Severity]; ok {
			vsum.Summary[v.Severity] = num + 1
		} else {
			vsum.Summary[v.Severity] = 1
		}
	}
	sum.Summary = vsum

	return sum, nil
}
