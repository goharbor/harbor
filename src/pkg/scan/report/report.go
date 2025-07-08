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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

// Merger is a helper function to merge report together
type Merger func(r1, r2 any) (any, error)

// SupportedMergers declares mappings between mime type and report merger func.
var SupportedMergers = map[string]Merger{
	v1.MimeTypeNativeReport:               MergeNativeReport,
	v1.MimeTypeGenericVulnerabilityReport: MergeNativeReport,
}

// Merge merge report r1 and r2
func Merge(mimeType string, r1, r2 any) (any, error) {
	m, ok := SupportedMergers[mimeType]
	if !ok {
		return nil, errors.Errorf("no report merger bound with mime type %s", mimeType)
	}

	return m(r1, r2)
}

// MergeNativeReport merge report r1 and r2
func MergeNativeReport(r1, r2 any) (any, error) {
	nr1, ok := r1.(*vuln.Report)
	if !ok {
		return nil, errors.New("native report required")
	}

	nr2, ok := r2.(*vuln.Report)
	if !ok {
		return nil, errors.New("native report required")
	}

	return nr1.Merge(nr2), nil
}

// Reports slice of scan.Reports pointer
type Reports []*scan.Report

// ResolveData resolve the data from the reports and merge them together
func (l Reports) ResolveData(mimeType string) (any, error) {
	var result any

	for _, rp := range l {
		// Resolve scan report data only when it is ready and its mime type equal the given one
		if len(rp.Report) == 0 || rp.MimeType != mimeType {
			continue
		}

		vrp, err := ResolveData(rp.MimeType, []byte(rp.Report), WithArtifactDigest(rp.Digest))
		if err != nil {
			return nil, err
		}

		if result == nil {
			result = vrp
		} else {
			r, err := Merge(rp.MimeType, result, vrp)
			if err != nil {
				return nil, err
			}

			result = r
		}
	}

	return result, nil
}
