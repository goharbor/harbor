/*
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

// Package resolver implements the xds resolver, that does LDS and RDS to find
// the cluster to use.
package resolver

import (
	"errors"
	"fmt"
	"strings"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/internal"
	"google.golang.org/grpc/internal/grpclog"
	"google.golang.org/grpc/internal/grpcrand"
	"google.golang.org/grpc/internal/grpcsync"
	"google.golang.org/grpc/internal/pretty"
	iresolver "google.golang.org/grpc/internal/resolver"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/xds/internal/xdsclient"
	"google.golang.org/grpc/xds/internal/xdsclient/bootstrap"
	"google.golang.org/grpc/xds/internal/xdsclient/xdsresource"
)

const xdsScheme = "xds"

// newBuilderForTesting creates a new xds resolver builder using a specific xds
// bootstrap config, so tests can use multiple xds clients in different
// ClientConns at the same time.
func newBuilderForTesting(config []byte) (resolver.Builder, error) {
	return &xdsResolverBuilder{
		newXDSClient: func() (xdsclient.XDSClient, func(), error) {
			return xdsclient.NewWithBootstrapContentsForTesting(config)
		},
	}, nil
}

// For overriding in unittests.
var newXDSClient = func() (xdsclient.XDSClient, func(), error) { return xdsclient.New() }

func init() {
	resolver.Register(&xdsResolverBuilder{})
	internal.NewXDSResolverWithConfigForTesting = newBuilderForTesting
}

type xdsResolverBuilder struct {
	newXDSClient func() (xdsclient.XDSClient, func(), error)
}

// Build helps implement the resolver.Builder interface.
//
// The xds bootstrap process is performed (and a new xds client is built) every
// time an xds resolver is built.
func (b *xdsResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (_ resolver.Resolver, retErr error) {
	r := &xdsResolver{
		cc:             cc,
		closed:         grpcsync.NewEvent(),
		updateCh:       make(chan suWithError, 1),
		activeClusters: make(map[string]*clusterInfo),
		channelID:      grpcrand.Uint64(),
	}
	defer func() {
		if retErr != nil {
			r.Close()
		}
	}()
	r.logger = prefixLogger(r)
	r.logger.Infof("Creating resolver for target: %+v", target)

	newXDSClient := newXDSClient
	if b.newXDSClient != nil {
		newXDSClient = b.newXDSClient
	}

	client, close, err := newXDSClient()
	if err != nil {
		return nil, fmt.Errorf("xds: failed to create xds-client: %v", err)
	}
	r.xdsClient = client
	r.xdsClientClose = close
	bootstrapConfig := client.BootstrapConfig()
	if bootstrapConfig == nil {
		return nil, errors.New("bootstrap configuration is empty")
	}

	// If xds credentials were specified by the user, but bootstrap configs do
	// not contain any certificate provider configuration, it is better to fail
	// right now rather than failing when attempting to create certificate
	// providers after receiving an CDS response with security configuration.
	var creds credentials.TransportCredentials
	switch {
	case opts.DialCreds != nil:
		creds = opts.DialCreds
	case opts.CredsBundle != nil:
		creds = opts.CredsBundle.TransportCredentials()
	}
	if xc, ok := creds.(interface{ UsesXDS() bool }); ok && xc.UsesXDS() {
		if len(bootstrapConfig.CertProviderConfigs) == 0 {
			return nil, errors.New("xds: xdsCreds specified but certificate_providers config missing in bootstrap file")
		}
	}

	// Find the client listener template to use from the bootstrap config:
	// - If authority is not set in the target, use the top level template
	// - If authority is set, use the template from the authority map.
	template := bootstrapConfig.ClientDefaultListenerResourceNameTemplate
	if authority := target.URL.Host; authority != "" {
		a := bootstrapConfig.Authorities[authority]
		if a == nil {
			return nil, fmt.Errorf("xds: authority %q is not found in the bootstrap file", authority)
		}
		if a.ClientListenerResourceNameTemplate != "" {
			// This check will never be false, because
			// ClientListenerResourceNameTemplate is required to start with
			// xdstp://, and has a default value (not an empty string) if unset.
			template = a.ClientListenerResourceNameTemplate
		}
	}
	endpoint := target.URL.Path
	if endpoint == "" {
		endpoint = target.URL.Opaque
	}
	endpoint = strings.TrimPrefix(endpoint, "/")
	r.ldsResourceName = bootstrap.PopulateResourceTemplate(template, endpoint)

	// Register a watch on the xdsClient for the resource name determined above.
	cancelWatch := watchService(r.xdsClient, r.ldsResourceName, r.handleServiceUpdate, r.logger)
	r.logger.Infof("Watch started on resource name %v with xds-client %p", r.ldsResourceName, r.xdsClient)
	r.cancelWatch = func() {
		cancelWatch()
		r.logger.Infof("Watch cancel on resource name %v with xds-client %p", r.ldsResourceName, r.xdsClient)
	}

	go r.run()
	return r, nil
}

// Name helps implement the resolver.Builder interface.
func (*xdsResolverBuilder) Scheme() string {
	return xdsScheme
}

// suWithError wraps the ServiceUpdate and error received through a watch API
// callback, so that it can pushed onto the update channel as a single entity.
type suWithError struct {
	su          serviceUpdate
	emptyUpdate bool
	err         error
}

// xdsResolver implements the resolver.Resolver interface.
//
// It registers a watcher for ServiceConfig updates with the xdsClient object
// (which performs LDS/RDS queries for the same), and passes the received
// updates to the ClientConn.
type xdsResolver struct {
	cc              resolver.ClientConn
	closed          *grpcsync.Event
	logger          *grpclog.PrefixLogger
	ldsResourceName string

	// The underlying xdsClient which performs all xDS requests and responses.
	xdsClient      xdsclient.XDSClient
	xdsClientClose func()
	// A channel for the watch API callback to write service updates on to. The
	// updates are read by the run goroutine and passed on to the ClientConn.
	updateCh chan suWithError
	// cancelWatch is the function to cancel the watcher.
	cancelWatch func()

	// activeClusters is a map from cluster name to a ref count.  Only read or
	// written during a service update (synchronous).
	activeClusters map[string]*clusterInfo

	curConfigSelector *configSelector

	// A random number which uniquely identifies the channel which owns this
	// resolver.
	channelID uint64
}

// sendNewServiceConfig prunes active clusters, generates a new service config
// based on the current set of active clusters, and sends an update to the
// channel with that service config and the provided config selector.  Returns
// false if an error occurs while generating the service config and the update
// cannot be sent.
func (r *xdsResolver) sendNewServiceConfig(cs *configSelector) bool {
	// Delete entries from r.activeClusters with zero references;
	// otherwise serviceConfigJSON will generate a config including
	// them.
	r.pruneActiveClusters()

	if cs == nil && len(r.activeClusters) == 0 {
		// There are no clusters and we are sending a failing configSelector.
		// Send an empty config, which picks pick-first, with no address, and
		// puts the ClientConn into transient failure.
		r.cc.UpdateState(resolver.State{ServiceConfig: r.cc.ParseServiceConfig("{}")})
		return true
	}

	sc, err := serviceConfigJSON(r.activeClusters)
	if err != nil {
		// JSON marshal error; should never happen.
		r.logger.Errorf("%v", err)
		r.cc.ReportError(err)
		return false
	}
	r.logger.Infof("Received update on resource %v from xds-client %p, generated service config: %v", r.ldsResourceName, r.xdsClient, pretty.FormatJSON(sc))

	// Send the update to the ClientConn.
	state := iresolver.SetConfigSelector(resolver.State{
		ServiceConfig: r.cc.ParseServiceConfig(string(sc)),
	}, cs)
	r.cc.UpdateState(xdsclient.SetClient(state, r.xdsClient))
	return true
}

// run is a long running goroutine which blocks on receiving service updates
// and passes it on the ClientConn.
func (r *xdsResolver) run() {
	for {
		select {
		case <-r.closed.Done():
			return
		case update := <-r.updateCh:
			if update.err != nil {
				r.logger.Warningf("Watch error on resource %v from xds-client %p, %v", r.ldsResourceName, r.xdsClient, update.err)
				if xdsresource.ErrType(update.err) == xdsresource.ErrorTypeResourceNotFound {
					// If error is resource-not-found, it means the LDS
					// resource was removed. Ultimately send an empty service
					// config, which picks pick-first, with no address, and
					// puts the ClientConn into transient failure.  Before we
					// can do that, we may need to send a normal service config
					// along with an erroring (nil) config selector.
					r.sendNewServiceConfig(nil)
					// Stop and dereference the active config selector, if one exists.
					r.curConfigSelector.stop()
					r.curConfigSelector = nil
					continue
				}
				// Send error to ClientConn, and balancers, if error is not
				// resource not found.  No need to update resolver state if we
				// can keep using the old config.
				r.cc.ReportError(update.err)
				continue
			}
			if update.emptyUpdate {
				r.sendNewServiceConfig(r.curConfigSelector)
				continue
			}

			// Create the config selector for this update.
			cs, err := r.newConfigSelector(update.su)
			if err != nil {
				r.logger.Warningf("Error parsing update on resource %v from xds-client %p: %v", r.ldsResourceName, r.xdsClient, err)
				r.cc.ReportError(err)
				continue
			}

			if !r.sendNewServiceConfig(cs) {
				// JSON error creating the service config (unexpected); erase
				// this config selector and ignore this update, continuing with
				// the previous config selector.
				cs.stop()
				continue
			}

			// Decrement references to the old config selector and assign the
			// new one as the current one.
			r.curConfigSelector.stop()
			r.curConfigSelector = cs
		}
	}
}

// handleServiceUpdate is the callback which handles service updates. It writes
// the received update to the update channel, which is picked by the run
// goroutine.
func (r *xdsResolver) handleServiceUpdate(su serviceUpdate, err error) {
	if r.closed.HasFired() {
		// Do not pass updates to the ClientConn once the resolver is closed.
		return
	}
	// Remove any existing entry in updateCh and replace with the new one.
	select {
	case <-r.updateCh:
	default:
	}
	r.updateCh <- suWithError{su: su, err: err}
}

// ResolveNow is a no-op at this point.
func (*xdsResolver) ResolveNow(o resolver.ResolveNowOptions) {}

// Close closes the resolver, and also closes the underlying xdsClient.
func (r *xdsResolver) Close() {
	// Note that Close needs to check for nils even if some of them are always
	// set in the constructor. This is because the constructor defers Close() in
	// error cases, and the fields might not be set when the error happens.
	if r.cancelWatch != nil {
		r.cancelWatch()
	}
	if r.xdsClientClose != nil {
		r.xdsClientClose()
	}
	r.closed.Fire()
	r.logger.Infof("Shutdown")
}
