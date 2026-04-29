/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"time"
)

// EventReader is the interface for reading through an event from attributes.
type EventReader interface {
	// SpecVersion returns event.Context.GetSpecVersion().
	SpecVersion() string
	// Type returns event.Context.GetType().
	Type() string
	// Source returns event.Context.GetSource().
	Source() string
	// Subject returns event.Context.GetSubject().
	Subject() string
	// ID returns event.Context.GetID().
	ID() string
	// Time returns event.Context.GetTime().
	Time() time.Time
	// DataSchema returns event.Context.GetDataSchema().
	DataSchema() string
	// DataContentType returns event.Context.GetDataContentType().
	DataContentType() string
	// DataMediaType returns event.Context.GetDataMediaType().
	DataMediaType() string
	// DeprecatedDataContentEncoding returns event.Context.DeprecatedGetDataContentEncoding().
	DeprecatedDataContentEncoding() string

	// Extension Attributes

	// Extensions returns the event.Context.GetExtensions().
	// Extensions use the CloudEvents type system, details in package cloudevents/types.
	Extensions() map[string]interface{}

	// ExtensionAs returns event.Context.ExtensionAs(name, obj).
	//
	// DEPRECATED: Access extensions directly via the e.Extensions() map.
	// Use functions in the types package to convert extension values.
	// For example replace this:
	//
	//     var i int
	//     err := e.ExtensionAs("foo", &i)
	//
	// With this:
	//
	//     i, err := types.ToInteger(e.Extensions["foo"])
	//
	ExtensionAs(string, interface{}) error

	// Data Attribute

	// Data returns the raw data buffer
	// If the event was encoded with base64 encoding, Data returns the already decoded
	// byte array
	Data() []byte

	// DataAs attempts to populate the provided data object with the event payload.
	DataAs(interface{}) error
}

// EventWriter is the interface for writing through an event onto attributes.
// If an error is thrown by a sub-component, EventWriter caches the error
// internally and exposes errors with a call to event.Validate().
type EventWriter interface {
	// Context Attributes

	// SetSpecVersion performs event.Context.SetSpecVersion.
	SetSpecVersion(string)
	// SetType performs event.Context.SetType.
	SetType(string)
	// SetSource performs event.Context.SetSource.
	SetSource(string)
	// SetSubject( performs event.Context.SetSubject.
	SetSubject(string)
	// SetID performs event.Context.SetID.
	SetID(string)
	// SetTime performs event.Context.SetTime.
	SetTime(time.Time)
	// SetDataSchema performs event.Context.SetDataSchema.
	SetDataSchema(string)
	// SetDataContentType performs event.Context.SetDataContentType.
	SetDataContentType(string)
	// DeprecatedSetDataContentEncoding performs event.Context.DeprecatedSetDataContentEncoding.
	SetDataContentEncoding(string)

	// Extension Attributes

	// SetExtension performs event.Context.SetExtension.
	SetExtension(string, interface{})

	// SetData encodes the given payload with the given content type.
	// If the provided payload is a byte array, when marshalled to json it will be encoded as base64.
	// If the provided payload is different from byte array, datacodec.Encode is invoked to attempt a
	// marshalling to byte array.
	SetData(string, interface{}) error
}
