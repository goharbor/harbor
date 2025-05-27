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

package runner

import (
	"reflect"

	"github.com/goharbor/harbor/src/jobservice/job"
)

// Wrap returns a new job.Interface based on the wrapped job handler reference.
func Wrap(j any) job.Interface {
	theType := reflect.TypeOf(j)

	if theType.Kind() == reflect.Ptr {
		theType = theType.Elem()
	}

	// Crate new
	v := reflect.New(theType).Elem()
	return v.Addr().Interface().(job.Interface)
}
