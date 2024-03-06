// Copyright 2021 Google LLC All Rights Reserved.
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

package static

import (
	"bytes"
	"io"
	"sync"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
)

// NewLayer returns a layer containing the given bytes, with the given mediaType.
//
// Contents will not be compressed.
func NewLayer(b []byte, mt types.MediaType) v1.Layer {
	return &staticLayer{b: b, mt: mt}
}

type staticLayer struct {
	b  []byte
	mt types.MediaType

	once sync.Once
	h    v1.Hash
}

func (l *staticLayer) Digest() (v1.Hash, error) {
	var err error
	// Only calculate digest the first time we're asked.
	l.once.Do(func() {
		l.h, _, err = v1.SHA256(bytes.NewReader(l.b))
	})
	return l.h, err
}

func (l *staticLayer) DiffID() (v1.Hash, error) {
	return l.Digest()
}

func (l *staticLayer) Compressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(l.b)), nil
}

func (l *staticLayer) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(l.b)), nil
}

func (l *staticLayer) Size() (int64, error) {
	return int64(len(l.b)), nil
}

func (l *staticLayer) MediaType() (types.MediaType, error) {
	return l.mt, nil
}
