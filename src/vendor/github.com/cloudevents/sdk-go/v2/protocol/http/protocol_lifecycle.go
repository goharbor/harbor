/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/cloudevents/sdk-go/v2/protocol"
)

var _ protocol.Opener = (*Protocol)(nil)

func (p *Protocol) OpenInbound(ctx context.Context) error {
	p.reMu.Lock()
	defer p.reMu.Unlock()

	if p.Handler == nil {
		p.Handler = http.NewServeMux()
	}

	if !p.handlerRegistered {
		// handler.Handle might panic if the user tries to use the same path as the sdk.
		p.Handler.Handle(p.GetPath(), p)
		p.handlerRegistered = true
	}

	// After listener is invok
	listener, err := p.listen()
	if err != nil {
		return err
	}

	p.server = &http.Server{
		Addr:         listener.Addr().String(),
		Handler:      attachMiddleware(p.Handler, p.middleware),
		ReadTimeout:  DefaultTimeout,
		WriteTimeout: DefaultTimeout,
	}

	// Shutdown
	defer func() {
		_ = p.server.Close()
		p.server = nil
	}()

	errChan := make(chan error)
	go func() {
		errChan <- p.server.Serve(listener)
	}()

	// wait for the server to return or ctx.Done().
	select {
	case <-ctx.Done():
		// Try a graceful shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), p.ShutdownTimeout)
		defer cancel()

		shdwnErr := p.server.Shutdown(ctx)
		if shdwnErr != nil {
			shdwnErr = fmt.Errorf("shutting down HTTP server: %w", shdwnErr)
		}

		// Wait for server goroutine to exit
		rntmErr := <-errChan
		if rntmErr != nil && rntmErr != http.ErrServerClosed {
			rntmErr = fmt.Errorf("server failed during shutdown: %w", rntmErr)

			if shdwnErr != nil {
				return fmt.Errorf("combined error during shutdown of HTTP server: %w, %v",
					shdwnErr, rntmErr)
			}

			return rntmErr
		}

		return shdwnErr

	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("during runtime of HTTP server: %w", err)
		}
		return nil
	}
}

// GetListeningPort returns the listening port.
// Returns -1 if it's not listening.
func (p *Protocol) GetListeningPort() int {
	if listener := p.listener.Load(); listener != nil {
		if tcpAddr, ok := listener.(net.Listener).Addr().(*net.TCPAddr); ok {
			return tcpAddr.Port
		}
	}
	return -1
}

// listen if not already listening, update t.Port
func (p *Protocol) listen() (net.Listener, error) {
	if p.listener.Load() == nil {
		port := 8080
		if p.Port != -1 {
			port = p.Port
			if port < 0 || port > 65535 {
				return nil, fmt.Errorf("invalid port %d", port)
			}
		}
		var err error
		var listener net.Listener
		if listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port)); err != nil {
			return nil, err
		}
		p.listener.Store(listener)
		return listener, nil
	}
	return p.listener.Load().(net.Listener), nil
}

// GetPath returns the path the transport is hosted on. If the path is '/',
// the transport will handle requests on any URI. To discover the true path
// a request was received on, inspect the context from Receive(cxt, ...) with
// TransportContextFrom(ctx).
func (p *Protocol) GetPath() string {
	path := strings.TrimSpace(p.Path)
	if len(path) > 0 {
		return path
	}
	return "/" // default
}

// attachMiddleware attaches the HTTP middleware to the specified handler.
func attachMiddleware(h http.Handler, middleware []Middleware) http.Handler {
	for _, m := range middleware {
		h = m(h)
	}
	return h
}
