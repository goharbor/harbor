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

package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	_ "github.com/docker/distribution/manifest/ocischema" // register oci manifest unmarshal function
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/registry/auth"
	"github.com/goharbor/harbor/src/pkg/registry/interceptor"
	"github.com/goharbor/harbor/src/pkg/registry/interceptor/readonly"
)

var (
	// Cli is the global registry client instance, it targets to the backend docker registry
	Cli = func() Client {
		url, _ := config.RegistryURL()
		username, password := config.RegistryCredential()
		return NewClient(url, username, password, false, readonly.NewInterceptor())
	}()

	accepts = []string{
		v1.MediaTypeImageIndex,
		manifestlist.MediaTypeManifestList,
		v1.MediaTypeImageManifest,
		schema2.MediaTypeManifest,
		schema1.MediaTypeSignedManifest,
		schema1.MediaTypeManifest,
	}
)

// const definition
const (
	UserAgent = "harbor-registry-client"
	// DefaultHTTPClientTimeout is the default timeout for registry http client.
	DefaultHTTPClientTimeout = 30 * time.Minute
)

var (
	// registryHTTPClientTimeout is the timeout for registry http client.
	registryHTTPClientTimeout time.Duration
)

func init() {
	registryHTTPClientTimeout = DefaultHTTPClientTimeout
	// override it if read from environment variable, in minutes
	if env := os.Getenv("REGISTRY_HTTP_CLIENT_TIMEOUT"); len(env) > 0 {
		timeout, err := strconv.ParseInt(env, 10, 64)
		if err != nil {
			log.Errorf("Failed to parse REGISTRY_HTTP_CLIENT_TIMEOUT: %v, use default value: %v", err, DefaultHTTPClientTimeout)
		} else {
			if timeout > 0 {
				registryHTTPClientTimeout = time.Duration(timeout) * time.Minute
			}
		}
	}
}

// Client defines the methods that a registry client should implements
type Client interface {
	// Ping the base API endpoint "/v2/"
	Ping() (err error)
	// Catalog the repositories
	Catalog() (repositories []string, err error)
	// ListTags lists the tags under the specified repository
	ListTags(repository string) (tags []string, err error)
	// ManifestExist checks the existence of the manifest
	ManifestExist(repository, reference string) (exist bool, desc *distribution.Descriptor, err error)
	// PullManifest pulls the specified manifest
	PullManifest(repository, reference string, acceptedMediaTypes ...string) (manifest distribution.Manifest, digest string, err error)
	// PushManifest pushes the specified manifest
	PushManifest(repository, reference, mediaType string, payload []byte) (digest string, err error)
	// DeleteManifest deletes the specified manifest. The "reference" can be "tag" or "digest"
	DeleteManifest(repository, reference string) (err error)
	// BlobExist checks the existence of the specified blob
	BlobExist(repository, digest string) (exist bool, err error)
	// PullBlob pulls the specified blob. The caller must close the returned "blob"
	PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error)
	// PullBlobChunk pulls the specified blob, but by chunked
	PullBlobChunk(repository, digest string, blobSize, start, end int64) (size int64, blob io.ReadCloser, err error)
	// PushBlob pushes the specified blob
	PushBlob(repository, digest string, size int64, blob io.Reader) error
	// PushBlobChunk pushes the specified blob, but by chunked
	PushBlobChunk(repository, digest string, blobSize int64, chunk io.Reader, start, end int64, location string) (nextUploadLocation string, endRange int64, err error)
	// MountBlob mounts the blob from the source repository
	MountBlob(srcRepository, digest, dstRepository string) (err error)
	// DeleteBlob deletes the specified blob
	DeleteBlob(repository, digest string) (err error)
	// Copy the artifact from source repository to the destination. The "override"
	// is used to specify whether the destination artifact will be overridden if
	// its name is same with source but digest isn't
	Copy(srcRepository, srcReference, dstRepository, dstReference string, override bool) (err error)
	// Do send generic HTTP requests to the target registry service
	Do(req *http.Request) (*http.Response, error)
	// ListReferrers return all referrers
	ListReferrers(repository, ref string, rawQuery string) (*v1.Index, map[string][]string, error)
}

// NewClient creates a registry client with the default authorizer which determines the auth scheme
// of the registry automatically and calls the corresponding underlying authorizers(basic/bearer) to
// do the auth work. If a customized authorizer is needed, use "NewClientWithAuthorizer" instead
func NewClient(url, username, password string, insecure bool, interceptors ...interceptor.Interceptor) Client {
	authorizer := auth.NewAuthorizer(username, password, insecure)
	return NewClientWithAuthorizer(url, authorizer, insecure, "", interceptors...)
}

// NewClientWithCACert creates a registry client with custom CA certificate
func NewClientWithCACert(url, username, password string, insecure bool, caCert string, interceptors ...interceptor.Interceptor) Client {
	authorizer := auth.NewAuthorizer(username, password, insecure, caCert)
	return NewClientWithAuthorizer(url, authorizer, insecure, caCert, interceptors...)
}

// NewClientWithAuthorizer creates a registry client with the provided authorizer
func NewClientWithAuthorizer(url string, authorizer lib.Authorizer, insecure bool, caCert string, interceptors ...interceptor.Interceptor) Client {
	// When CACertificate is set, it takes precedence and Insecure is ignored
	transport := commonhttp.GetHTTPTransport(
		commonhttp.WithInsecure(insecure),
		commonhttp.WithCACert(caCert),
	)

	return &client{
		url:          url,
		authorizer:   authorizer,
		interceptors: interceptors,
		client: &http.Client{
			Transport: transport,
			Timeout:   registryHTTPClientTimeout,
		},
	}
}

type client struct {
	url          string
	authorizer   lib.Authorizer
	interceptors []interceptor.Interceptor
	client       *http.Client
}

func (c *client) Ping() error {
	req, err := http.NewRequest(http.MethodGet, buildPingURL(c.url), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) Catalog() ([]string, error) {
	var repositories []string
	url := buildCatalogURL(c.url)
	for {
		repos, next, err := c.catalog(url)
		if err != nil {
			return nil, err
		}
		repositories = append(repositories, repos...)

		url = next
		// no next page, end the loop
		if len(url) == 0 {
			break
		}
		// relative URL
		if !strings.Contains(url, "://") {
			url = c.url + url
		}
	}
	return repositories, nil
}

func (c *client) catalog(url string) ([]string, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	repositories := struct {
		Repositories []string `json:"repositories"`
	}{}
	if err := json.Unmarshal(body, &repositories); err != nil {
		return nil, "", err
	}
	return repositories.Repositories, next(resp.Header.Get("Link")), nil
}

func (c *client) ListTags(repository string) ([]string, error) {
	var tags []string
	url := buildTagListURL(c.url, repository)
	for {
		tgs, next, err := c.listTags(url)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tgs...)

		url = next
		// no next page, end the loop
		if len(url) == 0 {
			break
		}
		// relative URL
		if !strings.Contains(url, "://") {
			url = c.url + url
		}
	}
	return tags, nil
}

func (c *client) listTags(url string) ([]string, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	tgs := struct {
		Tags []string `json:"tags"`
	}{}
	if err := json.Unmarshal(body, &tgs); err != nil {
		return nil, "", err
	}
	return tgs.Tags, next(resp.Header.Get("Link")), nil
}

func (c *client) ManifestExist(repository, reference string) (bool, *distribution.Descriptor, error) {
	req, err := http.NewRequest(http.MethodHead, buildManifestURL(c.url, repository, reference), nil)
	if err != nil {
		return false, nil, err
	}
	for _, mediaType := range accepts {
		req.Header.Add("Accept", mediaType)
	}
	resp, err := c.do(req)
	if err != nil {
		if errors.IsErr(err, errors.NotFoundCode) {
			return false, nil, nil
		}
		return false, nil, err
	}
	defer resp.Body.Close()
	dig := resp.Header.Get("Docker-Content-Digest")
	contentType := resp.Header.Get("Content-Type")
	contentLen := resp.Header.Get("Content-Length")
	lenth, _ := strconv.Atoi(contentLen)
	return true, &distribution.Descriptor{Digest: digest.Digest(dig), MediaType: contentType, Size: int64(lenth)}, nil
}

func (c *client) PullManifest(repository, reference string, acceptedMediaTypes ...string) (
	distribution.Manifest, string, error) {
	req, err := http.NewRequest(http.MethodGet, buildManifestURL(c.url, repository, reference), nil)
	if err != nil {
		return nil, "", err
	}
	if len(acceptedMediaTypes) == 0 {
		acceptedMediaTypes = accepts
	}
	for _, mediaType := range acceptedMediaTypes {
		req.Header.Add("Accept", mediaType)
	}
	resp, err := c.do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	mediaType := resp.Header.Get("Content-Type")
	manifest, _, err := distribution.UnmarshalManifest(mediaType, payload)
	if err != nil {
		return nil, "", err
	}
	digest := resp.Header.Get("Docker-Content-Digest")
	return manifest, digest, nil
}

func (c *client) PushManifest(repository, reference, mediaType string, payload []byte) (string, error) {
	req, err := http.NewRequest(http.MethodPut, buildManifestURL(c.url, repository, reference),
		bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", mediaType)
	resp, err := c.do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return resp.Header.Get("Docker-Content-Digest"), nil
}

func (c *client) DeleteManifest(repository, reference string) error {
	_, err := digest.Parse(reference)
	if err != nil {
		// the reference is tag, get the digest first
		exist, desc, err := c.ManifestExist(repository, reference)
		if err != nil {
			return err
		}
		if !exist {
			return errors.New(nil).WithCode(errors.NotFoundCode).
				WithMessagef("%s:%s not found", repository, reference)
		}
		reference = string(desc.Digest)
	}
	req, err := http.NewRequest(http.MethodDelete, buildManifestURL(c.url, repository, reference), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) BlobExist(repository, digest string) (bool, error) {
	req, err := http.NewRequest(http.MethodHead, buildBlobURL(c.url, repository, digest), nil)
	if err != nil {
		return false, err
	}
	resp, err := c.do(req)
	if err != nil {
		if errors.IsErr(err, errors.NotFoundCode) {
			return false, nil
		}
		return false, err
	}
	defer resp.Body.Close()
	return true, nil
}

func (c *client) PullBlob(repository, digest string) (int64, io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, buildBlobURL(c.url, repository, digest), nil)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Add("Accept-Encoding", "identity")
	resp, err := c.do(req)
	if err != nil {
		return 0, nil, err
	}

	var size int64
	n := resp.Header.Get("Content-Length")
	// no content-length is acceptable, which can taken from manifests
	if len(n) > 0 {
		size, err = strconv.ParseInt(n, 10, 64)
		if err != nil {
			defer resp.Body.Close()
			return 0, nil, err
		}
	}

	return size, resp.Body, nil
}

// PullBlobChunk pulls the specified blob, but by chunked, refer to https://github.com/opencontainers/distribution-spec/blob/main/spec.md#pull for more details.
func (c *client) PullBlobChunk(repository, digest string, _ int64, start, end int64) (int64, io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, buildBlobURL(c.url, repository, digest), nil)
	if err != nil {
		return 0, nil, err
	}

	req.Header.Add("Accept-Encoding", "identity")
	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := c.do(req)
	if err != nil {
		return 0, nil, err
	}

	var size int64
	n := resp.Header.Get("Content-Length")
	// no content-length is acceptable, which can taken from manifests
	if len(n) > 0 {
		size, err = strconv.ParseInt(n, 10, 64)
		if err != nil {
			defer resp.Body.Close()
			return 0, nil, err
		}
	}

	return size, resp.Body, nil
}

func (c *client) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	location, _, err := c.initiateBlobUpload(repository)
	if err != nil {
		return err
	}
	return c.monolithicBlobUpload(location, digest, size, blob)
}

// PushBlobChunk pushes the specified blob, but by chunked, refer to https://github.com/opencontainers/distribution-spec/blob/main/spec.md#push for more details.
func (c *client) PushBlobChunk(repository, digest string, blobSize int64, chunk io.Reader, start, end int64, location string) (string, int64, error) {
	var err error
	// first chunk need to initialize blob upload location
	if start == 0 {
		location, _, err = c.initiateBlobUpload(repository)
		if err != nil {
			return location, end, err
		}
	}

	// the range is from 0 to (blobSize-1), so (end == blobSize-1) means it is last chunk
	lastChunk := end == blobSize-1
	url, err := buildChunkBlobUploadURL(c.url, location, digest, lastChunk)
	if err != nil {
		return location, end, err
	}

	// use PUT instead of PATCH for last chunk which can reduce a final request
	method := http.MethodPatch
	if lastChunk {
		method = http.MethodPut
	}
	req, err := http.NewRequest(method, url, chunk)
	if err != nil {
		return location, end, err
	}

	req.Header.Set("Content-Length", fmt.Sprintf("%d", end-start+1))
	req.Header.Set("Content-Range", fmt.Sprintf("%d-%d", start, end))
	resp, err := c.do(req)
	if err != nil {
		// if push chunk error, we should query the upload progress for new location and end range.
		newLocation, newEnd, err1 := c.getUploadStatus(url)
		if err1 == nil {
			return newLocation, newEnd, err
		}
		// end should return start-1 to re-push this chunk
		return location, start - 1, fmt.Errorf("failed to get upload status: %w", err1)
	}

	defer resp.Body.Close()
	// return the location for next chunk upload
	return resp.Header.Get("Location"), end, nil
}

func (c *client) getUploadStatus(location string) (string, int64, error) {
	req, err := http.NewRequest(http.MethodGet, location, nil)
	if err != nil {
		return location, -1, err
	}

	resp, err := c.do(req)
	if err != nil {
		return location, -1, err
	}

	defer resp.Body.Close()

	_, end, err := parseContentRange(resp.Header.Get("Range"))
	if err != nil {
		return location, -1, err
	}

	return resp.Header.Get("Location"), end, nil
}

func parseContentRange(cr string) (int64, int64, error) {
	ranges := strings.Split(cr, "-")
	if len(ranges) != 2 {
		return -1, -1, fmt.Errorf("invalid content range format, %s", cr)
	}
	start, err := strconv.ParseInt(ranges[0], 10, 64)
	if err != nil {
		return -1, -1, err
	}
	end, err := strconv.ParseInt(ranges[1], 10, 64)
	if err != nil {
		return -1, -1, err
	}

	return start, end, nil
}

func (c *client) initiateBlobUpload(repository string) (string, string, error) {
	req, err := http.NewRequest(http.MethodPost, buildInitiateBlobUploadURL(c.url, repository), nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Length", "0")
	resp, err := c.do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	return resp.Header.Get("Location"), resp.Header.Get("Docker-Upload-UUID"), nil
}

func (c *client) monolithicBlobUpload(location, digest string, size int64, data io.Reader) error {
	url, err := buildMonolithicBlobUploadURL(c.url, location, digest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, url, data)
	if err != nil {
		return err
	}
	req.ContentLength = size
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) MountBlob(srcRepository, digest, dstRepository string) error {
	req, err := http.NewRequest(http.MethodPost, buildMountBlobURL(c.url, dstRepository, digest, srcRepository), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Length", "0")
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) DeleteBlob(repository, digest string) error {
	req, err := http.NewRequest(http.MethodDelete, buildBlobURL(c.url, repository, digest), nil)
	if err != nil {
		return err
	}
	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *client) Copy(srcRepo, srcRef, dstRepo, dstRef string, override bool) error {
	// pull the manifest from the source repository
	manifest, srcDgt, err := c.PullManifest(srcRepo, srcRef)
	if err != nil {
		return err
	}

	// check the existence of the artifact on the destination repository
	exist, desc, err := c.ManifestExist(dstRepo, dstRef)
	if err != nil {
		return err
	}
	if exist {
		// the same artifact already exists
		if desc != nil && srcDgt == string(desc.Digest) {
			return nil
		}
		// the same name artifact exists, but not allowed to override
		if !override {
			return errors.New(nil).WithCode(errors.PreconditionCode).
				WithMessage("the same name but different digest artifact exists, but the override is set to false")
		}
	}

	for _, descriptor := range manifest.References() {
		digest := descriptor.Digest.String()
		switch descriptor.MediaType {
		// skip foreign layer
		case schema2.MediaTypeForeignLayer:
			continue
		// manifest or index
		case v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList,
			v1.MediaTypeImageManifest, schema2.MediaTypeManifest,
			schema1.MediaTypeSignedManifest, schema1.MediaTypeManifest:
			if err = c.Copy(srcRepo, digest, dstRepo, digest, false); err != nil {
				return err
			}
		// common layer
		default:
			exist, err := c.BlobExist(dstRepo, digest)
			if err != nil {
				return err
			}
			// the layer already exist, skip
			if exist {
				continue
			}
			// when the copy happens inside the same registry, use mount
			if err = c.MountBlob(srcRepo, digest, dstRepo); err != nil {
				return err
			}
		}
	}

	mediaType, payload, err := manifest.Payload()
	if err != nil {
		return err
	}
	// push manifest to the destination repository
	if _, err = c.PushManifest(dstRepo, dstRef, mediaType, payload); err != nil {
		return err
	}

	return nil
}

func (c *client) Do(req *http.Request) (*http.Response, error) {
	return c.do(req)
}

func (c *client) do(req *http.Request) (*http.Response, error) {
	for _, interceptor := range c.interceptors {
		if err := interceptor.Intercept(req); err != nil {
			return nil, err
		}
	}
	if c.authorizer != nil {
		if err := c.authorizer.Modify(req); err != nil {
			return nil, err
		}
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		code := errors.GeneralCode
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			code = errors.UnAuthorizedCode
		case http.StatusForbidden:
			code = errors.ForbiddenCode
		case http.StatusNotFound:
			code = errors.NotFoundCode
		case http.StatusTooManyRequests:
			code = errors.RateLimitCode
		}
		return nil, errors.New(nil).WithCode(code).
			WithMessagef("http status code: %d, body: %s", resp.StatusCode, string(body))
	}
	return resp, nil
}

func (c *client) ListReferrers(repository, ref string, rawQuery string) (*v1.Index, map[string][]string, error) {
	remoteURL := buildReferrersURL(c.url, repository, ref, rawQuery)
	req, err := http.NewRequest(http.MethodGet, remoteURL, nil)
	if err != nil {
		return nil, nil, err
	}
	log.Debugf("upstream url %v", remoteURL)

	if c.authorizer == nil {
		log.Debug("registry client authorizer is nil")
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// HTTP Status check
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, nil, errors.New(nil).WithCode(errors.NotFoundCode).
				WithMessagef("referrers for %s:%s not found", repository, ref)
		}
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("upstream returned status %d: %s", resp.StatusCode, string(body))
	}

	// Decode JSON into ocispec.Index
	var index v1.Index
	decoder := json.NewDecoder(resp.Body)
	// copy the header to headerMap
	headerMap := make(map[string][]string)
	for k, v := range resp.Header {
		headerMap[k] = v
	}
	log.Debugf("headerMap from upstream %v", headerMap)
	if err := decoder.Decode(&index); err != nil {
		return nil, nil, err
	}

	return &index, headerMap, nil
}

// parse the next page link from the link header
func next(link string) string {
	links := lib.ParseLinks(link)
	for _, lk := range links {
		if lk.Rel == "next" {
			return lk.URL
		}
	}
	return ""
}

func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}

func buildCatalogURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/_catalog?n=1000", endpoint)
}

func buildTagListURL(endpoint, repository string) string {
	return fmt.Sprintf("%s/v2/%s/tags/list", endpoint, repository)
}

func buildManifestURL(endpoint, repository, reference string) string {
	return fmt.Sprintf("%s/v2/%s/manifests/%s", endpoint, repository, reference)
}

func buildBlobURL(endpoint, repository, reference string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/%s", endpoint, repository, reference)
}

func buildMountBlobURL(endpoint, repository, digest, from string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/uploads/?mount=%s&from=%s", endpoint, repository, digest, from)
}

func buildInitiateBlobUploadURL(endpoint, repository string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/uploads/", endpoint, repository)
}

func buildReferrersURL(endpoint, repository, ref, rawQuery string) string {
	url := fmt.Sprintf("%s/v2/%s/referrers/%s", endpoint, repository, ref)
	if len(rawQuery) > 0 {
		url = url + "?" + rawQuery
	}
	return url
}

func buildChunkBlobUploadURL(endpoint, location, digest string, lastChunk bool) (string, error) {
	url, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	q := url.Query()
	if lastChunk {
		q.Set("digest", digest)
	}
	url.RawQuery = q.Encode()
	if url.IsAbs() {
		return url.String(), nil
	}
	// the "relativeurls" is enabled in registry
	return endpoint + url.String(), nil
}

func buildMonolithicBlobUploadURL(endpoint, location, digest string) (string, error) {
	url, err := url.Parse(location)
	if err != nil {
		return "", err
	}
	q := url.Query()
	q.Set("digest", digest)
	url.RawQuery = q.Encode()
	if url.IsAbs() {
		return url.String(), nil
	}
	// the "relativeurls" is enabled in registry
	return endpoint + url.String(), nil
}
