/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"time"

	"github.com/cloudevents/sdk-go/v2/event"

	"github.com/google/uuid"
)

// EventDefaulter is the function signature for extensions that are able
// to perform event defaulting.
type EventDefaulter func(ctx context.Context, event event.Event) event.Event

// DefaultIDToUUIDIfNotSet will inspect the provided event and assign a UUID to
// context.ID if it is found to be empty.
func DefaultIDToUUIDIfNotSet(ctx context.Context, event event.Event) event.Event {
	if event.Context != nil {
		if event.ID() == "" {
			event.Context = event.Context.Clone()
			event.SetID(uuid.New().String())
		}
	}
	return event
}

// DefaultTimeToNowIfNotSet will inspect the provided event and assign a new
// Timestamp to context.Time if it is found to be nil or zero.
func DefaultTimeToNowIfNotSet(ctx context.Context, event event.Event) event.Event {
	if event.Context != nil {
		if event.Time().IsZero() {
			event.Context = event.Context.Clone()
			event.SetTime(time.Now())
		}
	}
	return event
}

// NewDefaultDataContentTypeIfNotSet returns a defaulter that will inspect the
// provided event and set the provided content type if content type is found
// to be empty.
func NewDefaultDataContentTypeIfNotSet(contentType string) EventDefaulter {
	return func(ctx context.Context, event event.Event) event.Event {
		if event.Context != nil {
			if event.DataContentType() == "" {
				event.SetDataContentType(contentType)
			}
		}
		return event
	}
}
