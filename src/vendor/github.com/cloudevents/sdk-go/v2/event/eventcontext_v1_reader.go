/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package event

import (
	"fmt"
	"strings"
	"time"
)

// GetSpecVersion implements EventContextReader.GetSpecVersion
func (ec EventContextV1) GetSpecVersion() string {
	return CloudEventsVersionV1
}

// GetDataContentType implements EventContextReader.GetDataContentType
func (ec EventContextV1) GetDataContentType() string {
	if ec.DataContentType != nil {
		return *ec.DataContentType
	}
	return ""
}

// GetDataMediaType implements EventContextReader.GetDataMediaType
func (ec EventContextV1) GetDataMediaType() (string, error) {
	if ec.DataContentType != nil {
		dct := *ec.DataContentType
		i := strings.IndexRune(dct, ';')
		if i == -1 {
			return dct, nil
		}
		return strings.TrimSpace(dct[0:i]), nil
	}
	return "", nil
}

// GetType implements EventContextReader.GetType
func (ec EventContextV1) GetType() string {
	return ec.Type
}

// GetSource implements EventContextReader.GetSource
func (ec EventContextV1) GetSource() string {
	return ec.Source.String()
}

// GetSubject implements EventContextReader.GetSubject
func (ec EventContextV1) GetSubject() string {
	if ec.Subject != nil {
		return *ec.Subject
	}
	return ""
}

// GetTime implements EventContextReader.GetTime
func (ec EventContextV1) GetTime() time.Time {
	if ec.Time != nil {
		return ec.Time.Time
	}
	return time.Time{}
}

// GetID implements EventContextReader.GetID
func (ec EventContextV1) GetID() string {
	return ec.ID
}

// GetDataSchema implements EventContextReader.GetDataSchema
func (ec EventContextV1) GetDataSchema() string {
	if ec.DataSchema != nil {
		return ec.DataSchema.String()
	}
	return ""
}

// DeprecatedGetDataContentEncoding implements EventContextReader.DeprecatedGetDataContentEncoding
func (ec EventContextV1) DeprecatedGetDataContentEncoding() string {
	return ""
}

// GetExtensions implements EventContextReader.GetExtensions
func (ec EventContextV1) GetExtensions() map[string]interface{} {
	if len(ec.Extensions) == 0 {
		return nil
	}
	// For now, convert the extensions of v1.0 to the pre-v1.0 style.
	ext := make(map[string]interface{}, len(ec.Extensions))
	for k, v := range ec.Extensions {
		ext[k] = v
	}
	return ext
}

// GetExtension implements EventContextReader.GetExtension
func (ec EventContextV1) GetExtension(key string) (interface{}, error) {
	v, ok := caseInsensitiveSearch(key, ec.Extensions)
	if !ok {
		return "", fmt.Errorf("%q not found", key)
	}
	return v, nil
}
