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
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Manager defines the related scanner API endpoints
type Manager interface {
	// List returns a list of currently configured scanner registrations.
	// Query parameters are optional
	List(query *q.Query) ([]*scanner.Registration, error)

	// Create creates a new scanner registration with the given data.
	// Returns the scanner registration identifier.
	Create(registration *scanner.Registration) (string, error)

	// Get returns the details of the specified scanner registration.
	Get(registrationUUID string) (*scanner.Registration, error)

	// Update updates the specified scanner registration.
	Update(registration *scanner.Registration) error

	// Delete deletes the specified scanner registration.
	Delete(registrationUUID string) error

	// SetAsDefault marks the specified scanner registration as default.
	// The implementation is supposed to unset any registration previously set as default.
	SetAsDefault(registrationUUID string) error

	// GetDefault returns the default scanner registration or `nil` if there are no registrations configured.
	GetDefault() (*scanner.Registration, error)
}

// basicManager is the default implementation of Manager
type basicManager struct{}

// New a basic manager
func New() Manager {
	return &basicManager{}
}

// Create ...
func (bm *basicManager) Create(registration *scanner.Registration) (string, error) {
	if registration == nil {
		return "", errors.New("nil registration to create")
	}

	// Inject new UUID
	uid, err := uuid.NewUUID()
	if err != nil {
		return "", errors.Wrap(err, "new UUID: create registration")
	}
	registration.UUID = uid.String()

	if err := registration.Validate(true); err != nil {
		return "", errors.Wrap(err, "create registration")
	}

	if _, err := scanner.AddRegistration(registration); err != nil {
		return "", errors.Wrap(err, "dao: create registration")
	}

	return uid.String(), nil
}

// Get ...
func (bm *basicManager) Get(registrationUUID string) (*scanner.Registration, error) {
	if len(registrationUUID) == 0 {
		return nil, errors.New("empty uuid of registration")
	}

	return scanner.GetRegistration(registrationUUID)
}

// Update ...
func (bm *basicManager) Update(registration *scanner.Registration) error {
	if registration == nil {
		return errors.New("nil registration to update")
	}

	if err := registration.Validate(true); err != nil {
		return errors.Wrap(err, "update registration")
	}

	return scanner.UpdateRegistration(registration)
}

// Delete ...
func (bm *basicManager) Delete(registrationUUID string) error {
	if len(registrationUUID) == 0 {
		return errors.New("empty UUID to delete")
	}

	return scanner.DeleteRegistration(registrationUUID)
}

// List ...
func (bm *basicManager) List(query *q.Query) ([]*scanner.Registration, error) {
	return scanner.ListRegistrations(query)
}

// SetAsDefault ...
func (bm *basicManager) SetAsDefault(registrationUUID string) error {
	if len(registrationUUID) == 0 {
		return errors.New("empty UUID to set default")
	}

	return scanner.SetDefaultRegistration(registrationUUID)
}

// GetDefault ...
func (bm *basicManager) GetDefault() (*scanner.Registration, error) {
	return scanner.GetDefaultRegistration()
}
