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
	// JobStoppedErrorCode is code for jobStoppedError
	JobStoppedErrorCode = 10000 + iota
	// JobCancelledErrorCode is code for jobCancelledError
	JobCancelledErrorCode
	// ReadRequestBodyErrorCode is code for the error of reading http request body error
	ReadRequestBodyErrorCode
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
	return New(ReadRequestBodyErrorCode, "Read request body failed with error", err.Error())
}

// HandleJSONDataError is error wrapper for the error of handling json data.
func HandleJSONDataError(err error) error {
	return New(HandleJSONDataErrorCode, "Handle json data failed with error", err.Error())
}

// MissingBackendHandlerError is error wrapper for the error of missing backend controller.
func MissingBackendHandlerError(err error) error {
	return New(MissingBackendHandlerErrorCode, "Missing backend controller to handle the requests", err.Error())
}

// LaunchJobError is error wrapper for the error of launching job failed.
func LaunchJobError(err error) error {
	return New(LaunchJobErrorCode, "Launch job failed with error", err.Error())
}

// CheckStatsError is error wrapper for the error of checking stats failed
func CheckStatsError(err error) error {
	return New(CheckStatsErrorCode, "Check stats of server failed with error", err.Error())
}

// GetJobStatsError is error wrapper for the error of getting job stats
func GetJobStatsError(err error) error {
	return New(GetJobStatsErrorCode, "Get job stats failed with error", err.Error())
}

// StopJobError is error for the case of stopping job failed
func StopJobError(err error) error {
	return New(StopJobErrorCode, "Stop job failed with error", err.Error())
}

// RetryJobError is error for the case of retrying job failed
func RetryJobError(err error) error {
	return New(RetryJobErrorCode, "Retry job failed with error", err.Error())
}

// UnknownActionNameError is error for the case of getting unknown job action
func UnknownActionNameError(err error) error {
	return New(UnknownActionNameErrorCode, "Unknown job action name", err.Error())
}

// GetJobLogError is error for the case of getting job log failed
func GetJobLogError(err error) error {
	return New(GetJobLogErrorCode, "Failed to get the job log", err.Error())
}

// UnauthorizedError is error for the case of unauthorized accessing
func UnauthorizedError(err error) error {
	return New(UnAuthorizedErrorCode, "Unauthorized", err.Error())
}

// jobStoppedError is designed for the case of stopping job.
type jobStoppedError struct {
	baseError
}

// JobStoppedError is error wrapper for the case of stopping job.
func JobStoppedError() error {
	return jobStoppedError{
		baseError{
			Code: JobStoppedErrorCode,
			Err:  "Job is stopped",
		},
	}
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

// IsJobStoppedError return true if the error is jobStoppedError
func IsJobStoppedError(err error) bool {
	_, ok := err.(jobStoppedError)
	return ok
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
