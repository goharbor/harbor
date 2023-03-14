/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudevents/sdk-go/v2/event"
)

// Format marshals and unmarshals structured events to bytes.
type Format interface {
	// MediaType identifies the format
	MediaType() string
	// Marshal event to bytes
	Marshal(*event.Event) ([]byte, error)
	// Unmarshal bytes to event
	Unmarshal([]byte, *event.Event) error
}

// Prefix for event-format media types.
const Prefix = "application/cloudevents"

// IsFormat returns true if mediaType begins with "application/cloudevents"
func IsFormat(mediaType string) bool { return strings.HasPrefix(mediaType, Prefix) }

// JSON is the built-in "application/cloudevents+json" format.
var JSON = jsonFmt{}

type jsonFmt struct{}

func (jsonFmt) MediaType() string { return event.ApplicationCloudEventsJSON }

func (jsonFmt) Marshal(e *event.Event) ([]byte, error) { return json.Marshal(e) }
func (jsonFmt) Unmarshal(b []byte, e *event.Event) error {
	return json.Unmarshal(b, e)
}

// built-in formats
var formats map[string]Format

func init() {
	formats = map[string]Format{}
	Add(JSON)
}

// Lookup returns the format for contentType, or nil if not found.
func Lookup(contentType string) Format {
	i := strings.IndexRune(contentType, ';')
	if i == -1 {
		i = len(contentType)
	}
	contentType = strings.TrimSpace(strings.ToLower(contentType[0:i]))
	return formats[contentType]
}

func unknown(mediaType string) error {
	return fmt.Errorf("unknown event format media-type %#v", mediaType)
}

// Add a new Format. It can be retrieved by Lookup(f.MediaType())
func Add(f Format) { formats[f.MediaType()] = f }

// Marshal an event to bytes using the mediaType event format.
func Marshal(mediaType string, e *event.Event) ([]byte, error) {
	if f := formats[mediaType]; f != nil {
		return f.Marshal(e)
	}
	return nil, unknown(mediaType)
}

// Unmarshal bytes to an event using the mediaType event format.
func Unmarshal(mediaType string, b []byte, e *event.Event) error {
	if f := formats[mediaType]; f != nil {
		return f.Unmarshal(b, e)
	}
	return unknown(mediaType)
}
