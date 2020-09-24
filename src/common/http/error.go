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

package http

import (
	"encoding/json"
	"fmt"
)

// Error wrap HTTP status code and message as an error
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error ...
func (e *Error) Error() string {
	return fmt.Sprintf("http error: code %d, message %s", e.Code, e.Message)
}

// String wraps the error msg to the well formatted error message
func (e *Error) String() string {
	data, err := json.Marshal(&e)
	if err != nil {
		return e.Message
	}
	return string(data)
}
