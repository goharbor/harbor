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
	"regexp"
	"sync"

	"github.com/goharbor/harbor/src/pkg/notifier/event"
)

// Resolver the interface to resolve Metadata to CommonEvent
type Resolver interface {
	Resolve(*Metadata, *event.Event) error
	PreCheck(ctx context.Context, url string, method string) (bool, string)
}

type registration struct {
	pattern  *regexp.Regexp
	resolver Resolver
}

var registrations = map[string]registration{}

var mu = &sync.Mutex{}

// RegisterResolver register a resolver for a specific URL pattern
func RegisterResolver(urlPattern string, resolver Resolver) {
	p := regexp.MustCompile(urlPattern)
	mu.Lock()
	defer mu.Unlock()
	registrations[urlPattern] = registration{pattern: p, resolver: resolver}
}

// Resolvers returns a copy of the registered resolvers keyed by URL pattern
func Resolvers() map[string]Resolver {
	mu.Lock()
	defer mu.Unlock()
	resolvers := make(map[string]Resolver, len(registrations))
	for urlPattern, reg := range registrations {
		resolvers[urlPattern] = reg.resolver
	}
	return resolvers
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
	for _, reg := range registrations {
		if reg.pattern.MatchString(c.RequestURL) {
			return reg.resolver.Resolve(c, event)
		}
	}
	return nil
}

// PreCheck check if current event is matched and return the prefetched resource name when it is delete operation
func (c *Metadata) PreCheckMetadata() (bool, string) {
	for _, reg := range registrations {
		if reg.pattern.MatchString(c.RequestURL) {
			return reg.resolver.PreCheck(c.Ctx, c.RequestURL, c.RequestMethod)
		}
	}
	return false, ""
}
