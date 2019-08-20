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

package quota

import (
	"sort"
	"strings"

	"github.com/goharbor/harbor/src/pkg/types"
)

func isSafe(hardLimits types.ResourceList, currentUsed types.ResourceList, newUsed types.ResourceList) error {
	var errs Errors

	for resource, value := range newUsed {
		hardLimit, found := hardLimits[resource]
		if !found {
			errs = errs.Add(NewResourceNotFoundError(resource))
			continue
		}

		if hardLimit == types.UNLIMITED || value == currentUsed[resource] {
			continue
		}

		if value > hardLimit {
			errs = errs.Add(NewResourceOverflowError(resource, hardLimit, currentUsed[resource], value))
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func prettyPrintResourceNames(a []types.ResourceName) string {
	values := []string{}
	for _, value := range a {
		values = append(values, string(value))
	}
	sort.Strings(values)
	return strings.Join(values, ",")
}
