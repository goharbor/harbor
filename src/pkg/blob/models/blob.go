// Copyright Project Harbor Authors
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

package models

import (
	"github.com/goharbor/harbor/src/common/models"
)

// TODO: move ArtifactAndBlob, Blob and ProjectBlob to here

// ArtifactAndBlob alias ArtifactAndBlob model
type ArtifactAndBlob = models.ArtifactAndBlob

// Blob alias Blob model
type Blob = models.Blob

// ProjectBlob alias ProjectBlob model
type ProjectBlob = models.ProjectBlob

// ListParams list params
type ListParams struct {
	ArtifactDigest string   // list blobs which associated with the artifact
	BlobDigests    []string // list blobs which digest in the digests
}
