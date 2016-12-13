// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package v1 implements the first version of the Clair API.
package v1

import (
	"github.com/julienschmidt/httprouter"

	"github.com/coreos/clair/api/context"
)

// NewRouter creates an HTTP router for version 1 of the Clair API.
func NewRouter(ctx *context.RouteContext) *httprouter.Router {
	router := httprouter.New()

	// Layers
	router.POST("/layers", context.HTTPHandler(postLayer, ctx))
	router.GET("/layers/:layerName", context.HTTPHandler(getLayer, ctx))
	router.DELETE("/layers/:layerName", context.HTTPHandler(deleteLayer, ctx))

	// Namespaces
	router.GET("/namespaces", context.HTTPHandler(getNamespaces, ctx))

	// Vulnerabilities
	router.GET("/namespaces/:namespaceName/vulnerabilities", context.HTTPHandler(getVulnerabilities, ctx))
	router.POST("/namespaces/:namespaceName/vulnerabilities", context.HTTPHandler(postVulnerability, ctx))
	router.GET("/namespaces/:namespaceName/vulnerabilities/:vulnerabilityName", context.HTTPHandler(getVulnerability, ctx))
	router.PUT("/namespaces/:namespaceName/vulnerabilities/:vulnerabilityName", context.HTTPHandler(putVulnerability, ctx))
	router.DELETE("/namespaces/:namespaceName/vulnerabilities/:vulnerabilityName", context.HTTPHandler(deleteVulnerability, ctx))

	// Fixes
	router.GET("/namespaces/:namespaceName/vulnerabilities/:vulnerabilityName/fixes", context.HTTPHandler(getFixes, ctx))
	router.PUT("/namespaces/:namespaceName/vulnerabilities/:vulnerabilityName/fixes/:fixName", context.HTTPHandler(putFix, ctx))
	router.DELETE("/namespaces/:namespaceName/vulnerabilities/:vulnerabilityName/fixes/:fixName", context.HTTPHandler(deleteFix, ctx))

	// Notifications
	router.GET("/notifications/:notificationName", context.HTTPHandler(getNotification, ctx))
	router.DELETE("/notifications/:notificationName", context.HTTPHandler(deleteNotification, ctx))

	// Metrics
	router.GET("/metrics", context.HTTPHandler(getMetrics, ctx))

	return router
}
