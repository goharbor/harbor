package gock

import (
	"net/http"
	"sync"
)

// Mock represents the required interface that must
// be implemented by HTTP mock instances.
type Mock interface {
	// Disable disables the current mock manually.
	Disable()

	// Done returns true if the current mock is disabled.
	Done() bool

	// Request returns the mock Request instance.
	Request() *Request

	// Response returns the mock Response instance.
	Response() *Response

	// Match matches the given http.Request with the current mock.
	Match(*http.Request) (bool, error)

	// AddMatcher adds a new matcher function.
	AddMatcher(MatchFunc)

	// SetMatcher uses a new matcher implementation.
	SetMatcher(Matcher)
}

// Mocker implements a Mock capable interface providing
// a default mock configuration used internally to store mocks.
type Mocker struct {
	// disabler stores a disabler for thread safety checking current mock is disabled
	disabler *disabler

	// mutex stores the mock mutex for thread safety.
	mutex sync.Mutex

	// matcher stores a Matcher capable instance to match the given http.Request.
	matcher Matcher

	// request stores the mock Request to match.
	request *Request

	// response stores the mock Response to use in case of match.
	response *Response
}

type disabler struct {
	// disabled stores if the current mock is disabled.
	disabled bool

	// mutex stores the disabler mutex for thread safety.
	mutex sync.RWMutex
}

func (d *disabler) isDisabled() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.disabled
}

func (d *disabler) Disable() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.disabled = true
}

// NewMock creates a new HTTP mock based on the given request and response instances.
// It's mostly used internally.
func NewMock(req *Request, res *Response) *Mocker {
	mock := &Mocker{
		disabler: new(disabler),
		request:  req,
		response: res,
		matcher:  DefaultMatcher.Clone(),
	}
	res.Mock = mock
	req.Mock = mock
	req.Response = res
	return mock
}

// Disable disables the current mock manually.
func (m *Mocker) Disable() {
	m.disabler.Disable()
}

// Done returns true in case that the current mock
// instance is disabled and therefore must be removed.
func (m *Mocker) Done() bool {
	// prevent deadlock with m.mutex
	if m.disabler.isDisabled() {
		return true
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	return !m.request.Persisted && m.request.Counter == 0
}

// Request returns the Request instance
// configured for the current HTTP mock.
func (m *Mocker) Request() *Request {
	return m.request
}

// Response returns the Response instance
// configured for the current HTTP mock.
func (m *Mocker) Response() *Response {
	return m.response
}

// Match matches the given http.Request with the current Request
// mock expectation, returning true if matches.
func (m *Mocker) Match(req *http.Request) (bool, error) {
	if m.disabler.isDisabled() {
		return false, nil
	}

	// Filter
	for _, filter := range m.request.Filters {
		if !filter(req) {
			return false, nil
		}
	}

	// Map
	for _, mapper := range m.request.Mappers {
		if treq := mapper(req); treq != nil {
			req = treq
		}
	}

	// Match
	matches, err := m.matcher.Match(req, m.request)
	if matches {
		m.decrement()
	}

	return matches, err
}

// SetMatcher sets a new matcher implementation
// for the current mock expectation.
func (m *Mocker) SetMatcher(matcher Matcher) {
	m.matcher = matcher
}

// AddMatcher adds a new matcher function
// for the current mock expectation.
func (m *Mocker) AddMatcher(fn MatchFunc) {
	m.matcher.Add(fn)
}

// decrement decrements the current mock Request counter.
func (m *Mocker) decrement() {
	if m.request.Persisted {
		return
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.request.Counter--
	if m.request.Counter == 0 {
		m.disabler.Disable()
	}
}
