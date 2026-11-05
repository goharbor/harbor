// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package commonevent

import (
	"context"
	"net/url"
	"regexp"
	"sync"

	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// Resolver the interface to resolve Metadata to CommonEvent
type Resolver interface {
	Resolve(*Metadata, *event.Event) error
	PreCheck(ctx context.Context, url string, method string) (bool, string)
}

var urlResolvers = map[string]Resolver{}

var mu = &sync.Mutex{}

// RegisterResolver register a resolver for a specific URL pattern
func RegisterResolver(urlPattern string, resolver Resolver) {
	mu.Lock()
	urlResolvers[urlPattern] = resolver
	mu.Unlock()
}

// UnregisterResolver removes the resolver registered for a URL pattern
func UnregisterResolver(urlPattern string) {
	mu.Lock()
	delete(urlResolvers, urlPattern)
	mu.Unlock()
}

// Resolvers returns a snapshot of the registered resolvers, safe to range over
// while resolvers are registered or unregistered concurrently
func Resolvers() map[string]Resolver {
	mu.Lock()
	defer mu.Unlock()
	snapshot := make(map[string]Resolver, len(urlResolvers))
	for urlPattern, resolver := range urlResolvers {
		snapshot[urlPattern] = resolver
	}
	return snapshot
}

// Metadata the raw data of event
type Metadata struct {
	// Ctx ...
	Ctx context.Context
	// Username requester username
	Username string
	// RequestPayload http request payload
	RequestPayload string
	// RequestMethod
	RequestMethod string
	// ResponseCode response code
	ResponseCode int
	// RequestURL request URL
	RequestURL string
	// IsResourceName indicates the request declared, via the X-Is-Resource-Name
	// header, that name/ID path parameters are resource names
	IsResourceName bool
	// IPAddress IP address of the request
	IPAddress string
	// ResponseLocation response location
	ResponseLocation string
	// ResourceName resource name
	ResourceName string
	// Payload request payload
	Payload string
}

// Resolve parse the audit information from CommonEventMetadata
func (c *Metadata) Resolve(event *event.Event) error {
	resolver, metadata, ok := c.resolver()
	if !ok {
		return nil
	}
	return resolver.Resolve(metadata, event)
}

// PreCheck check if current event is matched and return the prefetched resource name when it is delete operation
func (c *Metadata) PreCheckMetadata() (bool, string) {
	resolver, metadata, ok := c.resolver()
	if ok {
		return resolver.PreCheck(metadata.Ctx, metadata.RequestURL, metadata.RequestMethod)
	}
	return false, ""
}

func (c *Metadata) resolver() (Resolver, *Metadata, bool) {
	metadata := *c
	if requestURL, err := url.Parse(c.RequestURL); err == nil {
		metadata.RequestURL = requestURL.Path
	}

	for urlPattern, resolver := range Resolvers() {
		match := regexp.MustCompile(urlPattern).FindStringIndex(metadata.RequestURL)
		if match != nil && match[0] == 0 && match[1] == len(metadata.RequestURL) {
			return resolver, &metadata, true
		}
	}
	return nil, &metadata, false
}
