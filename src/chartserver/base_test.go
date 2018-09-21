package chartserver

import (
	"net/http/httptest"
	"net/url"

	"github.com/goharbor/harbor/src/testing"
)

// createMockObjects create mock objects for chart repo related testing.
func createMockObjects() (*httptest.Server, *Controller, error) {
	s := httptest.NewServer(testing.MockChartRepoHandler)
	backendURL, err := url.Parse(s.URL)
	if err != nil {
		s.Close()
		return nil, nil, err
	}

	mockController, err := NewController(backendURL)
	if err != nil {
		s.Close()
		return nil, nil, err
	}

	return s, mockController, nil
}

// Http client
var httpClient = NewChartClient(nil)
