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

package sweeper

// Interface defines the operations a sweeper should have
type Interface interface {
	// Sweep the outdated log entries if necessary
	//
	// If failed, an non-nil error will return
	// If succeeded, count of sweepped log entries is returned
	Sweep() (int, error)

	// Return the sweeping duration with day unit.
	Duration() int
}
