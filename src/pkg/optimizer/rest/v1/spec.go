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

package v1

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	// HTTPAcceptHeader represents the HTTP accept header
	HTTPAcceptHeader = "Accept"
	// HTTPContentType represents the HTTP content-type header
	HTTPContentType = "Content-Type"
	// MimeTypeOCIArtifact defines the mime type for OCI image manifest
	MimeTypeOCIArtifact = "application/vnd.oci.image.manifest.v1+json"
	// MimeTypeOCIIndex defines the mime type for OCI image index. BuildKit provenance
	// attestations hang off the index, so optimizer adapters must consume it.
	MimeTypeOCIIndex = "application/vnd.oci.image.index.v1+json"
	// MimeTypeDockerArtifact defines the mime type for docker manifest
	MimeTypeDockerArtifact = "application/vnd.docker.distribution.manifest.v2+json"
	// MimeTypeDockerManifestList defines the mime type for docker manifest list
	MimeTypeDockerManifestList = "application/vnd.docker.distribution.manifest.list.v2+json"
	// MimeTypeAdapterMeta defines the mime type for optimizer adapter metadata
	MimeTypeAdapterMeta = "application/vnd.harbor.optimizer.metadata+json; version=1.0"
	// MimeTypeOptimizeRequest defines the mime type for optimize request
	MimeTypeOptimizeRequest = "application/vnd.harbor.optimizer.request+json; version=1.0"
	// MimeTypeOptimizeResponse defines the mime type for optimize response
	MimeTypeOptimizeResponse = "application/vnd.harbor.optimizer.response+json; version=1.0"
	// MimeTypeOptimizationReport defines the mime type for the optimization report
	MimeTypeOptimizationReport = "application/vnd.harbor.optimizer.report+json; version=1.0"

	apiPrefix = "/api/v1"
)

// RequestResolver is a function template to modify the API request, e.g: add headers
type RequestResolver func(req *http.Request)

// Definition for API
type Definition struct {
	// URL of the API
	URL string
	// Resolver for the request
	Resolver RequestResolver
}

// Spec of the API
// Contains URL and possible headers.
type Spec struct {
	baseRoute string
}

// NewSpec news V1 spec
func NewSpec(base string) *Spec {
	s := &Spec{}

	if len(base) > 0 {
		if strings.HasSuffix(base, "/") {
			s.baseRoute = base[:len(base)-1]
		} else {
			s.baseRoute = base
		}
	}

	s.baseRoute = fmt.Sprintf("%s%s", s.baseRoute, apiPrefix)

	return s
}

// Metadata API
func (s *Spec) Metadata() Definition {
	return Definition{
		URL: fmt.Sprintf("%s%s", s.baseRoute, "/metadata"),
		Resolver: func(req *http.Request) {
			req.Header.Add(HTTPAcceptHeader, MimeTypeAdapterMeta)
		},
	}
}

// SubmitOptimize API
func (s *Spec) SubmitOptimize() Definition {
	return Definition{
		URL: fmt.Sprintf("%s%s", s.baseRoute, "/optimize"),
		Resolver: func(req *http.Request) {
			req.Header.Add(HTTPContentType, MimeTypeOptimizeRequest)
			req.Header.Add(HTTPAcceptHeader, MimeTypeOptimizeResponse)
		},
	}
}

// GetOptimizationReport API
func (s *Spec) GetOptimizationReport(requestID string) Definition {
	path := fmt.Sprintf("/optimize/%s/report", requestID)

	return Definition{
		URL: fmt.Sprintf("%s%s", s.baseRoute, path),
		Resolver: func(req *http.Request) {
			req.Header.Add(HTTPAcceptHeader, MimeTypeOptimizationReport)
		},
	}
}
