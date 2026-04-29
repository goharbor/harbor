/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package spec

import (
	"strings"

	"github.com/cloudevents/sdk-go/v2/event"
)

// Version provides meta-data for a single spec-version.
type Version interface {
	// String name of the version, e.g. "1.0"
	String() string
	// Prefix for attribute names.
	Prefix() string
	// Attribute looks up a prefixed attribute name (case insensitive).
	// Returns nil if not found.
	Attribute(prefixedName string) Attribute
	// Attribute looks up the attribute from kind.
	// Returns nil if not found.
	AttributeFromKind(kind Kind) Attribute
	// Attributes returns all the context attributes for this version.
	Attributes() []Attribute
	// Convert translates a context to this version.
	Convert(event.EventContextConverter) event.EventContext
	// NewContext returns a new context for this version.
	NewContext() event.EventContext
	// SetAttribute sets named attribute to value.
	//
	// Name is case insensitive.
	// Does nothing if name does not start with prefix.
	SetAttribute(context event.EventContextWriter, name string, value interface{}) error
}

// Versions contains all known versions with the same attribute prefix.
type Versions struct {
	prefix string
	all    []Version
	m      map[string]Version
}

// Versions returns the list of all known versions, most recent first.
func (vs *Versions) Versions() []Version { return vs.all }

// Version returns the named version.
func (vs *Versions) Version(name string) Version {
	return vs.m[name]
}

// Latest returns the latest Version
func (vs *Versions) Latest() Version { return vs.all[0] }

// PrefixedSpecVersionName returns the specversion attribute PrefixedName
func (vs *Versions) PrefixedSpecVersionName() string { return vs.prefix + "specversion" }

// Prefix is the lowercase attribute name prefix.
func (vs *Versions) Prefix() string { return vs.prefix }

type attribute struct {
	accessor
	name    string
	version Version
}

func (a *attribute) PrefixedName() string { return a.version.Prefix() + a.name }
func (a *attribute) Name() string         { return a.name }
func (a *attribute) Version() Version     { return a.version }

type version struct {
	prefix  string
	context event.EventContext
	convert func(event.EventContextConverter) event.EventContext
	attrMap map[string]Attribute
	attrs   []Attribute
}

func (v *version) Attribute(name string) Attribute { return v.attrMap[strings.ToLower(name)] }
func (v *version) Attributes() []Attribute         { return v.attrs }
func (v *version) String() string                  { return v.context.GetSpecVersion() }
func (v *version) Prefix() string                  { return v.prefix }
func (v *version) NewContext() event.EventContext  { return v.context.Clone() }

// HasPrefix is a case-insensitive prefix check.
func (v *version) HasPrefix(name string) bool {
	return strings.HasPrefix(strings.ToLower(name), v.prefix)
}

func (v *version) Convert(c event.EventContextConverter) event.EventContext { return v.convert(c) }

func (v *version) SetAttribute(c event.EventContextWriter, name string, value interface{}) error {
	if a := v.Attribute(name); a != nil { // Standard attribute
		return a.Set(c, value)
	}
	name = strings.ToLower(name)
	var err error
	if v.HasPrefix(name) { // Extension attribute
		return c.SetExtension(strings.TrimPrefix(name, v.prefix), value)
	}
	return err
}

func (v *version) AttributeFromKind(kind Kind) Attribute {
	for _, a := range v.Attributes() {
		if a.Kind() == kind {
			return a
		}
	}
	return nil
}

func newVersion(
	prefix string,
	context event.EventContext,
	convert func(event.EventContextConverter) event.EventContext,
	attrs ...*attribute,
) *version {
	v := &version{
		prefix:  strings.ToLower(prefix),
		context: context,
		convert: convert,
		attrMap: map[string]Attribute{},
		attrs:   make([]Attribute, len(attrs)),
	}
	for i, a := range attrs {
		a.version = v
		v.attrs[i] = a
		v.attrMap[strings.ToLower(a.PrefixedName())] = a
	}
	return v
}

// WithPrefix returns a set of versions with prefix added to all attribute names.
func WithPrefix(prefix string) *Versions {
	attr := func(name string, kind Kind) *attribute {
		return &attribute{accessor: acc[kind], name: name}
	}
	vs := &Versions{
		m:      map[string]Version{},
		prefix: prefix,
		all: []Version{
			newVersion(prefix, event.EventContextV1{}.AsV1(),
				func(c event.EventContextConverter) event.EventContext { return c.AsV1() },
				attr("id", ID),
				attr("source", Source),
				attr("specversion", SpecVersion),
				attr("type", Type),
				attr("datacontenttype", DataContentType),
				attr("dataschema", DataSchema),
				attr("subject", Subject),
				attr("time", Time),
			),
			newVersion(prefix, event.EventContextV03{}.AsV03(),
				func(c event.EventContextConverter) event.EventContext { return c.AsV03() },
				attr("specversion", SpecVersion),
				attr("type", Type),
				attr("source", Source),
				attr("schemaurl", DataSchema),
				attr("subject", Subject),
				attr("id", ID),
				attr("time", Time),
				attr("datacontenttype", DataContentType),
			),
		},
	}
	for _, v := range vs.all {
		vs.m[v.String()] = v
	}
	return vs
}

// New returns a set of versions
func New() *Versions { return WithPrefix("") }

// Built-in un-prefixed versions.
var (
	VS  *Versions
	V03 Version
	V1  Version
)

func init() {
	VS = New()
	V03 = VS.Version("0.3")
	V1 = VS.Version("1.0")
}
