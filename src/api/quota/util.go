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
	"strconv"
)

const (
	// ProjectReference reference type for project
	ProjectReference = "project"
)

// ReferenceID returns reference id for the interface
func ReferenceID(i interface{}) string {
	switch s := i.(type) {
	case string:
		return s
	case int64:
		return strconv.FormatInt(s, 10)
	case fmt.Stringer:
		return s.String()
	case error:
		return s.Error()
	default:
		return fmt.Sprintf("%v", i)
	}
}
