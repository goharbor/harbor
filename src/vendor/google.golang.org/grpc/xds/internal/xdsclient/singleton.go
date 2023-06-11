/*
 *
 * Copyright 2020 gRPC authors.
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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/internal/envconfig"
	"google.golang.org/grpc/internal/grpcsync"
	"google.golang.org/grpc/xds/internal/xdsclient/bootstrap"
)

const (
	defaultWatchExpiryTimeout         = 15 * time.Second
	defaultIdleAuthorityDeleteTimeout = 5 * time.Minute
)

var (
	// This is the client returned by New(). It contains one client implementation,
	// and maintains the refcount.
	singletonMu     sync.Mutex
	singletonClient *clientRefCounted

	// The following functions are no-ops in the actual code, but can be
	// overridden in tests to give them visibility into certain events.
	singletonClientImplCreateHook = func() {}
	singletonClientImplCloseHook  = func() {}
)

// To override in tests.
var bootstrapNewConfig = bootstrap.NewConfig

func clientRefCountedClose() {
	singletonMu.Lock()
	defer singletonMu.Unlock()

	if singletonClient.decrRef() != 0 {
		return
	}
	singletonClient.clientImpl.close()
	singletonClientImplCloseHook()
	singletonClient = nil
}

func newRefCountedWithConfig(fallbackConfig *bootstrap.Config) (XDSClient, func(), error) {
	singletonMu.Lock()
	defer singletonMu.Unlock()

	if singletonClient != nil {
		singletonClient.incrRef()
		return singletonClient, grpcsync.OnceFunc(clientRefCountedClose), nil

	}

	// Use fallbackConfig only if bootstrap env vars are unspecified.
	var config *bootstrap.Config
	if envconfig.XDSBootstrapFileName == "" && envconfig.XDSBootstrapFileContent == "" {
		if fallbackConfig == nil {
			return nil, nil, fmt.Errorf("xds: bootstrap env vars are unspecified and provided fallback config is nil")
		}
		config = fallbackConfig
	} else {
		var err error
		config, err = bootstrapNewConfig()
		if err != nil {
			return nil, nil, fmt.Errorf("xds: failed to read bootstrap file: %v", err)
		}
	}

	// Create the new client implementation.
	c, err := newWithConfig(config, defaultWatchExpiryTimeout, defaultIdleAuthorityDeleteTimeout)
	if err != nil {
		return nil, nil, err
	}
	singletonClient = &clientRefCounted{clientImpl: c, refCount: 1}
	singletonClientImplCreateHook()

	logger.Infof("xDS node ID: %s", config.NodeProto.GetId())
	return singletonClient, grpcsync.OnceFunc(clientRefCountedClose), nil
}

// clientRefCounted is ref-counted, and to be shared by the xds resolver and
// balancer implementations, across multiple ClientConns and Servers.
type clientRefCounted struct {
	*clientImpl

	refCount int32 // accessed atomically
}

func (c *clientRefCounted) incrRef() int32 {
	return atomic.AddInt32(&c.refCount, 1)
}

func (c *clientRefCounted) decrRef() int32 {
	return atomic.AddInt32(&c.refCount, -1)
}
