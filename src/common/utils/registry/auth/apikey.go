package auth

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/http/modifier"
)

type apiKeyType = string

const (
	// APIKeyInHeader sets auth content in header
	APIKeyInHeader apiKeyType = "header"
	// APIKeyInQuery sets auth content in url query
	APIKeyInQuery apiKeyType = "query"
)

type apiKeyAuthorizer struct {
	key, value, in apiKeyType
}

// NewAPIKeyAuthorizer returns a apikey authorizer
func NewAPIKeyAuthorizer(key, value, in apiKeyType) modifier.Modifier {
	return &apiKeyAuthorizer{
		key:   key,
		value: value,
		in:    in,
	}
}

// Modify implements modifier.Modifier
func (a *apiKeyAuthorizer) Modify(r *http.Request) error {
	switch a.in {
	case APIKeyInHeader:
		r.Header.Set(a.key, a.value)
		return nil
	case APIKeyInQuery:
		query := r.URL.Query()
		query.Add(a.key, a.value)
		r.URL.RawQuery = query.Encode()
		return nil
	}
	return fmt.Errorf("set api key in %s is invalid", a.in)
}
