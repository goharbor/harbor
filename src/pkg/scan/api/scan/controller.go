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
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// Controller provides the related operations for triggering scan.
// TODO: Here the artifact object is reused the v1 one which is sent to the adapter,
//  it should be pointed to the general artifact object in future once it's ready.
type Controller interface {
	// Ping pings Scanner Adapter to test EndpointURL and Authorization settings.
	// The implementation is supposed to call the GetMetadata method on scanner.Client.
	// Returns `nil` if connection succeeded, a non `nil` error otherwise.
	//
	//  Arguments:
	//    registration *scanner.Registration : scanner registration to ping
	//
	//  Returns:
	//    error  : non nil error if any errors occurred
	Ping(registration *scanner.Registration) error

	// Scan the given artifact
	//
	//   Arguments:
	//     artifact *v1.Artifact : artifact to be scanned
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	Scan(artifact *v1.Artifact) error

	// GetReport gets the reports for the given artifact identified by the digest
	//
	//   Arguments:
	//     artifact *v1.Artifact : the scanned artifact
	//
	//   Returns:
	//     []*scan.Report : scan results by different scanner vendors
	//     error          : non nil error if any errors occurred
	GetReport(artifact *v1.Artifact) ([]*scan.Report, error)

	// Get the scan log for the specified artifact with the given digest
	//
	//   Arguments:
	//     digest string : the digest of the artifact
	//
	//   Returns:
	//     []byte : the log text stream
	//     error  : non nil error if any errors occurred
	GetScanLog(digest string) ([]byte, error)

	// HandleJobHooks handle the hook events from the job service
	// e.g : status change of the scan job or scan result
	//
	//   Arguments:
	//     trackID int64            : ID for the result record
	//     change *job.StatusChange : change event from the job service
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	HandleJobHooks(trackID int64, change *job.StatusChange) error
}
