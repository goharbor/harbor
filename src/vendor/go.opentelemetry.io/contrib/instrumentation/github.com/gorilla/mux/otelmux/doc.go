// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

// Package otelmux instruments the github.com/gorilla/mux package.
//
// Currently only the routing of a received message can be instrumented. To do
// it, use the Middleware function.
package otelmux // import "go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
