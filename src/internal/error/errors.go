package error

import (
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/log"
	"strings"
)

// Error ...
type Error struct {
	Cause   error  `json:"-"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error returns a human readable error.
func (e *Error) Error() string {
	return fmt.Sprintf("%v, %s, %s", e.Cause, e.Code, e.Message)
}

// WithMessage ...
func (e *Error) WithMessage(msg string) *Error {
	e.Message = msg
	return e
}

// WithCode ...
func (e *Error) WithCode(code string) *Error {
	e.Code = code
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
		var err error
		switch e.(type) {
		case *Error:
			err = e.(*Error)
		default:
			err = UnknownError(e).WithMessage(err.Error())
		}
		tmpErrs.Errors = append(tmpErrs.Errors, *err.(*Error))
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
	ForbiddenCode = "FORBIDDER"
	// PreconditionCode ...
	PreconditionCode = "PRECONDITION"
	// GeneralCode ...
	GeneralCode = "UNKNOWN"
)

// New ...
func New(err error) *Error {
	if _, ok := err.(*Error); ok {
		err = err.(*Error).Unwrap()
	}
	return &Error{
		Cause:   err,
		Message: err.Error(),
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
	return New(err).WithCode(PreconditionCode).WithMessage("preconfition")
}

// UnknownError ...
func UnknownError(err error) *Error {
	return New(err).WithCode(GeneralCode).WithMessage("unknown")
}

// IsErr ...
func IsErr(err error, code string) bool {
	_, ok := err.(*Error)
	if !ok {
		return false
	}
	return strings.Compare(err.(*Error).Code, code) == 0
}
