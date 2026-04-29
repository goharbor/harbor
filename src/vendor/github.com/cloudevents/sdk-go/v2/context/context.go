/*
 Copyright 2021 The CloudEvents Authors
 SPDX-License-Identifier: Apache-2.0
*/

package context

import (
	"context"
	"net/url"
	"time"
)

// Opaque key type used to store target
type targetKeyType struct{}

var targetKey = targetKeyType{}

// WithTarget returns back a new context with the given target. Target is intended to be transport dependent.
// For http transport, `target` should be a full URL and will be injected into the outbound http request.
func WithTarget(ctx context.Context, target string) context.Context {
	return context.WithValue(ctx, targetKey, target)
}

// TargetFrom looks in the given context and returns `target` as a parsed url if found and valid, otherwise nil.
func TargetFrom(ctx context.Context) *url.URL {
	c := ctx.Value(targetKey)
	if c != nil {
		if s, ok := c.(string); ok && s != "" {
			if target, err := url.Parse(s); err == nil {
				return target
			}
		}
	}
	return nil
}

// Opaque key type used to store topic
type topicKeyType struct{}

var topicKey = topicKeyType{}

// WithTopic returns back a new context with the given topic. Topic is intended to be transport dependent.
// For pubsub transport, `topic` should be a Pub/Sub Topic ID.
func WithTopic(ctx context.Context, topic string) context.Context {
	return context.WithValue(ctx, topicKey, topic)
}

// TopicFrom looks in the given context and returns `topic` as a string if found and valid, otherwise "".
func TopicFrom(ctx context.Context) string {
	c := ctx.Value(topicKey)
	if c != nil {
		if s, ok := c.(string); ok {
			return s
		}
	}
	return ""
}

// Opaque key type used to store retry parameters
type retriesKeyType struct{}

var retriesKey = retriesKeyType{}

// WithRetriesConstantBackoff returns back a new context with retries parameters using constant backoff strategy.
// MaxTries is the maximum number for retries and delay is the time interval between retries
func WithRetriesConstantBackoff(ctx context.Context, delay time.Duration, maxTries int) context.Context {
	return WithRetryParams(ctx, &RetryParams{
		Strategy: BackoffStrategyConstant,
		Period:   delay,
		MaxTries: maxTries,
	})
}

// WithRetriesLinearBackoff returns back a new context with retries parameters using linear backoff strategy.
// MaxTries is the maximum number for retries and delay*tries is the time interval between retries
func WithRetriesLinearBackoff(ctx context.Context, delay time.Duration, maxTries int) context.Context {
	return WithRetryParams(ctx, &RetryParams{
		Strategy: BackoffStrategyLinear,
		Period:   delay,
		MaxTries: maxTries,
	})
}

// WithRetriesExponentialBackoff returns back a new context with retries parameters using exponential backoff strategy.
// MaxTries is the maximum number for retries and period is the amount of time to wait, used as `period * 2^retries`.
func WithRetriesExponentialBackoff(ctx context.Context, period time.Duration, maxTries int) context.Context {
	return WithRetryParams(ctx, &RetryParams{
		Strategy: BackoffStrategyExponential,
		Period:   period,
		MaxTries: maxTries,
	})
}

// WithRetryParams returns back a new context with retries parameters.
func WithRetryParams(ctx context.Context, rp *RetryParams) context.Context {
	return context.WithValue(ctx, retriesKey, rp)
}

// RetriesFrom looks in the given context and returns the retries parameters if found.
// Otherwise returns the default retries configuration (ie. no retries).
func RetriesFrom(ctx context.Context) *RetryParams {
	c := ctx.Value(retriesKey)
	if c != nil {
		if s, ok := c.(*RetryParams); ok {
			return s
		}
	}
	return &DefaultRetryParams
}
