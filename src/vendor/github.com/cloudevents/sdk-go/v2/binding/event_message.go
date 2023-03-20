/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package binding

import (
	"bytes"
	"context"

	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/event"
)

type eventFormatKey int

const (
	formatEventStructured eventFormatKey = iota
)

// EventMessage type-converts a event.Event object to implement Message.
// This allows local event.Event objects to be sent directly via Sender.Send()
//     s.Send(ctx, binding.EventMessage(e))
// When an event is wrapped into a EventMessage, the original event could be
// potentially mutated. If you need to use the Event again, after wrapping it into
// an Event message, you should copy it before
type EventMessage event.Event

func ToMessage(e *event.Event) Message {
	return (*EventMessage)(e)
}

func (m *EventMessage) ReadEncoding() Encoding {
	return EncodingEvent
}

func (m *EventMessage) ReadStructured(ctx context.Context, builder StructuredWriter) error {
	f := GetOrDefaultFromCtx(ctx, formatEventStructured, format.JSON).(format.Format)
	b, err := f.Marshal((*event.Event)(m))
	if err != nil {
		return err
	}
	return builder.SetStructuredEvent(ctx, f, bytes.NewReader(b))
}

func (m *EventMessage) ReadBinary(ctx context.Context, b BinaryWriter) (err error) {
	err = eventContextToBinaryWriter(m.Context, b)
	if err != nil {
		return err
	}
	// Pass the body
	body := (*event.Event)(m).Data()
	if len(body) > 0 {
		err = b.SetData(bytes.NewBuffer(body))
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *EventMessage) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	sv := spec.VS.Version(m.Context.GetSpecVersion())
	a := sv.AttributeFromKind(k)
	if a != nil {
		return a, a.Get(m.Context)
	}
	return nil, nil
}

func (m *EventMessage) GetExtension(name string) interface{} {
	ext, _ := m.Context.GetExtension(name)
	return ext
}

func eventContextToBinaryWriter(c event.EventContext, b BinaryWriter) (err error) {
	// Pass all attributes
	sv := spec.VS.Version(c.GetSpecVersion())
	for _, a := range sv.Attributes() {
		value := a.Get(c)
		if value != nil {
			err = b.SetAttribute(a, value)
		}
		if err != nil {
			return err
		}
	}
	// Pass all extensions
	for k, v := range c.GetExtensions() {
		err = b.SetExtension(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (*EventMessage) Finish(error) error { return nil }

var _ Message = (*EventMessage)(nil)               // Test it conforms to the interface
var _ MessageMetadataReader = (*EventMessage)(nil) // Test it conforms to the interface

// UseFormatForEvent configures which format to use when marshalling the event to structured mode
func UseFormatForEvent(ctx context.Context, f format.Format) context.Context {
	return context.WithValue(ctx, formatEventStructured, f)
}
