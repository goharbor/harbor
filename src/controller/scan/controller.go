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

package scan

import (
	"context"

	"github.com/goharbor/harbor/src/controller/artifact"
	sca "github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/report"
)

// Controller provides the related operations for triggering scan.
// TODO: Here the artifact object is reused the v1 one which is sent to the adapter,
//  it should be pointed to the general artifact object in future once it's ready.
type Controller interface {
	// Scan the given artifact
	//
	//   Arguments:
	//     ctx context.Context : the context for this method
	//     artifact *artifact.Artifact : artifact to be scanned
	//     options ...Option     : options for triggering a scan
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	Scan(ctx context.Context, artifact *artifact.Artifact, options ...Option) error

	// GetReport gets the reports for the given artifact identified by the digest
	//
	//   Arguments:
	//     ctx context.Context : the context for this method
	//     artifact *v1.Artifact : the scanned artifact
	//     mimeTypes []string    : the mime types of the reports
	//
	//   Returns:
	//     []*scan.Report : scan results by different scanner vendors
	//     error          : non nil error if any errors occurred
	GetReport(ctx context.Context, artifact *artifact.Artifact, mimeTypes []string) ([]*scan.Report, error)

	// GetSummary gets the summaries of the reports with given types.
	//
	//   Arguments:
	//     ctx context.Context : the context for this method
	//     artifact *artifact.Artifact    : the scanned artifact
	//     mimeTypes []string       : the mime types of the reports
	//     options ...report.Option : optional report options, specify if needed
	//
	//   Returns:
	//     map[string]interface{} : report summaries indexed by mime types
	//     error                  : non nil error if any errors occurred
	GetSummary(ctx context.Context, artifact *artifact.Artifact, mimeTypes []string, options ...report.Option) (map[string]interface{}, error)

	// Get the scan log for the specified artifact with the given digest
	//
	//   Arguments:
	//     ctx context.Context : the context for this method
	//     uuid string : the UUID of the scan report
	//
	//   Returns:
	//     []byte : the log text stream
	//     error  : non nil error if any errors occurred
	GetScanLog(ctx context.Context, uuid string) ([]byte, error)

	// Delete the reports related with the specified digests
	//
	//  Arguments:
	//    digests ...string : specify one or more digests whose reports will be deleted
	//
	//  Returns:
	//    error        : non nil error if any errors occurred
	DeleteReports(ctx context.Context, digests ...string) error

	// UpdateReport update the report
	//
	//   Arguments:
	//     ctx context.Context : the context for this method
	//     report *sca.CheckInReport : the scan report
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	UpdateReport(ctx context.Context, report *sca.CheckInReport) error

	// Scan all the artifacts
	//
	//   Arguments:
	//     ctx context.Context  : the context for this method
	//     trigger string       : the trigger mode to start the scan all job
	//     async bool           : scan all the artifacts in background
	//     triggerRevision int64: for identifying the duplicated trigger from the same schedule, refer to https://github.com/goharbor/harbor/issues/14683 for more detail
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	ScanAll(ctx context.Context, trigger string, async bool, triggerRevision ...int64) (int64, error)
}
