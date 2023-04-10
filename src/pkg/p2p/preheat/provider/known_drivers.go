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

package provider

import (
	"sort"
	"strings"
)

const (
	// DriverDragonfly represents the driver for dragonfly
	DriverDragonfly = "dragonfly"
	// DriverKraken represents the driver for kraken
	DriverKraken = "kraken"
)

// knownDrivers is static driver Factory registry
var knownDrivers = map[string]Factory{
	DriverDragonfly: DragonflyFactory,
	DriverKraken:    KrakenFactory,
}

// ListProviders returns all the registered drivers.
func ListProviders() ([]*Metadata, error) {
	results := make([]*Metadata, 0)

	for _, f := range knownDrivers {
		drv, err := f(nil)
		if err != nil {
			return nil, err
		}

		results = append(results, drv.Self())
	}

	// Sort results
	if len(results) > 1 {
		sort.SliceStable(results, func(i, j int) bool {
			return strings.Compare(results[i].ID, results[j].ID) < 0
		})
	}

	return results, nil
}

// GetProvider returns the driver factory identified by the ID.
//
// If exists, bool flag will be set to be true and a non-nil reference will be returned.
func GetProvider(ID string) (Factory, bool) {
	f, ok := knownDrivers[ID]

	return f, ok
}
