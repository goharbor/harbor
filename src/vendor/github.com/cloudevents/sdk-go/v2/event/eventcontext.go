/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import "time"

// EventContextReader are the methods required to be a reader of context
// attributes.
type EventContextReader interface {
	// GetSpecVersion returns the native CloudEvents Spec version of the event
	// context.
	GetSpecVersion() string
	// GetType returns the CloudEvents type from the context.
	GetType() string
	// GetSource returns the CloudEvents source from the context.
	GetSource() string
	// GetSubject returns the CloudEvents subject from the context.
	GetSubject() string
	// GetID returns the CloudEvents ID from the context.
	GetID() string
	// GetTime returns the CloudEvents creation time from the context.
	GetTime() time.Time
	// GetDataSchema returns the CloudEvents schema URL (if any) from the
	// context.
	GetDataSchema() string
	// GetDataContentType returns content type on the context.
	GetDataContentType() string
	// DeprecatedGetDataContentEncoding returns content encoding on the context.
	DeprecatedGetDataContentEncoding() string

	// GetDataMediaType returns the MIME media type for encoded data, which is
	// needed by both encoding and decoding. This is a processed form of
	// GetDataContentType and it may return an error.
	GetDataMediaType() (string, error)

	// DEPRECATED: Access extensions directly via the GetExtensions()
	// For example replace this:
	//
	//     var i int
	//     err := ec.ExtensionAs("foo", &i)
	//
	// With this:
	//
	//     i, err := types.ToInteger(ec.GetExtensions["foo"])
	//
	ExtensionAs(string, interface{}) error

	// GetExtensions returns the full extensions map.
	//
	// Extensions use the CloudEvents type system, details in package cloudevents/types.
	GetExtensions() map[string]interface{}

	// GetExtension returns the extension associated with with the given key.
	// The given key is case insensitive. If the extension can not be found,
	// an error will be returned.
	GetExtension(string) (interface{}, error)
}

// EventContextWriter are the methods required to be a writer of context
// attributes.
type EventContextWriter interface {
	// SetType sets the type of the context.
	SetType(string) error
	// SetSource sets the source of the context.
	SetSource(string) error
	// SetSubject sets the subject of the context.
	SetSubject(string) error
	// SetID sets the ID of the context.
	SetID(string) error
	// SetTime sets the time of the context.
	SetTime(time time.Time) error
	// SetDataSchema sets the schema url of the context.
	SetDataSchema(string) error
	// SetDataContentType sets the data content type of the context.
	SetDataContentType(string) error
	// DeprecatedSetDataContentEncoding sets the data context encoding of the context.
	DeprecatedSetDataContentEncoding(string) error

	// SetExtension sets the given interface onto the extension attributes
	// determined by the provided name.
	//
	// This function fails in V1 if the name doesn't respect the regex ^[a-zA-Z0-9]+$
	//
	// Package ./types documents the types that are allowed as extension values.
	SetExtension(string, interface{}) error
}

// EventContextConverter are the methods that allow for event version
// conversion.
type EventContextConverter interface {
	// AsV03 provides a translation from whatever the "native" encoding of the
	// CloudEvent was to the equivalent in v0.3 field names, moving fields to or
	// from extensions as necessary.
	AsV03() *EventContextV03

	// AsV1 provides a translation from whatever the "native" encoding of the
	// CloudEvent was to the equivalent in v1.0 field names, moving fields to or
	// from extensions as necessary.
	AsV1() *EventContextV1
}

// EventContext is conical interface for a CloudEvents Context.
type EventContext interface {
	// EventContextConverter allows for conversion between versions.
	EventContextConverter

	// EventContextReader adds methods for reading context.
	EventContextReader

	// EventContextWriter adds methods for writing to context.
	EventContextWriter

	// Validate the event based on the specifics of the CloudEvents spec version
	// represented by this event context.
	Validate() ValidationError

	// Clone clones the event context.
	Clone() EventContext

	// String returns a pretty-printed representation of the EventContext.
	String() string
}
