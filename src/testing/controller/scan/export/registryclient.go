package export

import (
	"github.com/docker/distribution"
	"github.com/stretchr/testify/mock"
	"io"
	"net/http"
)

type RegistryClient struct {
	mock.Mock
}

func (_m *RegistryClient) Ping() (err error) {
	panic("implement me")
}

func (_m *RegistryClient) Catalog() (repositories []string, err error) {
	panic("implement me")
}

func (_m *RegistryClient) ListTags(repository string) (tags []string, err error) {
	panic("implement me")
}

func (_m *RegistryClient) ManifestExist(repository, reference string) (exist bool, desc *distribution.Descriptor, err error) {
	panic("implement me")
}

func (_m *RegistryClient) PullManifest(repository, reference string, acceptedMediaTypes ...string) (manifest distribution.Manifest, digest string, err error) {
	panic("implement me")
}

func (_m *RegistryClient) PushManifest(repository, reference, mediaType string, payload []byte) (digest string, err error) {
	panic("implement me")
}

func (_m *RegistryClient) DeleteManifest(repository, reference string) (err error) {
	panic("implement me")
}

func (_m *RegistryClient) BlobExist(repository, digest string) (exist bool, err error) {
	panic("implement me")
}

func (_m *RegistryClient) PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error) {
	ret := _m.Called(repository, digest)
	return ret.Get(0).(int64), ret.Get(1).(io.ReadCloser), ret.Error(2)
}

func (_m *RegistryClient) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	ret := _m.Called(repository, digest, size, blob)
	return ret.Error(0)
}

func (_m *RegistryClient) MountBlob(srcRepository, digest, dstRepository string) (err error) {
	panic("implement me")
}

func (_m *RegistryClient) DeleteBlob(repository, digest string) (err error) {
	ret := _m.Called(repository, digest)
	return ret.Error(0)
}

func (_m *RegistryClient) Copy(srcRepository, srcReference, dstRepository, dstReference string, override bool) (err error) {
	panic("implement me")
}

func (_m *RegistryClient) Do(req *http.Request) (*http.Response, error) {
	panic("implement me")
}
