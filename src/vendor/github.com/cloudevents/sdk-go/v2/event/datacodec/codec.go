/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package datacodec

import (
	"context"
	"fmt"

	"github.com/cloudevents/sdk-go/v2/event/datacodec/json"
	"github.com/cloudevents/sdk-go/v2/event/datacodec/text"
	"github.com/cloudevents/sdk-go/v2/event/datacodec/xml"
)

// Decoder is the expected function signature for decoding `in` to `out`.
// If Event sent the payload as base64, Decoder assumes that `in` is the
// decoded base64 byte array.
type Decoder func(ctx context.Context, in []byte, out interface{}) error

// Encoder is the expected function signature for encoding `in` to bytes.
// Returns an error if the encoder has an issue encoding `in`.
type Encoder func(ctx context.Context, in interface{}) ([]byte, error)

var decoder map[string]Decoder
var encoder map[string]Encoder

func init() {
	decoder = make(map[string]Decoder, 10)
	encoder = make(map[string]Encoder, 10)

	AddDecoder("", json.Decode)
	AddDecoder("application/json", json.Decode)
	AddDecoder("text/json", json.Decode)
	AddDecoder("application/xml", xml.Decode)
	AddDecoder("text/xml", xml.Decode)
	AddDecoder("text/plain", text.Decode)

	AddEncoder("", json.Encode)
	AddEncoder("application/json", json.Encode)
	AddEncoder("text/json", json.Encode)
	AddEncoder("application/xml", xml.Encode)
	AddEncoder("text/xml", xml.Encode)
	AddEncoder("text/plain", text.Encode)
}

// AddDecoder registers a decoder for a given content type. The codecs will use
// these to decode the data payload from a cloudevent.Event object.
func AddDecoder(contentType string, fn Decoder) {
	decoder[contentType] = fn
}

// AddEncoder registers an encoder for a given content type. The codecs will
// use these to encode the data payload for a cloudevent.Event object.
func AddEncoder(contentType string, fn Encoder) {
	encoder[contentType] = fn
}

// Decode looks up and invokes the decoder registered for the given content
// type. An error is returned if no decoder is registered for the given
// content type.
func Decode(ctx context.Context, contentType string, in []byte, out interface{}) error {
	if fn, ok := decoder[contentType]; ok {
		return fn(ctx, in, out)
	}
	return fmt.Errorf("[decode] unsupported content type: %q", contentType)
}

// Encode looks up and invokes the encoder registered for the given content
// type. An error is returned if no encoder is registered for the given
// content type.
func Encode(ctx context.Context, contentType string, in interface{}) ([]byte, error) {
	if fn, ok := encoder[contentType]; ok {
		return fn(ctx, in)
	}
	return nil, fmt.Errorf("[encode] unsupported content type: %q", contentType)
}
