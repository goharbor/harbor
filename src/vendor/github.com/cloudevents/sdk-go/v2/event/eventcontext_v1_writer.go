/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/cloudevents/sdk-go/v2/types"
)

// Adhere to EventContextWriter
var _ EventContextWriter = (*EventContextV1)(nil)

// SetDataContentType implements EventContextWriter.SetDataContentType
func (ec *EventContextV1) SetDataContentType(ct string) error {
	ct = strings.TrimSpace(ct)
	if ct == "" {
		ec.DataContentType = nil
	} else {
		ec.DataContentType = &ct
	}
	return nil
}

// SetType implements EventContextWriter.SetType
func (ec *EventContextV1) SetType(t string) error {
	t = strings.TrimSpace(t)
	ec.Type = t
	return nil
}

// SetSource implements EventContextWriter.SetSource
func (ec *EventContextV1) SetSource(u string) error {
	pu, err := url.Parse(u)
	if err != nil {
		return err
	}
	ec.Source = types.URIRef{URL: *pu}
	return nil
}

// SetSubject implements EventContextWriter.SetSubject
func (ec *EventContextV1) SetSubject(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		ec.Subject = nil
	} else {
		ec.Subject = &s
	}
	return nil
}

// SetID implements EventContextWriter.SetID
func (ec *EventContextV1) SetID(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("id is required to be a non-empty string")
	}
	ec.ID = id
	return nil
}

// SetTime implements EventContextWriter.SetTime
func (ec *EventContextV1) SetTime(t time.Time) error {
	if t.IsZero() {
		ec.Time = nil
	} else {
		ec.Time = &types.Timestamp{Time: t}
	}
	return nil
}

// SetDataSchema implements EventContextWriter.SetDataSchema
func (ec *EventContextV1) SetDataSchema(u string) error {
	u = strings.TrimSpace(u)
	if u == "" {
		ec.DataSchema = nil
		return nil
	}
	pu, err := url.Parse(u)
	if err != nil {
		return err
	}
	ec.DataSchema = &types.URI{URL: *pu}
	return nil
}

// DeprecatedSetDataContentEncoding implements EventContextWriter.DeprecatedSetDataContentEncoding
func (ec *EventContextV1) DeprecatedSetDataContentEncoding(e string) error {
	return errors.New("deprecated: SetDataContentEncoding is not supported in v1.0 of CloudEvents")
}
