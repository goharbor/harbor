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
	"strconv"
	"strings"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
	"github.com/cloudevents/sdk-go/v2/types"
)

// WriteResponseWriter writes out to the the provided httpResponseWriter with the message m.
// Using context you can tweak the encoding processing (more details on binding.Write documentation).
func WriteResponseWriter(ctx context.Context, m binding.Message, status int, rw http.ResponseWriter, transformers ...binding.Transformer) error {
	if status < 200 || status >= 600 {
		status = http.StatusOK
	}
	writer := &httpResponseWriter{rw: rw, status: status}

	_, err := binding.Write(
		ctx,
		m,
		writer,
		writer,
		transformers...,
	)
	return err
}

type httpResponseWriter struct {
	rw     http.ResponseWriter
	status int
	body   io.Reader
}

func (b *httpResponseWriter) SetStructuredEvent(ctx context.Context, format format.Format, event io.Reader) error {
	b.rw.Header().Set(ContentType, format.MediaType())
	b.body = event
	return b.finalizeWriter()
}

func (b *httpResponseWriter) Start(ctx context.Context) error {
	return nil
}

func (b *httpResponseWriter) SetAttribute(attribute spec.Attribute, value interface{}) error {
	mapping := attributeHeadersMapping[attribute.Name()]
	if value == nil {
		delete(b.rw.Header(), mapping)
	}

	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.rw.Header()[mapping] = append(b.rw.Header()[mapping], s)
	return nil
}

func (b *httpResponseWriter) SetExtension(name string, value interface{}) error {
	if value == nil {
		delete(b.rw.Header(), extNameToHeaderName(name))
	}
	// Http headers, everything is a string!
	s, err := types.Format(value)
	if err != nil {
		return err
	}
	b.rw.Header()[extNameToHeaderName(name)] = []string{s}
	return nil
}

func (b *httpResponseWriter) SetData(reader io.Reader) error {
	b.body = reader
	return nil
}

func (b *httpResponseWriter) finalizeWriter() error {
	if b.body != nil {
		// Try to figure it out if we have a content-length
		contentLength := -1
		switch v := b.body.(type) {
		case *bytes.Buffer:
			contentLength = v.Len()
		case *bytes.Reader:
			contentLength = v.Len()
		case *strings.Reader:
			contentLength = v.Len()
		}

		if contentLength != -1 {
			b.rw.Header().Add("Content-length", strconv.Itoa(contentLength))
		}

		// Finalize the headers.
		b.rw.WriteHeader(b.status)

		// Write body.
		_, err := io.Copy(b.rw, b.body)
		if err != nil {
			return err
		}
	} else {
		// Finalize the headers.
		b.rw.WriteHeader(b.status)
	}
	return nil
}

func (b *httpResponseWriter) End(ctx context.Context) error {
	return b.finalizeWriter()
}

var _ binding.StructuredWriter = (*httpResponseWriter)(nil) // Test it conforms to the interface
var _ binding.BinaryWriter = (*httpResponseWriter)(nil)     // Test it conforms to the interface
