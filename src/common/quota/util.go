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
	"fmt"

	"github.com/goharbor/harbor/src/pkg/types"
)

type unsafe struct {
	message string
}

func (err *unsafe) Error() string {
	return err.message
}

func newUnsafe(message string) error {
	return &unsafe{message: message}
}

// IsUnsafeError returns true when the err is unsafe error
func IsUnsafeError(err error) bool {
	_, ok := err.(*unsafe)
	return ok
}

func isSafe(hardLimits types.ResourceList, used types.ResourceList) error {
	for key, value := range used {
		if value < 0 {
			return newUnsafe(fmt.Sprintf("bad used value: %d", value))
		}

		if hard, found := hardLimits[key]; found {
			if hard == types.UNLIMITED {
				continue
			}

			if value > hard {
				return newUnsafe(fmt.Sprintf("over the quota: used %d but only hard %d", value, hard))
			}
		} else {
			return newUnsafe(fmt.Sprintf("hard limit not found: %s", key))
		}

	}

	return nil
}
