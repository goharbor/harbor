/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
)

// WriteRequest fills the provided httpRequest with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteRequest(ctx context.Context, m binding.Message, httpRequest *http.Request, transformers ...binding.Transformer) error {
	structuredWriter := (*httpRequestWriter)(httpRequest)
	binaryWriter := (*httpRequestWriter)(httpRequest)

	_, err := binding.Write(
		ctx,
		m,
		structuredWriter,
		binaryWriter,
		transformers...,
	)
	return err
}

type httpRequestWriter http.Request

func (b *httpRequestWriter) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	b.Header.Set(ContentType, format.MediaType())
	return b.setBody(event)
}

func (b *httpRequestWriter) Start(ctx context.Context) error {
	return nil
}

func (b *httpRequestWriter) End(ctx context.Context) error {
	return nil
}

func (b *httpRequestWriter) SetData(data io.Reader) error {
	return b.setBody(data)
}

// setBody is a cherry-pick of the implementation in http.NewRequestWithContext
func (b *httpRequestWriter) setBody(body io.Reader) error {
	rc, ok := body.(io.ReadCloser)
	if !ok && body != nil {
		rc = io.NopCloser(body)
	}
	b.Body = rc
	if body != nil {
		switch v := body.(type) {
		case *bytes.Buffer:
			b.ContentLength = int64(v.Len())
			buf := v.Bytes()
			b.GetBody = func() (io.ReadCloser, error) {
				r := bytes.NewReader(buf)
				return io.NopCloser(r), nil
			}
		case *bytes.Reader:
			b.ContentLength = int64(v.Len())
			snapshot := *v
			b.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return io.NopCloser(&r), nil
			}
		case *strings.Reader:
			b.ContentLength = int64(v.Len())
			snapshot := *v
			b.GetBody = func() (io.ReadCloser, error) {
				r := snapshot
				return io.NopCloser(&r), nil
			}
		default:
			// This is where we'd set it to -1 (at least
			// if body != NoBody) to mean unknown, but
			// that broke people during the Go 1.8 testing
			// period. People depend on it being 0 I
			// guess. Maybe retry later. See Issue 18117.
		}
		// For client requests, Request.ContentLength of 0
		// means either actually 0, or unknown. The only way
		// to explicitly say that the ContentLength is zero is
		// to set the Body to nil. But turns out too much code
		// depends on NewRequest returning a non-nil Body,
		// so we use a well-known ReadCloser variable instead
		// and have the http package also treat that sentinel
		// variable to mean explicitly zero.
		if b.GetBody != nil && b.ContentLength == 0 {
			b.Body = http.NoBody
			b.GetBody = func() (io.ReadCloser, error) { return http.NoBody, nil }
		}
	}
	return nil
}

func (b *httpRequestWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	mapping := attributeHeadersMapping[attribute.Name()]
	if value == nil {
		delete(b.Header, mapping)
		return nil
	}

	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Header[mapping] = append(b.Header[mapping], s)
	return nil
}

func (b *httpRequestWriter) SetExtension(name string, value interface{}) error {
	if value == nil {
		delete(b.Header, extNameToHeaderName(name))
		return nil
	}
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.Header[extNameToHeaderName(name)] = []string{s}
	return nil
}

var (
	_ binding.StructuredWriter = (*httpRequestWriter)(nil) // Test it conforms to the interface
	_ binding.BinaryWriter     = (*httpRequestWriter)(nil) // Test it conforms to the interface
)
