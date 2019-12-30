package ocidriver

import (
	"fmt"
	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"io"
	"sort"
)

// OciRepository ...
type OciRepository struct {
	Registry *OciRegistry

	Name string
}

// List ...
func (rp *OciRepository) List( /*filter*/ ) ([]*OciImage, error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return nil, errors.New("bad repository " + rp.Name)
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	tags, err := repo.Tags(nil).All(nil)

	sort.Strings(tags)
	var ts []*OciImage
	for _, tag := range tags {
		ts = append(ts, &OciImage{
			Registry:   rp.Registry,
			repository: rp,
			Name:       tag,
		})
	}
	return ts, nil
}

// GetBlob ...
func (rp *OciRepository) GetBlob(d digest.Digest) (size int64, data io.ReadCloser, err error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return 0, nil, err
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	if err != nil {
		return 0, nil, err
	}
	b, err := repo.Blobs(nil).Open(nil, d)
	if err != nil {
		return 0, nil, err
	}
	size, err = b.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, nil, err
	}
	_, err = b.Seek(0, io.SeekStart)
	return size, b, err
}

// GetImageByTag ...
func (rp *OciRepository) GetImageByTag(tag string) (isList bool, manifest *OciManifest, manifestList *OciManifestIndex, manifestv1 *OciManifestv1, err error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return false, nil, nil, nil, err
	}
	ref, err := reference.WithTag(repoNameRef, tag)
	if err != nil {
		return false, nil, nil, nil, err
	}
	return rp.getImage(ref)
}

// GetImageByDigest ...
func (rp *OciRepository) GetImageByDigest(d digest.Digest) (isList bool, manifest *OciManifest, manifestList *OciManifestIndex, manifestv1 *OciManifestv1, err error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return false, nil, nil, nil, err
	}
	ref, err := reference.WithDigest(repoNameRef, d)
	if err != nil {
		return false, nil, nil, nil, err
	}
	return rp.getImage(ref)
}

func (rp *OciRepository) getImage(ref reference.Named) (isList bool, manifest *OciManifest, manifestList *OciManifestIndex, manifestv1 *OciManifestv1, err error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return false, nil, nil, nil, err
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	if err != nil {
		return false, nil, nil, nil, err
	}
	ms, err := repo.Manifests(nil)
	if err != nil {
		return false, nil, nil, nil, err
	}
	var m1 distribution.Manifest
	if digested, isDigested := ref.(reference.Canonical); isDigested {
		m1, err = ms.Get(nil, digested.Digest())
	} else if tagged, isTagged := ref.(reference.NamedTagged); isTagged {
		m1, err = ms.Get(nil, "", distribution.WithTag(tagged.Tag()))
	} else {
		return false, nil, nil, nil, fmt.Errorf("internal error: reference has neither a tag nor a digest: %s", reference.FamiliarString(ref))
	}

	if err != nil {
		return false, nil, nil, nil, err
	}

	switch v := m1.(type) {
	case *schema1.SignedManifest:
		return false, nil, nil, &OciManifestv1{m1.(*schema1.SignedManifest), rp}, nil
	case *schema2.DeserializedManifest:
		return false, &OciManifest{m1.(*schema2.DeserializedManifest), rp}, nil, nil, nil
	case *manifestlist.DeserializedManifestList:
		return true, nil, &OciManifestIndex{m1.(*manifestlist.DeserializedManifestList), rp}, nil, nil
	default:
		return false, nil, nil, nil, fmt.Errorf("bad manifest type %v", v)
	}
}

// PushBlob ...
func (rp *OciRepository) PushBlob(data io.Reader) (distribution.Descriptor, error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	bs := repo.Blobs(nil)
	writer, err := bs.Create(nil)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	dgstr := digest.Canonical.Digester()
	n, err := io.Copy(writer, io.TeeReader(data, dgstr.Hash()))
	if err != nil {
		return distribution.Descriptor{}, err
	}

	desc := distribution.Descriptor{
		Size:   n,
		Digest: dgstr.Digest(),
	}

	return writer.Commit(nil, desc)
}

// PushManifest ...
func (rp *OciRepository) PushManifest(tag, mediaType string, manifestPayload []byte) (digest.Digest, error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return "", err
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	if err != nil {
		return "", err
	}
	ms, err := repo.Manifests(nil)
	if err != nil {
		return "", err
	}

	m, _, err := distribution.UnmarshalManifest(mediaType, manifestPayload)

	d, err := ms.Put(nil, m, distribution.WithTag(tag))
	return d, err
}

// DeleteManifest ...
func (rp *OciRepository) DeleteManifest(dgst digest.Digest) error {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return err
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	if err != nil {
		return err
	}
	ms, err := repo.Manifests(nil)
	if err != nil {
		return err
	}

	return ms.Delete(nil, dgst)
}

// ManifestExist ...
func (rp *OciRepository) ManifestExist(dgst digest.Digest) (bool, error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return false, err
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	if err != nil {
		return false, err
	}
	ms, err := repo.Manifests(nil)
	if err != nil {
		return false, err
	}

	return ms.Exists(nil, dgst)
}

// MountBlob ...
func (rp *OciRepository) MountBlob(dgst, from string) (distribution.Descriptor, error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	bs := repo.Blobs(nil)
	remoteRef, err := reference.WithName(from)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	canonicalRef, err := reference.WithDigest(remoteRef, digest.FromString(dgst))
	if err != nil {
		return distribution.Descriptor{}, err
	}
	_, err = bs.Create(nil, client.WithMountFrom(canonicalRef))
	if err != nil {
		if ebm, ok := err.(distribution.ErrBlobMounted); ok {
			return ebm.Descriptor, nil
		}
		return distribution.Descriptor{}, err
	}
	return distribution.Descriptor{}, errors.Errorf("No such blob %s %s", from, dgst)
}

// BlobExist ...
func (rp *OciRepository) BlobExist(d digest.Digest) (distribution.Descriptor, error) {
	repoNameRef, err := reference.WithName(rp.Name)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	repo, err := client.NewRepository(repoNameRef, rp.Registry.baseURL, rp.Registry.transport)
	if err != nil {
		return distribution.Descriptor{}, err
	}
	return repo.Blobs(nil).Stat(nil, d)
}
