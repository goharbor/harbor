package error

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/utils/log"
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
func (e *Error) WithMessage(format string, v ...interface{}) *Error {
	e.Message = fmt.Sprintf(format, v...)
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
			err = UnknownError(e).WithMessage(e.Error())
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
)

// New ...
func New(err error) *Error {
	e := &Error{}
	if err != nil {
		e.Cause = err
		e.Message = err.Error()
		if ee, ok := err.(*Error); ok {
			e.Cause = ee
		}
	}
	return e
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
	return New(err).WithCode(PreconditionCode).WithMessage("precondition failed")
}

// UnknownError ...
func UnknownError(err error) *Error {
	return New(err).WithCode(GeneralCode).WithMessage("unknown")
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
