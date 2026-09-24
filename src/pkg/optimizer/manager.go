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

	"github.com/google/uuid"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/optimizer/dao"
)

// Mgr is the global manager for optimizer registrations
var Mgr = New()

// Manager defines the optimizer registration operations
type Manager interface {
	// Count returns the total count of optimizer registrations according to the query.
	Count(ctx context.Context, query *q.Query) (int64, error)

	// List returns a list of currently configured optimizer registrations.
	List(ctx context.Context, query *q.Query) ([]*dao.Registration, error)

	// Create creates a new optimizer registration with the given data.
	// Returns the registration identifier.
	Create(ctx context.Context, registration *dao.Registration) (string, error)

	// Get returns the details of the specified optimizer registration.
	Get(ctx context.Context, registrationUUID string) (*dao.Registration, error)

	// Update updates the specified optimizer registration.
	Update(ctx context.Context, registration *dao.Registration) error

	// Delete deletes the specified optimizer registration.
	Delete(ctx context.Context, registrationUUID string) error

	// SetAsDefault marks the specified optimizer registration as default.
	// The implementation is supposed to unset any registration previously set as default.
	SetAsDefault(ctx context.Context, registrationUUID string) error

	// GetDefault returns the default optimizer registration or `nil` if there are no
	// registrations configured.
	GetDefault(ctx context.Context) (*dao.Registration, error)

	// DefaultOptimizerUUID returns the default optimizer UUID.
	DefaultOptimizerUUID(ctx context.Context) (string, error)
}

// basicManager is the default implementation of Manager
type basicManager struct{}

// New a basic manager
func New() Manager {
	return &basicManager{}
}

func (bm *basicManager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return dao.GetTotalOfRegistrations(ctx, query)
}

// Create ...
func (bm *basicManager) Create(ctx context.Context, registration *dao.Registration) (string, error) {
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

	if err := encryptCredential(registration); err != nil {
		return "", errors.Wrap(err, "encrypt credential: create registration")
	}

	if _, err := dao.AddRegistration(ctx, registration); err != nil {
		return "", errors.Wrap(err, "dao: create registration")
	}

	return uid.String(), nil
}

// Get ...
func (bm *basicManager) Get(ctx context.Context, registrationUUID string) (*dao.Registration, error) {
	if len(registrationUUID) == 0 {
		return nil, errors.New("empty uuid of registration")
	}

	r, err := dao.GetRegistration(ctx, registrationUUID)
	if err != nil {
		return nil, err
	}
	if r != nil {
		if err := decryptCredential(r); err != nil {
			return nil, errors.Wrap(err, "decrypt credential: get registration")
		}
	}
	return r, nil
}

// Update ...
func (bm *basicManager) Update(ctx context.Context, registration *dao.Registration) error {
	if registration == nil {
		return errors.New("nil registration to update")
	}

	if err := registration.Validate(true); err != nil {
		return errors.Wrap(err, "update registration")
	}

	if err := encryptCredential(registration); err != nil {
		return errors.Wrap(err, "encrypt credential: update registration")
	}

	return dao.UpdateRegistration(ctx, registration)
}

// Delete ...
func (bm *basicManager) Delete(ctx context.Context, registrationUUID string) error {
	if len(registrationUUID) == 0 {
		return errors.New("empty UUID to delete")
	}

	return dao.DeleteRegistration(ctx, registrationUUID)
}

// List ...
func (bm *basicManager) List(ctx context.Context, query *q.Query) ([]*dao.Registration, error) {
	regs, err := dao.ListRegistrations(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, r := range regs {
		if err := decryptCredential(r); err != nil {
			return nil, errors.Wrap(err, "decrypt credential: list registrations")
		}
	}
	return regs, nil
}

// SetAsDefault ...
func (bm *basicManager) SetAsDefault(ctx context.Context, registrationUUID string) error {
	if len(registrationUUID) == 0 {
		return errors.New("empty UUID to set default")
	}

	return dao.SetDefaultRegistration(ctx, registrationUUID)
}

// GetDefault ...
func (bm *basicManager) GetDefault(ctx context.Context) (*dao.Registration, error) {
	r, err := dao.GetDefaultRegistration(ctx)
	if err != nil {
		return nil, err
	}
	if r != nil {
		if err := decryptCredential(r); err != nil {
			return nil, errors.Wrap(err, "decrypt credential: get default registration")
		}
	}
	return r, nil
}

// DefaultOptimizerUUID returns the default optimizer uuid.
func (bm *basicManager) DefaultOptimizerUUID(ctx context.Context) (string, error) {
	reg, err := bm.GetDefault(ctx)
	if err != nil {
		return "", err
	}
	if reg == nil {
		return "", nil
	}
	return reg.UUID, nil
}

// encryptCredential encrypts AccessCredential before persisting to the database.
func encryptCredential(r *dao.Registration) error {
	if len(r.AccessCredential) == 0 {
		return nil
	}
	encrypted, err := config.EncryptSecret(r.AccessCredential)
	if err != nil {
		return err
	}
	r.AccessCredential = encrypted
	return nil
}

// decryptCredential decrypts AccessCredential after reading from the database.
func decryptCredential(r *dao.Registration) error {
	if len(r.AccessCredential) == 0 {
		return nil
	}
	decrypted, err := config.DecryptSecret(r.AccessCredential)
	if err != nil {
		return err
	}
	r.AccessCredential = decrypted
	return nil
}
