/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"
	"time"
)

var _ EventWriter = (*Event)(nil)

// SetSpecVersion implements EventWriter.SetSpecVersion
func (e *Event) SetSpecVersion(v string) {
	switch v {
	case CloudEventsVersionV03:
		if e.Context == nil {
			e.Context = &EventContextV03{}
		} else {
			e.Context = e.Context.AsV03()
		}
	case CloudEventsVersionV1:
		if e.Context == nil {
			e.Context = &EventContextV1{}
		} else {
			e.Context = e.Context.AsV1()
		}
	default:
		e.fieldError("specversion", fmt.Errorf("a valid spec version is required: [%s, %s]",
			CloudEventsVersionV03, CloudEventsVersionV1))
		return
	}
	e.fieldOK("specversion")
}

// SetType implements EventWriter.SetType
func (e *Event) SetType(t string) {
	if err := e.Context.SetType(t); err != nil {
		e.fieldError("type", err)
	} else {
		e.fieldOK("type")
	}
}

// SetSource implements EventWriter.SetSource
func (e *Event) SetSource(s string) {
	if err := e.Context.SetSource(s); err != nil {
		e.fieldError("source", err)
	} else {
		e.fieldOK("source")
	}
}

// SetSubject implements EventWriter.SetSubject
func (e *Event) SetSubject(s string) {
	if err := e.Context.SetSubject(s); err != nil {
		e.fieldError("subject", err)
	} else {
		e.fieldOK("subject")
	}
}

// SetID implements EventWriter.SetID
func (e *Event) SetID(id string) {
	if err := e.Context.SetID(id); err != nil {
		e.fieldError("id", err)
	} else {
		e.fieldOK("id")
	}
}

// SetTime implements EventWriter.SetTime
func (e *Event) SetTime(t time.Time) {
	if err := e.Context.SetTime(t); err != nil {
		e.fieldError("time", err)
	} else {
		e.fieldOK("time")
	}
}

// SetDataSchema implements EventWriter.SetDataSchema
func (e *Event) SetDataSchema(s string) {
	if err := e.Context.SetDataSchema(s); err != nil {
		e.fieldError("dataschema", err)
	} else {
		e.fieldOK("dataschema")
	}
}

// SetDataContentType implements EventWriter.SetDataContentType
func (e *Event) SetDataContentType(ct string) {
	if err := e.Context.SetDataContentType(ct); err != nil {
		e.fieldError("datacontenttype", err)
	} else {
		e.fieldOK("datacontenttype")
	}
}

// SetDataContentEncoding is deprecated. Implements EventWriter.SetDataContentEncoding.
func (e *Event) SetDataContentEncoding(enc string) {
	if err := e.Context.DeprecatedSetDataContentEncoding(enc); err != nil {
		e.fieldError("datacontentencoding", err)
	} else {
		e.fieldOK("datacontentencoding")
	}
}

// SetExtension implements EventWriter.SetExtension
func (e *Event) SetExtension(name string, obj interface{}) {
	if err := e.Context.SetExtension(name, obj); err != nil {
		e.fieldError("extension:"+name, err)
	} else {
		e.fieldOK("extension:" + name)
	}
}
