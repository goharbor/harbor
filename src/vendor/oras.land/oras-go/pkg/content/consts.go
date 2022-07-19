/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package content

import (
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// DefaultBlobMediaType specifies the default blob media type
	DefaultBlobMediaType = ocispec.MediaTypeImageLayer
	// DefaultBlobDirMediaType specifies the default blob directory media type
	DefaultBlobDirMediaType = ocispec.MediaTypeImageLayerGzip
)

const (
	// TempFilePattern specifies the pattern to create temporary files
	TempFilePattern = "oras"
)

const (
	// AnnotationDigest is the annotation key for the digest of the uncompressed content
	AnnotationDigest = "io.deis.oras.content.digest"
	// AnnotationUnpack is the annotation key for indication of unpacking
	AnnotationUnpack = "io.deis.oras.content.unpack"
)

const (
	// OCIImageIndexFile is the file name of the index from the OCI Image Layout Specification
	// Reference: https://github.com/opencontainers/image-spec/blob/master/image-layout.md#indexjson-file
	OCIImageIndexFile = "index.json"
)

const (
	// DefaultBlocksize default size of each slice of bytes read in each write through in gunzipand untar.
	// Simply uses the same size as io.Copy()
	DefaultBlocksize = 32768
)

const (
	// what you get for a blank digest
	BlankHash = digest.Digest("sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855")
)
