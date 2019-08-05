package auth

import (
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuthorizer(t *testing.T) {
	type suite struct {
		key   string
		value string
		in    string
	}

	var (
		s          suite
		authorizer modifier.Modifier
		request    *http.Request
		err        error
	)

	// set in header
	s = suite{key: "Authorization", value: "Basic abc", in: "header"}
	authorizer = NewAPIKeyAuthorizer(s.key, s.value, s.in)
	request, err = http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.Nil(t, err)
	err = authorizer.Modify(request)
	assert.Nil(t, err)
	assert.Equal(t, s.value, request.Header.Get(s.key))

	// set in query
	s = suite{key: "private_token", value: "abc", in: "query"}
	authorizer = NewAPIKeyAuthorizer(s.key, s.value, s.in)
	request, err = http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.Nil(t, err)
	err = authorizer.Modify(request)
	assert.Nil(t, err)
	assert.Equal(t, s.value, request.URL.Query().Get(s.key))

	// set in invalid location
	s = suite{key: "", value: "", in: "invalid"}
	authorizer = NewAPIKeyAuthorizer(s.key, s.value, s.in)
	request, err = http.NewRequest(http.MethodGet, "http://example.com", nil)
	assert.Nil(t, err)
	err = authorizer.Modify(request)
	assert.NotNil(t, err)
}
