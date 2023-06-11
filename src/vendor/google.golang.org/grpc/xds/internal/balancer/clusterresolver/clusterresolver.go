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

// Package clusterresolver contains the implementation of the
// xds_cluster_resolver_experimental LB policy which resolves endpoint addresses
// using a list of one or more discovery mechanisms.
package clusterresolver

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/internal/buffer"
	"google.golang.org/grpc/internal/grpclog"
	"google.golang.org/grpc/internal/grpcsync"
	"google.golang.org/grpc/internal/pretty"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"google.golang.org/grpc/xds/internal/balancer/priority"
	"google.golang.org/grpc/xds/internal/balancer/ringhash"
	"google.golang.org/grpc/xds/internal/xdsclient"
	"google.golang.org/grpc/xds/internal/xdsclient/xdsresource"
)

// Name is the name of the cluster_resolver balancer.
const Name = "cluster_resolver_experimental"

var (
	errBalancerClosed = errors.New("cdsBalancer is closed")
	newChildBalancer  = func(bb balancer.Builder, cc balancer.ClientConn, o balancer.BuildOptions) balancer.Balancer {
		return bb.Build(cc, o)
	}
)

func init() {
	balancer.Register(bb{})
}

type bb struct{}

// Build helps implement the balancer.Builder interface.
func (bb) Build(cc balancer.ClientConn, opts balancer.BuildOptions) balancer.Balancer {
	priorityBuilder := balancer.Get(priority.Name)
	if priorityBuilder == nil {
		logger.Errorf("%q LB policy is needed but not registered", priority.Name)
		return nil
	}
	priorityConfigParser, ok := priorityBuilder.(balancer.ConfigParser)
	if !ok {
		logger.Errorf("%q LB policy does not implement a config parser", priority.Name)
		return nil
	}

	b := &clusterResolverBalancer{
		bOpts:    opts,
		updateCh: buffer.NewUnbounded(),
		closed:   grpcsync.NewEvent(),
		done:     grpcsync.NewEvent(),

		priorityBuilder:      priorityBuilder,
		priorityConfigParser: priorityConfigParser,
	}
	b.logger = prefixLogger(b)
	b.logger.Infof("Created")

	b.resourceWatcher = newResourceResolver(b)
	b.cc = &ccWrapper{
		ClientConn:      cc,
		resourceWatcher: b.resourceWatcher,
	}

	go b.run()
	return b
}

func (bb) Name() string {
	return Name
}

func (bb) ParseConfig(c json.RawMessage) (serviceconfig.LoadBalancingConfig, error) {
	var cfg LBConfig
	if err := json.Unmarshal(c, &cfg); err != nil {
		return nil, fmt.Errorf("unable to unmarshal balancer config %s into cluster-resolver config, error: %v", string(c), err)
	}
	if lbp := cfg.XDSLBPolicy; lbp != nil && !strings.EqualFold(lbp.Name, roundrobin.Name) && !strings.EqualFold(lbp.Name, ringhash.Name) {
		return nil, fmt.Errorf("unsupported child policy with name %q, not one of {%q,%q}", lbp.Name, roundrobin.Name, ringhash.Name)
	}
	return &cfg, nil
}

// ccUpdate wraps a clientConn update received from gRPC.
type ccUpdate struct {
	state balancer.ClientConnState
	err   error
}

// scUpdate wraps a subConn update received from gRPC. This is directly passed
// on to the child policy.
type scUpdate struct {
	subConn balancer.SubConn
	state   balancer.SubConnState
}

type exitIdle struct{}

// clusterResolverBalancer resolves endpoint addresses using a list of one or
// more discovery mechanisms.
type clusterResolverBalancer struct {
	cc              balancer.ClientConn
	bOpts           balancer.BuildOptions
	updateCh        *buffer.Unbounded // Channel for updates from gRPC.
	resourceWatcher *resourceResolver
	logger          *grpclog.PrefixLogger
	closed          *grpcsync.Event
	done            *grpcsync.Event

	priorityBuilder      balancer.Builder
	priorityConfigParser balancer.ConfigParser

	config          *LBConfig
	configRaw       *serviceconfig.ParseResult
	xdsClient       xdsclient.XDSClient    // xDS client to watch EDS resource.
	attrsWithClient *attributes.Attributes // Attributes with xdsClient attached to be passed to the child policies.

	child               balancer.Balancer
	priorities          []priorityConfig
	watchUpdateReceived bool
}

// handleClientConnUpdate handles a ClientConnUpdate received from gRPC.
//
// A good update results in creation of endpoint resolvers for the configured
// discovery mechanisms. An update with an error results in cancellation of any
// existing endpoint resolution and propagation of the same to the child policy.
func (b *clusterResolverBalancer) handleClientConnUpdate(update *ccUpdate) {
	if err := update.err; err != nil {
		b.handleErrorFromUpdate(err, true)
		return
	}

	b.logger.Infof("Received new balancer config: %v", pretty.ToJSON(update.state.BalancerConfig))
	cfg, _ := update.state.BalancerConfig.(*LBConfig)
	if cfg == nil {
		b.logger.Warningf("Ignoring unsupported balancer configuration of type: %T", update.state.BalancerConfig)
		return
	}

	b.config = cfg
	b.configRaw = update.state.ResolverState.ServiceConfig
	b.resourceWatcher.updateMechanisms(cfg.DiscoveryMechanisms)

	// The child policy is created only after all configured discovery
	// mechanisms have been successfully returned endpoints. If that is not the
	// case, we return early.
	if !b.watchUpdateReceived {
		return
	}
	b.updateChildConfig()
}

// handleResourceUpdate handles a resource update or error from the resource
// resolver by propagating the same to the child LB policy.
func (b *clusterResolverBalancer) handleResourceUpdate(update *resourceUpdate) {
	if err := update.err; err != nil {
		b.handleErrorFromUpdate(err, false)
		return
	}

	b.watchUpdateReceived = true
	b.priorities = update.priorities

	// An update from the resource resolver contains resolved endpoint addresses
	// for all configured discovery mechanisms ordered by priority. This is used
	// to generate configuration for the priority LB policy.
	b.updateChildConfig()
}

// updateChildConfig builds child policy configuration using endpoint addresses
// returned by the resource resolver and child policy configuration provided by
// parent LB policy.
//
// A child policy is created if one doesn't already exist. The newly built
// configuration is then pushed to the child policy.
func (b *clusterResolverBalancer) updateChildConfig() {
	if b.child == nil {
		b.child = newChildBalancer(b.priorityBuilder, b.cc, b.bOpts)
	}

	childCfgBytes, addrs, err := buildPriorityConfigJSON(b.priorities, b.config.XDSLBPolicy)
	if err != nil {
		b.logger.Warningf("Failed to build child policy config: %v", err)
		return
	}
	childCfg, err := b.priorityConfigParser.ParseConfig(childCfgBytes)
	if err != nil {
		b.logger.Warningf("Failed to parse child policy config. This should never happen because the config was generated: %v", err)
		return
	}
	b.logger.Infof("Built child policy config: %v", pretty.ToJSON(childCfg))

	if err := b.child.UpdateClientConnState(balancer.ClientConnState{
		ResolverState: resolver.State{
			Addresses:     addrs,
			ServiceConfig: b.configRaw,
			Attributes:    b.attrsWithClient,
		},
		BalancerConfig: childCfg,
	}); err != nil {
		b.logger.Warningf("Failed to push config to child policy: %v", err)
	}
}

// handleErrorFromUpdate handles errors from the parent LB policy and endpoint
// resolvers. fromParent is true if error is from the parent LB policy. In both
// cases, the error is propagated to the child policy, if one exists.
func (b *clusterResolverBalancer) handleErrorFromUpdate(err error, fromParent bool) {
	b.logger.Warningf("Received error: %v", err)

	// A resource-not-found error from the parent LB policy means that the LDS
	// or CDS resource was removed. This should result in endpoint resolvers
	// being stopped here.
	//
	// A resource-not-found error from the EDS endpoint resolver means that the
	// EDS resource was removed. No action needs to be taken for this, and we
	// should continue watching the same EDS resource.
	if fromParent && xdsresource.ErrType(err) == xdsresource.ErrorTypeResourceNotFound {
		b.resourceWatcher.stop()
	}

	if b.child != nil {
		b.child.ResolverError(err)
		return
	}
	b.cc.UpdateState(balancer.State{
		ConnectivityState: connectivity.TransientFailure,
		Picker:            base.NewErrPicker(err),
	})
}

// run is a long-running goroutine that handles updates from gRPC and endpoint
// resolvers. The methods handling the individual updates simply push them onto
// a channel which is read and acted upon from here.
func (b *clusterResolverBalancer) run() {
	for {
		select {
		case u := <-b.updateCh.Get():
			b.updateCh.Load()
			switch update := u.(type) {
			case *ccUpdate:
				b.handleClientConnUpdate(update)
			case *scUpdate:
				// SubConn updates are simply handed over to the underlying
				// child balancer.
				if b.child == nil {
					b.logger.Errorf("Received a SubConn update {%+v} with no child policy", update)
					break
				}
				b.child.UpdateSubConnState(update.subConn, update.state)
			case exitIdle:
				if b.child == nil {
					b.logger.Errorf("xds: received ExitIdle with no child balancer")
					break
				}
				// This implementation assumes the child balancer supports
				// ExitIdle (but still checks for the interface's existence to
				// avoid a panic if not).  If the child does not, no subconns
				// will be connected.
				if ei, ok := b.child.(balancer.ExitIdler); ok {
					ei.ExitIdle()
				}
			}
		case u := <-b.resourceWatcher.updateChannel:
			b.handleResourceUpdate(u)

		// Close results in stopping the endpoint resolvers and closing the
		// underlying child policy and is the only way to exit this goroutine.
		case <-b.closed.Done():
			b.resourceWatcher.stop()

			if b.child != nil {
				b.child.Close()
				b.child = nil
			}
			// This is the *ONLY* point of return from this function.
			b.logger.Infof("Shutdown")
			b.done.Fire()
			return
		}
	}
}

// Following are methods to implement the balancer interface.

func (b *clusterResolverBalancer) UpdateClientConnState(state balancer.ClientConnState) error {
	if b.closed.HasFired() {
		b.logger.Warningf("Received update from gRPC {%+v} after close", state)
		return errBalancerClosed
	}

	if b.xdsClient == nil {
		c := xdsclient.FromResolverState(state.ResolverState)
		if c == nil {
			return balancer.ErrBadResolverState
		}
		b.xdsClient = c
		b.attrsWithClient = state.ResolverState.Attributes
	}

	b.updateCh.Put(&ccUpdate{state: state})
	return nil
}

// ResolverError handles errors reported by the xdsResolver.
func (b *clusterResolverBalancer) ResolverError(err error) {
	if b.closed.HasFired() {
		b.logger.Warningf("Received resolver error {%v} after close", err)
		return
	}
	b.updateCh.Put(&ccUpdate{err: err})
}

// UpdateSubConnState handles subConn updates from gRPC.
func (b *clusterResolverBalancer) UpdateSubConnState(sc balancer.SubConn, state balancer.SubConnState) {
	if b.closed.HasFired() {
		b.logger.Warningf("Received subConn update {%v, %v} after close", sc, state)
		return
	}
	b.updateCh.Put(&scUpdate{subConn: sc, state: state})
}

// Close closes the cdsBalancer and the underlying child balancer.
func (b *clusterResolverBalancer) Close() {
	b.closed.Fire()
	<-b.done.Done()
}

func (b *clusterResolverBalancer) ExitIdle() {
	b.updateCh.Put(exitIdle{})
}

// ccWrapper overrides ResolveNow(), so that re-resolution from the child
// policies will trigger the DNS resolver in cluster_resolver balancer.
type ccWrapper struct {
	balancer.ClientConn
	resourceWatcher *resourceResolver
}

func (c *ccWrapper) ResolveNow(resolver.ResolveNowOptions) {
	c.resourceWatcher.resolveNow()
}
