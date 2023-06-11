//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sigstore/cosign/pkg/oci"
)

// These enable mocking for unit testing without faking an entire registry.
var (
	remoteImage = remote.Image
	remoteIndex = remote.Index
	remoteGet   = remote.Get
	remoteWrite = remote.Write
)

// SignedEntity provides access to a remote reference, and its signatures.
// The SignedEntity will be one of SignedImage or SignedImageIndex.
func SignedEntity(ref name.Reference, options ...Option) (oci.SignedEntity, error) {
	o := makeOptions(ref.Context(), options...)

	got, err := remoteGet(ref, o.ROpt...)
	var te *transport.Error
	if errors.As(err, &te) && te.StatusCode == http.StatusNotFound {
		return nil, errors.New("entity not found in registry")
	} else if err != nil {
		return nil, err
	}

	switch got.MediaType {
	case types.OCIImageIndex, types.DockerManifestList:
		ii, err := got.ImageIndex()
		if err != nil {
			return nil, err
		}
		return &index{
			v1Index: ii,
			ref:     ref.Context().Digest(got.Digest.String()),
			opt:     o,
		}, nil

	case types.OCIManifestSchema1, types.DockerManifestSchema2:
		i, err := got.Image()
		if err != nil {
			return nil, err
		}
		return &image{
			Image: i,
			opt:   o,
		}, nil

	default:
		return nil, fmt.Errorf("unknown mime type: %v", got.MediaType)
	}
}

// normalize turns image digests into tags with optional prefix & suffix:
// sha256:d34db33f -> [prefix]sha256-d34db33f[.suffix]
func normalize(h v1.Hash, prefix string, suffix string) string {
	if suffix == "" {
		return fmt.Sprint(prefix, h.Algorithm, "-", h.Hex)
	}
	return fmt.Sprint(prefix, h.Algorithm, "-", h.Hex, ".", suffix)
}

// SignatureTag returns the name.Tag that associated signatures with a particular digest.
func SignatureTag(ref name.Reference, opts ...Option) (name.Tag, error) {
	o := makeOptions(ref.Context(), opts...)
	return suffixTag(ref, o.SignatureSuffix, o)
}

// AttestationTag returns the name.Tag that associated attestations with a particular digest.
func AttestationTag(ref name.Reference, opts ...Option) (name.Tag, error) {
	o := makeOptions(ref.Context(), opts...)
	return suffixTag(ref, o.AttestationSuffix, o)
}

// SBOMTag returns the name.Tag that associated SBOMs with a particular digest.
func SBOMTag(ref name.Reference, opts ...Option) (name.Tag, error) {
	o := makeOptions(ref.Context(), opts...)
	return suffixTag(ref, o.SBOMSuffix, o)
}

func suffixTag(ref name.Reference, suffix string, o *options) (name.Tag, error) {
	var h v1.Hash
	if digest, ok := ref.(name.Digest); ok {
		var err error
		h, err = v1.NewHash(digest.DigestStr())
		if err != nil { // This is effectively impossible.
			return name.Tag{}, err
		}
	} else {
		desc, err := remoteGet(ref, o.ROpt...)
		if err != nil {
			return name.Tag{}, err
		}
		h = desc.Digest
	}
	return o.TargetRepository.Tag(normalize(h, o.TagPrefix, suffix)), nil
}

type digestable interface {
	Digest() (v1.Hash, error)
}

// signatures is a shared implementation of the oci.Signed* Signatures method.
func signatures(digestable digestable, o *options) (oci.Signatures, error) {
	h, err := digestable.Digest()
	if err != nil {
		return nil, err
	}
	return Signatures(o.TargetRepository.Tag(normalize(h, o.TagPrefix, o.SignatureSuffix)), o.OriginalOptions...)
}

// attestations is a shared implementation of the oci.Signed* Attestations method.
func attestations(digestable digestable, o *options) (oci.Signatures, error) {
	h, err := digestable.Digest()
	if err != nil {
		return nil, err
	}
	return Signatures(o.TargetRepository.Tag(normalize(h, o.TagPrefix, o.AttestationSuffix)), o.OriginalOptions...)
}

// attachment is a shared implementation of the oci.Signed* Attachment method.
func attachment(digestable digestable, attName string, o *options) (oci.File, error) {
	h, err := digestable.Digest()
	if err != nil {
		return nil, err
	}
	img, err := SignedImage(o.TargetRepository.Tag(normalize(h, o.TagPrefix, attName)), o.OriginalOptions...)
	if err != nil {
		return nil, err
	}
	ls, err := img.Layers()
	if err != nil {
		return nil, err
	}
	if len(ls) != 1 {
		return nil, fmt.Errorf("expected exactly one layer in attachment, got %d", len(ls))
	}

	return &attached{
		SignedImage: img,
		layer:       ls[0],
	}, nil
}

type attached struct {
	oci.SignedImage
	layer v1.Layer
}

var _ oci.File = (*attached)(nil)

// FileMediaType implements oci.File
func (f *attached) FileMediaType() (types.MediaType, error) {
	return f.layer.MediaType()
}

// Payload implements oci.File
func (f *attached) Payload() ([]byte, error) {
	// remote layers are believed to be stored
	// compressed, but we don't compress attachments
	// so use "Compressed" to access the raw byte
	// stream.
	rc, err := f.layer.Compressed()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
