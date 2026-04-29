package gock

import (
	"errors"
	"net/http"
	"sync"
)

// var mutex *sync.Mutex = &sync.Mutex{}

var (
	// DefaultTransport stores the default mock transport used by gock.
	DefaultTransport = NewTransport()

	// NativeTransport stores the native net/http default transport
	// in order to restore it when needed.
	NativeTransport = http.DefaultTransport
)

var (
	// ErrCannotMatch store the error returned in case of no matches.
	ErrCannotMatch = errors.New("gock: cannot match any request")
)

// Transport implements http.RoundTripper, which fulfills single http requests issued by
// an http.Client.
//
// gock's Transport encapsulates a given or default http.Transport for further
// delegation, if needed.
type Transport struct {
	// mutex is used to make transport thread-safe of concurrent uses across goroutines.
	mutex sync.Mutex

	// Transport encapsulates the original http.RoundTripper transport interface for delegation.
	Transport http.RoundTripper
}

// NewTransport creates a new *Transport with no responders.
func NewTransport() *Transport {
	return &Transport{Transport: NativeTransport}
}

// RoundTrip receives HTTP requests and routes them to the appropriate responder.  It is required to
// implement the http.RoundTripper interface.  You will not interact with this directly, instead
// the *http.Client you are using will call it for you.
func (m *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Just act as a proxy if not intercepting
	if !Intercepting() {
		return m.Transport.RoundTrip(req)
	}

	m.mutex.Lock()
	defer Clean()

	var err error
	var res *http.Response

	// Match mock for the incoming http.Request
	mock, err := MatchMock(req)
	if err != nil {
		m.mutex.Unlock()
		return nil, err
	}

	// Invoke the observer with the intercepted http.Request and matched mock
	if config.Observer != nil {
		config.Observer(req, mock)
	}

	// Verify if should use real networking
	networking := shouldUseNetwork(req, mock)
	if !networking && mock == nil {
		m.mutex.Unlock()
		trackUnmatchedRequest(req)
		return nil, ErrCannotMatch
	}

	// Ensure me unlock the mutex before building the response
	m.mutex.Unlock()

	// Perform real networking via original transport
	if networking {
		res, err = m.Transport.RoundTrip(req)
		// In no mock matched, continue with the response
		if err != nil || mock == nil {
			return res, err
		}
	}

	return Responder(req, mock.Response(), res)
}

// CancelRequest is a no-op function.
func (m *Transport) CancelRequest(req *http.Request) {}

func shouldUseNetwork(req *http.Request, mock Mock) bool {
	if mock != nil && mock.Response().UseNetwork {
		return true
	}
	if !config.Networking {
		return false
	}
	if len(config.NetworkingFilters) == 0 {
		return true
	}
	for _, filter := range config.NetworkingFilters {
		if !filter(req) {
			return false
		}
	}
	return true
}
