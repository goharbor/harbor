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
	"strconv"
)

// BoolValue returns the value of the bool pointer or false if the pointer is nil
func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}

// Int64Value returns the value of the int64 pointer or 0 if the pointer is nil
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}

// StringValue returns the value of the string pointer or "" if the pointer is nil
func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// ToBool convert interface to bool
func ToBool(v interface{}) bool {
	switch b := v.(type) {
	case bool:
		return b
	case nil:
		return false
	case int:
		return v.(int) != 0
	case int64:
		return v.(int64) != 0
	case string:
		r, _ := strconv.ParseBool(v.(string))
		return r
	default:
		return false
	}
}
