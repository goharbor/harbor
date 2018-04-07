package utils

import (
	"fmt"
	"net/http"

	"github.com/docker/distribution/registry/auth/token"
	httpauth "github.com/vmware/harbor/src/common/http/modifier/auth"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
)

// NewRepositoryClient creates a repository client with standard token authorizer
func NewRepositoryClient(endpoint string, insecure bool, credential auth.Credential,
	tokenServiceEndpoint, repository string) (*registry.Repository, error) {

	transport := registry.GetHTTPTransport(insecure)

	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential, tokenServiceEndpoint)

	uam := &UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}

	return registry.NewRepository(repository, endpoint, &http.Client{
		Transport: registry.NewTransport(transport, authorizer, uam),
	})
}

// NewRepositoryClientForJobservice creates a repository client that can only be used to
// access the internal registry
func NewRepositoryClientForJobservice(repository, internalRegistryURL, secret, internalTokenServiceURL string) (*registry.Repository, error) {
	transport := registry.GetHTTPTransport()
	credential := httpauth.NewSecretAuthorizer(secret)

	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential, internalTokenServiceURL)

	uam := &UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}

	return registry.NewRepository(repository, internalRegistryURL, &http.Client{
		Transport: registry.NewTransport(transport, authorizer, uam),
	})
}

// UserAgentModifier adds the "User-Agent" header to the request
type UserAgentModifier struct {
	UserAgent string
}

// Modify adds user-agent header to the request
func (u *UserAgentModifier) Modify(req *http.Request) error {
	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), u.UserAgent)
	return nil
}

// BuildBlobURL ...
func BuildBlobURL(endpoint, repository, digest string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/%s", endpoint, repository, digest)
}

//GetTokenForRepo is used for job handler to get a token for clair.
func GetTokenForRepo(repository, secret, internalTokenServiceURL string) (string, error) {
	credential := httpauth.NewSecretAuthorizer(secret)
	t, err := auth.GetToken(internalTokenServiceURL, true, credential,
		[]*token.ResourceActions{&token.ResourceActions{
			Type:    "repository",
			Name:    repository,
			Actions: []string{"pull"},
		}})
	if err != nil {
		return "", err
	}

	return t.Token, nil
}
