// Copyright 2018 The Harbor Authors. All rights reserved.

//Package errs define some system errors with specified types.
package errs

import (
	"encoding/json"
)

const (
	//JobStoppedErrorCode is code for jobStoppedError
	JobStoppedErrorCode = 10000 + iota
	//ReadRequestBodyErrorCode is code for the error of reading http request body error
	ReadRequestBodyErrorCode
	//HandleJSONDataErrorCode is code for the error of handling json data error
	HandleJSONDataErrorCode
	//MissingBackendHandlerErrorCode is code for the error of missing backend controller
	MissingBackendHandlerErrorCode
	//LaunchJobErrorCode is code for the error of launching job
	LaunchJobErrorCode
	//CheckStatsErrorCode is code for the error of checking stats of worker pool
	CheckStatsErrorCode
	//GetJobStatsErrorCode is code for the error of getting stats of enqueued job
	GetJobStatsErrorCode
)

//baseError ...
type baseError struct {
	Code        uint16 `json:"code"`
	Err         string `json:"message"`
	Description string `json:"details"`
}

//Error is implementation of error interface.
func (be baseError) Error() string {
	if data, err := json.Marshal(be); err == nil {
		return string(data)
	}

	return "{}"
}

//New customized errors
func New(code uint16, err string, description string) error {
	return baseError{
		Code:        code,
		Err:         err,
		Description: description,
	}
}

//ReadRequestBodyError is error wrapper for the error of reading request body.
func ReadRequestBodyError(err error) error {
	return New(ReadRequestBodyErrorCode, "Read request body failed with error", err.Error())
}

//HandleJSONDataError is error wrapper for the error of handling json data.
func HandleJSONDataError(err error) error {
	return New(HandleJSONDataErrorCode, "Handle json data failed with error", err.Error())
}

//MissingBackendHandlerError is error wrapper for the error of missing backend controller.
func MissingBackendHandlerError(err error) error {
	return New(MissingBackendHandlerErrorCode, "Missing backend controller to handle the requests", err.Error())
}

//LaunchJobError is error wrapper for the error of launching job failed.
func LaunchJobError(err error) error {
	return New(LaunchJobErrorCode, "Launch job failed with error", err.Error())
}

//CheckStatsError is error wrapper for the error of checking stats failed
func CheckStatsError(err error) error {
	return New(CheckStatsErrorCode, "Check stats of server failed with error", err.Error())
}

//GetJobStatsError is error wrapper for the error of getting job stats
func GetJobStatsError(err error) error {
	return New(GetJobStatsErrorCode, "Get job stats failed with error", err.Error())
}

//jobStoppedError is designed for the case of stopping job.
type jobStoppedError struct {
	baseError
}

//jobCancelledError is designed for the case of cancelling job.
type jobCancelledError struct {
	baseError
}
