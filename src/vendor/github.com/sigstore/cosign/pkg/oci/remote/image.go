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
	"net/http"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/sigstore/cosign/pkg/oci"
)

var ErrImageNotFound = errors.New("image not found in registry")

// SignedImage provides access to a remote image reference, and its signatures.
func SignedImage(ref name.Reference, options ...Option) (oci.SignedImage, error) {
	o := makeOptions(ref.Context(), options...)
	ri, err := remoteImage(ref, o.ROpt...)
	var te *transport.Error
	if errors.As(err, &te) && te.StatusCode == http.StatusNotFound {
		return nil, ErrImageNotFound
	} else if err != nil {
		return nil, err
	}
	return &image{
		Image: ri,
		opt:   o,
	}, nil
}

type image struct {
	v1.Image
	opt *options
}

var _ oci.SignedImage = (*image)(nil)

// Signatures implements oci.SignedImage
func (i *image) Signatures() (oci.Signatures, error) {
	return signatures(i, i.opt)
}

// Attestations implements oci.SignedImage
func (i *image) Attestations() (oci.Signatures, error) {
	return attestations(i, i.opt)
}

// Attestations implements oci.SignedImage
func (i *image) Attachment(name string) (oci.File, error) {
	return attachment(i, name, i.opt)
}
