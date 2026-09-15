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

package harborsatellite

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/docker/distribution"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	regadapter "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var (
	_ regadapter.Adapter          = (*adapter)(nil)
	_ regadapter.ArtifactRegistry = (*adapter)(nil)
)
var ErrNotImplemented = errors.New("not implemented")

type Result struct {
	Group     string     `json:"group"`
	Registry  string     `json:"registry"`
	Artifacts []Artifact `json:"artifacts"`
}

type Artifact struct {
	Repository string   `json:"repository"`
	Tags       []string `json:"tag"`
	Labels     []string `json:"labels"`
	Type       string   `json:"type"`
	Digest     string   `json:"digest"`
	Deleted    bool     `json:"deleted"`
}

func init() {
	err := regadapter.RegisterFactory(model.RegistryHarborSatellite, &factory{})
	if err != nil {
		return
	}
}

type factory struct{}

// Create ...
func (f *factory) Create(r *model.Registry) (regadapter.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

type adapter struct {
	httpClient *http.Client
}

func (a adapter) RoundTrip(request *http.Request) (*http.Response, error) {
	u, err := url.Parse(config.InternalCoreURL())
	if err != nil {
		return nil, fmt.Errorf("unable to parse internal core url: %v", err)
	}

	// replace request's host with core's address
	request.Host = config.InternalCoreURL()
	request.URL.Host = u.Host

	request.URL.Scheme = u.Scheme
	// adds auth headers
	_ = secret.AddToRequest(request, config.JobserviceSecret())

	return a.httpClient.Do(request)
}

func (a adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{}, nil
}

func (a adapter) PrepareForPush(resources []*model.Resource) error {
	var (
		artifacts      []Artifact
		registry       *model.Registry
		destinationURL string
		groupName      string
	)

	for _, r := range resources {
		if r.Metadata == nil {
			continue
		}
		if r.Metadata.Repository == nil {
			continue
		}
		if r.Registry == nil {
			continue
		}
		if r.ExtendedInfo == nil {
			return fmt.Errorf("extended_info map is nil")
		}

		if registry == nil {
			registry = r.Registry
		}
		if destinationURL == "" {
			destURL, ok := r.ExtendedInfo["destinationURL"].(string)
			if ok {
				destinationURL = destURL
			} else {
				return fmt.Errorf("destination_url not a string or missing")
			}
		}

		if groupName == "" {
			grp, ok := r.ExtendedInfo["groupName"].(string)
			if ok {
				groupName = grp
			} else {
				return fmt.Errorf("groupName not a string or missing")
			}
		}

		for _, at := range r.Metadata.Artifacts {
			artifacts = append(artifacts, Artifact{
				Repository: r.Metadata.Repository.Name,
				Deleted:    r.Deleted,
				Tags:       at.Tags,
				Labels:     at.Labels,
				Type:       at.Type,
				Digest:     at.Digest,
			})
		}
	}

	if registry == nil {
		return fmt.Errorf("no registry information found")
	}

	result := &Result{
		Group:     groupName,
		Registry:  registry.URL,
		Artifacts: artifacts,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return errors.Wrap(err, "failed to marshal result")
	}

	// Create a POST request
	req, err := http.NewRequest("POST", destinationURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set the content type header
	req.Header.Set("Content-Type", "application/json")

	// Send the request using http.Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

func (a adapter) HealthCheck() (string, error) {
	return model.Healthy, nil
}

func (a adapter) FetchArtifacts(_ []*model.Filter) ([]*model.Resource, error) {
	return nil, nil
}

func (a adapter) ManifestExist(_, _ string) (exist bool, desc *distribution.Descriptor, err error) {
	return true, nil, nil
}

func (a adapter) PullManifest(_, _ string, _ ...string) (manifest distribution.Manifest, digest string, err error) {
	return nil, "", ErrNotImplemented
}

func (a adapter) PushManifest(_, _, _ string, _ []byte) (string, error) {
	return "", nil
}

func (a adapter) DeleteManifest(_, _ string) error {
	return ErrNotImplemented
}

func (a adapter) BlobExist(_, _ string) (exist bool, err error) {
	return true, nil
}

func (a adapter) PullBlob(_, _ string) (size int64, blob io.ReadCloser, err error) {
	return 0, nil, ErrNotImplemented
}

func (a adapter) PullBlobChunk(_, _ string, _, _, _ int64) (size int64, blob io.ReadCloser, err error) {
	return 0, nil, ErrNotImplemented
}

func (a adapter) PushBlobChunk(_, _ string, _ int64, _ io.Reader, _, _ int64, _ string) (nextUploadLocation string, endRange int64, err error) {
	return "", 0, ErrNotImplemented
}

func (a adapter) PushBlob(_, _ string, _ int64, _ io.Reader) error {
	return nil
}

func (a adapter) MountBlob(_, _, _ string) (err error) {
	return nil
}

func (a adapter) CanBeMount(_ string) (mount bool, repository string, err error) {
	return false, "", ErrNotImplemented
}

func (a adapter) DeleteTag(_, _ string) error {
	return ErrNotImplemented
}

func (a adapter) ListTags(_ string) (tags []string, err error) {
	return nil, nil
}

func newAdapter(_ *model.Registry) (regadapter.Adapter, error) {
	return &adapter{
		httpClient: &http.Client{},
	}, nil
}
