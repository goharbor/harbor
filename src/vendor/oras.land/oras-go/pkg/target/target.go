/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package target

import (
	"github.com/containerd/containerd/remotes"
)

// Target represents a place to which one can send/push or retrieve/pull artifacts.
// Anything that implements the Target interface can be used as a place to send or
// retrieve artifacts.
type Target interface {
	remotes.Resolver
}
