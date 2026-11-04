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
	"crypto/rand"
	"math/big"
)

// ShuffleStringSlice shuffles the string slice in place using crypto-secure randomness.
func ShuffleStringSlice(slice []string) {
	for i := len(slice) - 1; i > 0; i-- {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			// fallback to deterministic shuffle on error (should not happen)
			continue
		}
		j := int(jBig.Int64())
		slice[i], slice[j] = slice[j], slice[i]
	}
}
