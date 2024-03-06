/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package spec

import (
	"github.com/cloudevents/sdk-go/v2/event"
)

type matchExactVersion struct {
	version
}

func (v *matchExactVersion) Attribute(name string) Attribute { return v.attrMap[name] }

var _ Version = (*matchExactVersion)(nil)

func newMatchExactVersionVersion(
	prefix string,
	attributeNameMatchMapper func(string) string,
	context event.EventContext,
	convert func(event.EventContextConverter) event.EventContext,
	attrs ...*attribute,
) *matchExactVersion {
	v := &matchExactVersion{
		version: version{
			prefix:  prefix,
			context: context,
			convert: convert,
			attrMap: map[string]Attribute{},
			attrs:   make([]Attribute, len(attrs)),
		},
	}
	for i, a := range attrs {
		a.version = v
		v.attrs[i] = a
		v.attrMap[attributeNameMatchMapper(a.name)] = a
	}
	return v
}

// WithPrefixMatchExact returns a set of versions with prefix added to all attribute names.
func WithPrefixMatchExact(attributeNameMatchMapper func(string) string, prefix string) *Versions {
	attr := func(name string, kind Kind) *attribute {
		return &attribute{accessor: acc[kind], name: name}
	}
	vs := &Versions{
		m:      map[string]Version{},
		prefix: prefix,
		all: []Version{
			newMatchExactVersionVersion(prefix, attributeNameMatchMapper, event.EventContextV1{}.AsV1(),
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
			newMatchExactVersionVersion(prefix, attributeNameMatchMapper, event.EventContextV03{}.AsV03(),
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
