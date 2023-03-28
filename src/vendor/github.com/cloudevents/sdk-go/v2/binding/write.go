/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import (
	"context"

	"github.com/cloudevents/sdk-go/v2/event"
)

type eventEncodingKey int

const (
	skipDirectStructuredEncoding eventEncodingKey = iota
	skipDirectBinaryEncoding
	preferredEventEncoding
)

// DirectWrite invokes the encoders. structuredWriter and binaryWriter could be nil if the protocol doesn't support it.
// transformers can be nil and this function guarantees that they are invoked only once during the encoding process.
// This function MUST be invoked only if message.ReadEncoding() == EncodingBinary or message.ReadEncoding() == EncodingStructured
//
// Returns:
// * EncodingStructured, nil if message is correctly encoded in structured encoding
// * EncodingBinary, nil if message is correctly encoded in binary encoding
// * EncodingStructured, err if message was structured but error happened during the encoding
// * EncodingBinary, err if message was binary but error happened during the encoding
// * EncodingUnknown, ErrUnknownEncoding if message is not a structured or a binary Message
func DirectWrite(
	ctx context.Context,
	message MessageReader,
	structuredWriter StructuredWriter,
	binaryWriter BinaryWriter,
	transformers ...Transformer,
) (Encoding, error) {
	if structuredWriter != nil && len(transformers) == 0 && !GetOrDefaultFromCtx(ctx, skipDirectStructuredEncoding, false).(bool) {
		if err := message.ReadStructured(ctx, structuredWriter); err == nil {
			return EncodingStructured, nil
		} else if err != ErrNotStructured {
			return EncodingStructured, err
		}
	}

	if binaryWriter != nil && !GetOrDefaultFromCtx(ctx, skipDirectBinaryEncoding, false).(bool) && message.ReadEncoding() == EncodingBinary {
		return EncodingBinary, writeBinaryWithTransformer(ctx, message, binaryWriter, transformers)
	}

	return EncodingUnknown, ErrUnknownEncoding
}

// Write executes the full algorithm to encode a Message using transformers:
// 1. It first tries direct encoding using DirectWrite
// 2. If no direct encoding is possible, it uses ToEvent to generate an Event representation
// 3. From the Event, the message is encoded back to the provided structured or binary encoders
// You can tweak the encoding process using the context decorators WithForceStructured, WithForceStructured, etc.
// transformers can be nil and this function guarantees that they are invoked only once during the encoding process.
// Returns:
// * EncodingStructured, nil if message is correctly encoded in structured encoding
// * EncodingBinary, nil if message is correctly encoded in binary encoding
// * EncodingUnknown, ErrUnknownEncoding if message.ReadEncoding() == EncodingUnknown
// * _, err if error happened during the encoding
func Write(
	ctx context.Context,
	message MessageReader,
	structuredWriter StructuredWriter,
	binaryWriter BinaryWriter,
	transformers ...Transformer,
) (Encoding, error) {
	enc := message.ReadEncoding()
	var err error
	// Skip direct encoding if the event is an event message
	if enc != EncodingEvent {
		enc, err = DirectWrite(ctx, message, structuredWriter, binaryWriter, transformers...)
		if enc != EncodingUnknown {
			// Message directly encoded, nothing else to do here
			return enc, err
		}
	}

	var e *event.Event
	e, err = ToEvent(ctx, message, transformers...)
	if err != nil {
		return enc, err
	}

	message = (*EventMessage)(e)

	if GetOrDefaultFromCtx(ctx, preferredEventEncoding, EncodingBinary).(Encoding) == EncodingStructured {
		if structuredWriter != nil {
			return EncodingStructured, message.ReadStructured(ctx, structuredWriter)
		}
		if binaryWriter != nil {
			return EncodingBinary, writeBinary(ctx, message, binaryWriter)
		}
	} else {
		if binaryWriter != nil {
			return EncodingBinary, writeBinary(ctx, message, binaryWriter)
		}
		if structuredWriter != nil {
			return EncodingStructured, message.ReadStructured(ctx, structuredWriter)
		}
	}

	return EncodingUnknown, ErrUnknownEncoding
}

// WithSkipDirectStructuredEncoding skips direct structured to structured encoding during the encoding process
func WithSkipDirectStructuredEncoding(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, skipDirectStructuredEncoding, skip)
}

// WithSkipDirectBinaryEncoding skips direct binary to binary encoding during the encoding process
func WithSkipDirectBinaryEncoding(ctx context.Context, skip bool) context.Context {
	return context.WithValue(ctx, skipDirectBinaryEncoding, skip)
}

// WithPreferredEventEncoding defines the preferred encoding from event to message during the encoding process
func WithPreferredEventEncoding(ctx context.Context, enc Encoding) context.Context {
	return context.WithValue(ctx, preferredEventEncoding, enc)
}

// WithForceStructured forces structured encoding during the encoding process
func WithForceStructured(ctx context.Context) context.Context {
	return context.WithValue(context.WithValue(ctx, preferredEventEncoding, EncodingStructured), skipDirectBinaryEncoding, true)
}

// WithForceBinary forces binary encoding during the encoding process
func WithForceBinary(ctx context.Context) context.Context {
	return context.WithValue(context.WithValue(ctx, preferredEventEncoding, EncodingBinary), skipDirectStructuredEncoding, true)
}

// GetOrDefaultFromCtx gets a configuration value from the provided context
func GetOrDefaultFromCtx(ctx context.Context, key interface{}, def interface{}) interface{} {
	if val := ctx.Value(key); val != nil {
		return val
	} else {
		return def
	}
}

func writeBinaryWithTransformer(
	ctx context.Context,
	message MessageReader,
	binaryWriter BinaryWriter,
	transformers Transformers,
) error {
	err := binaryWriter.Start(ctx)
	if err != nil {
		return err
	}
	err = message.ReadBinary(ctx, binaryWriter)
	if err != nil {
		return err
	}
	err = transformers.Transform(message.(MessageMetadataReader), binaryWriter)
	if err != nil {
		return err
	}
	return binaryWriter.End(ctx)
}

func writeBinary(
	ctx context.Context,
	message MessageReader,
	binaryWriter BinaryWriter,
) error {
	err := binaryWriter.Start(ctx)
	if err != nil {
		return err
	}
	err = message.ReadBinary(ctx, binaryWriter)
	if err != nil {
		return err
	}
	return binaryWriter.End(ctx)
}
