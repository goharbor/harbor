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

package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/goharbor/harbor/src/common/utils/log"
)

var (
	// As alias function of `errors.As`
	As = errors.As
	// Is alias function of `errors.Is`
	Is = errors.Is
)

// Error ...
type Error struct {
	Cause   error  `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns a human readable error.
func (e *Error) Error() string {
	var parts []string

	var causeStr string
	if e.Cause != nil {
		causeStr = e.Cause.Error()
		parts = append(parts, causeStr)
	}

	if e.Code != "" {
		parts = append(parts, e.Code)
	}

	if e.Message != causeStr {
		parts = append(parts, e.Message)
	}

	return strings.Join(parts, ", ")
}

// WithMessage ...
func (e *Error) WithMessage(format string, v ...interface{}) *Error {
	e.Message = fmt.Sprintf(format, v...)
	return e
}

// WithCode ...
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// WithCause ...
func (e *Error) WithCause(err error) *Error {
	e.Cause = err
	return e
}

// Unwrap ...
func (e *Error) Unwrap() error { return e.Cause }

// Errors ...
type Errors []error

var _ error = Errors{}

// Error converts slice of error
func (errs Errors) Error() string {
	var tmpErrs struct {
		Errors []Error `json:"errors,omitempty"`
	}

	for _, e := range errs {
		err, ok := e.(*Error)
		if !ok {
			err = UnknownError(e).WithMessage(e.Error())
		}
		if err.Code == "" {
			err.Code = GeneralCode
		}

		tmpErrs.Errors = append(tmpErrs.Errors, *err)
	}

	msg, err := json.Marshal(tmpErrs)
	if err != nil {
		log.Error(err)
		return "{}"
	}
	return string(msg)
}

// Len returns the current number of errors.
func (errs Errors) Len() int {
	return len(errs)
}

// NewErrs ...
func NewErrs(err error) Errors {
	return Errors{err}
}

const (
	// NotFoundCode is code for the error of no object found
	NotFoundCode = "NOT_FOUND"
	// ConflictCode ...
	ConflictCode = "CONFLICT"
	// UnAuthorizedCode ...
	UnAuthorizedCode = "UNAUTHORIZED"
	// BadRequestCode ...
	BadRequestCode = "BAD_REQUEST"
	// ForbiddenCode ...
	ForbiddenCode = "FORBIDDEN"
	// PreconditionCode ...
	PreconditionCode = "PRECONDITION"
	// GeneralCode ...
	GeneralCode = "UNKNOWN"
	// DENIED it's used by middleware(readonly, vul and content trust) and returned to docker client to index the request is denied.
	DENIED = "DENIED"
	// PROJECTPOLICYVIOLATION ...
	PROJECTPOLICYVIOLATION = "PROJECTPOLICYVIOLATION"
	// ViolateForeignKeyConstraintCode is the error code for violating foreign key constraint error
	ViolateForeignKeyConstraintCode = "VIOLATE_FOREIGN_KEY_CONSTRAINT"
	// DIGESTINVALID ...
	DIGESTINVALID = "DIGEST_INVALID"
)

// New ...
func New(in interface{}) *Error {
	var err error
	switch in := in.(type) {
	case error:
		err = in
	case *Error:
		err = in.Cause
	default:
		err = fmt.Errorf("%v", in)
	}
	return &Error{
		Cause:   err,
		Message: err.Error(),
	}
}

// Wrap ...
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	e := &Error{
		Cause:   err,
		Message: message,
	}
	return e
}

// Wrapf ...
func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	e := &Error{
		Cause: err,
	}
	return e.WithMessage(format, args...)
}

// Errorf ...
func Errorf(format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
	}
}

// NotFoundError is error for the case of object not found
func NotFoundError(err error) *Error {
	return New(err).WithCode(NotFoundCode).WithMessage("resource not found")
}

// ConflictError is error for the case of object conflict
func ConflictError(err error) *Error {
	return New(err).WithCode(ConflictCode).WithMessage("resource conflict")
}

// DeniedError is error for the case of denied
func DeniedError(err error) *Error {
	return New(err).WithCode(DENIED).WithMessage("denied")
}

// UnauthorizedError is error for the case of unauthorized accessing
func UnauthorizedError(err error) *Error {
	return New(err).WithCode(UnAuthorizedCode).WithMessage("unauthorized")
}

// BadRequestError is error for the case of bad request
func BadRequestError(err error) *Error {
	return New(err).WithCode(BadRequestCode).WithMessage("bad request")
}

// ForbiddenError is error for the case of forbidden
func ForbiddenError(err error) *Error {
	return New(err).WithCode(ForbiddenCode).WithMessage("forbidden")
}

// PreconditionFailedError is error for the case of precondition failed
func PreconditionFailedError(err error) *Error {
	return New(err).WithCode(PreconditionCode).WithMessage("precondition failed")
}

// UnknownError ...
func UnknownError(err error) *Error {
	return New(err).WithCode(GeneralCode).WithMessage("unknown")
}

// Cause gets the root error
func Cause(err error) error {
	for err != nil {
		cause, ok := err.(*Error)
		if !ok {
			break
		}
		err = cause.Cause
	}
	return err
}

// IsErr checks whether the err chain contains error matches the code
func IsErr(err error, code string) bool {
	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}

// IsNotFoundErr returns true when the error is NotFoundError
func IsNotFoundErr(err error) bool {
	return IsErr(err, NotFoundCode)
}

// IsConflictErr checks whether the err chain contains conflict error
func IsConflictErr(err error) bool {
	return IsErr(err, ConflictCode)
}

// ErrCode returns code of err
func ErrCode(err error) string {
	if err == nil {
		return ""
	}

	var e *Error
	if ok := errors.As(err, &e); ok && e.Code != "" {
		return e.Code
	} else if ok && e.Cause != nil {
		return ErrCode(e.Cause)
	}

	return GeneralCode
}
