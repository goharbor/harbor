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
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/lib/errors"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanv2"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

//GetNativeV1ReportFromResolvedData returns the native V1 scan report from the resolved
//interface data.
func GetNativeV1ReportFromResolvedData(ctx job.Context, rp interface{}) (*vuln.Report, error) {
	report, ok := rp.(*vuln.Report)
	if !ok {
		return nil, errors.New("Data cannot be converted to v1 report format")
	}
	ctx.GetLogger().Infof("Converted raw data to report. Count of Vulnerabilities in report : %d", len(report.Vulnerabilities))
	return report, nil
}

//ConvertV1ReportToV2Report converts the Report instance compatble with V1 schema to a Report and VulnerabilityRecord instance
//compatible with the V2 schema
func ConvertV1ReportToV2Report(reportV1 *scan.Report) (string, error) {
	reportV2 := new(scanv2.Report)
	reportV2.UUID = reportV1.UUID
	reportV2.Digest = reportV1.Digest
	reportV2.StartTime = reportV1.StartTime
	reportV2.EndTime = reportV1.EndTime
	reportV2.JobID = reportV1.JobID
	reportV2.MimeType = reportV1.MimeType
	reportV2.TrackID = reportV1.TrackID
	reportV2.Status = reportV1.Status
	reportV2.StatusCode = reportV1.StatusCode
	reportV2.StatusRevision = reportV1.StatusRevision
	reportV2.RegistrationUUID = reportV1.RegistrationUUID
	reportV2.Requester = reportV1.Requester

	//parse the raw report with the V1 schema of the report to the normalized structures
	var rawReport vuln.Report
	if err := json.Unmarshal([]byte(reportV1.Report), &rawReport); err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Error when convert V1 report to V2"))
	}

	return reportV2.UUID, nil
}
