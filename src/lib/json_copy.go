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
	"encoding/json"
)

// JSONCopy copy from src to dst with json marshal and unmarshal.
// NOTE: copy from one struct to another struct may miss some values depend on
// the json tag in the fields, you should know what will happened when call this function.
func JSONCopy(dst, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dst)
}
