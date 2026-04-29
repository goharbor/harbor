/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"errors"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/protocol"
)

// NewResult returns a fully populated http Result that should be used as
// a transport.Result.
func NewResult(statusCode int, messageFmt string, args ...interface{}) protocol.Result {
	return &Result{
		StatusCode: statusCode,
		Format:     messageFmt,
		Args:       args,
	}
}

// Result wraps the fields required to make adjustments for http Responses.
type Result struct {
	StatusCode int
	Format     string
	Args       []interface{}
}

// make sure Result implements error.
var _ error = (*Result)(nil)

// Is returns if the target error is a Result type checking target.
func (e *Result) Is(target error) bool {
	if o, ok := target.(*Result); ok {
		return e.StatusCode == o.StatusCode
	}

	// Special case for nil == ACK
	if o, ok := target.(*protocol.Receipt); ok {
		if e == nil && o.ACK {
			return true
		}
	}

	// Allow for wrapped errors.
	if e != nil {
		err := fmt.Errorf(e.Format, e.Args...)
		return errors.Is(err, target)
	}
	return false
}

// Error returns the string that is formed by using the format string with the
// provided args.
func (e *Result) Error() string {
	return fmt.Sprintf("%d: %v", e.StatusCode, fmt.Errorf(e.Format, e.Args...))
}
