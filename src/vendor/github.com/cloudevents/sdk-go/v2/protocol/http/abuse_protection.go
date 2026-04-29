/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type WebhookConfig struct {
	AllowedMethods  []string // defaults to POST
	AllowedRate     *int
	AutoACKCallback bool
	AllowedOrigins  []string
}

const (
	DefaultAllowedRate = 1000
	DefaultTimeout     = time.Second * 600
)

// TODO: implement rate limiting.
// Throttling is indicated by requests being rejected using HTTP status code 429 Too Many Requests.
// TODO: use this if Webhook Request Origin has been turned on.
// Inbound requests should be rejected if Allowed Origins is required by SDK.

func (p *Protocol) OptionsHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodOptions || p.WebhookConfig == nil {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	headers := make(http.Header)

	// The spec does not say we need to validate the origin, just the request origin.
	// After the handshake, we will validate the origin.
	if origin, ok := p.ValidateRequestOrigin(req); !ok {
		rw.WriteHeader(http.StatusBadRequest)
		return
	} else {
		headers.Set("WebHook-Allowed-Origin", origin)
	}

	allowedRateRequired := false
	if _, ok := req.Header[http.CanonicalHeaderKey("WebHook-Request-Rate")]; ok {
		// must send WebHook-Allowed-Rate
		allowedRateRequired = true
	}

	if p.WebhookConfig.AllowedRate != nil {
		headers.Set("WebHook-Allowed-Rate", strconv.Itoa(*p.WebhookConfig.AllowedRate))
	} else if allowedRateRequired {
		headers.Set("WebHook-Allowed-Rate", strconv.Itoa(DefaultAllowedRate))
	}

	if len(p.WebhookConfig.AllowedMethods) > 0 {
		headers.Set("Allow", strings.Join(p.WebhookConfig.AllowedMethods, ", "))
	} else {
		headers.Set("Allow", http.MethodPost)
	}

	cb := req.Header.Get("WebHook-Request-Callback")
	if cb != "" {
		if p.WebhookConfig.AutoACKCallback {
			go func() {
				reqAck, err := http.NewRequest(http.MethodPost, cb, nil)
				if err != nil {
					cecontext.LoggerFrom(req.Context()).Errorw("OPTIONS handler failed to create http request attempting to ack callback.", zap.Error(err), zap.String("callback", cb))
					return
				}

				// Write out the headers.
				for k := range headers {
					reqAck.Header.Set(k, headers.Get(k))
				}

				_, err = http.DefaultClient.Do(reqAck)
				if err != nil {
					cecontext.LoggerFrom(req.Context()).Errorw("OPTIONS handler failed to ack callback.", zap.Error(err), zap.String("callback", cb))
					return
				}
			}()
			return
		} else {
			cecontext.LoggerFrom(req.Context()).Infof("ACTION REQUIRED: Please validate web hook request callback: %q", cb)
			// TODO: what to do pending https://github.com/cloudevents/spec/issues/617
			return
		}
	}

	// Write out the headers.
	for k := range headers {
		rw.Header().Set(k, headers.Get(k))
	}
}

func (p *Protocol) ValidateRequestOrigin(req *http.Request) (string, bool) {
	return p.validateOrigin(req.Header.Get("WebHook-Request-Origin"))
}

func (p *Protocol) ValidateOrigin(req *http.Request) (string, bool) {
	return p.validateOrigin(req.Header.Get("Origin"))
}

func (p *Protocol) validateOrigin(ro string) (string, bool) {
	cecontext.LoggerFrom(context.TODO()).Infow("Validating origin.", zap.String("origin", ro))

	for _, ao := range p.WebhookConfig.AllowedOrigins {
		if ao == "*" {
			return ao, true
		}
		// TODO: it is not clear what the rules for allowed hosts are.
		// Need to find docs for this. For now, test for prefix.
		if strings.HasPrefix(ro, ao) {
			return ao, true
		}
	}

	return ro, false
}
