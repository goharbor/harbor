/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding"
)

// Sender sends messages.
type Sender interface {
	// Send a message.
	//
	// Send returns when the "outbound" message has been sent. The Sender may
	// still be expecting acknowledgment or holding other state for the message.
	//
	// m.Finish() is called when sending is finished (both succeeded or failed):
	// expected acknowledgments (or errors) have been received, the Sender is
	// no longer holding any state for the message.
	// m.Finish() may be called during or after Send().
	//
	// transformers are applied when the message is written on the wire.
	Send(ctx context.Context, m binding.Message, transformers ...binding.Transformer) error
}

// SendCloser is a Sender that can be closed.
type SendCloser interface {
	Sender
	Closer
}

// Requester sends a message and receives a response
//
// Optional interface that may be implemented by protocols that support
// request/response correlation.
type Requester interface {
	// Request sends m like Sender.Send() but also arranges to receive a response.
	Request(ctx context.Context, m binding.Message, transformers ...binding.Transformer) (binding.Message, error)
}

// RequesterCloser is a Requester that can be closed.
type RequesterCloser interface {
	Requester
	Closer
}
