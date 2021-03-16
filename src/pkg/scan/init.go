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
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	sc "github.com/goharbor/harbor/src/pkg/scan/scanner"
)

var (
	scannerManager = sc.New()
)

// EnsureScanners ensures that the scanners with the specified endpoints URLs exist in the system.
func EnsureScanners(ctx context.Context, wantedScanners []scanner.Registration) (err error) {
	if len(wantedScanners) == 0 {
		return
	}
	names := make([]string, len(wantedScanners))
	for i, ws := range wantedScanners {
		names[i] = ws.Name
	}

	list, err := scannerManager.List(ctx, q.New(q.KeyWords{"name__in": names}))
	if err != nil {
		return errors.Errorf("listing scanners: %v", err)
	}
	existingScanners := make(map[string]*scanner.Registration)
	for _, li := range list {
		existingScanners[li.Name] = li
	}

	for _, ws := range wantedScanners {
		scanner, exists := existingScanners[ws.Name]
		if !exists {
			if _, err := scannerManager.Create(ctx, &ws); err != nil {
				return errors.Errorf("creating registration %s at %s failed: %v", ws.Name, ws.URL, err)
			}
			log.Infof("Successfully registered %s scanner at %s", ws.Name, ws.URL)
		} else if scanner.URL != ws.URL {
			scanner.URL = ws.URL
			if err := scannerManager.Update(ctx, scanner); err != nil {
				return errors.Errorf("updating registration %s to %s failed: %v", ws.Name, ws.URL, err)
			}
			log.Infof("Successfully updated %s scanner to %s", ws.Name, ws.URL)
		} else {
			log.Infof("Scanner registration already exists: %s", ws.URL)
		}
	}

	return
}

// EnsureDefaultScanner ensures that the scanner with the specified URL is set as default in the system.
func EnsureDefaultScanner(ctx context.Context, scannerName string) (err error) {
	defaultScanner, err := scannerManager.GetDefault(ctx)
	if err != nil {
		err = errors.Errorf("getting default scanner: %v", err)
		return
	}
	if defaultScanner != nil {
		log.Infof("Skipped setting %s as the default scanner. The default scanner is already set to %s", scannerName, defaultScanner.URL)
		return
	}
	scanners, err := scannerManager.List(ctx, q.New(q.KeyWords{"name": scannerName}))
	if err != nil {
		err = errors.Errorf("listing scanners: %v", err)
		return
	}
	if len(scanners) != 1 {
		return errors.Errorf("expected only one scanner with name %v but got %d", scannerName, len(scanners))
	}
	err = scannerManager.SetAsDefault(ctx, scanners[0].UUID)
	if err != nil {
		err = errors.Errorf("setting %s as default scanner: %v", scannerName, err)
	}
	return
}

// RemoveImmutableScanners removes immutable scanner Registrations with the specified endpoint URLs.
func RemoveImmutableScanners(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}
	query := q.New(q.KeyWords{"immutable": true, "name__in": names})

	// TODO Instead of executing 1 to N SQL queries we might want to delete multiple rows with scannerManager.DeleteByImmutableAndURLIn(true, []string{})
	registrations, err := scannerManager.List(ctx, query)
	if err != nil {
		return errors.Errorf("listing scanners: %v", err)
	}

	for _, reg := range registrations {
		if err := scannerManager.Delete(ctx, reg.UUID); err != nil {
			return errors.Errorf("deleting scanner: %s: %v", reg.UUID, err)
		}
	}

	return nil
}
