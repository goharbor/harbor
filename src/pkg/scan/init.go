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
)

var (
	scannerManager = sc.New()
)

// EnsureScanner ensure the scanner which specially name exists in the system
func EnsureScanner(registration *scanner.Registration) error {
	q := &q.Query{
		Keywords: map[string]interface{}{"url": registration.URL},
	}

	// Check if the registration with the url already existing.
	registrations, err := scannerManager.List(q)
	if err != nil {
		return err
	}

	if len(registrations) == 0 {
		defaultReg, err := scannerManager.GetDefault()
		if err != nil {
			return fmt.Errorf("failed to get the default scanner, error: %v", err)
		}

		// Set the registration to be default one when no default registration exist in the system
		registration.IsDefault = defaultReg == nil

		if _, err := scannerManager.Create(registration); err != nil {
			return err
		}

		log.Infof("initialized scanner named %s", registration.Name)
	}

	return nil
}
