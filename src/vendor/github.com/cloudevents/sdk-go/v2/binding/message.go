/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

// MessageReader defines the read-related portion of the Message interface.
//
// The ReadStructured and ReadBinary methods allows to perform an optimized encoding of a Message to a specific data structure.
//
// If MessageReader.ReadEncoding() can be equal to EncodingBinary, then the implementation of MessageReader
// MUST also implement MessageMetadataReader.
//
// A Sender should try each method of interest and fall back to binding.ToEvent() if none are supported.
// An out of the box algorithm is provided for writing a message: binding.Write().
type MessageReader interface {
	// Return the type of the message Encoding.
	// The encoding should be preferably computed when the message is constructed.
	ReadEncoding() Encoding

	// ReadStructured transfers a structured-mode event to a StructuredWriter.
	// It must return ErrNotStructured if message is not in structured mode.
	//
	// Returns a different err if something wrong happened while trying to read the structured event.
	// In this case, the caller must Finish the message with appropriate error.
	//
	// This allows Senders to avoid re-encoding messages that are
	// already in suitable structured form.
	ReadStructured(context.Context, StructuredWriter) error

	// ReadBinary transfers a binary-mode event to an BinaryWriter.
	// It must return ErrNotBinary if message is not in binary mode.
	//
	// The implementation of ReadBinary must not control the lifecycle with BinaryWriter.Start() and BinaryWriter.End(),
	// because the caller must control the lifecycle.
	//
	// Returns a different err if something wrong happened while trying to read the binary event
	// In this case, the caller must Finish the message with appropriate error
	//
	// This allows Senders to avoid re-encoding messages that are
	// already in suitable binary form.
	ReadBinary(context.Context, BinaryWriter) error
}

// MessageMetadataReader defines how to read metadata from a binary/event message
//
// If a message implementing MessageReader is encoded as binary (MessageReader.ReadEncoding() == EncodingBinary)
// or it's an EventMessage, then it's safe to assume that it also implements this interface
type MessageMetadataReader interface {
	// GetAttribute returns:
	//
	// * attribute, value: if the message contains an attribute of that attribute kind
	// * attribute, nil: if the message spec version supports the attribute kind, but doesn't have any value
	// * nil, nil: if the message spec version doesn't support the attribute kind
	GetAttribute(attributeKind spec.Kind) (spec.Attribute, interface{})
	// GetExtension returns the value of that extension, if any.
	GetExtension(name string) interface{}
}

// Message is the interface to a binding-specific message containing an event.
//
// # Reliable Delivery
//
// There are 3 reliable qualities of service for messages:
//
// 0/at-most-once/unreliable: messages can be dropped silently.
//
// 1/at-least-once: messages are not dropped without signaling an error
// to the sender, but they may be duplicated in the event of a re-send.
//
// 2/exactly-once: messages are never dropped (without error) or
// duplicated, as long as both sending and receiving ends maintain
// some binding-specific delivery state. Whether this is persisted
// depends on the configuration of the binding implementations.
//
// The Message interface supports QoS 0 and 1, the ExactlyOnceMessage interface
// supports QoS 2
//
// Message includes the MessageReader interface to read messages. Every binding.Message implementation *must* specify if the message can be accessed one or more times.
//
// When a Message can be forgotten by the entity who produced the message, Message.Finish() *must* be invoked.
type Message interface {
	MessageReader

	// Finish *must* be called when message from a Receiver can be forgotten by
	// the receiver. A QoS 1 sender should not call Finish() until it gets an acknowledgment of
	// receipt on the underlying transport.  For QoS 2 see ExactlyOnceMessage.
	//
	// Note that, depending on the Message implementation, forgetting to Finish the message
	// could produce memory/resources leaks!
	//
	// Passing a non-nil err indicates sending or processing failed.
	// A non-nil return indicates that the message was not accepted
	// by the receivers peer.
	Finish(error) error
}

// ExactlyOnceMessage is implemented by received Messages
// that support QoS 2.  Only transports that support QoS 2 need to
// implement or use this interface.
type ExactlyOnceMessage interface {
	Message

	// Received is called by a forwarding QoS2 Sender when it gets
	// acknowledgment of receipt (e.g. AMQP 'accept' or MQTT PUBREC)
	//
	// The receiver must call settle(nil) when it get's the ack-of-ack
	// (e.g. AMQP 'settle' or MQTT PUBCOMP) or settle(err) if the
	// transfer fails.
	//
	// Finally the Sender calls Finish() to indicate the message can be
	// discarded.
	//
	// If sending fails, or if the sender does not support QoS 2, then
	// Finish() may be called without any call to Received()
	Received(settle func(error))
}

// MessageContext interface exposes the internal context that a message might contain
// Only some Message implementations implement this interface.
type MessageContext interface {
	// Get the context associated with this message
	Context() context.Context
}

// MessageWrapper interface is used to walk through a decorated Message and unwrap it.
type MessageWrapper interface {
	Message
	MessageMetadataReader

	// Method to get the wrapped message
	GetWrappedMessage() Message
}

func UnwrapMessage(message Message) Message {
	m := message
	for m != nil {
		switch mt := m.(type) {
		case MessageWrapper:
			m = mt.GetWrappedMessage()
		default:
			return m
		}
	}
	return m
}
