/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

func (p *Protocol) do(ctx context.Context, req *http.Request) (binding.Message, error) {
	params := cecontext.RetriesFrom(ctx)

	switch params.Strategy {
	case cecontext.BackoffStrategyConstant, cecontext.BackoffStrategyLinear, cecontext.BackoffStrategyExponential:
		return p.doWithRetry(ctx, params, req)
	case cecontext.BackoffStrategyNone:
		fallthrough
	default:
		return p.doOnce(req)
	}
}

func (p *Protocol) doOnce(req *http.Request) (binding.Message, protocol.Result) {
	resp, err := p.Client.Do(req)
	if err != nil {
		return nil, protocol.NewReceipt(false, "%w", err)
	}

	var result protocol.Result
	if resp.StatusCode/100 == 2 {
		result = protocol.ResultACK
	} else {
		result = protocol.ResultNACK
	}

	return NewMessage(resp.Header, resp.Body), NewResult(resp.StatusCode, "%w", result)
}

func (p *Protocol) doWithRetry(ctx context.Context, params *cecontext.RetryParams, req *http.Request) (binding.Message, error) {
	start := time.Now()
	retry := 0
	results := make([]protocol.Result, 0)

	var (
		body []byte
		err  error
	)

	if req != nil && req.Body != nil {
		defer func() {
			if err = req.Body.Close(); err != nil {
				cecontext.LoggerFrom(ctx).Warnw("could not close request body", zap.Error(err))
			}
		}()
		body, err = io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		resetBody(req, body)
	}

	for {
		msg, result := p.doOnce(req)

		// Fast track common case.
		if protocol.IsACK(result) {
			return msg, NewRetriesResult(result, retry, start, results)
		}

		var httpResult *Result
		if errors.As(result, &httpResult) {
			sc := httpResult.StatusCode
			if !p.isRetriableFunc(sc) {
				cecontext.LoggerFrom(ctx).Debugw("status code not retryable, will not try again",
					zap.Error(httpResult),
					zap.Int("statusCode", sc))
				return msg, NewRetriesResult(result, retry, start, results)
			}
		}

		// total tries = retry + 1
		if err = params.Backoff(ctx, retry+1); err != nil {
			// do not try again.
			cecontext.LoggerFrom(ctx).Debugw("backoff error, will not try again", zap.Error(err))
			return msg, NewRetriesResult(result, retry, start, results)
		}

		retry++
		resetBody(req, body)
		results = append(results, result)
		if msg != nil {
			// avoid leak, forget message, ignore error
			_ = msg.Finish(nil)
		}
	}
}

// reset body to allow it to be read multiple times, e.g. when retrying http
// requests
func resetBody(req *http.Request, body []byte) {
	if req == nil || req.Body == nil {
		return
	}

	req.Body = io.NopCloser(bytes.NewReader(body))

	// do not modify existing GetBody function
	if req.GetBody == nil {
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
	}
}
