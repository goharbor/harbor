package auth

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/http/modifier"
)

type apiKeyAuthorizer struct {
	key   string
	value string
	in    string
}

// NewAPIKeyAuthorizer returns a apikey authorizer
func NewAPIKeyAuthorizer(key, value, in string) modifier.Modifier {
	return &apiKeyAuthorizer{
		key:   key,
		value: value,
		in:    in,
	}
}

// Modify implements modifier.Modifier
func (a *apiKeyAuthorizer) Modify(r *http.Request) error {
	switch a.in {
	case "header":
		r.Header.Set(a.key, a.value)
		return nil
	case "query":
		query := r.URL.Query()
		query.Add(a.key, a.value)
		r.URL.RawQuery = query.Encode()
		return nil
	}
	return fmt.Errorf("Set api key in %s is invalid", a.in)
}
