/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import "errors"

// Encoding enum specifies the type of encodings supported by binding interfaces
type Encoding int

const (
	// Binary encoding as specified in https://github.com/cloudevents/spec/blob/master/spec.md#message
	EncodingBinary Encoding = iota
	// Structured encoding as specified in https://github.com/cloudevents/spec/blob/master/spec.md#message
	EncodingStructured
	// Message is an instance of EventMessage or it contains EventMessage nested (through MessageWrapper)
	EncodingEvent
	// When the encoding is unknown (which means that the message is a non-event)
	EncodingUnknown

	// EncodingBatch is an instance of JSON Batched Events
	EncodingBatch
)

func (e Encoding) String() string {
	switch e {
	case EncodingBinary:
		return "binary"
	case EncodingStructured:
		return "structured"
	case EncodingEvent:
		return "event"
	case EncodingBatch:
		return "batch"
	case EncodingUnknown:
		return "unknown"
	}
	return ""
}

// ErrUnknownEncoding specifies that the Message is not an event or it is encoded with an unknown encoding
var ErrUnknownEncoding = errors.New("unknown Message encoding")

// ErrNotStructured returned by Message.Structured for non-structured messages.
var ErrNotStructured = errors.New("message is not in structured mode")

// ErrNotBinary returned by Message.Binary for non-binary messages.
var ErrNotBinary = errors.New("message is not in binary mode")
