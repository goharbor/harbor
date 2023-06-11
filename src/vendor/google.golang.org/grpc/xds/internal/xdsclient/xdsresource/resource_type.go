/*
 *
 * Copyright 2022 gRPC authors.
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
 */

// Package xdsresource implements the xDS data model layer.
//
// Provides resource-type specific functionality to unmarshal xDS protos into
// internal data structures that contain only fields gRPC is interested in.
// These internal data structures are passed to components in the xDS stack
// (resolver/balancers/server) that have expressed interest in receiving
// updates to specific resources.
package xdsresource

import (
	"google.golang.org/grpc/xds/internal"
	"google.golang.org/grpc/xds/internal/xdsclient/bootstrap"
	"google.golang.org/grpc/xds/internal/xdsclient/xdsresource/version"
	"google.golang.org/protobuf/types/known/anypb"
)

func init() {
	internal.ResourceTypeMapForTesting = make(map[string]interface{})
	internal.ResourceTypeMapForTesting[version.V3ListenerURL] = listenerType
	internal.ResourceTypeMapForTesting[version.V3RouteConfigURL] = routeConfigType
	internal.ResourceTypeMapForTesting[version.V3ClusterURL] = clusterType
	internal.ResourceTypeMapForTesting[version.V3EndpointsURL] = endpointsType
}

// Producer contains a single method to discover resource configuration from a
// remote management server using xDS APIs.
//
// The xdsclient package provides a concrete implementation of this interface.
type Producer interface {
	// WatchResource uses xDS to discover the resource associated with the
	// provided resource name. The resource type implementation determines how
	// xDS requests are sent out and how responses are deserialized and
	// validated. Upon receipt of a response from the management server, an
	// appropriate callback on the watcher is invoked.
	WatchResource(rType Type, resourceName string, watcher ResourceWatcher) (cancel func())
}

// ResourceWatcher wraps the callbacks to be invoked for different events
// corresponding to the resource being watched.
type ResourceWatcher interface {
	// OnUpdate is invoked to report an update for the resource being watched.
	// The ResourceData parameter needs to be type asserted to the appropriate
	// type for the resource being watched.
	OnUpdate(ResourceData)

	// OnError is invoked under different error conditions including but not
	// limited to the following:
	//	- authority mentioned in the resource is not found
	//	- resource name parsing error
	//	- resource deserialization error
	//	- resource validation error
	//	- ADS stream failure
	//	- connection failure
	OnError(error)

	// OnResourceDoesNotExist is invoked for a specific error condition where
	// the requested resource is not found on the xDS management server.
	OnResourceDoesNotExist()
}

// TODO: Once the implementation is complete, rename this interface as
// ResourceType and get rid of the existing ResourceType enum.

// Type wraps all resource-type specific functionality. Each supported resource
// type will provide an implementation of this interface.
type Type interface {
	// TypeURL is the xDS type URL of this resource type for v3 transport.
	TypeURL() string

	// TypeEnum is an enumerated value for this resource type. This can be used
	// for logging/debugging purposes, as well in cases where the resource type
	// is to be uniquely identified but the actual functionality provided by the
	// resource type is not required.
	//
	// TODO: once Type is renamed to ResourceType, rename ResourceType to
	// ResourceTypeEnum.
	TypeEnum() ResourceType

	// AllResourcesRequiredInSotW indicates whether this resource type requires
	// that all resources be present in every SotW response from the server. If
	// true, a response that does not include a previously seen resource will be
	// interpreted as a deletion of that resource.
	AllResourcesRequiredInSotW() bool

	// Decode deserializes and validates an xDS resource serialized inside the
	// provided `Any` proto, as received from the xDS management server.
	//
	// If protobuf deserialization fails or resource validation fails,
	// returns a non-nil error. Otherwise, returns a fully populated
	// DecodeResult.
	Decode(*DecodeOptions, *anypb.Any) (*DecodeResult, error)
}

// ResourceData contains the configuration data sent by the xDS management
// server, associated with the resource being watched. Every resource type must
// provide an implementation of this interface to represent the configuration
// received from the xDS management server.
type ResourceData interface {
	isResourceData()

	// Equal returns true if the passed in resource data is equal to that of the
	// receiver.
	Equal(ResourceData) bool

	// ToJSON returns a JSON string representation of the resource data.
	ToJSON() string

	Raw() *anypb.Any
}

// DecodeOptions wraps the options required by ResourceType implementation for
// decoding configuration received from the xDS management server.
type DecodeOptions struct {
	// BootstrapConfig contains the bootstrap configuration passed to the
	// top-level xdsClient. This contains useful data for resource validation.
	BootstrapConfig *bootstrap.Config
}

// DecodeResult is the result of a decode operation.
type DecodeResult struct {
	// Name is the name of the resource being watched.
	Name string
	// Resource contains the configuration associated with the resource being
	// watched.
	Resource ResourceData
}

// resourceTypeState wraps the static state associated with concrete resource
// type implementations, which can then embed this struct and get the methods
// implemented here for free.
type resourceTypeState struct {
	typeURL                    string
	typeEnum                   ResourceType
	allResourcesRequiredInSotW bool
}

func (r resourceTypeState) TypeURL() string {
	return r.typeURL
}

func (r resourceTypeState) TypeEnum() ResourceType {
	return r.typeEnum
}

func (r resourceTypeState) AllResourcesRequiredInSotW() bool {
	return r.allResourcesRequiredInSotW
}
