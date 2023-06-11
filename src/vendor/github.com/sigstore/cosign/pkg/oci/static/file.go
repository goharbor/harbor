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

package static

import (
	"io"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sigstore/cosign/pkg/oci"
	"github.com/sigstore/cosign/pkg/oci/signed"
)

// NewFile constructs a new v1.Image with the provided payload.
func NewFile(payload []byte, opts ...Option) (oci.File, error) {
	o, err := makeOptions(opts...)
	if err != nil {
		return nil, err
	}
	base := mutate.MediaType(empty.Image, types.OCIManifestSchema1)
	base = mutate.ConfigMediaType(base, o.ConfigMediaType)
	layer := &staticLayer{
		b:    payload,
		opts: o,
	}
	img, err := mutate.Append(base, mutate.Addendum{
		Layer: layer,
	})
	if err != nil {
		return nil, err
	}

	// Add annotations from options
	img = mutate.Annotations(img, o.Annotations).(v1.Image)

	// Set the Created date to time of execution
	img, err = mutate.CreatedAt(img, v1.Time{Time: time.Now()})
	if err != nil {
		return nil, err
	}
	return &file{
		SignedImage: signed.Image(img),
		layer:       layer,
	}, nil
}

type file struct {
	oci.SignedImage
	layer v1.Layer
}

var _ oci.File = (*file)(nil)

// FileMediaType implements oci.File
func (f *file) FileMediaType() (types.MediaType, error) {
	return f.layer.MediaType()
}

// Payload implements oci.File
func (f *file) Payload() ([]byte, error) {
	rc, err := f.layer.Uncompressed()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	return io.ReadAll(rc)
}
