/*
 Copyright 2022 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
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
