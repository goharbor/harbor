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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

// Options provides options for getting the report w/ summary.
type Options struct {
	// If it is set, the returned report will contains artifact digest for the vulnerabilities
	ArtifactDigest string
}

// Option for getting the report w/ summary with func template way.
type Option func(options *Options)

// WithArtifactDigest is an option of setting artifact digest
func WithArtifactDigest(artifactDigest string) Option {
	return func(options *Options) {
		options.ArtifactDigest = artifactDigest
	}
}

// SummaryMerger is a helper function to merge summary together
type SummaryMerger func(s1, s2 any) (any, error)

// SupportedSummaryMergers declares mappings between mime type and summary merger func.
var SupportedSummaryMergers = map[string]SummaryMerger{
	v1.MimeTypeNativeReport:               MergeNativeSummary,
	v1.MimeTypeGenericVulnerabilityReport: MergeNativeSummary,
}

// MergeSummary merge summary s1 and s2
func MergeSummary(mimeType string, s1, s2 any) (any, error) {
	m, ok := SupportedSummaryMergers[mimeType]
	if !ok {
		return nil, errors.Errorf("no summary merger bound with mime type %s", mimeType)
	}

	return m(s1, s2)
}

// MergeNativeSummary merge vuln.NativeReportSummary together
func MergeNativeSummary(s1, s2 any) (any, error) {
	nrs1, ok := s1.(*vuln.NativeReportSummary)
	if !ok {
		return nil, errors.New("native report summary required")
	}

	nrs2, ok := s2.(*vuln.NativeReportSummary)
	if !ok {
		return nil, errors.New("native report summary required")
	}

	return nrs1.Merge(nrs2), nil
}

// SupportedGenerators declares mappings between mime type and summary generator func.
var SupportedGenerators = map[string]SummaryGenerator{
	v1.MimeTypeNativeReport:               GenerateNativeSummary,
	v1.MimeTypeGenericVulnerabilityReport: GenerateNativeSummary,
}

// GenerateSummary is a helper function to generate report
// summary based on the given report.
func GenerateSummary(r *scan.Report, options ...Option) (any, error) {
	g, ok := SupportedGenerators[r.MimeType]
	if !ok {
		return nil, errors.Errorf("no generator bound with mime type %s", r.MimeType)
	}

	return g(r, options...)
}

// SummaryGenerator is a func template which used to generated report
// summary for relevant mime type.
type SummaryGenerator func(r *scan.Report, options ...Option) (any, error)

// GenerateNativeSummary generates the report summary for the native report.
func GenerateNativeSummary(r *scan.Report, _ ...Option) (any, error) {
	sum := &vuln.NativeReportSummary{}
	sum.ReportID = r.UUID
	sum.StartTime = r.StartTime
	sum.EndTime = r.EndTime
	sum.Duration = r.EndTime.Unix() - r.StartTime.Unix()
	if sum.Duration < 0 {
		sum.Duration = 0
	}

	sum.ScanStatus = job.ErrorStatus.String()
	if job.Status(r.Status).Code() != -1 {
		sum.ScanStatus = r.Status
	}

	sum.TotalCount = 1

	// If the status is not success, there will not be any report.
	if r.Status != job.SuccessStatus.String() {
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

	sum.CompleteCount = 1
	sum.CompletePercent = 100
	sum.Severity = rp.Severity
	sum.Scanner = rp.Scanner

	sum.UpdateSeveritySummary(rp.GetVulnerabilityItemList())

	return sum, nil
}
