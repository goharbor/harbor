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

package xdsclient

import (
	"google.golang.org/grpc/xds/internal/xdsclient/xdsresource"
)

func appendMaps(dst, src map[string]map[string]xdsresource.UpdateWithMD) {
	// Iterate through the resource types.
	for rType, srcResources := range src {
		// Lookup/create the resource type specific map in the destination.
		dstResources := dst[rType]
		if dstResources == nil {
			dstResources = make(map[string]xdsresource.UpdateWithMD)
			dst[rType] = dstResources
		}

		// Iterate through the resources within the resource type in the source,
		// and copy them over to the destination.
		for name, update := range srcResources {
			dstResources[name] = update
		}
	}
}

// DumpResources returns the status and contents of all xDS resources.
func (c *clientImpl) DumpResources() map[string]map[string]xdsresource.UpdateWithMD {
	c.authorityMu.Lock()
	defer c.authorityMu.Unlock()
	dumps := make(map[string]map[string]xdsresource.UpdateWithMD)
	for _, a := range c.authorities {
		dump := a.dumpResources()
		appendMaps(dumps, dump)
	}
	return dumps
}
