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
	// MimeTypeOCIArtifact defines the mime type for OCI artifact
	MimeTypeOCIArtifact = "application/vnd.oci.image.manifest.v1+json"
	// MimeTypeDockerArtifact defines the mime type for docker artifact
	MimeTypeDockerArtifact = "application/vnd.docker.distribution.manifest.v2+json"
	// MimeTypeNativeReport defines the mime type for native report
	MimeTypeNativeReport = "application/vnd.scanner.adapter.vuln.report.harbor+json; version=1.0"
	// MimeTypeRawReport defines the mime type for raw report
	MimeTypeRawReport = "application/vnd.scanner.adapter.vuln.report.raw"
	// MimeTypeAdapterMeta defines the mime type for adapter metadata
	MimeTypeAdapterMeta = "application/vnd.scanner.adapter.metadata+json; version=1.0"
	// MimeTypeScanRequest defines the mime type for scan request
	MimeTypeScanRequest = "application/vnd.scanner.adapter.scan.request+json; version=1.0"
	// MimeTypeScanResponse defines the mime type for scan response
	MimeTypeScanResponse = "application/vnd.scanner.adapter.scan.response+json; version=1.0"
)

// RequestResolver is a function template to modify the API request, e.g: add headers
type RequestResolver func(req *http.Request)

// Definition for API
type Definition struct {
	// URL of the API
	URL string
	// Resolver fro the request
	Resolver RequestResolver
}

// Spec of the API
// Contains URL and possible headers.
type Spec struct {
	baseRoute string
}

// NewSpec news V1 spec
func NewSpec(base string) *Spec {
	baseRoute := "http://localhost"

	if len(base) > 0 {
		if strings.HasSuffix(base, "/") {
			baseRoute = base[:len(base)-1]
		} else {
			baseRoute = base
		}
	}

	return &Spec{
		baseRoute: baseRoute,
	}
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

// SubmitScan API
func (s *Spec) SubmitScan() Definition {
	return Definition{
		URL: fmt.Sprintf("%s%s", s.baseRoute, "/scan"),
		Resolver: func(req *http.Request) {
			req.Header.Add(HTTPContentType, MimeTypeScanRequest)
			req.Header.Add(HTTPAcceptHeader, MimeTypeScanResponse)
		},
	}
}

// GetScanReport API
func (s *Spec) GetScanReport(scanReqID string, mimeType string) Definition {
	path := fmt.Sprintf("/scan/%s/report", scanReqID)

	return Definition{
		URL: fmt.Sprintf("%s%s", s.baseRoute, path),
		Resolver: func(req *http.Request) {
			req.Header.Add(HTTPAcceptHeader, mimeType)
		},
	}
}
