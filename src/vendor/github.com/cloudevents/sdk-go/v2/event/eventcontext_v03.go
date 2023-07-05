/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"encoding/json"
	"fmt"
	"mime"
	"sort"
	"strings"

	"github.com/cloudevents/sdk-go/v2/types"
)

const (
	// CloudEventsVersionV03 represents the version 0.3 of the CloudEvents spec.
	CloudEventsVersionV03 = "0.3"
)

var specV03Attributes = map[string]struct{}{
	"type":                {},
	"source":              {},
	"subject":             {},
	"id":                  {},
	"time":                {},
	"schemaurl":           {},
	"datacontenttype":     {},
	"datacontentencoding": {},
}

// EventContextV03 represents the non-data attributes of a CloudEvents v0.3
// event.
type EventContextV03 struct {
	// Type - The type of the occurrence which has happened.
	Type string `json:"type"`
	// Source - A URI describing the event producer.
	Source types.URIRef `json:"source"`
	// Subject - The subject of the event in the context of the event producer
	// (identified by `source`).
	Subject *string `json:"subject,omitempty"`
	// ID of the event; must be non-empty and unique within the scope of the producer.
	ID string `json:"id"`
	// Time - A Timestamp when the event happened.
	Time *types.Timestamp `json:"time,omitempty"`
	// DataSchema - A link to the schema that the `data` attribute adheres to.
	SchemaURL *types.URIRef `json:"schemaurl,omitempty"`
	// GetDataMediaType - A MIME (RFC2046) string describing the media type of `data`.
	DataContentType *string `json:"datacontenttype,omitempty"`
	// DeprecatedDataContentEncoding describes the content encoding for the `data` attribute. Valid: nil, `Base64`.
	DataContentEncoding *string `json:"datacontentencoding,omitempty"`
	// Extensions - Additional extension metadata beyond the base spec.
	Extensions map[string]interface{} `json:"-"`
}

// Adhere to EventContext
var _ EventContext = (*EventContextV03)(nil)

// ExtensionAs implements EventContext.ExtensionAs
func (ec EventContextV03) ExtensionAs(name string, obj interface{}) error {
	value, ok := ec.Extensions[name]
	if !ok {
		return fmt.Errorf("extension %q does not exist", name)
	}

	// Try to unmarshal extension if we find it as a RawMessage.
	switch v := value.(type) {
	case json.RawMessage:
		if err := json.Unmarshal(v, obj); err == nil {
			// if that worked, return with obj set.
			return nil
		}
	}
	// else try as a string ptr.

	// Only support *string for now.
	switch v := obj.(type) {
	case *string:
		if valueAsString, ok := value.(string); ok {
			*v = valueAsString
			return nil
		} else {
			return fmt.Errorf("invalid type for extension %q", name)
		}
	default:
		return fmt.Errorf("unknown extension type %T", obj)
	}
}

// SetExtension adds the extension 'name' with value 'value' to the CloudEvents
// context. This function fails if the name uses a reserved event context key.
func (ec *EventContextV03) SetExtension(name string, value interface{}) error {
	if ec.Extensions == nil {
		ec.Extensions = make(map[string]interface{})
	}

	if _, ok := specV03Attributes[strings.ToLower(name)]; ok {
		return fmt.Errorf("bad key %q: CloudEvents spec attribute MUST NOT be overwritten by extension", name)
	}

	if value == nil {
		delete(ec.Extensions, name)
		if len(ec.Extensions) == 0 {
			ec.Extensions = nil
		}
		return nil
	} else {
		v, err := types.Validate(value)
		if err == nil {
			ec.Extensions[name] = v
		}
		return err
	}
}

// Clone implements EventContextConverter.Clone
func (ec EventContextV03) Clone() EventContext {
	ec03 := ec.AsV03()
	ec03.Source = types.Clone(ec.Source).(types.URIRef)
	if ec.Time != nil {
		ec03.Time = types.Clone(ec.Time).(*types.Timestamp)
	}
	if ec.SchemaURL != nil {
		ec03.SchemaURL = types.Clone(ec.SchemaURL).(*types.URIRef)
	}
	ec03.Extensions = ec.cloneExtensions()
	return ec03
}

func (ec *EventContextV03) cloneExtensions() map[string]interface{} {
	old := ec.Extensions
	if old == nil {
		return nil
	}
	new := make(map[string]interface{}, len(ec.Extensions))
	for k, v := range old {
		new[k] = types.Clone(v)
	}
	return new
}

// AsV03 implements EventContextConverter.AsV03
func (ec EventContextV03) AsV03() *EventContextV03 {
	return &ec
}

// AsV1 implements EventContextConverter.AsV1
func (ec EventContextV03) AsV1() *EventContextV1 {
	ret := EventContextV1{
		ID:              ec.ID,
		Time:            ec.Time,
		Type:            ec.Type,
		DataContentType: ec.DataContentType,
		Source:          types.URIRef{URL: ec.Source.URL},
		Subject:         ec.Subject,
		Extensions:      make(map[string]interface{}),
	}
	if ec.SchemaURL != nil {
		ret.DataSchema = &types.URI{URL: ec.SchemaURL.URL}
	}

	// DataContentEncoding was removed in 1.0, so put it in an extension for 1.0.
	if ec.DataContentEncoding != nil {
		_ = ret.SetExtension(DataContentEncodingKey, *ec.DataContentEncoding)
	}

	if ec.Extensions != nil {
		for k, v := range ec.Extensions {
			k = strings.ToLower(k)
			ret.Extensions[k] = v
		}
	}
	if len(ret.Extensions) == 0 {
		ret.Extensions = nil
	}
	return &ret
}

// Validate returns errors based on requirements from the CloudEvents spec.
// For more details, see https://github.com/cloudevents/spec/blob/master/spec.md
// As of Feb 26, 2019, commit 17c32ea26baf7714ad027d9917d03d2fff79fc7e
// + https://github.com/cloudevents/spec/pull/387 -> datacontentencoding
// + https://github.com/cloudevents/spec/pull/406 -> subject
func (ec EventContextV03) Validate() ValidationError {
	errors := map[string]error{}

	// type
	// Type: String
	// Constraints:
	//  REQUIRED
	//  MUST be a non-empty string
	//  SHOULD be prefixed with a reverse-DNS name. The prefixed domain dictates the organization which defines the semantics of this event type.
	eventType := strings.TrimSpace(ec.Type)
	if eventType == "" {
		errors["type"] = fmt.Errorf("MUST be a non-empty string")
	}

	// source
	// Type: URI-reference
	// Constraints:
	//  REQUIRED
	source := strings.TrimSpace(ec.Source.String())
	if source == "" {
		errors["source"] = fmt.Errorf("REQUIRED")
	}

	// subject
	// Type: String
	// Constraints:
	//  OPTIONAL
	//  MUST be a non-empty string
	if ec.Subject != nil {
		subject := strings.TrimSpace(*ec.Subject)
		if subject == "" {
			errors["subject"] = fmt.Errorf("if present, MUST be a non-empty string")
		}
	}

	// id
	// Type: String
	// Constraints:
	//  REQUIRED
	//  MUST be a non-empty string
	//  MUST be unique within the scope of the producer
	id := strings.TrimSpace(ec.ID)
	if id == "" {
		errors["id"] = fmt.Errorf("MUST be a non-empty string")

		// no way to test "MUST be unique within the scope of the producer"
	}

	// time
	// Type: Timestamp
	// Constraints:
	//  OPTIONAL
	//  If present, MUST adhere to the format specified in RFC 3339
	// --> no need to test this, no way to set the time without it being valid.

	// schemaurl
	// Type: URI
	// Constraints:
	//  OPTIONAL
	//  If present, MUST adhere to the format specified in RFC 3986
	if ec.SchemaURL != nil {
		schemaURL := strings.TrimSpace(ec.SchemaURL.String())
		// empty string is not RFC 3986 compatible.
		if schemaURL == "" {
			errors["schemaurl"] = fmt.Errorf("if present, MUST adhere to the format specified in RFC 3986")
		}
	}

	// datacontenttype
	// Type: String per RFC 2046
	// Constraints:
	//  OPTIONAL
	//  If present, MUST adhere to the format specified in RFC 2046
	if ec.DataContentType != nil {
		dataContentType := strings.TrimSpace(*ec.DataContentType)
		if dataContentType == "" {
			errors["datacontenttype"] = fmt.Errorf("if present, MUST adhere to the format specified in RFC 2046")
		} else {
			_, _, err := mime.ParseMediaType(dataContentType)
			if err != nil {
				errors["datacontenttype"] = fmt.Errorf("if present, MUST adhere to the format specified in RFC 2046")
			}
		}
	}

	// datacontentencoding
	// Type: String per RFC 2045 Section 6.1
	// Constraints:
	//  The attribute MUST be set if the data attribute contains string-encoded binary data.
	//    Otherwise the attribute MUST NOT be set.
	//  If present, MUST adhere to RFC 2045 Section 6.1
	if ec.DataContentEncoding != nil {
		dataContentEncoding := strings.ToLower(strings.TrimSpace(*ec.DataContentEncoding))
		if dataContentEncoding != Base64 {
			errors["datacontentencoding"] = fmt.Errorf("if present, MUST adhere to RFC 2045 Section 6.1")
		}
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// String returns a pretty-printed representation of the EventContext.
func (ec EventContextV03) String() string {
	b := strings.Builder{}

	b.WriteString("Context Attributes,\n")

	b.WriteString("  specversion: " + CloudEventsVersionV03 + "\n")
	b.WriteString("  type: " + ec.Type + "\n")
	b.WriteString("  source: " + ec.Source.String() + "\n")
	if ec.Subject != nil {
		b.WriteString("  subject: " + *ec.Subject + "\n")
	}
	b.WriteString("  id: " + ec.ID + "\n")
	if ec.Time != nil {
		b.WriteString("  time: " + ec.Time.String() + "\n")
	}
	if ec.SchemaURL != nil {
		b.WriteString("  schemaurl: " + ec.SchemaURL.String() + "\n")
	}
	if ec.DataContentType != nil {
		b.WriteString("  datacontenttype: " + *ec.DataContentType + "\n")
	}
	if ec.DataContentEncoding != nil {
		b.WriteString("  datacontentencoding: " + *ec.DataContentEncoding + "\n")
	}

	if ec.Extensions != nil && len(ec.Extensions) > 0 {
		b.WriteString("Extensions,\n")
		keys := make([]string, 0, len(ec.Extensions))
		for k := range ec.Extensions {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, key := range keys {
			b.WriteString(fmt.Sprintf("  %s: %v\n", key, ec.Extensions[key]))
		}
	}

	return b.String()
}
