package list_export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/lib/errors"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	regadapter "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"io"
	"net/http"
)

var (
	_ adp.Adapter          = (*adapter)(nil)
	_ adp.ArtifactRegistry = (*adapter)(nil)
)
var ErrNotImplemented = errors.New("not implemented")

type Result struct {
	Registry  string     `json:"registry"`
	Artifacts []Artifact `json:"artifacts"`
}

type Artifact struct {
	Repository string   `json:"repository"`
	Tags       []string `json:"tag"`
	Type       string   `json:"type"`
	Digest     string   `json:"digest"`
	Deleted    bool     `json:"deleted"`
}

func init() {
	err := regadapter.RegisterFactory(model.RegistryArtifactListExport, &factory{})
	if err != nil {
		return
	}
}

type factory struct {
}

// Create ...
func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	return newAdapter(r)
}

// AdapterPattern ...
func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}

type adapter struct {
}

func (a adapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{}, nil
}

func (a adapter) PrepareForPush(resources []*model.Resource) error {

	var (
		artifacts []Artifact
		registry  *model.Registry
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

		if registry == nil {
			registry = r.Registry
		}

		for _, at := range r.Metadata.Artifacts {

			artifacts = append(artifacts, Artifact{
				Repository: r.Metadata.Repository.Name,
				Deleted:    r.Deleted,
				Tags:       at.Tags,
				Type:       at.Type,
				Digest:     at.Digest,
			})
		}
	}

	if registry == nil {
		return fmt.Errorf("no registry information found")
	}

	result := &Result{
		Registry:  registry.Name,
		Artifacts: artifacts,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return errors.Wrap(err, "failed to marshal result")
	}

	responseBody := bytes.NewBuffer(data)
	resp, err := http.Post(registry.URL, "application/json", responseBody)
	if err != nil {
		return errors.Wrap(err, "failed to post result")
	}
	defer resp.Body.Close()

	return nil
}

func (a adapter) HealthCheck() (string, error) {
	return model.Healthy, nil
}

func (a adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	return nil, nil
}

func (a adapter) ManifestExist(repository, reference string) (exist bool, desc *distribution.Descriptor, err error) {
	return true, nil, nil
}

func (a adapter) PullManifest(repository, reference string, accepttedMediaTypes ...string) (manifest distribution.Manifest, digest string, err error) {
	return nil, "", ErrNotImplemented
}

func (a adapter) PushManifest(repository, reference, mediaType string, payload []byte) (string, error) {
	//fmt.Println("push manifest", repository, reference)
	return "", nil
}

func (a adapter) DeleteManifest(repository, reference string) error {
	return ErrNotImplemented
}

func (a adapter) BlobExist(repository, digest string) (exist bool, err error) {
	return true, nil
}

func (a adapter) PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error) {
	return 0, nil, ErrNotImplemented
}

func (a adapter) PullBlobChunk(repository, digest string, blobSize, start, end int64) (size int64, blob io.ReadCloser, err error) {
	return 0, nil, ErrNotImplemented
}

func (a adapter) PushBlobChunk(repository, digest string, size int64, chunk io.Reader, start, end int64, location string) (nextUploadLocation string, endRange int64, err error) {
	return "", 0, ErrNotImplemented
}

func (a adapter) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	return nil
}

func (a adapter) MountBlob(srcRepository, digest, dstRepository string) (err error) {
	return nil
}

func (a adapter) CanBeMount(digest string) (mount bool, repository string, err error) {
	return false, "", ErrNotImplemented
}

func (a adapter) DeleteTag(repository, tag string) error {
	return ErrNotImplemented
}

func (a adapter) ListTags(repository string) (tags []string, err error) {
	return nil, nil
}

func newAdapter(registry *model.Registry) (adp.Adapter, error) {
	return &adapter{}, nil
}
