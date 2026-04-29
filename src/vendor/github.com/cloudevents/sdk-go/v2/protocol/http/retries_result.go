/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"fmt"
	"time"

	"github.com/cloudevents/sdk-go/v2/protocol"
)

// NewRetriesResult returns a http RetriesResult that should be used as
// a transport.Result without retries
func NewRetriesResult(result protocol.Result, retries int, startTime time.Time, attempts []protocol.Result) protocol.Result {
	rr := &RetriesResult{
		Result:   result,
		Retries:  retries,
		Duration: time.Since(startTime),
	}
	if len(attempts) > 0 {
		rr.Attempts = attempts
	}
	return rr
}

// RetriesResult wraps the fields required to make adjustments for http Responses.
type RetriesResult struct {
	// The last result
	protocol.Result

	// Retries is the number of times the request was tried
	Retries int

	// Duration records the time spent retrying. Exclude the successful request (if any)
	Duration time.Duration

	// Attempts of all failed requests. Exclude last result.
	Attempts []protocol.Result
}

// make sure RetriesResult implements error.
var _ error = (*RetriesResult)(nil)

// Is returns if the target error is a RetriesResult type checking target.
func (e *RetriesResult) Is(target error) bool {
	return protocol.ResultIs(e.Result, target)
}

// Error returns the string that is formed by using the format string with the
// provided args.
func (e *RetriesResult) Error() string {
	if e.Retries == 0 {
		return e.Result.Error()
	}
	return fmt.Sprintf("%s (%dx)", e.Result.Error(), e.Retries)
}
