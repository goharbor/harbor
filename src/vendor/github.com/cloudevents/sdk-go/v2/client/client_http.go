/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"github.com/cloudevents/sdk-go/v2/protocol/http"
)

// NewHTTP provides the good defaults for the common case using an HTTP
// Protocol client.
// The WithTimeNow, and WithUUIDs client options are also applied to the
// client, all outbound events will have a time and id set if not already
// present.
func NewHTTP(opts ...http.Option) (Client, error) {
	p, err := http.New(opts...)
	if err != nil {
		return nil, err
	}

	c, err := New(p, WithTimeNow(), WithUUIDs())
	if err != nil {
		return nil, err
	}

	return c, nil
}

// NewDefault has been replaced by NewHTTP
// Deprecated. To get the same as NewDefault provided, please use NewHTTP with
// the observability service passed as an option, or client.NewClientHTTP from
// package github.com/cloudevents/sdk-go/observability/opencensus/v2/client
var NewDefault = NewHTTP
