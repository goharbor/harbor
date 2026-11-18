// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	// DefaultEndpoint is the public Hugging Face Hub.
	DefaultEndpoint = "https://huggingface.co"

	listPageSize      = 1000
	maxRetryAfter     = time.Minute
	defaultRetryAfter = time.Second
)

var downloadHeaderTimeout = 2 * time.Minute

// Options configures a Client.
type Options struct {
	// Token is the Hub access token. Empty means anonymous.
	Token         string
	Insecure      bool
	CACertificate string
	// Timeout bounds API calls. Downloads are not bounded by it.
	Timeout time.Duration
}

// Client talks to the Hugging Face Hub HTTP API.
type Client struct {
	endpoint *url.URL
	base     string
	hasToken bool
	api      *common_http.Client
	download *common_http.Client
}

// New creates a Client for the Hub at endpoint.
func New(endpoint string, opts Options) (*Client, error) {
	u, err := url.Parse(strings.TrimSuffix(endpoint, "/"))
	if err != nil {
		return nil, fmt.Errorf("invalid hugging face endpoint %q: %v", endpoint, err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid hugging face endpoint %q", endpoint)
	}
	transport := common_http.GetHTTPTransport(
		common_http.WithInsecure(opts.Insecure),
		common_http.WithCACert(opts.CACertificate),
	)
	var authorizer *bearer
	if opts.Token != "" {
		authorizer = &bearer{token: opts.Token}
	}
	newClient := func(timeout time.Duration) *common_http.Client {
		hc := &http.Client{Transport: transport, Timeout: timeout}
		if authorizer == nil {
			return common_http.NewClient(hc)
		}
		return common_http.NewClient(hc, authorizer)
	}
	return &Client{
		endpoint: u,
		base:     u.String(),
		hasToken: opts.Token != "",
		api:      newClient(opts.Timeout),
		// http.Client.Timeout also covers reading the body and would cut multi-GB
		// streams; downloads are bounded by a response header timeout in Open instead.
		download: newClient(0),
	}, nil
}

// Endpoint returns the Hub URL the client talks to.
func (c *Client) Endpoint() string {
	return c.base
}

// HasToken reports whether requests carry an access token.
func (c *Client) HasToken() bool {
	return c.hasToken
}

// Snapshot returns the file listing of the model at rev, which can be a ref or a commit.
func (c *Client) Snapshot(ctx context.Context, modelID, rev string) (*Snapshot, error) {
	var resp revisionResponse
	path := "/api/models/" + escapePath(modelID) + "/revision/" + url.PathEscape(rev)
	if _, err := c.getJSON(ctx, path, url.Values{"blobs": {"true"}}, &resp); err != nil {
		return nil, err
	}
	s, err := resp.toSnapshot()
	if err != nil {
		return nil, err
	}
	if IsCommit(rev) && s.Commit != rev {
		return nil, fmt.Errorf("hub returned commit %s for model %s at commit %s", s.Commit, modelID, rev)
	}
	return s, nil
}

// Refs returns the branches and tags of the model.
func (c *Client) Refs(ctx context.Context, modelID string) ([]Ref, error) {
	var resp refsResponse
	if _, err := c.getJSON(ctx, "/api/models/"+escapePath(modelID)+"/refs", nil, &resp); err != nil {
		return nil, err
	}
	var refs []Ref
	for _, group := range [][]refEntity{resp.Branches, resp.Tags} {
		for _, r := range group {
			if !IsCommit(r.TargetCommit) {
				return nil, fmt.Errorf("ref %s of model %s has invalid commit %q", r.Name, modelID, r.TargetCommit)
			}
			refs = append(refs, Ref{Name: r.Name, Commit: r.TargetCommit})
		}
	}
	return refs, nil
}

// ListModels returns the model IDs owned by author, following the Link pagination.
// The Hub matches author case-sensitively, see Author.
func (c *Client) ListModels(ctx context.Context, author string) ([]string, error) {
	next := c.url("/api/models", url.Values{"author": {author}, "limit": {strconv.Itoa(listPageSize)}})
	var ids []string
	for next != "" {
		var page []modelEntity
		header, err := c.getJSONURL(ctx, next, &page)
		if err != nil {
			return nil, err
		}
		for _, m := range page {
			ids = append(ids, m.ID)
		}
		next, err = c.nextLink(header)
		if err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// Author returns the canonical name of an organization or user, matched case-insensitively.
func (c *Client) Author(ctx context.Context, name string) (string, error) {
	var account accountEntity
	_, err := c.getJSON(ctx, "/api/organizations/"+url.PathEscape(name)+"/overview", nil, &account)
	if errors.IsNotFoundErr(err) {
		_, err = c.getJSON(ctx, "/api/users/"+url.PathEscape(name)+"/overview", nil, &account)
	}
	if err != nil {
		return "", err
	}
	if account.Name == "" {
		return "", fmt.Errorf("hub returned no name for account %s", name)
	}
	return account.Name, nil
}

// WhoAmI validates the access token.
func (c *Client) WhoAmI(ctx context.Context) error {
	_, err := c.getJSON(ctx, "/api/whoami-v2", nil, &json.RawMessage{})
	return err
}

// Ping checks that the model API answers, without credentials being required.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.getJSON(ctx, "/api/models", url.Values{"limit": {"1"}}, &json.RawMessage{})
	return err
}

// Open streams length bytes of a file at commit, starting at offset. A negative length reads to the end.
func (c *Client) Open(ctx context.Context, modelID, commit, path string, offset, length int64) (io.ReadCloser, error) {
	if !IsCommit(commit) {
		return nil, fmt.Errorf("invalid commit %q", commit)
	}
	header := http.Header{}
	ranged := offset > 0 || length >= 0
	if ranged {
		if length == 0 {
			return io.NopCloser(strings.NewReader("")), nil
		}
		end := ""
		if length > 0 {
			end = strconv.FormatInt(offset+length-1, 10)
		}
		header.Set("Range", fmt.Sprintf("bytes=%d-%s", offset, end))
	}
	target := c.url("/"+escapePath(modelID)+"/resolve/"+commit+"/"+escapePath(path), nil)

	ctx, cancel := context.WithCancel(ctx)
	timer := time.AfterFunc(downloadHeaderTimeout, cancel)
	resp, err := c.do(ctx, c.download, target, header)
	if !timer.Stop() {
		if err == nil {
			resp.Body.Close()
		}
		err = fmt.Errorf("timed out after %s waiting for response headers from hub for %s of model %s: %v", downloadHeaderTimeout, path, modelID, err)
	}
	if err != nil {
		cancel()
		return nil, err
	}
	want := http.StatusOK
	// A server may answer a range covering the whole file with 200 and the full body.
	wholeFile := offset == 0 && length > 0 && resp.StatusCode == http.StatusOK && resp.ContentLength == length
	if ranged && !wholeFile {
		want = http.StatusPartialContent
	}
	if resp.StatusCode != want {
		resp.Body.Close()
		cancel()
		return nil, fmt.Errorf("hub answered %d instead of %d for %s of model %s", resp.StatusCode, want, path, modelID)
	}
	if length > 0 && resp.ContentLength >= 0 && resp.ContentLength != length {
		resp.Body.Close()
		cancel()
		return nil, fmt.Errorf("hub returned %d bytes instead of %d for %s of model %s", resp.ContentLength, length, path, modelID)
	}
	return &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}, nil
}

type cancelOnClose struct {
	io.ReadCloser
	cancel context.CancelFunc
}

func (c *cancelOnClose) Close() error {
	defer c.cancel()
	return c.ReadCloser.Close()
}

// url joins the endpoint with an already escaped path.
func (c *Client) url(escapedPath string, query url.Values) string {
	target := c.base + escapedPath
	if len(query) > 0 {
		target += "?" + query.Encode()
	}
	return target
}

func (c *Client) getJSON(ctx context.Context, path string, query url.Values, v any) (http.Header, error) {
	return c.getJSONURL(ctx, c.url(path, query), v)
}

func (c *Client) getJSONURL(ctx context.Context, target string, v any) (http.Header, error) {
	header := http.Header{}
	header.Set("Accept", "application/json")
	resp, err := c.do(ctx, c.api, target, header)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return nil, fmt.Errorf("failed to decode hub response of %s: %v", resp.Request.URL.Path, err)
	}
	return resp.Header, nil
}

// do sends a GET and returns a 2xx response. A 429 is retried once after Retry-After.
func (c *Client) do(ctx context.Context, hc *common_http.Client, target string, header http.Header) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
		if err != nil {
			return nil, err
		}
		for k, v := range header {
			req.Header[k] = v
		}
		resp, err := hc.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
			return resp, nil
		}
		if resp.StatusCode == http.StatusTooManyRequests && attempt == 0 {
			if wait, ok := retryAfter(resp.Header); ok {
				_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxErrorBody))
				resp.Body.Close()
				select {
				case <-ctx.Done():
					return nil, ctx.Err()
				case <-time.After(wait):
				}
				continue
			}
		}
		err = statusError(resp, c.hasToken)
		resp.Body.Close()
		return nil, err
	}
}

func retryAfter(h http.Header) (time.Duration, bool) {
	v := strings.TrimSpace(h.Get("Retry-After"))
	if v == "" {
		return defaultRetryAfter, true
	}
	var wait time.Duration
	if secs, err := strconv.Atoi(v); err == nil {
		wait = time.Duration(secs) * time.Second
	} else if t, err := http.ParseTime(v); err == nil {
		wait = time.Until(t)
	} else {
		return 0, false
	}
	if wait < 0 {
		wait = 0
	}
	return wait, wait <= maxRetryAfter
}

// nextLink returns the rel="next" URL of an RFC 8288 Link header. It must stay on the
// endpoint host so the token is never sent elsewhere.
func (c *Client) nextLink(h http.Header) (string, error) {
	for _, value := range h.Values("Link") {
		for _, link := range lib.ParseLinks(value) {
			if link.Rel != "next" {
				continue
			}
			u, err := c.endpoint.Parse(link.URL)
			if err != nil {
				return "", fmt.Errorf("invalid next link %q: %v", link.URL, err)
			}
			if u.Scheme != c.endpoint.Scheme || u.Host != c.endpoint.Host {
				return "", fmt.Errorf("next link %q leaves the hub endpoint", u.Redacted())
			}
			return u.String(), nil
		}
	}
	return "", nil
}

func escapePath(p string) string {
	segments := strings.Split(p, "/")
	for i, s := range segments {
		segments[i] = url.PathEscape(s)
	}
	return strings.Join(segments, "/")
}

type bearer struct {
	token string
}

func (b *bearer) Modify(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+b.token)
	return nil
}
