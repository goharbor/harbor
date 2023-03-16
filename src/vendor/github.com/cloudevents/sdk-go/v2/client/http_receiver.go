/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	thttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"go.uber.org/zap"
	"net/http"
)

func NewHTTPReceiveHandler(ctx context.Context, p *thttp.Protocol, fn interface{}) (*EventReceiver, error) {
	invoker, err := newReceiveInvoker(fn, noopObservabilityService{}, nil) //TODO(slinkydeveloper) maybe not nil?
	if err != nil {
		return nil, err
	}

	return &EventReceiver{
		p:       p,
		invoker: invoker,
	}, nil
}

type EventReceiver struct {
	p       *thttp.Protocol
	invoker Invoker
}

func (r *EventReceiver) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Prepare to handle the message if there's one (context cancellation will ensure this closes)
	go func() {
		ctx := req.Context()
		msg, respFn, err := r.p.Respond(ctx)
		if err != nil {
			cecontext.LoggerFrom(context.TODO()).Debugw("failed to call Respond", zap.Error(err))
		} else if err := r.invoker.Invoke(ctx, msg, respFn); err != nil {
			cecontext.LoggerFrom(context.TODO()).Debugw("failed to call Invoke", zap.Error(err))
		}
	}()
	r.p.ServeHTTP(rw, req)
}
