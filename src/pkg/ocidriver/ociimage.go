package ocidriver

import (
	"github.com/opencontainers/go-digest"
)

// OciImage ...
type OciImage struct {
	Registry   *OciRegistry
	repository *OciRepository

	Name   string
	Digest digest.Digest
}

// GetManifest ...
func (i *OciImage) GetManifest() (isList bool, manifest *OciManifest, manifestList *OciManifestIndex, manifestv1 *OciManifestv1, err error) {
	if i.Digest != "" {
		return i.repository.GetImageByDigest(i.Digest)
	}
	return i.repository.GetImageByTag(i.Name)
}
