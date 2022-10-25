/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package remote

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	errdef "oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/registry"
	"oras.land/oras-go/pkg/registry/remote/auth"
	"oras.land/oras-go/pkg/registry/remote/internal/errutil"
)

// Client is an interface for a HTTP client.
type Client interface {
	// Do sends an HTTP request and returns an HTTP response.
	//
	// Unlike http.RoundTripper, Client can attempt to interpret the response
	// and handle higher-level protocol details such as redirects and
	// authentication.
	//
	// Like http.RoundTripper, Client should not modify the request, and must
	// always close the request body.
	Do(*http.Request) (*http.Response, error)
}

// Repository is an HTTP client to a remote repository.
type Repository struct {
	// Client is the underlying HTTP client used to access the remote registry.
	// If nil, auth.DefaultClient is used.
	Client Client

	// Reference references the remote repository.
	Reference registry.Reference

	// PlainHTTP signals the transport to access the remote repository via HTTP
	// instead of HTTPS.
	PlainHTTP bool

	// ManifestMediaTypes is used in `Accept` header for resolving manifests from
	// references. It is also used in identifying manifests and blobs from
	// descriptors.
	// If an empty list is present, default manifest media types are used.
	ManifestMediaTypes []string

	// TagListPageSize specifies the page size when invoking the tag list API.
	// If zero, the page size is determined by the remote registry.
	// Reference: https://docs.docker.com/registry/spec/api/#tags
	TagListPageSize int

	// ReferrerListPageSize specifies the page size when invoking the Referrers
	// API.
	// If zero, the page size is determined by the remote registry.
	// Reference: https://github.com/oras-project/artifacts-spec/blob/main/manifest-referrers-api.md
	ReferrerListPageSize int

	// MaxMetadataBytes specifies a limit on how many response bytes are allowed
	// in the server's response to the metadata APIs, such as catalog list, tag
	// list, and referrers list.
	// If zero, a default (currently 4MiB) is used.
	MaxMetadataBytes int64
}

// NewRepository creates a client to the remote repository identified by a
// reference.
// Example: localhost:5000/hello-world
func NewRepository(reference string) (*Repository, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		return nil, err
	}
	return &Repository{
		Reference: ref,
	}, nil
}

// client returns an HTTP client used to access the remote repository.
// A default HTTP client is return if the client is not configured.
func (r *Repository) client() Client {
	if r.Client == nil {
		return auth.DefaultClient
	}
	return r.Client
}

// parseReference validates the reference.
// Both simplified or fully qualified references are accepted as input.
// A fully qualified reference is returned on success.
func (r *Repository) parseReference(reference string) (registry.Reference, error) {
	ref, err := registry.ParseReference(reference)
	if err != nil {
		ref = registry.Reference{
			Registry:   r.Reference.Registry,
			Repository: r.Reference.Repository,
			Reference:  reference,
		}
		if err = ref.ValidateReference(); err != nil {
			return registry.Reference{}, err
		}
		return ref, nil
	}
	if ref.Registry == r.Reference.Registry && ref.Repository == r.Reference.Repository {
		return ref, nil
	}
	return registry.Reference{}, fmt.Errorf("%w %q: expect %q", errdef.ErrInvalidReference, ref, r.Reference)
}

// Tags lists the tags available in the repository.
func (r *Repository) Tags(ctx context.Context, fn func(tags []string) error) error {
	ctx = withScopeHint(ctx, r.Reference, auth.ActionPull)
	url := buildRepositoryTagListURL(r.PlainHTTP, r.Reference)
	var err error
	for err == nil {
		url, err = r.tags(ctx, fn, url)
	}
	if err != errNoLink {
		return err
	}
	return nil
}

// tags returns a single page of tag list with the next link.
func (r *Repository) tags(ctx context.Context, fn func(tags []string) error, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if r.TagListPageSize > 0 {
		q := req.URL.Query()
		q.Set("n", strconv.Itoa(r.TagListPageSize))
		req.URL.RawQuery = q.Encode()
	}

	resp, err := r.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errutil.ParseErrorResponse(resp)
	}
	var page struct {
		Tags []string `json:"tags"`
	}
	lr := limitReader(resp.Body, r.MaxMetadataBytes)
	if err := json.NewDecoder(lr).Decode(&page); err != nil {
		return "", fmt.Errorf("%s %q: failed to decode response: %w", resp.Request.Method, resp.Request.URL, err)
	}
	if err := fn(page.Tags); err != nil {
		return "", err
	}

	return parseLink(resp)
}
