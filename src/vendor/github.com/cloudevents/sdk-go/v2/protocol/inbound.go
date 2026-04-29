/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
)

// Receiver receives messages.
type Receiver interface {
	// Receive blocks till a message is received or ctx expires.
	// Receive can be invoked safely from different goroutines.
	//
	// A non-nil error means the receiver is closed.
	// io.EOF means it closed cleanly, any other value indicates an error.
	// The caller is responsible for `Finish()` the returned message
	Receive(ctx context.Context) (binding.Message, error)
}

// ReceiveCloser is a Receiver that can be closed.
type ReceiveCloser interface {
	Receiver
	Closer
}

// ResponseFn is the function callback provided from Responder.Respond to allow
// for a receiver to "reply" to a message it receives.
// transformers are applied when the message is written on the wire.
type ResponseFn func(ctx context.Context, m binding.Message, r Result, transformers ...binding.Transformer) error

// Responder receives messages and is given a callback to respond.
type Responder interface {
	// Respond blocks till a message is received or ctx expires.
	// Respond can be invoked safely from different goroutines.
	//
	// A non-nil error means the receiver is closed.
	// io.EOF means it closed cleanly, any other value indicates an error.
	// The caller is responsible for `Finish()` the returned message,
	// while the protocol implementation is responsible for `Finish()` the response message.
	// The caller MUST invoke ResponseFn, in order to avoid leaks.
	// The correct flow for the caller is to finish the received message and then invoke the ResponseFn
	Respond(ctx context.Context) (binding.Message, ResponseFn, error)
}

// ResponderCloser is a Responder that can be closed.
type ResponderCloser interface {
	Responder
	Closer
}
