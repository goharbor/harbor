/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/binding"
)

// Option is the function signature required to be considered an client.Option.
type Option func(interface{}) error

// WithEventDefaulter adds an event defaulter to the end of the defaulter chain.
func WithEventDefaulter(fn EventDefaulter) Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			if fn == nil {
				return fmt.Errorf("client option was given an nil event defaulter")
			}
			c.eventDefaulterFns = append(c.eventDefaulterFns, fn)
		}
		return nil
	}
}

func WithForceBinary() Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			c.outboundContextDecorators = append(c.outboundContextDecorators, binding.WithForceBinary)
		}
		return nil
	}
}

func WithForceStructured() Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			c.outboundContextDecorators = append(c.outboundContextDecorators, binding.WithForceStructured)
		}
		return nil
	}
}

// WithUUIDs adds DefaultIDToUUIDIfNotSet event defaulter to the end of the
// defaulter chain.
func WithUUIDs() Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			c.eventDefaulterFns = append(c.eventDefaulterFns, DefaultIDToUUIDIfNotSet)
		}
		return nil
	}
}

// WithTimeNow adds DefaultTimeToNowIfNotSet event defaulter to the end of the
// defaulter chain.
func WithTimeNow() Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			c.eventDefaulterFns = append(c.eventDefaulterFns, DefaultTimeToNowIfNotSet)
		}
		return nil
	}
}

// WithTracePropagation enables trace propagation via the distributed tracing
// extension.
// Deprecated: this is now noop and will be removed in future releases.
// Don't use distributed tracing extension to propagate traces:
// https://github.com/cloudevents/spec/blob/v1.0.1/extensions/distributed-tracing.md#using-the-distributed-tracing-extension
func WithTracePropagation() Option {
	return func(i interface{}) error {
		return nil
	}
}

// WithPollGoroutines configures how much goroutines should be used to
// poll the Receiver/Responder/Protocol implementations.
// Default value is GOMAXPROCS
func WithPollGoroutines(pollGoroutines int) Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			c.pollGoroutines = pollGoroutines
		}
		return nil
	}
}

// WithObservabilityService configures the observability service to use
// to record traces and metrics
func WithObservabilityService(service ObservabilityService) Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			c.observabilityService = service
			c.inboundContextDecorators = append(c.inboundContextDecorators, service.InboundContextDecorators()...)
		}
		return nil
	}
}

// WithInboundContextDecorator configures a new inbound context decorator.
// Inbound context decorators are invoked to wrap additional informations from the binding.Message
// and propagate these informations in the context passed to the event receiver.
func WithInboundContextDecorator(dec func(context.Context, binding.Message) context.Context) Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			c.inboundContextDecorators = append(c.inboundContextDecorators, dec)
		}
		return nil
	}
}

// WithBlockingCallback makes the callback passed into StartReceiver is executed as a blocking call,
// i.e. in each poll go routine, the next event will not be received until the callback on current event completes.
// To make event processing serialized (no concurrency), use this option along with WithPollGoroutines(1)
func WithBlockingCallback() Option {
	return func(i interface{}) error {
		if c, ok := i.(*ceClient); ok {
			c.blockingCallback = true
		}
		return nil
	}
}
