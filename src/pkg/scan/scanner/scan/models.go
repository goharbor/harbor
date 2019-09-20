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

package scan

// Artifact represents an artifact stored in Registry.
type Artifact struct {
	// The full name of a Harbor repository containing the artifact, including the namespace.
	// For example, `library/oracle/nosql`.
	Repository string
	// The artifact's digest, consisting of an algorithm and hex portion.
	// For example, `sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b`,
	// represents sha256 based digest.
	Digest string
	// The mime type of the scanned artifact
	MimeType string
}

// Registry represents Registry connection settings.
type Registry struct {
	// A base URL of the Docker Registry v2 API exposed by Harbor.
	URL string
	// An optional value of the HTTP Authorization header sent with each request to the Docker Registry v2 API.
	// For example, `Bearer: JWTTOKENGOESHERE`.
	Authorization string
}

// Request represents a structure that is sent to a Scanner Adapter to initiate artifact scanning.
// Conducts all the details required to pull the artifact from a Harbor registry.
type Request struct {
	// Connection settings for the Docker Registry v2 API exposed by Harbor.
	Registry *Registry
	// Artifact to be scanned.
	Artifact *Artifact
}
