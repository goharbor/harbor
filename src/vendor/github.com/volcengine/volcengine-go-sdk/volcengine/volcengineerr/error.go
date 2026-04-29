// Package volcengineerr represents API error interface accessors for the SDK.
package volcengineerr

// Copy from https://github.com/aws/aws-sdk-go
// May have been modified by Beijing Volcanoengine Technology Ltd.

// An Error wraps lower level errors with code, message and an original error.
// The underlying concrete error type may also satisfy other interfaces which
// can be to used to obtain more specific information about the error.
//
// Calling Error() or String() will always include the full information about
// an error based on its underlying type.
//
type Error interface {
	// Satisfy the generic error interface.
	error

	// Code Returns the short phrase depicting the classification of the error.
	Code() string

	// Message Returns the error details message.
	Message() string

	// OrigErr Returns the original error if one was set.  Nil is returned if not set.
	OrigErr() error
}

// BatchError is a batch of errors which also wraps lower level errors with
// code, message, and original errors. Calling Error() will include all errors
// that occurred in the batch.
//
// Deprecated: Replaced with BatchedErrors. Only defined for backwards
// compatibility.
type BatchError interface {
	// Satisfy the generic error interface.
	error

	// Code Returns the short phrase depicting the classification of the error.
	Code() string

	// Message Returns the error details message.
	Message() string

	// OrigErrs Returns the original error if one was set.  Nil is returned if not set.
	OrigErrs() []error
}

// BatchedErrors is a batch of errors which also wraps lower level errors with
// code, message, and original errors. Calling Error() will include all errors
// that occurred in the batch.
//
// Replaces BatchError
type BatchedErrors interface {
	// Error Satisfy the base Error interface.
	Error

	// OrigErrs Returns the original error if one was set.  Nil is returned if not set.
	OrigErrs() []error
}

// New returns an Error object described by the code, message, and origErr.
//
// If origErr satisfies the Error interface it will not be wrapped within a new
// Error object and will instead be returned.
func New(code, message string, origErr error) Error {
	var errs []error
	if origErr != nil {
		errs = append(errs, origErr)
	}
	return newBaseError(code, message, errs)
}

// NewBatchError returns an BatchedErrors with a collection of errors as an
// array of errors.
func NewBatchError(code, message string, errs []error) BatchedErrors {
	return newBaseError(code, message, errs)
}

// A RequestFailure is an interface to extract request failure information from
// an Error such as the request ID of the failed request returned by a service.
// RequestFailures may not always have a requestID value if the request failed
// prior to reaching the service such as a connection error.
//
type RequestFailure interface {
	Error

	// StatusCode The status code of the HTTP response.
	StatusCode() int

	// RequestID The request ID returned by the service for a request failure. This will
	// be empty if no request ID is available such as the request failed due
	// to a connection error.
	RequestID() string
}

// NewRequestFailure returns a wrapped error with additional information for
// request status code, and service requestID.
//
// Should be used to wrap all request which involve service requests. Even if
// the request failed without a service response, but had an HTTP status code
// that may be meaningful.
func NewRequestFailure(err Error, statusCode int, reqID string, simple ...*bool) RequestFailure {
	return newRequestError(err, statusCode, reqID, simple...)
}

// UnmarshalError provides the interface for the SDK failing to unmarshal data.
type UnmarshalError interface {
	volcengineerror
	Bytes() []byte
}

// NewUnmarshalError returns an initialized UnmarshalError error wrapper adding
// the bytes that fail to unmarshal to the error.
func NewUnmarshalError(err error, msg string, bytes []byte) UnmarshalError {
	return &unmarshalError{
		volcengineerror: New("UnmarshalError", msg, err),
		bytes:           bytes,
	}
}
