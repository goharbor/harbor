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

package optimizer

import (
	"context"

	ar "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
	v1 "github.com/goharbor/harbor/src/pkg/optimizer/rest/v1"
)

// Controller provides the related operations of the optimizer adapter framework:
// registration lifecycle and triggering optimizations.
//
// Project-level optimizer selection (mirroring the projectScanner metadata pattern
// of the scanner framework) is intentionally not supported in v1 — the global
// default registration is used for all projects.
type Controller interface {
	// ListRegistrations returns a list of currently configured optimizer registrations.
	ListRegistrations(ctx context.Context, query *q.Query) ([]*dao.Registration, error)

	// GetTotalOfRegistrations returns the total count of registrations according to the query.
	GetTotalOfRegistrations(ctx context.Context, query *q.Query) (int64, error)

	// CreateRegistration creates a new optimizer registration with the given data.
	// The adapter is pinged before the registration is persisted.
	CreateRegistration(ctx context.Context, registration *dao.Registration) (string, error)

	// GetRegistration returns the details of the specified registration.
	GetRegistration(ctx context.Context, registrationUUID string) (*dao.Registration, error)

	// UpdateRegistration updates the specified registration.
	UpdateRegistration(ctx context.Context, registration *dao.Registration) error

	// DeleteRegistration deletes the specified registration and returns it.
	DeleteRegistration(ctx context.Context, registrationUUID string) (*dao.Registration, error)

	// SetDefaultRegistration marks the specified registration as default.
	SetDefaultRegistration(ctx context.Context, registrationUUID string) error

	// GetDefaultRegistration returns the default registration or nil if none is configured.
	GetDefaultRegistration(ctx context.Context) (*dao.Registration, error)

	// Ping pings the optimizer adapter to ensure the registration is reachable and
	// returns its validated metadata.
	Ping(ctx context.Context, registration *dao.Registration) (*v1.OptimizerAdapterMetadata, error)

	// GetMetadata returns the metadata of the adapter behind the specified registration.
	GetMetadata(ctx context.Context, registrationUUID string) (*v1.OptimizerAdapterMetadata, error)

	// Optimize submits the artifact to the default optimizer adapter asynchronously:
	// it creates an execution + task and a Pending dockerfile_optimization row, then
	// launches the jobservice job that talks to the adapter.
	// Returns the execution ID.
	Optimize(ctx context.Context, artifact *ar.Artifact, tag string) (int64, error)

	// CheckAvailability runs the same checks Optimize would perform before starting
	// a job (configured, enabled, reachable, capable of the artifact's mime type) and
	// returns nil if optimization could be started, or the same error Optimize would
	// return otherwise.
	CheckAvailability(ctx context.Context, artifact *ar.Artifact) error
}
