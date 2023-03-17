/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// ReceiveFull is the signature of a fn to be invoked for incoming cloudevents.
type ReceiveFull func(context.Context, event.Event) protocol.Result

type receiverFn struct {
	numIn   int
	numOut  int
	fnValue reflect.Value

	hasContextIn bool
	hasEventIn   bool

	hasEventOut  bool
	hasResultOut bool
}

const (
	inParamUsage  = "expected a function taking either no parameters, one or more of (context.Context, event.Event) ordered"
	outParamUsage = "expected a function returning one or mode of (*event.Event, protocol.Result) ordered"
)

var (
	contextType  = reflect.TypeOf((*context.Context)(nil)).Elem()
	eventType    = reflect.TypeOf((*event.Event)(nil)).Elem()
	eventPtrType = reflect.TypeOf((*event.Event)(nil)) // want the ptr type
	resultType   = reflect.TypeOf((*protocol.Result)(nil)).Elem()
)

// receiver creates a receiverFn wrapper class that is used by the client to
// validate and invoke the provided function.
// Valid fn signatures are:
// * func()
// * func() protocol.Result
// * func(context.Context)
// * func(context.Context) protocol.Result
// * func(event.Event)
// * func(event.Event) transport.Result
// * func(context.Context, event.Event)
// * func(context.Context, event.Event) protocol.Result
// * func(event.Event) *event.Event
// * func(event.Event) (*event.Event, protocol.Result)
// * func(context.Context, event.Event) *event.Event
// * func(context.Context, event.Event) (*event.Event, protocol.Result)
//
func receiver(fn interface{}) (*receiverFn, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, errors.New("must pass a function to handle events")
	}

	r := &receiverFn{
		fnValue: reflect.ValueOf(fn),
		numIn:   fnType.NumIn(),
		numOut:  fnType.NumOut(),
	}

	if err := r.validate(fnType); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *receiverFn) invoke(ctx context.Context, e *event.Event) (*event.Event, protocol.Result) {
	args := make([]reflect.Value, 0, r.numIn)

	if r.numIn > 0 {
		if r.hasContextIn {
			args = append(args, reflect.ValueOf(ctx))
		}
		if r.hasEventIn {
			args = append(args, reflect.ValueOf(*e))
		}
	}
	v := r.fnValue.Call(args)
	var respOut protocol.Result
	var eOut *event.Event
	if r.numOut > 0 {
		i := 0
		if r.hasEventOut {
			if eo, ok := v[i].Interface().(*event.Event); ok {
				eOut = eo
			}
			i++ // <-- note, need to inc i.
		}
		if r.hasResultOut {
			if resp, ok := v[i].Interface().(protocol.Result); ok {
				respOut = resp
			}
		}
	}
	return eOut, respOut
}

// Verifies that the inputs to a function have a valid signature
// Valid input is to be [0, all] of
// context.Context, event.Event in this order.
func (r *receiverFn) validateInParamSignature(fnType reflect.Type) error {
	r.hasContextIn = false
	r.hasEventIn = false

	switch fnType.NumIn() {
	case 2:
		// has to be (context.Context, event.Event)
		if !eventType.ConvertibleTo(fnType.In(1)) {
			return fmt.Errorf("%s; cannot convert parameter 2 to %s from event.Event", inParamUsage, fnType.In(1))
		} else {
			r.hasEventIn = true
		}
		fallthrough
	case 1:
		if !contextType.ConvertibleTo(fnType.In(0)) {
			if !eventType.ConvertibleTo(fnType.In(0)) {
				return fmt.Errorf("%s; cannot convert parameter 1 to %s from context.Context or event.Event", inParamUsage, fnType.In(0))
			} else if r.hasEventIn {
				return fmt.Errorf("%s; duplicate parameter of type event.Event", inParamUsage)
			} else {
				r.hasEventIn = true
			}
		} else {
			r.hasContextIn = true
		}
		fallthrough
	case 0:
		return nil

	default:
		return fmt.Errorf("%s; function has too many parameters (%d)", inParamUsage, fnType.NumIn())
	}
}

// Verifies that the outputs of a function have a valid signature
// Valid output signatures to be [0, all] of
// *event.Event, transport.Result in this order
func (r *receiverFn) validateOutParamSignature(fnType reflect.Type) error {
	r.hasEventOut = false
	r.hasResultOut = false

	switch fnType.NumOut() {
	case 2:
		// has to be (*event.Event, transport.Result)
		if !fnType.Out(1).ConvertibleTo(resultType) {
			return fmt.Errorf("%s; cannot convert parameter 2 from %s to event.Response", outParamUsage, fnType.Out(1))
		} else {
			r.hasResultOut = true
		}
		fallthrough
	case 1:
		if !fnType.Out(0).ConvertibleTo(resultType) {
			if !fnType.Out(0).ConvertibleTo(eventPtrType) {
				return fmt.Errorf("%s; cannot convert parameter 1 from %s to *event.Event or transport.Result", outParamUsage, fnType.Out(0))
			} else {
				r.hasEventOut = true
			}
		} else if r.hasResultOut {
			return fmt.Errorf("%s; duplicate parameter of type event.Response", outParamUsage)
		} else {
			r.hasResultOut = true
		}
		fallthrough
	case 0:
		return nil
	default:
		return fmt.Errorf("%s; function has too many return types (%d)", outParamUsage, fnType.NumOut())
	}
}

// validateReceiverFn validates that a function has the right number of in and
// out params and that they are of allowed types.
func (r *receiverFn) validate(fnType reflect.Type) error {
	if err := r.validateInParamSignature(fnType); err != nil {
		return err
	}
	if err := r.validateOutParamSignature(fnType); err != nil {
		return err
	}
	return nil
}
