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

package lib

import (
	"fmt"

	"github.com/goharbor/harbor/src/pkg/quota/types"
)

func ValidateQuotaLimit(storageLimit int64) error {
	if storageLimit <= 0 {
		if storageLimit != types.UNLIMITED {
			return fmt.Errorf("invalid non-positive value for quota limit, value=%v", storageLimit)
		}
	} else {
		// storageLimit > 0, there is a max capacity of limited storage
		if uint64(storageLimit) > types.MaxLimitedValue {
			return fmt.Errorf("exceeded 1024TB, which is 1125899906842624 Bytes, value=%v", storageLimit)
		}
	}
	return nil
}
