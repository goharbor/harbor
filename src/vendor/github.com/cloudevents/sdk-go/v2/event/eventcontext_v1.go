/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"
	"mime"
	"sort"
	"strings"

	"github.com/cloudevents/sdk-go/v2/types"
)

// WIP: AS OF SEP 20, 2019

const (
	// CloudEventsVersionV1 represents the version 1.0 of the CloudEvents spec.
	CloudEventsVersionV1 = "1.0"
)

var specV1Attributes = map[string]struct{}{
	"id":              {},
	"source":          {},
	"type":            {},
	"datacontenttype": {},
	"subject":         {},
	"time":            {},
	"specversion":     {},
	"dataschema":      {},
}

// EventContextV1 represents the non-data attributes of a CloudEvents v1.0
// event.
type EventContextV1 struct {
	// ID of the event; must be non-empty and unique within the scope of the producer.
	// +required
	ID string `json:"id"`
	// Source - A URI describing the event producer.
	// +required
	Source types.URIRef `json:"source"`
	// Type - The type of the occurrence which has happened.
	// +required
	Type string `json:"type"`

	// DataContentType - A MIME (RFC2046) string describing the media type of `data`.
	// +optional
	DataContentType *string `json:"datacontenttype,omitempty"`
	// Subject - The subject of the event in the context of the event producer
	// (identified by `source`).
	// +optional
	Subject *string `json:"subject,omitempty"`
	// Time - A Timestamp when the event happened.
	// +optional
	Time *types.Timestamp `json:"time,omitempty"`
	// DataSchema - A link to the schema that the `data` attribute adheres to.
	// +optional
	DataSchema *types.URI `json:"dataschema,omitempty"`

	// Extensions - Additional extension metadata beyond the base spec.
	// +optional
	Extensions map[string]interface{} `json:"-"`
}

// Adhere to EventContext
var _ EventContext = (*EventContextV1)(nil)

// ExtensionAs implements EventContext.ExtensionAs
func (ec EventContextV1) ExtensionAs(name string, obj interface{}) error {
	name = strings.ToLower(name)
	value, ok := ec.Extensions[name]
	if !ok {
		return fmt.Errorf("extension %q does not exist", name)
	}

	// Only support *string for now.
	if v, ok := obj.(*string); ok {
		if *v, ok = value.(string); ok {
			return nil
		}
	}
	return fmt.Errorf("unknown extension type %T", obj)
}

// SetExtension adds the extension 'name' with value 'value' to the CloudEvents
// context. This function fails if the name doesn't respect the regex
// ^[a-zA-Z0-9]+$ or if the name uses a reserved event context key.
func (ec *EventContextV1) SetExtension(name string, value interface{}) error {
	if err := validateExtensionName(name); err != nil {
		return err
	}

	if _, ok := specV1Attributes[strings.ToLower(name)]; ok {
		return fmt.Errorf("bad key %q: CloudEvents spec attribute MUST NOT be overwritten by extension", name)
	}

	name = strings.ToLower(name)
	if ec.Extensions == nil {
		ec.Extensions = make(map[string]interface{})
	}
	if value == nil {
		delete(ec.Extensions, name)
		if len(ec.Extensions) == 0 {
			ec.Extensions = nil
		}
		return nil
	} else {
		v, err := types.Validate(value) // Ensure it's a legal CE attribute value
		if err == nil {
			ec.Extensions[name] = v
		}
		return err
	}
}

// Clone implements EventContextConverter.Clone
func (ec EventContextV1) Clone() EventContext {
	ec1 := ec.AsV1()
	ec1.Source = types.Clone(ec.Source).(types.URIRef)
	if ec.Time != nil {
		ec1.Time = types.Clone(ec.Time).(*types.Timestamp)
	}
	if ec.DataSchema != nil {
		ec1.DataSchema = types.Clone(ec.DataSchema).(*types.URI)
	}
	ec1.Extensions = ec.cloneExtensions()
	return ec1
}

func (ec *EventContextV1) cloneExtensions() map[string]interface{} {
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
func (ec EventContextV1) AsV03() *EventContextV03 {
	ret := EventContextV03{
		ID:              ec.ID,
		Time:            ec.Time,
		Type:            ec.Type,
		DataContentType: ec.DataContentType,
		Source:          types.URIRef{URL: ec.Source.URL},
		Subject:         ec.Subject,
		Extensions:      make(map[string]interface{}),
	}

	if ec.DataSchema != nil {
		ret.SchemaURL = &types.URIRef{URL: ec.DataSchema.URL}
	}

	if ec.Extensions != nil {
		for k, v := range ec.Extensions {
			k = strings.ToLower(k)
			// DeprecatedDataContentEncoding was introduced in 0.3, removed in 1.0
			if strings.EqualFold(k, DataContentEncodingKey) {
				etv, ok := v.(string)
				if ok && etv != "" {
					ret.DataContentEncoding = &etv
				}
				continue
			}
			ret.Extensions[k] = v
		}
	}
	if len(ret.Extensions) == 0 {
		ret.Extensions = nil
	}
	return &ret
}

// AsV1 implements EventContextConverter.AsV1
func (ec EventContextV1) AsV1() *EventContextV1 {
	return &ec
}

// Validate returns errors based on requirements from the CloudEvents spec.
// For more details, see https://github.com/cloudevents/spec/blob/v1.0/spec.md.
func (ec EventContextV1) Validate() ValidationError {
	errors := map[string]error{}

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

	// source
	// Type: URI-reference
	// Constraints:
	//  REQUIRED
	//  MUST be a non-empty URI-reference
	//	An absolute URI is RECOMMENDED
	source := strings.TrimSpace(ec.Source.String())
	if source == "" {
		errors["source"] = fmt.Errorf("REQUIRED")
	}

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

	// The following attributes are optional but still have validation.

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
				errors["datacontenttype"] = fmt.Errorf("failed to parse RFC 2046 media type %w", err)
			}
		}
	}

	// dataschema
	// Type: URI
	// Constraints:
	//  OPTIONAL
	//  If present, MUST adhere to the format specified in RFC 3986
	if ec.DataSchema != nil {
		if !ec.DataSchema.Validate() {
			errors["dataschema"] = fmt.Errorf("if present, MUST adhere to the format specified in RFC 3986, Section 4.3. Absolute URI")
		}
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

	// time
	// Type: Timestamp
	// Constraints:
	//  OPTIONAL
	//  If present, MUST adhere to the format specified in RFC 3339
	// --> no need to test this, no way to set the time without it being valid.

	if len(errors) > 0 {
		return errors
	}
	return nil
}

// String returns a pretty-printed representation of the EventContext.
func (ec EventContextV1) String() string {
	b := strings.Builder{}

	b.WriteString("Context Attributes,\n")

	b.WriteString("  specversion: " + CloudEventsVersionV1 + "\n")
	b.WriteString("  type: " + ec.Type + "\n")
	b.WriteString("  source: " + ec.Source.String() + "\n")
	if ec.Subject != nil {
		b.WriteString("  subject: " + *ec.Subject + "\n")
	}
	b.WriteString("  id: " + ec.ID + "\n")
	if ec.Time != nil {
		b.WriteString("  time: " + ec.Time.String() + "\n")
	}
	if ec.DataSchema != nil {
		b.WriteString("  dataschema: " + ec.DataSchema.String() + "\n")
	}
	if ec.DataContentType != nil {
		b.WriteString("  datacontenttype: " + *ec.DataContentType + "\n")
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
