// Copyright 2021 The Sigstore Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package remote

import (
	"github.com/google/go-containerregistry/pkg/name"
)

// ResolveDigest returns the digest of the image at the reference.
//
// If the reference is by digest already, it simply extracts the digest.
// Otherwise, it looks up the digest from the registry.
func ResolveDigest(ref name.Reference, opts ...Option) (name.Digest, error) {
	o := makeOptions(ref.Context(), opts...)
	if d, ok := ref.(name.Digest); ok {
		return d, nil
	}
	desc, err := remoteGet(ref, o.ROpt...)
	if err != nil {
		return name.Digest{}, err
	}
	return ref.Context().Digest(desc.Digest.String()), nil
}
