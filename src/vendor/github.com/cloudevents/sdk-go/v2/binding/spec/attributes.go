/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package spec

import (
	"fmt"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/cloudevents/sdk-go/v2/types"
)

// Kind is a version-independent identifier for a CloudEvent context attribute.
type Kind uint8

const (
	// Required cloudevents attributes
	ID Kind = iota
	Source
	SpecVersion
	Type
	// Optional cloudevents attributes
	DataContentType
	DataSchema
	Subject
	Time
)
const nAttrs = int(Time) + 1

var kindNames = [nAttrs]string{
	"id",
	"source",
	"specversion",
	"type",
	"datacontenttype",
	"dataschema",
	"subject",
	"time",
}

// String is a human-readable string, for a valid attribute name use Attribute.Name
func (k Kind) String() string { return kindNames[k] }

// IsRequired returns true for attributes defined as "required" by the CE spec.
func (k Kind) IsRequired() bool { return k < DataContentType }

// Attribute is a named attribute accessor.
// The attribute name is specific to a Version.
type Attribute interface {
	Kind() Kind
	// Name of the attribute with respect to the current spec Version() with prefix
	PrefixedName() string
	// Name of the attribute with respect to the current spec Version()
	Name() string
	// Version of the spec that this attribute belongs to
	Version() Version
	// Get the value of this attribute from an event context
	Get(event.EventContextReader) interface{}
	// Set the value of this attribute on an event context
	Set(event.EventContextWriter, interface{}) error
	// Delete this attribute from and event context, when possible
	Delete(event.EventContextWriter) error
}

// accessor provides Kind, Get, Set.
type accessor interface {
	Kind() Kind
	Get(event.EventContextReader) interface{}
	Set(event.EventContextWriter, interface{}) error
	Delete(event.EventContextWriter) error
}

var acc = [nAttrs]accessor{
	&aStr{aKind(ID), event.EventContextReader.GetID, event.EventContextWriter.SetID},
	&aStr{aKind(Source), event.EventContextReader.GetSource, event.EventContextWriter.SetSource},
	&aStr{aKind(SpecVersion), event.EventContextReader.GetSpecVersion, func(writer event.EventContextWriter, s string) error { return nil }},
	&aStr{aKind(Type), event.EventContextReader.GetType, event.EventContextWriter.SetType},
	&aStr{aKind(DataContentType), event.EventContextReader.GetDataContentType, event.EventContextWriter.SetDataContentType},
	&aStr{aKind(DataSchema), event.EventContextReader.GetDataSchema, event.EventContextWriter.SetDataSchema},
	&aStr{aKind(Subject), event.EventContextReader.GetSubject, event.EventContextWriter.SetSubject},
	&aTime{aKind(Time), event.EventContextReader.GetTime, event.EventContextWriter.SetTime},
}

// aKind implements Kind()
type aKind Kind

func (kind aKind) Kind() Kind { return Kind(kind) }

type aStr struct {
	aKind
	get func(event.EventContextReader) string
	set func(event.EventContextWriter, string) error
}

func (a *aStr) Get(c event.EventContextReader) interface{} {
	if s := a.get(c); s != "" {
		return s
	}
	return nil // Treat blank as missing
}

func (a *aStr) Set(c event.EventContextWriter, v interface{}) error {
	s, err := types.ToString(v)
	if err != nil {
		return fmt.Errorf("invalid value for %s: %#v", a.Kind(), v)
	}
	return a.set(c, s)
}

func (a *aStr) Delete(c event.EventContextWriter) error {
	return a.set(c, "")
}

type aTime struct {
	aKind
	get func(event.EventContextReader) time.Time
	set func(event.EventContextWriter, time.Time) error
}

func (a *aTime) Get(c event.EventContextReader) interface{} {
	if v := a.get(c); !v.IsZero() {
		return v
	}
	return nil // Treat zero time as missing.
}

func (a *aTime) Set(c event.EventContextWriter, v interface{}) error {
	t, err := types.ToTime(v)
	if err != nil {
		return fmt.Errorf("invalid value for %s: %#v", a.Kind(), v)
	}
	return a.set(c, t)
}

func (a *aTime) Delete(c event.EventContextWriter) error {
	return a.set(c, time.Time{})
}
