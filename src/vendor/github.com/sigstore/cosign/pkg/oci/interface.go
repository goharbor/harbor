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

type SignedEntity interface {
	// Signatures returns the set of signatures currently associated with this
	// entity, or the empty equivalent if none are found.
	Signatures() (Signatures, error)

	// Attestations returns the set of attestations currently associated with this
	// entity, or the empty equivalent if none are found.
	// Attestations are just like a Signature, but they do not contain
	// Base64Signature because it's baked into the payload.
	Attestations() (Signatures, error)

	// Attachment returns a named entity associated with this entity, or error if not found.
	Attachment(name string) (File, error)
}
