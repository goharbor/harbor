/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"github.com/cloudevents/sdk-go/v2/binding"
	"net/http"
	"net/textproto"
	"strings"
	"unicode"

	"github.com/cloudevents/sdk-go/v2/binding/spec"
)

var attributeHeadersMapping map[string]string

type customHeaderKey int

const (
	headerKey customHeaderKey = iota
)

func init() {
	attributeHeadersMapping = make(map[string]string)
	for _, v := range specs.Versions() {
		for _, a := range v.Attributes() {
			if a.Kind() == spec.DataContentType {
				attributeHeadersMapping[a.Name()] = ContentType
			} else {
				attributeHeadersMapping[a.Name()] = textproto.CanonicalMIMEHeaderKey(prefix + a.Name())
			}
		}
	}
}

func extNameToHeaderName(name string) string {
	var b strings.Builder
	b.Grow(len(name) + len(prefix))
	b.WriteString(prefix)
	b.WriteRune(unicode.ToUpper(rune(name[0])))
	b.WriteString(name[1:])
	return b.String()
}

func HeaderFrom(ctx context.Context) http.Header {
	return binding.GetOrDefaultFromCtx(ctx, headerKey, make(http.Header)).(http.Header)
}

func WithCustomHeader(ctx context.Context, header http.Header) context.Context {
	return context.WithValue(ctx, headerKey, header)
}
