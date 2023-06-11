/*
 *
 * Copyright 2023 gRPC authors.
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
	"sync"

	"google.golang.org/grpc/xds/internal/xdsclient/xdsresource"
)

type edsResourceWatcher interface {
	WatchEndpoints(string, func(xdsresource.EndpointsUpdate, error)) func()
}

type edsDiscoveryMechanism struct {
	cancel           func()
	topLevelResolver topLevelResolver

	mu             sync.Mutex
	update         xdsresource.EndpointsUpdate
	updateReceived bool
}

func (er *edsDiscoveryMechanism) lastUpdate() (interface{}, bool) {
	er.mu.Lock()
	defer er.mu.Unlock()

	if !er.updateReceived {
		return nil, false
	}
	return er.update, true
}

func (er *edsDiscoveryMechanism) resolveNow() {
}

func (er *edsDiscoveryMechanism) stop() {
	er.cancel()
}

func (er *edsDiscoveryMechanism) handleEndpointsUpdate(update xdsresource.EndpointsUpdate, err error) {
	if err != nil {
		er.topLevelResolver.onError(err)
		return
	}

	er.mu.Lock()
	er.update = update
	er.updateReceived = true
	er.mu.Unlock()

	er.topLevelResolver.onUpdate()
}

// newEDSResolver returns an implementation of the endpointsResolver interface
// that uses EDS to resolve the given name to endpoints.
func newEDSResolver(nameToWatch string, watcher edsResourceWatcher, topLevelResolver topLevelResolver) *edsDiscoveryMechanism {
	ret := &edsDiscoveryMechanism{topLevelResolver: topLevelResolver}
	ret.cancel = watcher.WatchEndpoints(nameToWatch, ret.handleEndpointsUpdate)
	return ret
}
