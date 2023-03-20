/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package http

import (
	"context"
	"net/http"
)

type RateLimiter interface {
	// Allow attempts to take one token from the rate limiter for the specified
	// request. It returns ok when this operation was successful. In case ok is
	// false, reset will indicate the time in seconds when it is safe to perform
	// another attempt. An error is returned when this operation failed, e.g. due to
	// a backend error.
	Allow(ctx context.Context, r *http.Request) (ok bool, reset uint64, err error)
	// Close terminates rate limiter and cleans up any data structures or
	// connections that may remain open. After a store is stopped, Take() should
	// always return zero values.
	Close(ctx context.Context) error
}

type noOpLimiter struct{}

func (n noOpLimiter) Allow(ctx context.Context, r *http.Request) (bool, uint64, error) {
	return true, 0, nil
}

func (n noOpLimiter) Close(ctx context.Context) error {
	return nil
}
