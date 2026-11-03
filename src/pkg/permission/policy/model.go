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

package policy

// Namespace is the object space a project grant is written against.
//
// A role grants the same thing in every project it is held in, so one row has
// to answer for every project and the project id is a placeholder. :pid matches
// a single path segment, so a grant on the project itself stops there instead
// of reaching everything below it, which is what /project/* would do.
const Namespace = "/project/:pid"

// Object is the casbin object a resource is granted against.
func Object(resource string) string {
	if resource == "" {
		return Namespace
	}
	return Namespace + "/" + resource
}
