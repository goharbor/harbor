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
 *
 */

package xdsclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/internal/cache"
	"google.golang.org/grpc/internal/grpcsync"
	"google.golang.org/grpc/xds/internal/xdsclient/bootstrap"
)

// New returns a new xDS client configured by the bootstrap file specified in env
// variable GRPC_XDS_BOOTSTRAP or GRPC_XDS_BOOTSTRAP_CONFIG.
//
// The returned client is a reference counted singleton instance. This function
// creates a new client only when one doesn't already exist.
//
// The second return value represents a close function which releases the
// caller's reference on the returned client.  The caller is expected to invoke
// it once they are done using the client. The underlying client will be closed
// only when all references are released, and it is safe for the caller to
// invoke this close function multiple times.
func New() (XDSClient, func(), error) {
	return newRefCountedWithConfig(nil)
}

// NewWithConfig returns a new xDS client configured by the given config.
//
// The second return value represents a close function which releases the
// caller's reference on the returned client.  The caller is expected to invoke
// it once they are done using the client. The underlying client will be closed
// only when all references are released, and it is safe for the caller to
// invoke this close function multiple times.
//
// # Internal/Testing Only
//
// This function should ONLY be used for internal (c2p resolver) and/or testing
// purposese. DO NOT use this elsewhere. Use New() instead.
func NewWithConfig(config *bootstrap.Config) (XDSClient, func(), error) {
	return newRefCountedWithConfig(config)
}

// newWithConfig returns a new xdsClient with the given config.
func newWithConfig(config *bootstrap.Config, watchExpiryTimeout time.Duration, idleAuthorityDeleteTimeout time.Duration) (*clientImpl, error) {
	ctx, cancel := context.WithCancel(context.Background())
	c := &clientImpl{
		done:               grpcsync.NewEvent(),
		config:             config,
		watchExpiryTimeout: watchExpiryTimeout,
		serializer:         newCallbackSerializer(ctx),
		serializerClose:    cancel,
		resourceTypes:      newResourceTypeRegistry(),
		authorities:        make(map[string]*authority),
		idleAuthorities:    cache.NewTimeoutCache(idleAuthorityDeleteTimeout),
	}

	c.logger = prefixLogger(c)
	c.logger.Infof("Created client to xDS management server: %s", config.XDSServer)
	return c, nil
}

// NewWithConfigForTesting returns an xDS client for the specified bootstrap
// config, separate from the global singleton.
//
// The second return value represents a close function which the caller is
// expected to invoke once they are done using the client.  It is safe for the
// caller to invoke this close function multiple times.
//
// # Testing Only
//
// This function should ONLY be used for testing purposes.
// TODO(easwars): Document the new close func.
func NewWithConfigForTesting(config *bootstrap.Config, watchExpiryTimeout, authorityIdleTimeout time.Duration) (XDSClient, func(), error) {
	cl, err := newWithConfig(config, watchExpiryTimeout, authorityIdleTimeout)
	if err != nil {
		return nil, nil, err
	}
	return cl, grpcsync.OnceFunc(cl.close), nil
}

// NewWithBootstrapContentsForTesting returns an xDS client for this config,
// separate from the global singleton.
//
// The second return value represents a close function which the caller is
// expected to invoke once they are done using the client.  It is safe for the
// caller to invoke this close function multiple times.
//
// # Testing Only
//
// This function should ONLY be used for testing purposes.
func NewWithBootstrapContentsForTesting(contents []byte) (XDSClient, func(), error) {
	// Normalize the contents
	buf := bytes.Buffer{}
	err := json.Indent(&buf, contents, "", "")
	if err != nil {
		return nil, nil, fmt.Errorf("xds: error normalizing JSON: %v", err)
	}
	contents = bytes.TrimSpace(buf.Bytes())

	c, err := getOrMakeClientForTesting(contents)
	if err != nil {
		return nil, nil, err
	}
	return c, grpcsync.OnceFunc(func() {
		clientsMu.Lock()
		defer clientsMu.Unlock()
		if c.decrRef() == 0 {
			c.close()
			delete(clients, string(contents))
		}
	}), nil
}

// getOrMakeClientForTesting creates a new reference counted client (separate
// from the global singleton) for the given config, or returns an existing one.
// It takes care of incrementing the reference count for the returned client,
// and leaves the caller responsible for decrementing the reference count once
// the client is no longer needed.
func getOrMakeClientForTesting(config []byte) (*clientRefCounted, error) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	if c := clients[string(config)]; c != nil {
		c.incrRef()
		return c, nil
	}

	bcfg, err := bootstrap.NewConfigFromContentsForTesting(config)
	if err != nil {
		return nil, fmt.Errorf("bootstrap config %s: %v", string(config), err)
	}
	cImpl, err := newWithConfig(bcfg, defaultWatchExpiryTimeout, defaultIdleAuthorityDeleteTimeout)
	if err != nil {
		return nil, fmt.Errorf("creating xDS client: %v", err)
	}
	c := &clientRefCounted{clientImpl: cImpl, refCount: 1}
	clients[string(config)] = c
	return c, nil
}

var (
	clients   = map[string]*clientRefCounted{}
	clientsMu sync.Mutex
)
