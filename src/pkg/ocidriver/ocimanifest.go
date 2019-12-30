package ocidriver

import (
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
)

// OciManifestv1 ...
type OciManifestv1 struct {
	*schema1.SignedManifest

	repository *OciRepository
}

// OciManifest ...
type OciManifest struct {
	*schema2.DeserializedManifest

	repository *OciRepository
}

// OciManifestIndex ...
type OciManifestIndex struct {
	*manifestlist.DeserializedManifestList

	repository *OciRepository
}

// Get ...
func (l *OciManifestIndex) Get(index int) (isList bool, manifest *OciManifest, manifestList *OciManifestIndex, manifestv1 *OciManifestv1, err error) {
	return l.repository.GetImageByDigest(l.Manifests[index].Digest)
}
