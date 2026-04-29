/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"io"
	nethttp "net/http"
	"net/textproto"
	"strings"
	"unicode"

	"github.com/cloudevents/sdk-go/v2/binding"
	"github.com/cloudevents/sdk-go/v2/binding/format"
	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

const prefix = "Ce-"

var specs = spec.WithPrefixMatchExact(
	func(s string) string {
		if s == "datacontenttype" {
			return "Content-Type"
		} else {
			return textproto.CanonicalMIMEHeaderKey("Ce-" + s)
		}
	},
	"Ce-",
)

const ContentType = "Content-Type"
const ContentLength = "Content-Length"

// Message holds the Header and Body of a HTTP Request or Response.
// The Message instance *must* be constructed from NewMessage function.
// This message *cannot* be read several times. In order to read it more times, buffer it using binding/buffering methods
type Message struct {
	Header     nethttp.Header
	BodyReader io.ReadCloser
	OnFinish   func(error) error

	ctx context.Context

	format  format.Format
	version spec.Version
}

// Check if http.Message implements binding.Message
var _ binding.Message = (*Message)(nil)
var _ binding.MessageContext = (*Message)(nil)
var _ binding.MessageMetadataReader = (*Message)(nil)

// NewMessage returns a binding.Message with header and data.
// The returned binding.Message *cannot* be read several times. In order to read it more times, buffer it using binding/buffering methods
func NewMessage(header nethttp.Header, body io.ReadCloser) *Message {
	m := Message{Header: header}
	if body != nil {
		m.BodyReader = body
	}
	if m.format = format.Lookup(header.Get(ContentType)); m.format == nil {
		m.version = specs.Version(m.Header.Get(specs.PrefixedSpecVersionName()))
	}
	return &m
}

// NewMessageFromHttpRequest returns a binding.Message with header and data.
// The returned binding.Message *cannot* be read several times. In order to read it more times, buffer it using binding/buffering methods
func NewMessageFromHttpRequest(req *nethttp.Request) *Message {
	if req == nil {
		return nil
	}
	message := NewMessage(req.Header, req.Body)
	message.ctx = req.Context()
	return message
}

// NewMessageFromHttpResponse returns a binding.Message with header and data.
// The returned binding.Message *cannot* be read several times. In order to read it more times, buffer it using binding/buffering methods
func NewMessageFromHttpResponse(resp *nethttp.Response) *Message {
	if resp == nil {
		return nil
	}
	msg := NewMessage(resp.Header, resp.Body)
	return msg
}

func (m *Message) ReadEncoding() binding.Encoding {
	if m.version != nil {
		return binding.EncodingBinary
	}
	if m.format != nil {
		if m.format == format.JSONBatch {
			return binding.EncodingBatch
		}
		return binding.EncodingStructured
	}
	return binding.EncodingUnknown
}

func (m *Message) ReadStructured(ctx context.Context, encoder binding.StructuredWriter) error {
	if m.format == nil {
		return binding.ErrNotStructured
	} else {
		return encoder.SetStructuredEvent(ctx, m.format, m.BodyReader)
	}
}

func (m *Message) ReadBinary(ctx context.Context, encoder binding.BinaryWriter) (err error) {
	if m.version == nil {
		return binding.ErrNotBinary
	}

	for k, v := range m.Header {
		attr := m.version.Attribute(k)
		if attr != nil {
			err = encoder.SetAttribute(attr, v[0])
		} else if strings.HasPrefix(k, prefix) {
			// Trim Prefix + To lower
			var b strings.Builder
			b.Grow(len(k) - len(prefix))
			b.WriteRune(unicode.ToLower(rune(k[len(prefix)])))
			b.WriteString(k[len(prefix)+1:])
			err = encoder.SetExtension(b.String(), v[0])
		}
		if err != nil {
			return err
		}
	}

	if m.BodyReader != nil {
		err = encoder.SetData(m.BodyReader)
		if err != nil {
			return err
		}
	}

	return
}

func (m *Message) GetAttribute(k spec.Kind) (spec.Attribute, interface{}) {
	attr := m.version.AttributeFromKind(k)
	if attr != nil {
		h := m.Header[attributeHeadersMapping[attr.Name()]]
		if h != nil {
			return attr, h[0]
		}
		return attr, nil
	}
	return nil, nil
}

func (m *Message) GetExtension(name string) interface{} {
	h := m.Header[extNameToHeaderName(name)]
	if h != nil {
		return h[0]
	}
	return nil
}

func (m *Message) Context() context.Context {
	return m.ctx
}

func (m *Message) Finish(err error) error {
	if m.BodyReader != nil {
		_ = m.BodyReader.Close()
	}
	if m.OnFinish != nil {
		return m.OnFinish(err)
	}
	return nil
}
