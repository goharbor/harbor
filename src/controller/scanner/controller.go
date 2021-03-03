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

package scanner

import (
	"context"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// Registration ...
type Registration = scanner.Registration

// Controller provides the related operations of scanner for the upper API.
// All the capabilities of the scanner are defined here.
type Controller interface {
	// ListRegistrations returns a list of currently configured scanner registrations.
	// Query parameters are optional
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    query *q.Query : query parameters
	//
	//  Returns:
	//    []*scanner.Registration : scanner list of all the matched ones
	//    error                   : non nil error if any errors occurred
	ListRegistrations(ctx context.Context, query *q.Query) ([]*scanner.Registration, error)

	// GetTotalOfRegistrations returns the total count of scanner registrations according to the query.
	GetTotalOfRegistrations(ctx context.Context, query *q.Query) (int64, error)

	// CreateRegistration creates a new scanner registration with the given data.
	// Returns the scanner registration identifier.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    registration *scanner.Registration : scanner registration to create
	//
	//  Returns:
	//    string : the generated UUID of the new scanner
	//    error  : non nil error if any errors occurred
	CreateRegistration(ctx context.Context, registration *scanner.Registration) (string, error)

	// GetRegistration returns the details of the specified scanner registration.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    registrationUUID string : the UUID of the given scanner
	//
	//  Returns:
	//    *scanner.Registration : the required scanner
	//    error                 : non nil error if any errors occurred
	GetRegistration(ctx context.Context, registrationUUID string) (*scanner.Registration, error)

	// RegistrationExists checks if the provided registration is there.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    registrationUUID string : the UUID of the given scanner
	//
	//  Returns:
	//    true for existing or false for not existing
	RegistrationExists(ctx context.Context, registrationUUID string) bool

	// UpdateRegistration updates the specified scanner registration.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    registration *scanner.Registration : scanner registration to update
	//
	//  Returns:
	//    error  : non nil error if any errors occurred
	UpdateRegistration(ctx context.Context, registration *scanner.Registration) error

	// DeleteRegistration deletes the specified scanner registration.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    registrationUUID string : the UUID of the given scanner which is going to be deleted
	//
	//  Returns:
	//    *scanner.Registration : the deleted scanner
	//    error                 : non nil error if any errors occurred
	DeleteRegistration(ctx context.Context, registrationUUID string) (*scanner.Registration, error)

	// SetDefaultRegistration marks the specified scanner registration as default.
	// The implementation is supposed to unset any registration previously set as default.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    registrationUUID string : the UUID of the given scanner which is marked as default
	//
	//  Returns:
	//    error : non nil error if any errors occurred
	SetDefaultRegistration(ctx context.Context, registrationUUID string) error

	// SetRegistrationByProject sets scanner for the given project.
	//
	//  Arguments:
	//    ctx context.Context : the context.Context for this method
	//    projectID int64  : the ID of the given project
	//    scannerID string : the UUID of the the scanner
	//
	//  Returns:
	//    error : non nil error if any errors occurred
	SetRegistrationByProject(ctx context.Context, projectID int64, scannerID string) error

	// GetRegistrationByProject returns the configured scanner registration of the given project or
	// the system default registration if exists or `nil` if no system registrations set.
	//
	//   Arguments:
	//     ctx context.Context : the context.Context for this method
	//     projectID int64 : the ID of the given project
	//
	//   Returns:
	//     *scanner.Registration : the default scanner registration
	//     error                 : non nil error if any errors occurred
	GetRegistrationByProject(ctx context.Context, projectID int64, options ...Option) (*scanner.Registration, error)

	// Ping pings Scanner Adapter to test EndpointURL and Authorization settings.
	// The implementation is supposed to call the GetMetadata method on scanner.Client.
	// Returns `nil` if connection succeeded, a non `nil` error otherwise.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    registration *scanner.Registration : scanner registration to ping
	//
	//  Returns:
	//    *v1.ScannerAdapterMetadata : metadata returned by the scanner if successfully ping
	//    error                      : non nil error if any errors occurred
	Ping(ctx context.Context, registration *scanner.Registration) (*v1.ScannerAdapterMetadata, error)

	// GetMetadata returns the metadata of the given scanner.
	//
	//  Arguments:
	//    ctx context.Context : the context for this method
	//    registrationUUID string : the UUID of the given scanner which is marked as default
	//
	//  Returns:
	//    *v1.ScannerAdapterMetadata : metadata returned by the scanner if successfully ping
	//    error                      : non nil error if any errors occurred
	GetMetadata(ctx context.Context, registrationUUID string) (*v1.ScannerAdapterMetadata, error)
}
