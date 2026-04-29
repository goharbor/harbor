/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"

	"go.uber.org/zap"

	"github.com/cloudevents/sdk-go/v2/binding"
	cecontext "github.com/cloudevents/sdk-go/v2/context"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

// Client interface defines the runtime contract the CloudEvents client supports.
type Client interface {
	// Send will transmit the given event over the client's configured transport.
	Send(ctx context.Context, event event.Event) protocol.Result

	// Request will transmit the given event over the client's configured
	// transport and return any response event.
	Request(ctx context.Context, event event.Event) (*event.Event, protocol.Result)

	// StartReceiver will register the provided function for callback on receipt
	// of a cloudevent. It will also start the underlying protocol as it has
	// been configured.
	// This call is blocking.
	// Valid fn signatures are:
	// * func()
	// * func() error
	// * func(context.Context)
	// * func(context.Context) protocol.Result
	// * func(event.Event)
	// * func(event.Event) protocol.Result
	// * func(context.Context, event.Event)
	// * func(context.Context, event.Event) protocol.Result
	// * func(event.Event) *event.Event
	// * func(event.Event) (*event.Event, protocol.Result)
	// * func(context.Context, event.Event) *event.Event
	// * func(context.Context, event.Event) (*event.Event, protocol.Result)
	StartReceiver(ctx context.Context, fn interface{}) error
}

// New produces a new client with the provided transport object and applied
// client options.
func New(obj interface{}, opts ...Option) (Client, error) {
	c := &ceClient{
		// Running runtime.GOMAXPROCS(0) doesn't update the value, just returns the current one
		pollGoroutines:       runtime.GOMAXPROCS(0),
		observabilityService: noopObservabilityService{},
	}

	if p, ok := obj.(protocol.Sender); ok {
		c.sender = p
	}
	if p, ok := obj.(protocol.Requester); ok {
		c.requester = p
	}
	if p, ok := obj.(protocol.Responder); ok {
		c.responder = p
	}
	if p, ok := obj.(protocol.Receiver); ok {
		c.receiver = p
	}
	if p, ok := obj.(protocol.Opener); ok {
		c.opener = p
	}

	if err := c.applyOptions(opts...); err != nil {
		return nil, err
	}
	return c, nil
}

type ceClient struct {
	sender    protocol.Sender
	requester protocol.Requester
	receiver  protocol.Receiver
	responder protocol.Responder
	// Optional.
	opener protocol.Opener

	observabilityService ObservabilityService

	inboundContextDecorators  []func(context.Context, binding.Message) context.Context
	outboundContextDecorators []func(context.Context) context.Context
	invoker                   Invoker
	receiverMu                sync.Mutex
	eventDefaulterFns         []EventDefaulter
	pollGoroutines            int
	blockingCallback          bool
	ackMalformedEvent         bool
}

func (c *ceClient) applyOptions(opts ...Option) error {
	for _, fn := range opts {
		if err := fn(c); err != nil {
			return err
		}
	}
	return nil
}

func (c *ceClient) Send(ctx context.Context, e event.Event) protocol.Result {
	var err error
	if c.sender == nil {
		err = errors.New("sender not set")
		return err
	}

	for _, f := range c.outboundContextDecorators {
		ctx = f(ctx)
	}

	if len(c.eventDefaulterFns) > 0 {
		for _, fn := range c.eventDefaulterFns {
			e = fn(ctx, e)
		}
	}
	if err = e.Validate(); err != nil {
		return err
	}

	// Event has been defaulted and validated, record we are going to perform send.
	ctx, cb := c.observabilityService.RecordSendingEvent(ctx, e)
	err = c.sender.Send(ctx, (*binding.EventMessage)(&e))
	defer cb(err)
	return err
}

func (c *ceClient) Request(ctx context.Context, e event.Event) (*event.Event, protocol.Result) {
	var resp *event.Event
	var err error

	if c.requester == nil {
		err = errors.New("requester not set")
		return nil, err
	}
	for _, f := range c.outboundContextDecorators {
		ctx = f(ctx)
	}

	if len(c.eventDefaulterFns) > 0 {
		for _, fn := range c.eventDefaulterFns {
			e = fn(ctx, e)
		}
	}

	if err = e.Validate(); err != nil {
		return nil, err
	}

	// Event has been defaulted and validated, record we are going to perform request.
	ctx, cb := c.observabilityService.RecordRequestEvent(ctx, e)

	// If provided a requester, use it to do request/response.
	var msg binding.Message
	msg, err = c.requester.Request(ctx, (*binding.EventMessage)(&e))
	if msg != nil {
		defer func() {
			if err := msg.Finish(err); err != nil {
				cecontext.LoggerFrom(ctx).Warnw("failed calling message.Finish", zap.Error(err))
			}
		}()
	}
	if protocol.IsUndelivered(err) {
		return nil, err
	}

	// try to turn msg into an event, it might not work and that is ok.
	if rs, rserr := binding.ToEvent(ctx, msg); rserr != nil {
		cecontext.LoggerFrom(ctx).Debugw("response: failed calling ToEvent", zap.Error(rserr), zap.Any("resp", msg))
		// If the protocol returns no error, it is an ACK on the request, but we had
		// issues turning the response into an event, so make an ACK Result and pass
		// down the ToEvent error as well.
		err = protocol.NewReceipt(true, "failed to convert response into event: %v\n%w", rserr, err)
	} else {
		resp = rs
	}
	defer cb(err, resp)
	return resp, err
}

// StartReceiver sets up the given fn to handle Receive.
// See Client.StartReceiver for details. This is a blocking call.
func (c *ceClient) StartReceiver(ctx context.Context, fn interface{}) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	c.receiverMu.Lock()
	defer c.receiverMu.Unlock()

	if c.invoker != nil {
		return fmt.Errorf("client already has a receiver")
	}

	invoker, err := newReceiveInvoker(
		fn,
		c.observabilityService,
		c.inboundContextDecorators,
		c.eventDefaulterFns,
		c.ackMalformedEvent,
	)
	if err != nil {
		return err
	}
	if invoker.IsReceiver() && c.receiver == nil {
		return fmt.Errorf("mismatched receiver callback without protocol.Receiver supported by protocol")
	}
	if invoker.IsResponder() && c.responder == nil {
		return fmt.Errorf("mismatched receiver callback without protocol.Responder supported by protocol")
	}
	c.invoker = invoker

	if c.responder == nil && c.receiver == nil {
		return errors.New("responder nor receiver set")
	}

	defer func() {
		c.invoker = nil
	}()

	// Start Polling.
	wg := sync.WaitGroup{}
	for i := 0; i < c.pollGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				var msg binding.Message
				var respFn protocol.ResponseFn
				var err error

				if c.responder != nil {
					msg, respFn, err = c.responder.Respond(ctx)
				} else if c.receiver != nil {
					msg, err = c.receiver.Receive(ctx)
					respFn = noRespFn
				}

				if err == io.EOF { // Normal close
					return
				}

				if err != nil {
					cecontext.LoggerFrom(ctx).Warn("Error while receiving a message: ", err)
					continue
				}

				callback := func() {
					if err := c.invoker.Invoke(ctx, msg, respFn); err != nil {
						cecontext.LoggerFrom(ctx).Warn("Error while handling a message: ", err)
					}
				}

				if c.blockingCallback {
					callback()
				} else {
					// Do not block on the invoker.
					wg.Add(1)
					go func() {
						defer wg.Done()
						callback()
					}()
				}
			}
		}()
	}

	// Start the opener, if set.
	if c.opener != nil {
		if err = c.opener.OpenInbound(ctx); err != nil {
			err = fmt.Errorf("error while opening the inbound connection: %w", err)
			cancel()
		}
	}

	wg.Wait()

	return err
}

// noRespFn is used to simply forward the protocol.Result for receivers that aren't responders
func noRespFn(_ context.Context, _ binding.Message, r protocol.Result, _ ...binding.Transformer) error {
	return r
}
