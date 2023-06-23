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
	"encoding/json"
	"fmt"
	"net/url"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-playground/validator/v10"
	purl "github.com/package-url/packageurl-go"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v5"
	"golang.org/x/mod/semver"
)

var (
	// Validation helper as singleton
	validate *validator.Validate
)

func init() {
	// Register custom validators
	validate = validator.New()
	err := validate.RegisterValidation("event-type", ValidateEventType)
	panicOnError(err)
	err = validate.RegisterValidation("uri-reference", ValidateUriReference)
	panicOnError(err)
	err = validate.RegisterValidation("purl", ValidatePurl)
	panicOnError(err)
}

func panicOnError(err error) {
	if err != nil {
		panic(err.Error())
	}
}

// TODO(afrittoli) we may want to define something like:
// const CDEventsContentType = "application/cdevents+json"
// but it's not yet in the spec

// ParseType returns a CDEventType if eventType is a valid type
func ParseType(eventType string) (*CDEventType, error) {
	t, err := CDEventTypeFromString(eventType)
	if err != nil {
		return nil, err
	}
	_, ok := CDEventsByUnversionedTypes[t.UnversionedString()]
	if !ok {
		return nil, fmt.Errorf("unknown event type %s", t.UnversionedString())
	}
	if !semver.IsValid("v" + t.Version) {
		return nil, fmt.Errorf("invalid version format %s", t.Version)
	}
	return t, nil
}

func ValidateEventType(fl validator.FieldLevel) bool {
	_, err := ParseType(fl.Field().String())
	return err == nil
}

func ValidateUriReference(fl validator.FieldLevel) bool {
	_, err := url.Parse(fl.Field().String())
	return err == nil
}

func ValidatePurl(fl validator.FieldLevel) bool {
	_, err := purl.FromString(fl.Field().String())
	return err == nil
}

// AsCloudEvent renders a CDEvent as a CloudEvent
func AsCloudEvent(event CDEventReader) (*cloudevents.Event, error) {
	if event == nil {
		return nil, fmt.Errorf("nil CDEvent cannot be rendered as CloudEvent")
	}
	// Validate the event
	err := Validate(event)
	if err != nil {
		return nil, fmt.Errorf("cannot validate CDEvent %v", err)
	}
	ce := cloudevents.NewEvent()
	ce.SetSource(event.GetSource())
	ce.SetSubject(event.GetSubjectId())
	ce.SetType(event.GetType().String())
	err = ce.SetData(cloudevents.ApplicationJSON, event)
	return &ce, err
}

// AsJsonBytes renders a CDEvent as a JSON string
func AsJsonBytes(event CDEventReader) ([]byte, error) {
	if event == nil {
		return nil, nil
	}
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}

// AsJsonString renders a CDEvent as a JSON string
func AsJsonString(event CDEventReader) (string, error) {
	jsonBytes, err := AsJsonBytes(event)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// Validate checks the CDEvent against the JSON schema and validate constraints
func Validate(event CDEventReader) error {
	url, schema := event.GetSchema()
	sch, err := jsonschema.CompileString(url, schema)
	if err != nil {
		return fmt.Errorf("cannot compile jsonschema %s, %s", url, err)
	}
	var v interface{}
	jsonString, err := AsJsonString(event)
	if err != nil {
		return fmt.Errorf("cannot render the event %s as json %s", event, err)
	}
	if err := json.Unmarshal([]byte(jsonString), &v); err != nil {
		return fmt.Errorf("cannot unmarshal event json: %v", err)
	}
	// Validate the "jsonschema" tags
	err = sch.Validate(v)
	if err != nil {
		return err
	}
	// Validate the "validate" tags
	err = validate.Struct(event)
	if err != nil {
		return err
	}
	return nil
}

// Build a new CDEventReader from a JSON string
func NewFromJsonString(event string) (CDEvent, error) {
	return NewFromJsonBytes([]byte(event))
}

// Build a new CDEventReader from a JSON string as []bytes
func NewFromJsonBytes(event []byte) (CDEvent, error) {
	eventAux := &struct {
		Context Context `json:"context"`
	}{}
	err := json.Unmarshal(event, eventAux)
	if err != nil {
		return nil, err
	}
	eventType, err := ParseType(eventAux.Context.Type)
	if err != nil {
		return nil, err
	}
	receiver, ok := CDEventsByUnversionedTypes[eventType.UnversionedString()]
	if !ok {
		// This cannot really happen as ParseType checks if the type is known to the SDK
		return nil, fmt.Errorf("unknown event type %s", eventAux.Context.Type)
	}
	// Check if the receiver is compatible. It must have the same subject and predicate
	// and share the same major version.
	// If the minor version is different and the message received as a version that is
	// greater than the SDK one, some fields may be lost, as newer versions may add new
	// fields to the event specification.
	if !eventType.IsCompatible(receiver.GetType()) {
		return nil, fmt.Errorf("sdk event version %s not compatible with %s", receiver.GetType().Version, eventType.Version)
	}
	err = json.Unmarshal(event, receiver)
	if err != nil {
		return nil, err
	}
	return receiver, nil
}
