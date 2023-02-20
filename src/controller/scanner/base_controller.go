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
	"fmt"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/memory" // memory cache
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	rscanner "github.com/goharbor/harbor/src/pkg/scan/scanner"
)

const (
	proScannerMetaKey = "projectScanner"
	statusUnhealthy   = "unhealthy"
	statusHealthy     = "healthy"
)

// DefaultController is a singleton api controller for plug scanners
var DefaultController = New()

// New a basic controller
func New() Controller {
	return &basicController{
		manager:    rscanner.New(),
		proMetaMgr: pkg.ProjectMetaMgr,
		clientPool: v1.DefaultClientPool,
	}
}

// basicController is default implementation of api.Controller interface
type basicController struct {
	sync.Once

	// Managers for managing the scanner registrations
	manager rscanner.Manager
	// For operating the project level configured scanner
	proMetaMgr metadata.Manager
	// Client pool for talking to adapters
	clientPool v1.ClientPool
	// Cache of the scanner metadata
	cache cache.Cache
}

func (bc *basicController) Cache() cache.Cache {
	bc.Do(func() {
		bc.cache, _ = cache.New(cache.Memory, cache.Expiration(time.Second*30))
	})

	return bc.cache
}

// ListRegistrations ...
func (bc *basicController) ListRegistrations(ctx context.Context, query *q.Query) ([]*scanner.Registration, error) {
	l, err := bc.manager.List(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: list registrations")
	}

	return l, nil
}

// Count returns the total count of scanner registrations according to the query.
func (bc *basicController) GetTotalOfRegistrations(ctx context.Context, query *q.Query) (int64, error) {
	return bc.manager.Count(ctx, query)
}

// CreateRegistration ...
func (bc *basicController) CreateRegistration(ctx context.Context, registration *scanner.Registration) (string, error) {
	if isReservedName(registration.Name) {
		return "", errors.BadRequestError(nil).WithMessage(`name "%s" is reserved, please try a different name`, registration.Name)
	}

	// Check if the registration is available
	if _, err := bc.Ping(ctx, registration); err != nil {
		return "", errors.Wrap(err, "api controller: create registration")
	}

	// Check if there are any registrations already existing.
	l, err := bc.manager.List(ctx, &q.Query{
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

	return bc.manager.Create(ctx, registration)
}

// GetRegistration ...
func (bc *basicController) GetRegistration(ctx context.Context, registrationUUID string) (*scanner.Registration, error) {
	r, err := bc.manager.Get(ctx, registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: get registration")
	}

	return r, nil
}

// RegistrationExists ...
func (bc *basicController) RegistrationExists(ctx context.Context, registrationUUID string) bool {
	registration, err := bc.manager.Get(ctx, registrationUUID)

	// Just logged when an error occurred
	if err != nil {
		logger.Errorf("Check existence of registration error: %s", err)
	}

	return !(err == nil && registration == nil)
}

// UpdateRegistration ...
func (bc *basicController) UpdateRegistration(ctx context.Context, registration *scanner.Registration) error {
	if registration.IsDefault && registration.Disabled {
		return errors.Errorf("default registration %s can not be marked to deactivated", registration.UUID)
	}

	if isReservedName(registration.Name) {
		return errors.BadRequestError(nil).WithMessage(`name "%s" is reserved, please try a different name`, registration.Name)
	}

	return bc.manager.Update(ctx, registration)
}

// SetDefaultRegistration ...
func (bc *basicController) DeleteRegistration(ctx context.Context, registrationUUID string) (*scanner.Registration, error) {
	registration, err := bc.manager.Get(ctx, registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: delete registration")
	}

	if registration == nil {
		// Not found
		return nil, nil
	}

	if err := bc.manager.Delete(ctx, registrationUUID); err != nil {
		return nil, errors.Wrap(err, "api controller: delete registration")
	}

	return registration, nil
}

// SetDefaultRegistration ...
func (bc *basicController) SetDefaultRegistration(ctx context.Context, registrationUUID string) error {
	return bc.manager.SetAsDefault(ctx, registrationUUID)
}

// SetRegistrationByProject ...
func (bc *basicController) SetRegistrationByProject(ctx context.Context, projectID int64, registrationID string) error {
	if projectID == 0 {
		return errors.New("invalid project ID")
	}

	if len(registrationID) == 0 {
		return errors.New("missing scanner UUID")
	}

	// Only keep the UUID in the metadata of the given project
	// Scanner metadata existing?
	m, err := bc.proMetaMgr.Get(ctx, projectID, proScannerMetaKey)
	if err != nil {
		return errors.Wrap(err, "api controller: set project scanner")
	}

	// Update if exists
	if len(m) > 0 {
		// Compare and set new
		if registrationID != m[proScannerMetaKey] {
			m[proScannerMetaKey] = registrationID
			if err := bc.proMetaMgr.Update(ctx, projectID, m); err != nil {
				return errors.Wrap(err, "api controller: set project scanner")
			}
		}
	} else {
		meta := make(map[string]string, 1)
		meta[proScannerMetaKey] = registrationID
		if err := bc.proMetaMgr.Add(ctx, projectID, meta); err != nil {
			return errors.Wrap(err, "api controller: set project scanner")
		}
	}

	return nil
}

// GetRegistrationByProject ...
func (bc *basicController) GetRegistrationByProject(ctx context.Context, projectID int64, options ...Option) (*scanner.Registration, error) {
	if projectID == 0 {
		return nil, errors.New("invalid project ID")
	}

	// First, get it from the project metadata
	m, err := bc.proMetaMgr.Get(ctx, projectID, proScannerMetaKey)
	if err != nil {
		return nil, errors.Wrap(err, "api controller: get project scanner")
	}

	var registration *scanner.Registration
	if len(m) > 0 {
		if registrationID, ok := m[proScannerMetaKey]; ok && len(registrationID) > 0 {
			registration, err = bc.manager.Get(ctx, registrationID)
			if err != nil {
				return nil, errors.Wrap(err, "api controller: get project scanner")
			}

			if registration == nil {
				// Not found
				// Might be deleted by the admin, the project scanner ID reference should be cleared
				if err := bc.proMetaMgr.Delete(ctx, projectID, proScannerMetaKey); err != nil {
					return nil, errors.Wrap(err, "api controller: get project scanner")
				}
			}
		}
	}

	if registration == nil {
		// Second, get the default one
		registration, err = bc.manager.GetDefault(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "api controller: get project scanner")
		}
	}

	// No scanner configured
	if registration == nil {
		return nil, nil
	}

	opts := newOptions(options...)

	if opts.Ping {
		// Get metadata of the configured registration
		meta, err := bc.Ping(ctx, registration)
		if err != nil {
			// Not blocked, just logged it
			log.Error(errors.Wrap(err, "api controller: get project scanner"))
			registration.Health = statusUnhealthy
		} else {
			registration.Health = statusHealthy
			// Fill in some metadata
			registration.Adapter = meta.Scanner.Name
			registration.Vendor = meta.Scanner.Vendor
			registration.Version = meta.Scanner.Version

			registration.Metadata = meta
		}
	}

	return registration, nil
}

// Ping ...
func (bc *basicController) Ping(ctx context.Context, registration *scanner.Registration) (*v1.ScannerAdapterMetadata, error) {
	if registration == nil {
		return nil, errors.New("nil registration to ping")
	}

	var (
		err  error
		meta *v1.ScannerAdapterMetadata
	)

	if registration.ID > 0 {
		meta, err = bc.getScannerAdapterMetadataWithCache(ctx, registration)
	} else {
		meta, err = bc.getScannerAdapterMetadata(registration)
	}

	if err != nil {
		log.G(ctx).WithField("error", err).Error("failed to ping scanner")

		return nil, errors.Wrap(err, "scanner controller: ping")
	}

	if err := meta.Validate(); err != nil {
		return nil, err
	}

	return meta, nil
}

// GetMetadata ...
func (bc *basicController) GetMetadata(ctx context.Context, registrationUUID string) (*v1.ScannerAdapterMetadata, error) {
	if len(registrationUUID) == 0 {
		return nil, errors.New("empty registration uuid")
	}

	r, err := bc.manager.Get(ctx, registrationUUID)
	if err != nil {
		return nil, errors.Wrap(err, "scanner controller: get metadata")
	}

	if r == nil {
		return nil, errors.NotFoundError(nil).WithMessage("registration %s not found", registrationUUID)
	}

	return bc.Ping(ctx, r)
}

func (bc *basicController) getScannerAdapterMetadata(registration *scanner.Registration) (*v1.ScannerAdapterMetadata, error) {
	client, err := registration.Client(bc.clientPool)
	if err != nil {
		return nil, err
	}

	return client.GetMetadata()
}

func (bc *basicController) getScannerAdapterMetadataWithCache(ctx context.Context, registration *scanner.Registration) (*v1.ScannerAdapterMetadata, error) {
	key := fmt.Sprintf("reg:%d:metadata", registration.ID)

	var result MetadataResult
	err := cache.FetchOrSave(ctx, bc.Cache(), key, &result, func() (interface{}, error) {
		meta, err := bc.getScannerAdapterMetadata(registration)
		if err != nil {
			return &MetadataResult{Error: err.Error()}, nil
		}

		return &MetadataResult{Metadata: meta}, nil
	})

	if err != nil {
		return nil, err
	}

	return result.Unpack()
}

var (
	reservedNames = []string{"Trivy"}
)

func isReservedName(name string) bool {
	for _, reservedName := range reservedNames {
		if name == reservedName {
			return true
		}
	}

	return false
}

// MetadataResult metadata or error saved in cache
type MetadataResult struct {
	Metadata *v1.ScannerAdapterMetadata
	Error    string
}

// Unpack get ScannerAdapterMetadata and error from the result
func (m *MetadataResult) Unpack() (*v1.ScannerAdapterMetadata, error) {
	var err error
	if m.Error != "" {
		err = fmt.Errorf(m.Error)
	}

	return m.Metadata, err
}
