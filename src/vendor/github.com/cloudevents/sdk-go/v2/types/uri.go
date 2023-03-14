/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package types

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
)

// URI is a wrapper to url.URL. It is intended to enforce compliance with
// the CloudEvents spec for their definition of URI. Custom
// marshal methods are implemented to ensure the outbound URI object
// is a flat string.
type URI struct {
	url.URL
}

// ParseURI attempts to parse the given string as a URI.
func ParseURI(u string) *URI {
	if u == "" {
		return nil
	}
	pu, err := url.Parse(u)
	if err != nil {
		return nil
	}
	return &URI{URL: *pu}
}

// MarshalJSON implements a custom json marshal method used when this type is
// marshaled using json.Marshal.
func (u URI) MarshalJSON() ([]byte, error) {
	b := fmt.Sprintf("%q", u.String())
	return []byte(b), nil
}

// UnmarshalJSON implements the json unmarshal method used when this type is
// unmarshaled using json.Unmarshal.
func (u *URI) UnmarshalJSON(b []byte) error {
	var ref string
	if err := json.Unmarshal(b, &ref); err != nil {
		return err
	}
	r := ParseURI(ref)
	if r != nil {
		*u = *r
	}
	return nil
}

// MarshalXML implements a custom xml marshal method used when this type is
// marshaled using xml.Marshal.
func (u URI) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(u.String(), start)
}

// UnmarshalXML implements the xml unmarshal method used when this type is
// unmarshaled using xml.Unmarshal.
func (u *URI) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var ref string
	if err := d.DecodeElement(&ref, &start); err != nil {
		return err
	}
	r := ParseURI(ref)
	if r != nil {
		*u = *r
	}
	return nil
}

func (u URI) Validate() bool {
	return u.IsAbs()
}

// String returns the full string representation of the URI-Reference.
func (u *URI) String() string {
	if u == nil {
		return ""
	}
	return u.URL.String()
}
