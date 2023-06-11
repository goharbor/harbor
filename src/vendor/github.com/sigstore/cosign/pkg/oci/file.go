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

package oci

import "github.com/google/go-containerregistry/pkg/v1/types"

// File is a degenerate form of SignedImage that stores a single file as a v1.Layer
type File interface {
	SignedImage

	// FileMediaType retrieves the media type of the File
	FileMediaType() (types.MediaType, error)

	// Payload fetches the opaque data that is being signed.
	// This will always return data when there is no error.
	Payload() ([]byte, error)
}
