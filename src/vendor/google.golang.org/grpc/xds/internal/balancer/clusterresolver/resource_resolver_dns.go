/*
 *
 * Copyright 2021 gRPC authors.
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

package clusterresolver

import (
	"fmt"
	"net/url"
	"sync"

	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
)

var (
	newDNS = func(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
		// The dns resolver is registered by the grpc package. So, this call to
		// resolver.Get() is never expected to return nil.
		return resolver.Get("dns").Build(target, cc, opts)
	}
)

// dnsDiscoveryMechanism watches updates for the given DNS hostname.
//
// It implements resolver.ClientConn interface to work with the DNS resolver.
type dnsDiscoveryMechanism struct {
	target           string
	topLevelResolver topLevelResolver
	dnsR             resolver.Resolver

	mu             sync.Mutex
	addrs          []string
	updateReceived bool
}

// newDNSResolver creates an endpoints resolver which uses a DNS resolver under
// the hood.
//
// An error in parsing the provided target string or an error in creating a DNS
// resolver means that we will never be able to resolve the provided target
// strings to endpoints. The topLevelResolver propagates address updates to the
// clusterresolver LB policy **only** after it receives updates from all its
// child resolvers. Therefore, an error here means that the topLevelResolver
// will never send address updates to the clusterresolver LB policy.
//
// Calling the onError() callback will ensure that this error is
// propagated to the child policy which eventually move the channel to
// transient failure.
//
// The `dnsR` field is unset if we run into erros in this function. Therefore, a
// nil check is required wherever we access that field.
func newDNSResolver(target string, topLevelResolver topLevelResolver) *dnsDiscoveryMechanism {
	ret := &dnsDiscoveryMechanism{
		target:           target,
		topLevelResolver: topLevelResolver,
	}
	u, err := url.Parse("dns:///" + target)
	if err != nil {
		topLevelResolver.onError(fmt.Errorf("failed to parse dns hostname %q in clusterresolver LB policy", target))
		return ret
	}

	r, err := newDNS(resolver.Target{Scheme: "dns", URL: *u}, ret, resolver.BuildOptions{})
	if err != nil {
		topLevelResolver.onError(fmt.Errorf("failed to build DNS resolver for target %q: %v", target, err))
		return ret
	}
	ret.dnsR = r
	return ret
}

func (dr *dnsDiscoveryMechanism) lastUpdate() (interface{}, bool) {
	dr.mu.Lock()
	defer dr.mu.Unlock()

	if !dr.updateReceived {
		return nil, false
	}
	return dr.addrs, true
}

func (dr *dnsDiscoveryMechanism) resolveNow() {
	if dr.dnsR != nil {
		dr.dnsR.ResolveNow(resolver.ResolveNowOptions{})
	}
}

func (dr *dnsDiscoveryMechanism) stop() {
	if dr.dnsR != nil {
		dr.dnsR.Close()
	}
}

// dnsDiscoveryMechanism needs to implement resolver.ClientConn interface to receive
// updates from the real DNS resolver.

func (dr *dnsDiscoveryMechanism) UpdateState(state resolver.State) error {
	dr.mu.Lock()
	addrs := make([]string, len(state.Addresses))
	for i, a := range state.Addresses {
		addrs[i] = a.Addr
	}
	dr.addrs = addrs
	dr.updateReceived = true
	dr.mu.Unlock()

	dr.topLevelResolver.onUpdate()
	return nil
}

func (dr *dnsDiscoveryMechanism) ReportError(err error) {
	dr.topLevelResolver.onError(err)
}

func (dr *dnsDiscoveryMechanism) NewAddress(addresses []resolver.Address) {
	dr.UpdateState(resolver.State{Addresses: addresses})
}

func (dr *dnsDiscoveryMechanism) NewServiceConfig(string) {
	// This method is deprecated, and service config isn't supported.
}

func (dr *dnsDiscoveryMechanism) ParseServiceConfig(string) *serviceconfig.ParseResult {
	return &serviceconfig.ParseResult{Err: fmt.Errorf("service config not supported")}
}
