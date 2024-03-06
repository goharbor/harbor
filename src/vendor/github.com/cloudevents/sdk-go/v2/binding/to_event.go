/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/types"
)

// ErrCannotConvertToEvent is a generic error when a conversion of a Message to an Event fails
var ErrCannotConvertToEvent = errors.New("cannot convert message to event")

// ErrCannotConvertToEvents is a generic error when a conversion of a Message to a Batched Event fails
var ErrCannotConvertToEvents = errors.New("cannot convert message to batched events")

// ToEvent translates a Message with a valid Structured or Binary representation to an Event.
// This function returns the Event generated from the Message and the original encoding of the message or
// an error that points the conversion error.
// transformers can be nil and this function guarantees that they are invoked only once during the encoding process.
func ToEvent(ctx context.Context, message MessageReader, transformers ...Transformer) (*event.Event, error) {
	if message == nil {
		return nil, nil
	}

	messageEncoding := message.ReadEncoding()
	if messageEncoding == EncodingEvent {
		m := message
		for m != nil {
			switch mt := m.(type) {
			case *EventMessage:
				e := (*event.Event)(mt)
				return e, Transformers(transformers).Transform(mt, (*messageToEventBuilder)(e))
			case MessageWrapper:
				m = mt.GetWrappedMessage()
			default:
				break
			}
		}
		return nil, ErrCannotConvertToEvent
	}

	e := event.New()
	encoder := (*messageToEventBuilder)(&e)
	_, err := DirectWrite(
		context.Background(),
		message,
		encoder,
		encoder,
	)
	if err != nil {
		return nil, err
	}
	return &e, Transformers(transformers).Transform((*EventMessage)(&e), encoder)
}

// ToEvents translates a Batch Message and corresponding Reader data to a slice of Events.
// This function returns the Events generated from the body data, or an error that points
// to the conversion issue.
func ToEvents(ctx context.Context, message MessageReader, body io.Reader) ([]event.Event, error) {
	messageEncoding := message.ReadEncoding()
	if messageEncoding != EncodingBatch {
		return nil, ErrCannotConvertToEvents
	}

	// Since Format doesn't support batch Marshalling, and we know it's structured batch json, we'll go direct to the
	// json.UnMarshall(), since that is the best way to support batch operations for now.
	var events []event.Event
	return events, json.NewDecoder(body).Decode(&events)
}

type messageToEventBuilder event.Event

var _ StructuredWriter = (*messageToEventBuilder)(nil)
var _ BinaryWriter = (*messageToEventBuilder)(nil)

func (b *messageToEventBuilder) SetStructuredEvent(ctx context.Context, format format.Format, ev io.Reader) error {
	var buf bytes.Buffer
	_, err := io.Copy(&buf, ev)
	if err != nil {
		return err
	}
	return format.Unmarshal(buf.Bytes(), (*event.Event)(b))
}

func (b *messageToEventBuilder) Start(ctx context.Context) error {
	return nil
}

func (b *messageToEventBuilder) End(ctx context.Context) error {
	return nil
}

func (b *messageToEventBuilder) SetData(data io.Reader) error {
	buf, ok := data.(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
		_, err := io.Copy(buf, data)
		if err != nil {
			return err
		}
	}
	if buf.Len() > 0 {
		b.DataEncoded = buf.Bytes()
	}
	return nil
}

func (b *messageToEventBuilder) SetAttribute(attribute spec.Attribute, value interface{}) error {
	if value == nil {
		_ = attribute.Delete(b.Context)
		return nil
	}
	// If spec version we need to change to right context struct
	if attribute.Kind() == spec.SpecVersion {
		str, err := types.ToString(value)
		if err != nil {
			return err
		}
		switch str {
		case event.CloudEventsVersionV03:
			b.Context = b.Context.AsV03()
		case event.CloudEventsVersionV1:
			b.Context = b.Context.AsV1()
		default:
			return fmt.Errorf("unrecognized event version %s", str)
		}
		return nil
	}
	return attribute.Set(b.Context, value)
}

func (b *messageToEventBuilder) SetExtension(name string, value interface{}) error {
	if value == nil {
		return b.Context.SetExtension(name, nil)
	}
	value, err := types.Validate(value)
	if err != nil {
		return err
	}
	return b.Context.SetExtension(name, value)
}
