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

package errs

import (
	"encoding/json"
	"fmt"
)

const (
	// Common error code
	Common uint16 = 10000
	// Conflict error code
	Conflict uint16 = 10409
	// PreconditionFailed error code
	PreconditionFailed uint16 = 10412
)

// codeTexts behaviors as a hash map to look for the text for the given error code.
func codeTexts(code uint16) string {
	switch code {
	case Common:
		return "common"
	case Conflict:
		return "not found"
	case PreconditionFailed:
		return "Precondition failed"
	default:
		return "unknown"
	}
}

// Error with code
type Error struct {
	// Code of error
	Code uint16 `json:"code"`
	// Code represented by meaningful text
	TextCode string `json:"text_code"`
	// Message of error
	Message string `json:"message"`
	// Cause for error
	Cause error `json:"cause"`
}

// Error message
func (e *Error) Error() string {
	emsg := fmt.Sprintf("error: code %d:%s : %s", e.Code, e.TextCode, e.Message)
	if e.Cause != nil {
		emsg = fmt.Sprintf("%s : cause: %s", emsg, e.Cause.Error())
	}

	return emsg
}

// String outputs the error with well-formatted string.
func (e *Error) String() string {
	bytes, err := json.Marshal(e)
	if err != nil {
		// Fallback to normal string
		return e.Error()
	}

	return string(bytes)
}

// New common error.
func New(message string) error {
	return &Error{
		Code:     Common,
		TextCode: codeTexts(Common),
		Message:  message,
	}
}

// Wrap error with message.
func Wrap(err error, message string) error {
	return &Error{
		Code:     Common,
		TextCode: codeTexts(Common),
		Message:  message,
		Cause:    err,
	}
}

// Errorf new a message with the specified format and arguments
func Errorf(format string, args ...interface{}) error {
	return &Error{
		Code:     Common,
		TextCode: codeTexts(Common),
		Message:  fmt.Sprintf(format, args...),
	}
}

// WithCode sets specified code for the error
func WithCode(code uint16, err error) error {
	if err == nil {
		return err
	}

	e, ok := err.(*Error)
	if !ok {
		return err
	}

	e.Code = code
	e.TextCode = codeTexts(code)

	return e
}

// AsError checks if the given error has the given code
func AsError(err error, code uint16) bool {
	if err == nil {
		return false
	}

	e, ok := err.(*Error)

	return ok && e.Code == code
}
