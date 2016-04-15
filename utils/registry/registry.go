/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
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

package registry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/docker/distribution/manifest/schema2"
)

// Registry holds information of a registry entiry
type Registry struct {
	Endpoint *url.URL
	client   *http.Client
	ub       *uRLBuilder
}

type uRLBuilder struct {
	root *url.URL
}

// New returns an instance of Registry
func New(endpoint string, client *http.Client) (*Registry, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	return &Registry{
		Endpoint: u,
		client:   client,
		ub: &uRLBuilder{
			root: u,
		},
	}, nil
}

// ListTag ...
func (r *Registry) ListTag(name string) ([]string, error) {
	tags := []string{}
	req, err := http.NewRequest("GET", r.ub.buildTagListURL(name), nil)
	if err != nil {
		return tags, err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return tags, err
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

		tags = tagsResp.Tags

		return tags, nil
	}

	return tags, Error{
		StatusCode: resp.StatusCode,
		Message:    string(b),
	}

}

// ManifestExist ...
func (r *Registry) ManifestExist(name, reference string) (digest string, exist bool, err error) {
	req, err := http.NewRequest("HEAD", r.ub.buildManifestURL(name, reference), nil)
	if err != nil {
		return
	}

	// request Schema 2 manifest, if the registry does not support it,
	// Schema 1 manifest will be returned
	req.Header.Set(http.CanonicalHeaderKey("Accept"), schema2.MediaTypeManifest)

	resp, err := r.client.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusOK {
		exist = true
		digest = resp.Header.Get(http.CanonicalHeaderKey("Docker-Content-Digest"))
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		return
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	err = Error{
		StatusCode: resp.StatusCode,
		Message:    string(b),
	}
	return
}

// PullManifest ...
func (r *Registry) PullManifest(name, reference string) (digest, mediaType string, payload []byte, err error) {
	req, err := http.NewRequest("GET", r.ub.buildManifestURL(name, reference), nil)
	if err != nil {
		return
	}

	// request Schema 2 manifest, if the registry does not support it,
	// Schema 1 manifest will be returned
	req.Header.Set(http.CanonicalHeaderKey("Accept"), schema2.MediaTypeManifest)

	resp, err := r.client.Do(req)
	if err != nil {
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

	err = Error{
		StatusCode: resp.StatusCode,
		Message:    string(b),
	}

	return
}

// DeleteManifest ...
func (r *Registry) DeleteManifest(name, digest string) error {
	req, err := http.NewRequest("DELETE", r.ub.buildManifestURL(name, digest), nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusAccepted {
		return nil
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return Error{
		StatusCode: resp.StatusCode,
		Message:    string(b),
	}
}

// DeleteTag ...
func (r *Registry) DeleteTag(name, tag string) error {
	digest, exist, err := r.ManifestExist(name, tag)
	if err != nil {
		return err
	}

	if !exist {
		return Error{
			StatusCode: http.StatusNotFound,
		}
	}

	return r.DeleteManifest(name, digest)
}

// DeleteBlob ...
func (r *Registry) DeleteBlob(name, digest string) error {
	req, err := http.NewRequest("DELETE", r.ub.buildBlobURL(name, digest), nil)
	if err != nil {
		return err
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusAccepted {
		return nil
	}

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return Error{
		StatusCode: resp.StatusCode,
		Message:    string(b),
	}
}

func (u *uRLBuilder) buildTagListURL(name string) string {
	return fmt.Sprintf("%s/v2/%s/tags/list", u.root.String(), name)
}

func (u *uRLBuilder) buildManifestURL(name, reference string) string {
	return fmt.Sprintf("%s/v2/%s/manifests/%s", u.root.String(), name, reference)
}

func (u *uRLBuilder) buildBlobURL(name, reference string) string {
	return fmt.Sprintf("%s/v2/%s/blobs/%s", u.root.String(), name, reference)
}
