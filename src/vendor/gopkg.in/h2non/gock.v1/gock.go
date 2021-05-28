package gock

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"sync"
)

// mutex is used interally for locking thread-sensitive functions.
var mutex = &sync.Mutex{}

// config global singleton store.
var config = struct {
	Networking        bool
	NetworkingFilters []FilterRequestFunc
	Observer          ObserverFunc
}{}

// ObserverFunc is implemented by users to inspect the outgoing intercepted HTTP traffic
type ObserverFunc func(*http.Request, Mock)

// DumpRequest is a default implementation of ObserverFunc that dumps
// the HTTP/1.x wire representation of the http request
var DumpRequest ObserverFunc = func(request *http.Request, mock Mock) {
	bytes, _ := httputil.DumpRequestOut(request, true)
	fmt.Println(string(bytes))
	fmt.Printf("\nMatches: %v\n---\n", mock != nil)
}

// track unmatched requests so they can be tested for
var unmatchedRequests = []*http.Request{}

// New creates and registers a new HTTP mock with
// default settings and returns the Request DSL for HTTP mock
// definition and set up.
func New(uri string) *Request {
	Intercept()

	res := NewResponse()
	req := NewRequest()
	req.URLStruct, res.Error = url.Parse(normalizeURI(uri))

	// Create the new mock expectation
	exp := NewMock(req, res)
	Register(exp)

	return req
}

// Intercepting returns true if gock is currently able to intercept.
func Intercepting() bool {
	mutex.Lock()
	defer mutex.Unlock()
	return http.DefaultTransport == DefaultTransport
}

// Intercept enables HTTP traffic interception via http.DefaultTransport.
// If you are using a custom HTTP transport, you have to use `gock.Transport()`
func Intercept() {
	if !Intercepting() {
		mutex.Lock()
		http.DefaultTransport = DefaultTransport
		mutex.Unlock()
	}
}

// InterceptClient allows the developer to intercept HTTP traffic using
// a custom http.Client who uses a non default http.Transport/http.RoundTripper implementation.
func InterceptClient(cli *http.Client) {
	_, ok := cli.Transport.(*Transport)
	if ok {
		return // if transport already intercepted, just ignore it
	}
	trans := NewTransport()
	trans.Transport = cli.Transport
	cli.Transport = trans
}

// RestoreClient allows the developer to disable and restore the
// original transport in the given http.Client.
func RestoreClient(cli *http.Client) {
	trans, ok := cli.Transport.(*Transport)
	if !ok {
		return
	}
	cli.Transport = trans.Transport
}

// Disable disables HTTP traffic interception by gock.
func Disable() {
	mutex.Lock()
	defer mutex.Unlock()
	http.DefaultTransport = NativeTransport
}

// Off disables the default HTTP interceptors and removes
// all the registered mocks, even if they has not been intercepted yet.
func Off() {
	Flush()
	Disable()
}

// OffAll is like `Off()`, but it also removes the unmatched requests registry.
func OffAll() {
	Flush()
	Disable()
	CleanUnmatchedRequest()
}

// Observe provides a hook to support inspection of the request and matched mock
func Observe(fn ObserverFunc) {
	mutex.Lock()
	defer mutex.Unlock()
	config.Observer = fn
}

// EnableNetworking enables real HTTP networking
func EnableNetworking() {
	mutex.Lock()
	defer mutex.Unlock()
	config.Networking = true
}

// DisableNetworking disables real HTTP networking
func DisableNetworking() {
	mutex.Lock()
	defer mutex.Unlock()
	config.Networking = false
}

// NetworkingFilter determines if an http.Request should be triggered or not.
func NetworkingFilter(fn FilterRequestFunc) {
	mutex.Lock()
	defer mutex.Unlock()
	config.NetworkingFilters = append(config.NetworkingFilters, fn)
}

// DisableNetworkingFilters disables registered networking filters.
func DisableNetworkingFilters() {
	mutex.Lock()
	defer mutex.Unlock()
	config.NetworkingFilters = []FilterRequestFunc{}
}

// GetUnmatchedRequests returns all requests that have been received but haven't matched any mock
func GetUnmatchedRequests() []*http.Request {
	mutex.Lock()
	defer mutex.Unlock()
	return unmatchedRequests
}

// HasUnmatchedRequest returns true if gock has received any requests that didn't match a mock
func HasUnmatchedRequest() bool {
	return len(GetUnmatchedRequests()) > 0
}

// CleanUnmatchedRequest cleans the unmatched requests internal registry.
func CleanUnmatchedRequest() {
	mutex.Lock()
	defer mutex.Unlock()
	unmatchedRequests = []*http.Request{}
}

func trackUnmatchedRequest(req *http.Request) {
	mutex.Lock()
	defer mutex.Unlock()
	unmatchedRequests = append(unmatchedRequests, req)
}

func normalizeURI(uri string) string {
	if ok, _ := regexp.MatchString("^http[s]?", uri); !ok {
		return "http://" + uri
	}
	return uri
}
