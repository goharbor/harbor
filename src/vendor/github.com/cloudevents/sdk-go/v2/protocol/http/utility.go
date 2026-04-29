/*
 Copyright 2022 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"context"
	"encoding/json"
	nethttp "net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/event"
)

// NewEventFromHTTPRequest returns an Event.
func NewEventFromHTTPRequest(req *nethttp.Request) (*event.Event, error) {
	msg := NewMessageFromHttpRequest(req)
	return binding.ToEvent(context.Background(), msg)
}

// NewEventFromHTTPResponse returns an Event.
func NewEventFromHTTPResponse(resp *nethttp.Response) (*event.Event, error) {
	msg := NewMessageFromHttpResponse(resp)
	return binding.ToEvent(context.Background(), msg)
}

// NewEventsFromHTTPRequest returns a batched set of Events from a HTTP Request
func NewEventsFromHTTPRequest(req *nethttp.Request) ([]event.Event, error) {
	msg := NewMessageFromHttpRequest(req)
	return binding.ToEvents(context.Background(), msg, msg.BodyReader)
}

// NewEventsFromHTTPResponse returns a batched set of Events from a HTTP Response
func NewEventsFromHTTPResponse(resp *nethttp.Response) ([]event.Event, error) {
	msg := NewMessageFromHttpResponse(resp)
	return binding.ToEvents(context.Background(), msg, msg.BodyReader)
}

// NewHTTPRequestFromEvent creates a http.Request object that can be used with any http.Client for a singular event.
// This is an HTTP POST action to the provided url.
func NewHTTPRequestFromEvent(ctx context.Context, url string, event event.Event) (*nethttp.Request, error) {
	if err := event.Validate(); err != nil {
		return nil, err
	}

	req, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}
	if err := WriteRequest(ctx, (*binding.EventMessage)(&event), req); err != nil {
		return nil, err
	}

	return req, nil
}

// NewHTTPRequestFromEvents creates a http.Request object that can be used with any http.Client for sending
// a batched set of events. This is an HTTP POST action to the provided url.
func NewHTTPRequestFromEvents(ctx context.Context, url string, events []event.Event) (*nethttp.Request, error) {
	// Sending batch events is quite straightforward, as there is only JSON format, so a simple implementation.
	for _, e := range events {
		if err := e.Validate(); err != nil {
			return nil, err
		}
	}
	var buffer bytes.Buffer
	err := json.NewEncoder(&buffer).Encode(events)
	if err != nil {
		return nil, err
	}

	request, err := nethttp.NewRequestWithContext(ctx, nethttp.MethodPost, url, &buffer)
	if err != nil {
		return nil, err
	}

	request.Header.Set(ContentType, event.ApplicationCloudEventsBatchJSON)

	return request, nil
}

// IsHTTPBatch returns if the current http.Request or http.Response is a batch event operation, by checking the
// header `Content-Type` value.
func IsHTTPBatch(header nethttp.Header) bool {
	return header.Get(ContentType) == event.ApplicationCloudEventsBatchJSON
}
