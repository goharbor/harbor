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
	"fmt"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	sc "github.com/goharbor/harbor/src/pkg/scan/scanner"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

var (
	scannerManager = sc.New()
)

// EnsureScanner ensure the scanner which specially endpoint exists in the system
func EnsureScanner(registration *scanner.Registration, resolveConflicts ...bool) error {
	q := &q.Query{
		Keywords: map[string]interface{}{"url": registration.URL},
	}

	// Check if the registration with the url already existing.
	registrations, err := scannerManager.List(q)
	if err != nil {
		return err
	}

	if len(registrations) > 0 {
		return nil
	}

	var resolveConflict bool
	if len(resolveConflicts) > 0 {
		resolveConflict = resolveConflicts[0]
	}

	var defaultReg *scanner.Registration
	defaultReg, err = scannerManager.GetDefault()
	if err != nil {
		return fmt.Errorf("failed to get the default scanner, error: %v", err)
	}

	// Set the registration to be default one when no default registration exist in the system
	registration.IsDefault = defaultReg == nil

	for {
		_, err = scannerManager.Create(registration)
		if err != nil {
			if resolveConflict && errors.Cause(err) == types.ErrDupRows {
				var id uuid.UUID
				id, err = uuid.NewUUID()
				if err != nil {
					break
				}

				registration.Name = registration.Name + "-" + id.String()
				resolveConflict = false
				continue
			}
		}

		break
	}

	if err == nil {
		log.Infof("initialized scanner named %s", registration.Name)
	}

	return err
}

// RemoveImmutableScanner removes an immutable scanner with the specified endpoint URL.
func RemoveImmutableScanner(registrationURL string) error {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"immutable": true,
			"url":       registrationURL,
		},
	}

	registrations, err := scannerManager.List(query)
	if err != nil {
		return err
	}

	for _, reg := range registrations {
		if err := scannerManager.Delete(reg.UUID); err != nil {
			return err
		}
	}

	return nil
}
