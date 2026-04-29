/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"
	"strings"
)

type ValidationError map[string]error

func (e ValidationError) Error() string {
	b := strings.Builder{}
	for k, v := range e {
		b.WriteString(k)
		b.WriteString(": ")
		b.WriteString(v.Error())
		b.WriteRune('\n')
	}
	return b.String()
}

// Validate performs a spec based validation on this event.
// Validation is dependent on the spec version specified in the event context.
func (e Event) Validate() error {
	if e.Context == nil {
		return ValidationError{"specversion": fmt.Errorf("missing Event.Context")}
	}

	errs := map[string]error{}
	if e.FieldErrors != nil {
		for k, v := range e.FieldErrors {
			errs[k] = v
		}
	}

	if fieldErrors := e.Context.Validate(); fieldErrors != nil {
		for k, v := range fieldErrors {
			errs[k] = v
		}
	}

	if len(errs) > 0 {
		return ValidationError(errs)
	}
	return nil
}
