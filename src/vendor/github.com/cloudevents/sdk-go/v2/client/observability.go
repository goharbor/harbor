/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
)

// ObservabilityService is an interface users can implement to record metrics, create tracing spans, and plug other observability tools in the Client
type ObservabilityService interface {
	// InboundContextDecorators is a method that returns the InboundContextDecorators that must be mounted in the Client to properly propagate some tracing informations.
	InboundContextDecorators() []func(context.Context, binding.Message) context.Context

	// RecordReceivedMalformedEvent is invoked when an event was received but it's malformed or invalid.
	RecordReceivedMalformedEvent(ctx context.Context, err error)
	// RecordCallingInvoker is invoked before the user function is invoked.
	// The returned callback will be invoked after the user finishes to process the event with the eventual processing error
	// The error provided to the callback could be both a processing error, or a result
	RecordCallingInvoker(ctx context.Context, event *event.Event) (context.Context, func(errOrResult error))
	// RecordSendingEvent is invoked before the event is sent.
	// The returned callback will be invoked when the response is received
	// The error provided to the callback could be both a processing error, or a result
	RecordSendingEvent(ctx context.Context, event event.Event) (context.Context, func(errOrResult error))

	// RecordRequestEvent is invoked before the event is requested.
	// The returned callback will be invoked when the response is received
	RecordRequestEvent(ctx context.Context, event event.Event) (context.Context, func(errOrResult error, event *event.Event))
}

type noopObservabilityService struct{}

func (n noopObservabilityService) InboundContextDecorators() []func(context.Context, binding.Message) context.Context {
	return nil
}

func (n noopObservabilityService) RecordReceivedMalformedEvent(ctx context.Context, err error) {}

func (n noopObservabilityService) RecordCallingInvoker(ctx context.Context, event *event.Event) (context.Context, func(errOrResult error)) {
	return ctx, func(errOrResult error) {}
}

func (n noopObservabilityService) RecordSendingEvent(ctx context.Context, event event.Event) (context.Context, func(errOrResult error)) {
	return ctx, func(errOrResult error) {}
}

func (n noopObservabilityService) RecordRequestEvent(ctx context.Context, e event.Event) (context.Context, func(errOrResult error, event *event.Event)) {
	return ctx, func(errOrResult error, event *event.Event) {}
}
