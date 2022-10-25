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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"oras.land/oras-go/pkg/registry"
	"oras.land/oras-go/pkg/registry/remote/auth"
)

// defaultMaxMetadataBytes specifies the default limit on how many response
// bytes are allowed in the server's response to the metadata APIs.
// See also: Repository.MaxMetadataBytes
var defaultMaxMetadataBytes int64 = 4 * 1024 * 1024 // 4 MiB

// errNoLink is returned by parseLink() when no Link header is present.
var errNoLink = errors.New("no Link header in response")

// parseLink returns the URL of the response's "Link" header, if present.
func parseLink(resp *http.Response) (string, error) {
	link := resp.Header.Get("Link")
	if link == "" {
		return "", errNoLink
	}
	if link[0] != '<' {
		return "", fmt.Errorf("invalid next link %q: missing '<'", link)
	}
	if i := strings.IndexByte(link, '>'); i == -1 {
		return "", fmt.Errorf("invalid next link %q: missing '>'", link)
	} else {
		link = link[1:i]
	}

	linkURL, err := resp.Request.URL.Parse(link)
	if err != nil {
		return "", err
	}
	return linkURL.String(), nil
}

// limitReader returns a Reader that reads from r but stops with EOF after n
// bytes. If n is zero, defaultMaxMetadataBytes is used.
func limitReader(r io.Reader, n int64) io.Reader {
	if n == 0 {
		n = defaultMaxMetadataBytes
	}
	return io.LimitReader(r, n)
}

// withScopeHint adds a hinted scope to the context.
func withScopeHint(ctx context.Context, ref registry.Reference, actions ...string) context.Context {
	scope := auth.ScopeRepository(ref.Repository, actions...)
	return auth.AppendScopes(ctx, scope)
}
