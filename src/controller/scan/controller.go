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
	"github.com/goharbor/harbor/src/jobservice/job"
	allowlist "github.com/goharbor/harbor/src/pkg/allowlist/models"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
)

// Vulnerable ...
type Vulnerable struct {
	VulnerabilitiesCount int
	ScanStatus           string
	Severity             *vuln.Severity
	CVEBypassed          []string
}

// IsScanSuccess returns true when the artifact scanned success
func (v *Vulnerable) IsScanSuccess() bool {
	return v.ScanStatus == job.SuccessStatus.String()
}

// Controller provides the related operations for triggering scan.
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

	// Stop scan job of the given artifact
	//
	//   Arguments:
	//     ctx context.Context : the context for this method
	//     artifact *artifact.Artifact : the artifact whose scan job to be stopped
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	Stop(ctx context.Context, artifact *artifact.Artifact) error

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
	//
	//   Returns:
	//     map[string]interface{} : report summaries indexed by mime types
	//     error                  : non nil error if any errors occurred
	GetSummary(ctx context.Context, artifact *artifact.Artifact, mimeTypes []string) (map[string]interface{}, error)

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

	// Scan all the artifacts
	//
	//   Arguments:
	//     ctx context.Context : the context for this method
	//     trigger string      : the trigger mode to start the scan all job
	//     async bool          : scan all the artifacts in background
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	ScanAll(ctx context.Context, trigger string, async bool) (int64, error)

	// GetVulnerable returns the vulnerable of the artifact for the allowlist
	//
	//   Arguments:
	//     ctx context.Context : the context for this method
	//     artifact *artifact.Artifact : artifact to be scanned
	//
	//   Returns
	//      *Vulnerable : the vulnerable
	//     error        : non nil error if any errors occurred
	GetVulnerable(ctx context.Context, artifact *artifact.Artifact, allowlist allowlist.CVESet) (*Vulnerable, error)
}
