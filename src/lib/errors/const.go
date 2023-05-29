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
	// MethodNotAllowedCode ...
	MethodNotAllowedCode = "METHOD_NOT_ALLOWED"
	// RateLimitCode
	RateLimitCode = "TOO_MANY_REQUEST"
	// PreconditionCode ...
	PreconditionCode = "PRECONDITION"
	// GeneralCode ...
	GeneralCode = "UNKNOWN"
	// ChallengesUnsupportedCode ...
	ChallengesUnsupportedCode = "ChallengesUnsupportedCode"
	// DENIED it's used by middleware(readonly, vul and content trust) and returned to docker client to index the request is denied.
	DENIED = "DENIED"
	// PROJECTPOLICYVIOLATION ...
	PROJECTPOLICYVIOLATION = "PROJECTPOLICYVIOLATION"
	// ViolateForeignKeyConstraintCode is the error code for violating foreign key constraint error
	ViolateForeignKeyConstraintCode = "VIOLATE_FOREIGN_KEY_CONSTRAINT"
	// DIGESTINVALID ...
	DIGESTINVALID = "DIGEST_INVALID"
	// MANIFESTINVALID ...
	MANIFESTINVALID = "MANIFEST_INVALID"
	// UNSUPPORTED is for digest UNSUPPORTED error
	UNSUPPORTED = "UNSUPPORTED"
)

// NotFoundError is error for the case of object not found
func NotFoundError(err error) *Error {
	return New("resource not found").WithCode(NotFoundCode).WithCause(err)
}

// ConflictError is error for the case of object conflict
func ConflictError(err error) *Error {
	return New("resource conflict").WithCode(ConflictCode).WithCause(err)
}

// DeniedError is error for the case of denied
func DeniedError(err error) *Error {
	return New("denied").WithCode(DENIED).WithCause(err)
}

// UnauthorizedError is error for the case of unauthorized accessing
func UnauthorizedError(err error) *Error {
	return New("unauthorized").WithCode(UnAuthorizedCode).WithCause(err)
}

// BadRequestError is error for the case of bad request
func BadRequestError(err error) *Error {
	return New("bad request").WithCode(BadRequestCode).WithCause(err)
}

// ForbiddenError is error for the case of forbidden
func ForbiddenError(err error) *Error {
	return New("forbidden").WithCode(ForbiddenCode).WithCause(err)
}

// MethodNotAllowedError is error for the case of forbidden
func MethodNotAllowedError(err error) *Error {
	return New("method not allowed").WithCode(MethodNotAllowedCode).WithCause(err)
}

// PreconditionFailedError is error for the case of precondition failed
func PreconditionFailedError(err error) *Error {
	return New("precondition failed").WithCode(PreconditionCode).WithCause(err)
}

// UnknownError ...
func UnknownError(err error) *Error {
	return New("unknown").WithCode(GeneralCode).WithCause(err)
}

// IsNotFoundErr returns true when the error is NotFoundError
func IsNotFoundErr(err error) bool {
	return IsErr(err, NotFoundCode)
}

// IsConflictErr checks whether the err chain contains conflict error
func IsConflictErr(err error) bool {
	return IsErr(err, ConflictCode)
}

func IsChallengesUnsupportedErr(err error) bool {
	return IsErr(err, ChallengesUnsupportedCode)
}

// IsRateLimitError checks whether the err chains contains rate limit error
func IsRateLimitError(err error) bool {
	return IsErr(err, RateLimitCode)
}
