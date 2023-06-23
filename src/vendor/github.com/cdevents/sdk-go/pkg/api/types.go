/*
Copyright 2022 The CDEvents Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"golang.org/x/mod/semver"
)

const (
	EventTypeRoot             = "dev.cdevents"
	CDEventsSpecVersion       = "0.3.0"
	CDEventsSchemaURLTemplate = "https://cdevents.dev/%s/schema/%s-%s-event"
	CDEventsTypeRegex         = "^dev\\.cdevents\\.(?P<subject>[a-z]+)\\.(?P<predicate>[a-z]+)\\.(?P<version>.*)$"
)

var (
	CDEventsTypeCRegex = regexp.MustCompile(CDEventsTypeRegex)

	// CDEventsByUnversionedTypes maps non-versioned event types with events
	// set-pup at init time
	CDEventsByUnversionedTypes map[string]CDEvent
)

type Context struct {
	// Spec: https://cdevents.dev/docs/spec/#version
	// Description: The version of the CDEvents specification which the event
	// uses. This enables the interpretation of the context. Compliant event
	// producers MUST use a value of draft when referring to this version of the
	// specification.
	Version string `json:"version" jsonschema:"required"`

	// Spec: https://cdevents.dev/docs/spec/#id
	// Description: Identifier for an event. Subsequent delivery attempts of the
	// same event MAY share the same id. This attribute matches the syntax and
	// semantics of the id attribute of CloudEvents:
	// https://github.com/cloudevents/spec/blob/v1.0.2/cloudevents/spec.md#id
	Id string `json:"id" jsonschema:"required,minLength=1"`

	// Spec: https://cdevents.dev/docs/spec/#source
	// Description: defines the context in which an event happened. The main
	// purpose of the source is to provide global uniqueness for source + id.
	// The source MAY identify a single producer or a group of producer that
	// belong to the same application.
	Source string `json:"source" jsonschema:"required,minLength=1" validate:"uri-reference"`

	// Spec: https://cdevents.dev/docs/spec/#type
	// Description: defines the type of event, as combination of a subject and
	// predicate. Valid event types are defined in the vocabulary. All event
	// types should be prefixed with dev.cdevents.
	// One occurrence may have multiple events associated, as long as they have
	// different event types
	Type string `json:"type" jsonschema:"required,minLength=1" validate:"event-type"`

	// Spec: https://cdevents.dev/docs/spec/#timestamp
	// Description: Description: defines the time of the occurrence. When the
	// time of the occurrence is not available, the time when the event was
	// produced MAY be used. In case the transport layer should require a
	// re-transmission of the event, the timestamp SHOULD NOT be updated, i.e.
	// it should be the same for the same source + id combination.
	Timestamp time.Time `json:"timestamp" jsonschema:"required"`
}

type Reference struct {

	// Spec: https://cdevents.dev/docs/spec/#format-of-subjects
	// Description: Uniquely identifies the subject within the source
	Id string `json:"id" jsonschema:"required,minLength=1"`

	// Spec: https://cdevents.dev/docs/spec/#format-of-subjects
	// Description: defines the context in which an event happened. The main
	// purpose of the source is to provide global uniqueness for source + id.
	// The source MAY identify a single producer or a group of producer that
	// belong to the same application.
	Source string `json:"source,omitempty" validate:"uri-reference"`
}

type SubjectBase struct {
	Reference

	// The type of subject. Constraints what is a valid valid SubjectContent
	Type SubjectType `json:"type" jsonschema:"required,minLength=1"`
}

type SubjectType string

func (t SubjectType) String() string {
	return string(t)
}

type Subject interface {
	GetSubjectType() SubjectType
}

type CDEventType struct {
	Subject   string
	Predicate string

	// Version is a semantic version in the form <major>.<minor>.<patch>
	Version string
}

func (t CDEventType) String() string {
	return EventTypeRoot + "." + t.Subject + "." + t.Predicate + "." + t.Version
}

func (t CDEventType) UnversionedString() string {
	return EventTypeRoot + "." + t.Subject + "." + t.Predicate
}

func (t CDEventType) Short() string {
	return t.Subject + "_" + t.Predicate
}

// Two CDEventTypes are compatible if the subject and predicates
// are identical and they share the same major version
func (t CDEventType) IsCompatible(other CDEventType) bool {
	return t.Predicate == other.Predicate &&
		t.Subject == other.Subject &&
		semver.Major("v"+t.Version) == semver.Major("v"+other.Version)
}

func CDEventTypeFromString(cdeventType string) (*CDEventType, error) {
	parts := CDEventsTypeCRegex.FindStringSubmatch(cdeventType)
	if len(parts) != 4 {
		return nil, fmt.Errorf("cannot parse event type %s", cdeventType)
	}
	return &CDEventType{
		Subject:   parts[1],
		Predicate: parts[2],
		Version:   parts[3],
	}, nil
}

type CDEventReader interface {

	// The CDEventType "dev.cdevents.*"
	GetType() CDEventType

	// The CDEvents specification version implemented
	GetVersion() string

	// The event ID, unique for this event within the event producer (source)
	GetId() string

	// The source of the event
	GetSource() string

	// The time when the occurrence described in the event happened, or when
	// the event was produced if the former is not available
	GetTimestamp() time.Time

	// The ID of the subject, unique within the event producer (source), it may
	// by used in multiple events
	GetSubjectId() string

	// The source of the subject. Usually this matches the source of the event
	// but it may also be different.
	GetSubjectSource() string

	// The event specific subject. It is possible to use a type assertion with
	// the generic Subject to obtain an event specific implementation of Subject
	// for direct access to the content fields
	GetSubject() Subject

	// The URL and content of the schema file associated to the event type
	GetSchema() (string, string)

	// The custom data attached to the event
	// Depends on GetCustomDataContentType()
	// - When "application/json", un-marshalled data
	// - Else, raw []byte
	GetCustomData() (interface{}, error)

	// The raw custom data attached to the event
	GetCustomDataRaw() ([]byte, error)

	// Custom data un-marshalled into receiver, only if
	// GetCustomDataContentType() returns "application/json", else error
	GetCustomDataAs(receiver interface{}) error

	// Custom data content-type
	GetCustomDataContentType() string
}

type CDEventWriter interface {

	// The event ID, unique for this event within the event producer (source)
	SetId(id string)

	// The source of the event
	SetSource(source string)

	// The time when the occurrence described in the event happened, or when
	// the event was produced if the former is not available
	SetTimestamp(timestamp time.Time)

	// The ID of the subject, unique within the event producer (source), it may
	// by used in multiple events
	SetSubjectId(subjectId string)

	// The source of the subject. Usually this matches the source of the event
	// but it may also be different.
	SetSubjectSource(subjectSource string)

	// Set custom data. If contentType is "application/json", data can also be
	// anything that can be marshalled into json. For any other
	// content type, data must be passed as a []byte
	SetCustomData(contentType string, data interface{}) error
}

type CDEventCustomDataEncoding string

func (t CDEventCustomDataEncoding) String() string {
	return string(t)
}

// CDEventCustomData hosts the CDEvent custom data fields
//
// `CustomDataContentType` describes the content type of the data.
//
// `CustomData` contains the data:
//
// - When the content type is "application/json":
//
//   - if the CDEvent is produced via the golang API, the `CustomData`
//     can hold an un-marshalled golang interface{} of a specific type
//     or a marshalled byte slice
//
//   - if the CDEvent is consumed and thus un-marshalled from a []byte
//     the `CustomData` holds the data un-marshalled from []byte, into
//     a generic interface{}. It may be un-marshalled into a specific
//     golang type via the `GetCustomDataAs`
//
// - When the content type is anything else:
//
//   - if the CDEvent is produced via the golang API, the `CustomData`
//     hold an byte slice with the data passed via the API
//
//   - if the CDEvent is consumed and thus un-marshalled from a []byte
//     the `CustomData` holds the data base64 encoded
type CDEventCustomData struct {

	// CustomData added to the CDEvent. Format not specified by the SPEC.
	CustomData interface{} `json:"customData,omitempty" jsonschema:"oneof_type=object;string"`

	// CustomDataContentType for CustomData in a CDEvent.
	CustomDataContentType string `json:"customDataContentType,omitempty"`
}

type CDEvent interface {
	CDEventReader
	CDEventWriter
}

// Used to implement type specific GetCustomDataRaw()
func GetCustomDataRaw(contentType string, data interface{}) ([]byte, error) {
	switch data := data.(type) {
	case []byte:
		return data, nil
	default:
		if contentType != "application/json" && contentType != "" {
			return nil, fmt.Errorf("cannot use %v with content type %s", data, contentType)
		}
		// The content type is JSON, but the data is un-marshalled
		return json.Marshal(data)
	}
}

// Used to implement type specific GetCustomDataAs()
func GetCustomDataAs(e CDEventReader, receiver interface{}) error {
	contentType := e.GetCustomDataContentType()
	if contentType != "application/json" && contentType != "" {
		return fmt.Errorf("cannot unmarshal content-type %s", contentType)
	}
	data, err := e.GetCustomDataRaw()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, receiver)
}

// Used to implement type specific GetCustomData()
func GetCustomData(contentType string, data interface{}) (interface{}, error) {
	var v interface{}
	if contentType == "" {
		contentType = "application/json"
	}
	switch data := data.(type) {
	case []byte:
		// The data is JSON but still raw. Let's un-marshal it.
		if contentType == "application/json" {
			err := json.Unmarshal(data, &v)
			if err != nil {
				return nil, err
			}
			return v, nil
		}
		// The content type is not JSON, pass through raw data
		return data, nil
	case string:
		if contentType == "application/json" {
			return nil, fmt.Errorf("content type %s should not be a string: %s", contentType, data)
		}
		// If the data is not "application/json", and it's a string after
		// un-marshalling, we assume it's base64 encoded
		// NOTE(afrittoli) The standard un-marshaller would decode if the
		// receiving type was []byte, but we have interface because we need
		// to be able to store golang objects as well
		return b64.StdEncoding.DecodeString(data)
	default:
		if contentType != "application/json" {
			return nil, fmt.Errorf("cannot use %v with content type %s", data, contentType)
		}
		// The content type is JSON, pass through un-marshalled data
		return data, nil
	}
}

// Used to implement SetCustomData()
func CheckCustomData(contentType string, data interface{}) error {
	_, isBytes := data.([]byte)
	if !isBytes && contentType != "application/json" && contentType != "" {
		return fmt.Errorf("%s data must be set as []bytes, got %v", contentType, data)
	}
	return nil
}
