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

var urlResolvers = map[string]Resolver{}

var mu = &sync.Mutex{}

// RegisterResolver register a resolver for a specific URL pattern
func RegisterResolver(urlPattern string, resolver Resolver) {
	mu.Lock()
	urlResolvers[urlPattern] = resolver
	mu.Unlock()
}

// Resolvers get map of resolvers
func Resolvers() map[string]Resolver {
	return urlResolvers
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
	// IPAddress IP address of the request
	IPAddress string
	// ResponseLocation response location
	ResponseLocation string
	// ResourceName
	ResourceName string
}

// Resolve parse the audit information from CommonEventMetadata
func (c *Metadata) Resolve(event *event.Event) error {
	for url, r := range Resolvers() {
		p := regexp.MustCompile(url)
		if p.MatchString(c.RequestURL) {
			return r.Resolve(c, event)
		}
	}
	return nil
}

// PreCheck check if current event is matched and return the prefetched resource name when it is delete operation
func (c *Metadata) PreCheckMetadata() (bool, string) {
	for urlPattern, r := range Resolvers() {
		p := regexp.MustCompile(urlPattern)
		if p.MatchString(c.RequestURL) {
			return r.PreCheck(c.Ctx, c.RequestURL, c.RequestMethod)
		}
	}
	return false, ""
}
