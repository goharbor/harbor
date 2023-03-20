/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// WriteJson writes the in event in the provided writer.
// Note: this function assumes the input event is valid.
func WriteJson(in *Event, writer io.Writer) error {
	stream := jsoniter.ConfigFastest.BorrowStream(writer)
	defer jsoniter.ConfigFastest.ReturnStream(stream)
	stream.WriteObjectStart()

	var ext map[string]interface{}
	var dct *string
	var isBase64 bool

	// Write the context (without the extensions)
	switch eventContext := in.Context.(type) {
	case *EventContextV03:
		// Set a bunch of variables we need later
		ext = eventContext.Extensions
		dct = eventContext.DataContentType

		stream.WriteObjectField("specversion")
		stream.WriteString(CloudEventsVersionV03)
		stream.WriteMore()

		stream.WriteObjectField("id")
		stream.WriteString(eventContext.ID)
		stream.WriteMore()

		stream.WriteObjectField("source")
		stream.WriteString(eventContext.Source.String())
		stream.WriteMore()

		stream.WriteObjectField("type")
		stream.WriteString(eventContext.Type)

		if eventContext.Subject != nil {
			stream.WriteMore()
			stream.WriteObjectField("subject")
			stream.WriteString(*eventContext.Subject)
		}

		if eventContext.DataContentEncoding != nil {
			isBase64 = true
			stream.WriteMore()
			stream.WriteObjectField("datacontentencoding")
			stream.WriteString(*eventContext.DataContentEncoding)
		}

		if eventContext.DataContentType != nil {
			stream.WriteMore()
			stream.WriteObjectField("datacontenttype")
			stream.WriteString(*eventContext.DataContentType)
		}

		if eventContext.SchemaURL != nil {
			stream.WriteMore()
			stream.WriteObjectField("schemaurl")
			stream.WriteString(eventContext.SchemaURL.String())
		}

		if eventContext.Time != nil {
			stream.WriteMore()
			stream.WriteObjectField("time")
			stream.WriteString(eventContext.Time.String())
		}
	case *EventContextV1:
		// Set a bunch of variables we need later
		ext = eventContext.Extensions
		dct = eventContext.DataContentType
		isBase64 = in.DataBase64

		stream.WriteObjectField("specversion")
		stream.WriteString(CloudEventsVersionV1)
		stream.WriteMore()

		stream.WriteObjectField("id")
		stream.WriteString(eventContext.ID)
		stream.WriteMore()

		stream.WriteObjectField("source")
		stream.WriteString(eventContext.Source.String())
		stream.WriteMore()

		stream.WriteObjectField("type")
		stream.WriteString(eventContext.Type)

		if eventContext.Subject != nil {
			stream.WriteMore()
			stream.WriteObjectField("subject")
			stream.WriteString(*eventContext.Subject)
		}

		if eventContext.DataContentType != nil {
			stream.WriteMore()
			stream.WriteObjectField("datacontenttype")
			stream.WriteString(*eventContext.DataContentType)
		}

		if eventContext.DataSchema != nil {
			stream.WriteMore()
			stream.WriteObjectField("dataschema")
			stream.WriteString(eventContext.DataSchema.String())
		}

		if eventContext.Time != nil {
			stream.WriteMore()
			stream.WriteObjectField("time")
			stream.WriteString(eventContext.Time.String())
		}
	default:
		return fmt.Errorf("missing event context")
	}

	// Let's do a check on the error
	if stream.Error != nil {
		return fmt.Errorf("error while writing the event attributes: %w", stream.Error)
	}

	// Let's write the body
	if in.DataEncoded != nil {
		stream.WriteMore()

		// We need to figure out the media type first
		var mediaType string
		if dct == nil {
			mediaType = ApplicationJSON
		} else {
			// This code is required to extract the media type from the full content type string (which might contain encoding and stuff)
			contentType := *dct
			i := strings.IndexRune(contentType, ';')
			if i == -1 {
				i = len(contentType)
			}
			mediaType = strings.TrimSpace(strings.ToLower(contentType[0:i]))
		}

		isJson := mediaType == "" || mediaType == ApplicationJSON || mediaType == TextJSON

		// If isJson and no encoding to base64, we don't need to perform additional steps
		if isJson && !isBase64 {
			stream.WriteObjectField("data")
			_, err := stream.Write(in.DataEncoded)
			if err != nil {
				return fmt.Errorf("error while writing data: %w", err)
			}
		} else {
			if in.Context.GetSpecVersion() == CloudEventsVersionV1 && isBase64 {
				stream.WriteObjectField("data_base64")
			} else {
				stream.WriteObjectField("data")
			}
			// At this point of we need to write to base 64 string, or we just need to write the plain string
			if isBase64 {
				stream.WriteString(base64.StdEncoding.EncodeToString(in.DataEncoded))
			} else {
				stream.WriteString(string(in.DataEncoded))
			}
		}

	}

	// Let's do a check on the error
	if stream.Error != nil {
		return fmt.Errorf("error while writing the event data: %w", stream.Error)
	}

	for k, v := range ext {
		stream.WriteMore()
		stream.WriteObjectField(k)
		stream.WriteVal(v)
	}

	stream.WriteObjectEnd()

	// Let's do a check on the error
	if stream.Error != nil {
		return fmt.Errorf("error while writing the event extensions: %w", stream.Error)
	}
	return stream.Flush()
}

// MarshalJSON implements a custom json marshal method used when this type is
// marshaled using json.Marshal.
func (e Event) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	err := WriteJson(&e, &buf)
	return buf.Bytes(), err
}
