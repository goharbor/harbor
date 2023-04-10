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

package getter

// Interface defines operations of a log data getter
type Interface interface {
	// Retrieve the log data of the specified log entry
	//
	// logID string : the id of the log entry. e.g: file name a.log for file log
	//
	// If succeed, log data bytes will be returned
	// otherwise, a non nil error is returned
	Retrieve(logID string) ([]byte, error)
}
