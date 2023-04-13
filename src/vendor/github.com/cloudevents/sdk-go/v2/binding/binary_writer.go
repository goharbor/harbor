/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import (
	"context"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

// MessageMetadataWriter is used to set metadata when a binary Message is visited.
type MessageMetadataWriter interface {
	// Set a standard attribute.
	//
	// The value can either be the correct golang type for the attribute, or a canonical
	// string encoding, or nil. If value is nil, then the attribute should be deleted.
	// See package types to perform the needed conversions.
	SetAttribute(attribute spec.Attribute, value interface{}) error

	// Set an extension attribute.
	//
	// The value can either be the correct golang type for the attribute, or a canonical
	// string encoding, or nil. If value is nil, then the extension should be deleted.
	// See package types to perform the needed conversions.
	SetExtension(name string, value interface{}) error
}

// BinaryWriter is used to visit a binary Message and generate a new representation.
//
// Protocols that supports binary encoding should implement this interface to implement direct
// binary to binary encoding and event to binary encoding.
//
// Start() and End() methods must be invoked by the caller of Message.ReadBinary() every time
// the BinaryWriter implementation is used to visit a Message.
type BinaryWriter interface {
	MessageMetadataWriter

	// Method invoked at the beginning of the visit. Useful to perform initial memory allocations
	Start(ctx context.Context) error

	// SetData receives an io.Reader for the data attribute.
	// io.Reader is not invoked when the data attribute is empty
	SetData(data io.Reader) error

	// End method is invoked only after the whole encoding process ends successfully.
	// If it fails, it's never invoked. It can be used to finalize the message.
	End(ctx context.Context) error
}
