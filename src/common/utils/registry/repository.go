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
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils"
)

// Repository holds information of a repository entity
type Repository struct {
	Name     string
	Endpoint *url.URL
	client   *http.Client
}

// NewRepository returns an instance of Repository
func NewRepository(name, endpoint string, client *http.Client) (*Repository, error) {
	name = strings.TrimSpace(name)

	u, err := utils.ParseEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	repository := &Repository{
		Name:     name,
		Endpoint: u,
		client:   client,
	}

	return repository, nil
}

func parseError(err error) error {
	if urlErr, ok := err.(*url.Error); ok {
		if regErr, ok := urlErr.Err.(*commonhttp.Error); ok {
			return regErr
		}
	}
	return err
}

// ListTag ...
func (r *Repository) ListTag() ([]string, error) {
	tags := []string{}
	aurl := buildTagListURL(r.Endpoint.String(), r.Name)

	for len(aurl) > 0 {
		req, err := http.NewRequest("GET", aurl, nil)
		if err != nil {
			return tags, err
		}
		resp, err := r.client.Do(req)
		if err != nil {
			return nil, parseError(err)
		}

		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return tags, err
		}

		if resp.StatusCode == http.StatusOK {
			tagsResp := struct {
				Tags []string `json:"tags"`
			}{}

			if err := json.Unmarshal(b, &tagsResp); err != nil {
				return tags, err
			}

			tags = append(tags, tagsResp.Tags...)
			// Link: </v2/_catalog?last=library%2Fhello-world-25&n=100>; rel="next"
			// Link: <http://domain.com/v2/_catalog?last=library%2Fhello-world-25&n=100>; rel="next"
			link := resp.Header.Get("Link")
			if strings.HasSuffix(link, `rel="next"`) && strings.Index(link, "<") >= 0 && strings.Index(link, ">") >= 0 {
				aurl = link[strings.Index(link, "<")+1 : strings.Index(link, ">")]
				if strings.Index(aurl, ":") < 0 {
					aurl = r.Endpoint.String() + aurl
				}
			} else {
				aurl = ""
			}
		} else if resp.StatusCode == http.StatusNotFound {

			// TODO remove the logic if the bug of registry is fixed
			// It's a workaround for a bug of registry: when listing tags of
			// a repository which is being pushed, a "NAME_UNKNOWN" error will
			// been returned, while the catalog API can list this repository.
			return tags, nil
		} else {
			return tags, &commonhttp.Error{
				Code:    resp.StatusCode,
				Message: string(b),
			}
		}
	}
	tags = sort.Strings(tags)
	return tags, nil
}

// ManifestExist ...
func (r *Repository) ManifestExist(reference string) (digest string, exist bool, err error) {
	req, err := http.NewRequest("HEAD", buildManifestURL(r.Endpoint.String(), r.Name, reference), nil)
	if err != nil {
		return
	}

	req.Header.Add(http.CanonicalHeaderKey("Accept"), schema1.MediaTypeManifest)
	req.Header.Add(http.CanonicalHeaderKey("Accept"), schema2.MediaTypeManifest)

	resp, err := r.client.Do(req)
	if err != nil {
		err = parseError(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		exist = true
		digest = resp.Header.Get(http.CanonicalHeaderKey("Docker-Content-Digest"))
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
	return
}

// PullManifest ...
func (r *Repository) PullManifest(reference string, acceptMediaTypes []string) (digest, mediaType string, payload []byte, err error) {
	req, err := http.NewRequest("GET", buildManifestURL(r.Endpoint.String(), r.Name, reference), nil)
	if err != nil {
		return
	}

	for _, mediaType := range acceptMediaTypes {
		req.Header.Add(http.CanonicalHeaderKey("Accept"), mediaType)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		err = parseError(err)
		return
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusOK {
		digest = resp.Header.Get(http.CanonicalHeaderKey("Docker-Content-Digest"))
		mediaType = resp.Header.Get(http.CanonicalHeaderKey("Content-Type"))
		payload = b
		return
	}

	err = &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}

	return
}

// PushManifest ...
func (r *Repository) PushManifest(reference, mediaType string, payload []byte) (digest string, err error) {
	req, err := http.NewRequest("PUT", buildManifestURL(r.Endpoint.String(), r.Name, reference),
		bytes.NewReader(payload))
	if err != nil {
		return
	}
	req.Header.Set(http.CanonicalHeaderKey("Content-Type"), mediaType)

	resp, err := r.client.Do(req)
	if err != nil {
		err = parseError(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		digest = resp.Header.Get(http.CanonicalHeaderKey("Docker-Content-Digest"))
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}

	return
}

// DeleteManifest ...
func (r *Repository) DeleteManifest(digest string) error {
	req, err := http.NewRequest("DELETE", buildManifestURL(r.Endpoint.String(), r.Name, digest), nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return parseError(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}

// MountBlob ...
func (r *Repository) MountBlob(digest, from string) error {
	req, err := http.NewRequest("POST", buildMountBlobURL(r.Endpoint.String(), r.Name, digest, from), nil)
	req.Header.Set(http.CanonicalHeaderKey("Content-Length"), "0")

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		defer resp.Body.Close()
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(b),
		}
	}

	return nil
}

// DeleteTag ...
func (r *Repository) DeleteTag(tag string) error {
	digest, exist, err := r.ManifestExist(tag)
	if err != nil {
		return err
	}

	if !exist {
		return &commonhttp.Error{
			Code: http.StatusNotFound,
		}
	}

	return r.DeleteManifest(digest)
}

// BlobExist ...
func (r *Repository) BlobExist(digest string) (bool, error) {
	req, err := http.NewRequest("HEAD", buildBlobURL(r.Endpoint.String(), r.Name, digest), nil)
	if err != nil {
		return false, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return false, parseError(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return false, &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}

// PullBlob : client must close data if it is not nil
func (r *Repository) PullBlob(digest string) (size int64, data io.ReadCloser, err error) {
	req, err := http.NewRequest("GET", buildBlobURL(r.Endpoint.String(), r.Name, digest), nil)
	if err != nil {
		return
	}

	resp, err := r.client.Do(req)
	if err != nil {
		err = parseError(err)
		return
	}

	if resp.StatusCode == http.StatusOK {
		contengLength := resp.Header.Get(http.CanonicalHeaderKey("Content-Length"))
		size, err = strconv.ParseInt(contengLength, 10, 64)
		if err != nil {
			return
		}
		data = resp.Body
		return
	}
	// can not close the connect if the status code is 200
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}

	return
}

func (r *Repository) initiateBlobUpload(name string) (location, uploadUUID string, err error) {
	req, err := http.NewRequest("POST", buildInitiateBlobUploadURL(r.Endpoint.String(), r.Name), nil)
	req.Header.Set(http.CanonicalHeaderKey("Content-Length"), "0")

	resp, err := r.client.Do(req)
	if err != nil {
		err = parseError(err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		location = resp.Header.Get(http.CanonicalHeaderKey("Location"))
		uploadUUID = resp.Header.Get(http.CanonicalHeaderKey("Docker-Upload-UUID"))
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}

	return
}

func (r *Repository) monolithicBlobUpload(location, digest string, size int64, data io.Reader) error {
	url, err := buildMonolithicBlobUploadURL(r.Endpoint.String(), location, digest)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("PUT", url, data)
	if err != nil {
		return err
	}
	req.ContentLength = size

	resp, err := r.client.Do(req)
	if err != nil {
		return parseError(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}

// PushBlob ...
func (r *Repository) PushBlob(digest string, size int64, data io.Reader) error {
	location, _, err := r.initiateBlobUpload(r.Name)
	if err != nil {
		return err
	}
	return r.monolithicBlobUpload(location, digest, size, data)
}

// DeleteBlob ...
func (r *Repository) DeleteBlob(digest string) error {
	req, err := http.NewRequest("DELETE", buildBlobURL(r.Endpoint.String(), r.Name, digest), nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return parseError(err)
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return &commonhttp.Error{
		Code:    resp.StatusCode,
		Message: string(b),
	}
}

func buildPingURL(endpoint string) string {
	return fmt.Sprintf("%s/v2/", endpoint)
}

func buildTagListURL(endpoint, repoName string) string {
	return fmt.Sprintf("%s/v2/%s/tags/list", endpoint, repoName)
}

func buildManifestURL(endpoint, repoName, reference string) string {
	return fmt.Sprintf("%s/v2/%s/manifests/%s", endpoint, repoName, reference)
}

func buildBlobURL(endpoint, repoName, reference string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/%s", endpoint, repoName, reference)
}

func buildMountBlobURL(endpoint, repoName, digest, from string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/uploads/?mount=%s&from=%s", endpoint, repoName, digest, from)
}

func buildInitiateBlobUploadURL(endpoint, repoName string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/uploads/", endpoint, repoName)
}

func buildMonolithicBlobUploadURL(endpoint, location, digest string) (string, error) {
	relative, err := isRelativeURL(location)
	if err != nil {
		return "", err
	}
	// when the registry enables "relativeurls", the location returned
	// has no scheme and host part
	if relative {
		location = endpoint + location
	}
	query := ""
	if strings.ContainsRune(location, '?') {
		query = "&"
	} else {
		query = "?"
	}
	query += fmt.Sprintf("digest=%s", digest)
	return fmt.Sprintf("%s%s", location, query), nil
}

func isRelativeURL(endpoint string) (bool, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return false, err
	}
	return !u.IsAbs(), nil
}
