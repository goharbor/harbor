/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"

	nethttp "net/http"
	"net/url"
)

type requestKey struct{}

// RequestData holds the http.Request information subset that can be
// used to retrieve HTTP information for an incoming CloudEvent.
type RequestData struct {
	URL        *url.URL
	Header     nethttp.Header
	RemoteAddr string
	Host       string
}

// WithRequestDataAtContext uses the http.Request to add RequestData
// information to the Context.
func WithRequestDataAtContext(ctx context.Context, r *nethttp.Request) context.Context {
	if r == nil {
		return ctx
	}

	return context.WithValue(ctx, requestKey{}, &RequestData{
		URL:        r.URL,
		Header:     r.Header,
		RemoteAddr: r.RemoteAddr,
		Host:       r.Host,
	})
}

// RequestDataFromContext retrieves RequestData from the Context.
// If not set nil is returned.
func RequestDataFromContext(ctx context.Context) *RequestData {
	if req := ctx.Value(requestKey{}); req != nil {
		return req.(*RequestData)
	}
	return nil
}
