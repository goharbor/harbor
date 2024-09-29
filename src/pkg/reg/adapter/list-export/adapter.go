package list_export

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	regadapter "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/google/go-containerregistry/pkg/crane"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"
)

var (
	_ regadapter.Adapter          = (*adapter)(nil)
	_ regadapter.ArtifactRegistry = (*adapter)(nil)
)
var ErrNotImplemented = errors.New("not implemented")

type Result struct {
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
	err := regadapter.RegisterFactory(model.RegistryArtifactListExport, &factory{})
	if err != nil {
		return
	}
}

type factory struct {
}

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
		artifacts       []Artifact
		registry        *model.Registry
		destinationRepo string
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
			if destinationRepo == "" {
				destinationRepo = r.Metadata.Repository.Name
			}
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
		Registry:  registry.Name,
		Artifacts: artifacts,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return errors.Wrap(err, "failed to marshal result")
	}

	img, err := crane.Image(map[string][]byte{
		"artifacts.json": data,
	})
	if err != nil {
		return fmt.Errorf("image create failed: %v", err)
	}

	destinationRepo = fmt.Sprintf("%s/%s", path.Dir(destinationRepo), "state")
	//

	err = crane.Push(img, destinationRepo, crane.WithTransport(a))
	if err != nil {
		return fmt.Errorf("push image failed: %v", err)
	}

	err = crane.Tag(destinationRepo, fmt.Sprintf("%d", time.Now().Unix()), crane.WithTransport(a))
	if err != nil {
		return fmt.Errorf("tag image failed: %v", err)
	}

	err = crane.Tag(destinationRepo, "latest", crane.WithTransport(a))
	if err != nil {
		return fmt.Errorf("tag image failed: %v", err)
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

func newAdapter(_ *model.Registry) (regadapter.Adapter, error) {

	return &adapter{
		httpClient: &http.Client{},
	}, nil
}
