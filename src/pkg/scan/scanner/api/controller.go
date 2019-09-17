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

package api

import (
	"github.com/goharbor/harbor/src/pkg/q"
	dscan "github.com/goharbor/harbor/src/pkg/scan/scanner/dao/scan"
	"github.com/goharbor/harbor/src/pkg/scan/scanner/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/scanner/scan"
)

// Controller provides the related operations of scanner for the upper API.
// All the capabilities of the scanner are defined here.
type Controller interface {
	// ListRegistrations returns a list of currently configured scanner registrations.
	// Query parameters are optional
	//
	//  Arguments:
	//    query *q.Query : query parameters
	//
	//  Returns:
	//    []*scanner.Registration : scanner list of all the matched ones
	//    error                   : non nil error if any errors occurred
	ListRegistrations(query *q.Query) ([]*scanner.Registration, error)

	// CreateRegistration creates a new scanner registration with the given data.
	// Returns the scanner registration identifier.
	//
	//  Arguments:
	//    registration *scanner.Registration : scanner registration to create
	//
	//  Returns:
	//    string : the generated UUID of the new scanner
	//    error  : non nil error if any errors occurred
	CreateRegistration(registration *scanner.Registration) (string, error)

	// GetRegistration returns the details of the specified scanner registration.
	//
	//  Arguments:
	//    registrationUUID string : the UUID of the given scanner
	//
	//  Returns:
	//    *scanner.Registration : the required scanner
	//    error                 : non nil error if any errors occurred
	GetRegistration(registrationUUID string) (*scanner.Registration, error)

	// RegistrationExists checks if the provided registration is there.
	//
	//  Arguments:
	//    registrationUUID string : the UUID of the given scanner
	//
	//  Returns:
	//    true for existing or false for not existing
	RegistrationExists(registrationUUID string) bool

	// UpdateRegistration updates the specified scanner registration.
	//
	//  Arguments:
	//    registration *scanner.Registration : scanner registration to update
	//
	//  Returns:
	//    error  : non nil error if any errors occurred
	UpdateRegistration(registration *scanner.Registration) error

	// DeleteRegistration deletes the specified scanner registration.
	//
	//  Arguments:
	//    registrationUUID string : the UUID of the given scanner which is going to be deleted
	//
	//  Returns:
	//    *scanner.Registration : the deleted scanner
	//    error                 : non nil error if any errors occurred
	DeleteRegistration(registrationUUID string) (*scanner.Registration, error)

	// SetDefaultRegistration marks the specified scanner registration as default.
	// The implementation is supposed to unset any registration previously set as default.
	//
	//  Arguments:
	//    registrationUUID string : the UUID of the given scanner which is marked as default
	//
	//  Returns:
	//    error : non nil error if any errors occurred
	SetDefaultRegistration(registrationUUID string) error

	// SetRegistrationByProject sets scanner for the given project.
	//
	//  Arguments:
	//    projectID int64  : the ID of the given project
	//    scannerID string : the UUID of the the scanner
	//
	//  Returns:
	//    error : non nil error if any errors occurred
	SetRegistrationByProject(projectID int64, scannerID string) error

	// GetRegistrationByProject returns the configured scanner registration of the given project or
	// the system default registration if exists or `nil` if no system registrations set.
	//
	//   Arguments:
	//     projectID int64 : the ID of the given project
	//
	//   Returns:
	//     *scanner.Registration : the default scanner registration
	//     error                 : non nil error if any errors occurred
	GetRegistrationByProject(projectID int64) (*scanner.Registration, error)

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
	//     artifact *res.Artifact : artifact to be scanned
	//
	//   Returns:
	//     error  : non nil error if any errors occurred
	Scan(artifact *scan.Artifact) error

	// GetReport gets the reports for the given artifact identified by the digest
	//
	//   Arguments:
	//     artifact *res.Artifact : the scanned artifact
	//
	//   Returns:
	//     []*scan.Report : scan results by different scanner vendors
	//     error          : non nil error if any errors occurred
	GetReport(artifact *scan.Artifact) ([]*dscan.Report, error)

	// Get the scan log for the specified artifact with the given digest
	//
	//   Arguments:
	//     digest string : the digest of the artifact
	//
	//   Returns:
	//     []byte : the log text stream
	//     error  : non nil error if any errors occurred
	GetScanLog(digest string) ([]byte, error)
}
