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

// EnsureScanners ensures that the scanners with the specified endpoints URLs exist in the system.
func EnsureScanners(wantedScanners []scanner.Registration) (err error) {
	if len(wantedScanners) == 0 {
		return
	}
	endpointURLs := make([]string, len(wantedScanners))
	for i, ws := range wantedScanners {
		endpointURLs[i] = ws.URL
	}

	list, err := scannerManager.List(&q.Query{
		Keywords: map[string]interface{}{
			"ex_url__in": endpointURLs,
		},
	})
	if err != nil {
		return errors.Errorf("listing scanners: %v", err)
	}
	existingScanners := make(map[string]*scanner.Registration)
	for _, li := range list {
		existingScanners[li.URL] = li
	}

	for _, ws := range wantedScanners {
		if _, exists := existingScanners[ws.URL]; exists {
			log.Infof("Scanner registration already exists: %s", ws.URL)
			continue
		}
		err = createRegistration(&ws, true)
		if err != nil {
			return errors.Errorf("creating registration: %s: %v", ws.URL, err)
		}
		log.Infof("Successfully registered %s scanner at %s", ws.Name, ws.URL)
	}

	return
}

// EnsureDefaultScanner ensures that the scanner with the specified URL is set as default in the system.
func EnsureDefaultScanner(scannerURL string) (err error) {
	defaultScanner, err := scannerManager.GetDefault()
	if err != nil {
		err = errors.Errorf("getting default scanner: %v", err)
		return
	}
	if defaultScanner != nil && defaultScanner.URL == scannerURL {
		log.Infof("The default scanner is already set: %s", defaultScanner.URL)
		return
	}
	scanners, err := scannerManager.List(&q.Query{
		Keywords: map[string]interface{}{"url": scannerURL},
	})
	if err != nil {
		err = errors.Errorf("listing scanners: %v", err)
		return
	}
	if len(scanners) != 1 {
		return errors.Errorf("expected only one scanner with URL %v but got %d", scannerURL, len(scanners))
	}
	err = scannerManager.SetAsDefault(scanners[0].UUID)
	if err != nil {
		err = errors.Errorf("setting %s as default scanner: %v", scannerURL, err)
	}
	return
}

func createRegistration(registration *scanner.Registration, resolveConflict bool) (err error) {
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
	return
}

// RemoveImmutableScanners removes immutable scanner Registrations with the specified endpoint URLs.
func RemoveImmutableScanners(urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	query := &q.Query{
		Keywords: map[string]interface{}{
			"immutable":  true,
			"ex_url__in": urls,
		},
	}

	// TODO Instead of executing 1 to N SQL queries we might want to delete multiple rows with scannerManager.DeleteByImmutableAndURLIn(true, []string{})
	registrations, err := scannerManager.List(query)
	if err != nil {
		return errors.Errorf("listing scanners: %v", err)
	}

	for _, reg := range registrations {
		if err := scannerManager.Delete(reg.UUID); err != nil {
			return errors.Errorf("deleting scanner: %s: %v", reg.UUID, err)
		}
	}

	return nil
}
