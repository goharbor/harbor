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

// Package errs define some system errors with specified types.
package errs

import (
	"encoding/json"
	"fmt"
)

const (
	// ReadRequestBodyErrorCode is code for the error of reading http request body error
	ReadRequestBodyErrorCode = 10000 + iota
	// HandleJSONDataErrorCode is code for the error of handling json data error
	HandleJSONDataErrorCode
	// MissingBackendHandlerErrorCode is code for the error of missing backend controller
	MissingBackendHandlerErrorCode
	// LaunchJobErrorCode is code for the error of launching job
	LaunchJobErrorCode
	// CheckStatsErrorCode is code for the error of checking stats of worker worker
	CheckStatsErrorCode
	// GetJobStatsErrorCode is code for the error of getting stats of enqueued job
	GetJobStatsErrorCode
	// StopJobErrorCode is code for the error of stopping job
	StopJobErrorCode
	// RetryJobErrorCode is code for the error of retrying job
	RetryJobErrorCode
	// UnknownActionNameErrorCode is code for the case of unknown action name
	UnknownActionNameErrorCode
	// GetJobLogErrorCode is code for the error of getting job log
	GetJobLogErrorCode
	// NoObjectFoundErrorCode is code for the error of no object found
	NoObjectFoundErrorCode
	// UnAuthorizedErrorCode is code for the error of unauthorized accessing
	UnAuthorizedErrorCode
	// ResourceConflictsErrorCode is code for the error of resource conflicting
	ResourceConflictsErrorCode
	// BadRequestErrorCode is code for the error of bad request
	BadRequestErrorCode
	// GetScheduledJobsErrorCode is code for the error of getting scheduled jobs
	GetScheduledJobsErrorCode
	// GetPeriodicExecutionErrorCode is code for the error of getting periodic executions
	GetPeriodicExecutionErrorCode
)

// baseError ...
type baseError struct {
	Code        uint16 `json:"code"`
	Err         string `json:"message"`
	Description string `json:"details,omitempty"`
}

// Error is implementation of error interface.
func (be baseError) Error() string {
	if data, err := json.Marshal(be); err == nil {
		return string(data)
	}

	return "{}"
}

// New customized errors
func New(code uint16, err string, description string) error {
	return baseError{
		Code:        code,
		Err:         err,
		Description: description,
	}
}

// ReadRequestBodyError is error wrapper for the error of reading request body.
func ReadRequestBodyError(err error) error {
	return New(ReadRequestBodyErrorCode, "read request body failed with error", err.Error())
}

// HandleJSONDataError is error wrapper for the error of handling json data.
func HandleJSONDataError(err error) error {
	return New(HandleJSONDataErrorCode, "handle json data failed with error", err.Error())
}

// MissingBackendHandlerError is error wrapper for the error of missing backend controller.
func MissingBackendHandlerError(err error) error {
	return New(MissingBackendHandlerErrorCode, "missing backend controller to handle the requests", err.Error())
}

// LaunchJobError is error wrapper for the error of launching job failed.
func LaunchJobError(err error) error {
	return New(LaunchJobErrorCode, "launch job failed with error", err.Error())
}

// CheckStatsError is error wrapper for the error of checking stats failed
func CheckStatsError(err error) error {
	return New(CheckStatsErrorCode, "check stats of server failed with error", err.Error())
}

// GetJobStatsError is error wrapper for the error of getting job stats
func GetJobStatsError(err error) error {
	return New(GetJobStatsErrorCode, "get job stats failed with error", err.Error())
}

// StopJobError is error for the case of stopping job failed
func StopJobError(err error) error {
	return New(StopJobErrorCode, "stop job failed with error", err.Error())
}

// RetryJobError is error for the case of retrying job failed
func RetryJobError(err error) error {
	return New(RetryJobErrorCode, "retry job failed with error", err.Error())
}

// UnknownActionNameError is error for the case of getting unknown job action
func UnknownActionNameError(err error) error {
	return New(UnknownActionNameErrorCode, "unknown job action name", err.Error())
}

// GetJobLogError is error for the case of getting job log failed
func GetJobLogError(err error) error {
	return New(GetJobLogErrorCode, "failed to get the job log", err.Error())
}

// UnauthorizedError is error for the case of unauthorized accessing
func UnauthorizedError(err error) error {
	return New(UnAuthorizedErrorCode, "unauthorized", err.Error())
}

// GetScheduledJobsError is error for the case of getting scheduled jobs failed
func GetScheduledJobsError(err error) error {
	return New(GetScheduledJobsErrorCode, "failed to get scheduled jobs", err.Error())
}

// GetPeriodicExecutionError is error for the case of getting periodic jobs failed
func GetPeriodicExecutionError(err error) error {
	return New(GetPeriodicExecutionErrorCode, "failed to get periodic executions", err.Error())
}

// objectNotFound is designed for the case of no object found
type objectNotFoundError struct {
	baseError
}

// NoObjectFoundError is error wrapper for the case of no object found
func NoObjectFoundError(object string) error {
	return objectNotFoundError{
		baseError{
			Code:        NoObjectFoundErrorCode,
			Err:         "object is not found",
			Description: object,
		},
	}
}

// conflictError is designed for the case of resource conflicting
type conflictError struct {
	baseError
}

// ConflictError is error for the case of resource conflicting
func ConflictError(object string) error {
	return conflictError{
		baseError{
			Code:        ResourceConflictsErrorCode,
			Err:         "conflict",
			Description: fmt.Sprintf("the submitting resource is conflicted with existing one %s", object),
		},
	}
}

// badRequestError is designed for the case of bad request
type badRequestError struct {
	baseError
}

// BadRequestError returns the error of handing bad request case
func BadRequestError(object interface{}) error {
	return badRequestError{
		baseError{
			Code:        BadRequestErrorCode,
			Err:         "bad request",
			Description: fmt.Sprintf("%s", object),
		},
	}
}

// IsObjectNotFoundError return true if the error is objectNotFoundError
func IsObjectNotFoundError(err error) bool {
	_, ok := err.(objectNotFoundError)
	return ok
}

// IsConflictError returns true if the error is conflictError
func IsConflictError(err error) bool {
	_, ok := err.(conflictError)
	return ok
}

// IsBadRequestError returns true if the error is badRequestError
func IsBadRequestError(err error) bool {
	_, ok := err.(badRequestError)
	return ok
}
