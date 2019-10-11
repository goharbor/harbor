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
	"github.com/goharbor/harbor/src/core/promgr/metamgr"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	rscanner "github.com/goharbor/harbor/src/pkg/scan/scanner"
	"github.com/pkg/errors"
)

const (
	proScannerMetaKey = "projectScanner"
)

// DefaultController is a singleton api controller for plug scanners
var DefaultController = New()

// New a basic controller
func New() Controller {
	return &basicController{
		manager:    rscanner.New(),
		proMetaMgr: metamgr.NewDefaultProjectMetadataManager(),
	}
}

// basicController is default implementation of api.Controller interface
type basicController struct {
	// managers for managing the scanner registrations
	manager rscanner.Manager
	// for operating the project level configured scanner
	proMetaMgr metamgr.ProjectMetadataManager
}

// ListRegistrations ...
func (bc *basicController) ListRegistrations(query *q.Query) ([]*scanner.Registration, error) {
	return bc.manager.List(query)
}

// CreateRegistration ...
func (bc *basicController) CreateRegistration(registration *scanner.Registration) (string, error) {
	// TODO: Get metadata from the adapter service first
	// Check if there are any registrations already existing.
	l, err := bc.manager.List(&q.Query{
		PageSize:   1,
		PageNumber: 1,
	})
	if err != nil {
		return "", errors.Wrap(err, "api controller: create registration")
	}

	if len(l) == 0 && !registration.IsDefault {
		// Mark the 1st as default automatically
		registration.IsDefault = true
	}

	return bc.manager.Create(registration)
}

// GetRegistration ...
func (bc *basicController) GetRegistration(registrationUUID string) (*scanner.Registration, error) {
	return bc.manager.Get(registrationUUID)
}

// RegistrationExists ...
func (bc *basicController) RegistrationExists(registrationUUID string) bool {
	registration, err := bc.manager.Get(registrationUUID)

	// Just logged when an error occurred
	if err != nil {
		logger.Errorf("Check existence of registration error: %s", err)
	}

	return !(err == nil && registration == nil)
}

// UpdateRegistration ...
func (bc *basicController) UpdateRegistration(registration *scanner.Registration) error {
	return bc.manager.Update(registration)
}

// SetDefaultRegistration ...
func (bc *basicController) DeleteRegistration(registrationUUID string) (*scanner.Registration, error) {
	registration, err := bc.manager.Get(registrationUUID)
	if registration == nil && err == nil {
		// Not found
		return nil, nil
	}

	if err := bc.manager.Delete(registrationUUID); err != nil {
		return nil, errors.Wrap(err, "api controller: delete registration")
	}

	return registration, nil
}

// SetDefaultRegistration ...
func (bc *basicController) SetDefaultRegistration(registrationUUID string) error {
	return bc.manager.SetAsDefault(registrationUUID)
}

// SetRegistrationByProject ...
func (bc *basicController) SetRegistrationByProject(projectID int64, registrationID string) error {
	if projectID == 0 {
		return errors.New("invalid project ID")
	}

	if len(registrationID) == 0 {
		return errors.New("missing scanner UUID")
	}

	// Only keep the UUID in the metadata of the given project
	// Scanner metadata existing?
	m, err := bc.proMetaMgr.Get(projectID, proScannerMetaKey)
	if err != nil {
		return errors.Wrap(err, "api controller: set project scanner")
	}

	// Update if exists
	if len(m) > 0 {
		// Compare and set new
		if registrationID != m[proScannerMetaKey] {
			m[proScannerMetaKey] = registrationID
			if err := bc.proMetaMgr.Update(projectID, m); err != nil {
				return errors.Wrap(err, "api controller: set project scanner")
			}
		}
	} else {
		meta := make(map[string]string, 1)
		meta[proScannerMetaKey] = registrationID
		if err := bc.proMetaMgr.Add(projectID, meta); err != nil {
			return errors.Wrap(err, "api controller: set project scanner")
		}
	}

	return nil
}

// GetRegistrationByProject ...
func (bc *basicController) GetRegistrationByProject(projectID int64) (*scanner.Registration, error) {
	if projectID == 0 {
		return nil, errors.New("invalid project ID")
	}

	// First, get it from the project metadata
	m, err := bc.proMetaMgr.Get(projectID, proScannerMetaKey)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: get project scanner")
	}

	if len(m) > 0 {
		if registrationID, ok := m[proScannerMetaKey]; ok && len(registrationID) > 0 {
			registration, err := bc.manager.Get(registrationID)
			if err != nil {
				return nil, errors.Wrap(err, "api controller: get project scanner")
			}

			if registration == nil {
				// Not found
				// Might be deleted by the admin, the project scanner ID reference should be cleared
				if err := bc.proMetaMgr.Delete(projectID, proScannerMetaKey); err != nil {
					return nil, errors.Wrap(err, "api controller: get project scanner")
				}
			} else {
				return registration, nil
			}
		}
	}

	// Second, get the default one
	registration, err := bc.manager.GetDefault()

	// TODO: Check status by the client later
	return registration, err
}
