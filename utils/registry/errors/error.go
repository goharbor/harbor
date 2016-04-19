/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
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

package errors

import (
	"fmt"
)

// Error : if response's status code is not 200 or does not meet requirement,
// an Error instance will be returned
type Error struct {
	StatusCode int
	Message    string
}

// Error ...
func (e Error) Error() string {
	return fmt.Sprintf("%d %s", e.StatusCode, e.Message)
}

// ParseError parses err, if err is type Error, convert it to Error
func ParseError(err error) (Error, bool) {
	e, ok := err.(Error)
	return e, ok
}
