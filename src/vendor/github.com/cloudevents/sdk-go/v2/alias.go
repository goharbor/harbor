/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

// Package v2 reexports a subset of the SDK v2 API.
package v2

// Package cloudevents alias' common functions and types to improve discoverability and reduce
// the number of imports for simple HTTP clients.

import (
	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/cloudevents/sdk-go/v2/types"
)

// Client

type ClientOption = client.Option
type Client = client.Client

// Event

type Event = event.Event
type Result = protocol.Result

// Context

type EventContext = event.EventContext
type EventContextV1 = event.EventContextV1
type EventContextV03 = event.EventContextV03

// Custom Types

type Timestamp = types.Timestamp
type URIRef = types.URIRef

// HTTP Protocol

type HTTPOption = http.Option

type HTTPProtocol = http.Protocol

// Encoding

type Encoding = binding.Encoding

// Message

type Message = binding.Message

const (
	// ReadEncoding

	ApplicationXML                  = event.ApplicationXML
	ApplicationJSON                 = event.ApplicationJSON
	TextPlain                       = event.TextPlain
	ApplicationCloudEventsJSON      = event.ApplicationCloudEventsJSON
	ApplicationCloudEventsBatchJSON = event.ApplicationCloudEventsBatchJSON
	Base64                          = event.Base64

	// Event Versions

	VersionV1  = event.CloudEventsVersionV1
	VersionV03 = event.CloudEventsVersionV03

	// Encoding

	EncodingBinary     = binding.EncodingBinary
	EncodingStructured = binding.EncodingStructured
)

var (

	// ContentType Helpers

	StringOfApplicationJSON                 = event.StringOfApplicationJSON
	StringOfApplicationXML                  = event.StringOfApplicationXML
	StringOfTextPlain                       = event.StringOfTextPlain
	StringOfApplicationCloudEventsJSON      = event.StringOfApplicationCloudEventsJSON
	StringOfApplicationCloudEventsBatchJSON = event.StringOfApplicationCloudEventsBatchJSON
	StringOfBase64                          = event.StringOfBase64

	// Client Creation

	NewClient     = client.New
	NewClientHTTP = client.NewHTTP
	// Deprecated: please use New with the observability options.
	NewClientObserved = client.NewObserved
	// Deprecated: Please use NewClientHTTP with the observability options.
	NewDefaultClient      = client.NewDefault
	NewHTTPReceiveHandler = client.NewHTTPReceiveHandler

	// Client Options

	WithEventDefaulter = client.WithEventDefaulter
	WithUUIDs          = client.WithUUIDs
	WithTimeNow        = client.WithTimeNow
	// Deprecated: this is now noop and will be removed in future releases.
	WithTracePropagation = client.WithTracePropagation()

	// Event Creation

	NewEvent = event.New

	// Results

	NewResult = protocol.NewResult
	ResultIs  = protocol.ResultIs
	ResultAs  = protocol.ResultAs

	// Receipt helpers

	NewReceipt = protocol.NewReceipt

	ResultACK  = protocol.ResultACK
	ResultNACK = protocol.ResultNACK

	IsACK         = protocol.IsACK
	IsNACK        = protocol.IsNACK
	IsUndelivered = protocol.IsUndelivered

	// HTTP Results

	NewHTTPResult        = http.NewResult
	NewHTTPRetriesResult = http.NewRetriesResult

	// Message Creation

	ToMessage = binding.ToMessage

	// Event Creation

	NewEventFromHTTPRequest   = http.NewEventFromHTTPRequest
	NewEventFromHTTPResponse  = http.NewEventFromHTTPResponse
	NewEventsFromHTTPRequest  = http.NewEventsFromHTTPRequest
	NewEventsFromHTTPResponse = http.NewEventsFromHTTPResponse
	NewHTTPRequestFromEvent   = http.NewHTTPRequestFromEvent
	NewHTTPRequestFromEvents  = http.NewHTTPRequestFromEvents
	IsHTTPBatch               = http.IsHTTPBatch

	// HTTP Messages

	WriteHTTPRequest = http.WriteRequest

	// Context

	ContextWithTarget                    = context.WithTarget
	TargetFromContext                    = context.TargetFrom
	ContextWithRetriesConstantBackoff    = context.WithRetriesConstantBackoff
	ContextWithRetriesLinearBackoff      = context.WithRetriesLinearBackoff
	ContextWithRetriesExponentialBackoff = context.WithRetriesExponentialBackoff

	WithEncodingBinary     = binding.WithForceBinary
	WithEncodingStructured = binding.WithForceStructured

	// Custom Types

	ParseTimestamp = types.ParseTimestamp
	ParseURIRef    = types.ParseURIRef
	ParseURI       = types.ParseURI

	// HTTP Protocol

	NewHTTP = http.New

	// HTTP Protocol Options

	WithTarget          = http.WithTarget
	WithHeader          = http.WithHeader
	WithShutdownTimeout = http.WithShutdownTimeout
	//WithEncoding           = http.WithEncoding
	//WithStructuredEncoding = http.WithStructuredEncoding // TODO: expose new way
	WithPort                      = http.WithPort
	WithPath                      = http.WithPath
	WithMiddleware                = http.WithMiddleware
	WithListener                  = http.WithListener
	WithRoundTripper              = http.WithRoundTripper
	WithGetHandlerFunc            = http.WithGetHandlerFunc
	WithOptionsHandlerFunc        = http.WithOptionsHandlerFunc
	WithDefaultOptionsHandlerFunc = http.WithDefaultOptionsHandlerFunc
)
