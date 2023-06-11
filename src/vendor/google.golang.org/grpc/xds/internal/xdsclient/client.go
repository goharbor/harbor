/*
 *
 * Copyright 2019 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package xdsclient implements a full fledged gRPC client for the xDS API used
// by the xds resolver and balancer implementations.
package xdsclient

import (
	"google.golang.org/grpc/xds/internal/xdsclient/bootstrap"
	"google.golang.org/grpc/xds/internal/xdsclient/load"
	"google.golang.org/grpc/xds/internal/xdsclient/xdsresource"
)

// XDSClient is a full fledged gRPC client which queries a set of discovery APIs
// (collectively termed as xDS) on a remote management server, to discover
// various dynamic resources.
type XDSClient interface {
	WatchListener(string, func(xdsresource.ListenerUpdate, error)) func()
	WatchRouteConfig(string, func(xdsresource.RouteConfigUpdate, error)) func()
	WatchCluster(string, func(xdsresource.ClusterUpdate, error)) func()
	WatchEndpoints(string, func(xdsresource.EndpointsUpdate, error)) func()

	// WatchResource uses xDS to discover the resource associated with the
	// provided resource name. The resource type implementation determines how
	// xDS requests are sent out and how responses are deserialized and
	// validated. Upon receipt of a response from the management server, an
	// appropriate callback on the watcher is invoked.
	//
	// Most callers will not have a need to use this API directly. They will
	// instead use a resource-type-specific wrapper API provided by the relevant
	// resource type implementation.
	//
	// TODO: Once this generic client API is fully implemented and integrated,
	// delete the resource type specific watch APIs on this interface.
	WatchResource(rType xdsresource.Type, resourceName string, watcher xdsresource.ResourceWatcher) (cancel func())

	// DumpResources returns the status of the xDS resources. Returns a map of
	// resource type URLs to a map of resource names to resource state.
	DumpResources() map[string]map[string]xdsresource.UpdateWithMD

	ReportLoad(*bootstrap.ServerConfig) (*load.Store, func())

	BootstrapConfig() *bootstrap.Config
}
