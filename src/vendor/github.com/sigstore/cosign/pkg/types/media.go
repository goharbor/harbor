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

package types

const (
	JSONInputFormat = "json"
	XMLInputFormat  = "xml"
	TextInputFormat = "text"
)

const (
	CycloneDXXMLMediaType  = "application/vnd.cyclonedx+xml"
	CycloneDXJSONMediaType = "application/vnd.cyclonedx+json"
	SyftMediaType          = "application/vnd.syft+json"
	SimpleSigningMediaType = "application/vnd.dev.cosign.simplesigning.v1+json"
	SPDXMediaType          = "text/spdx"
	SPDXJSONMediaType      = "spdx+json"
	WasmLayerMediaType     = "application/vnd.wasm.content.layer.v1+wasm"
	WasmConfigMediaType    = "application/vnd.wasm.config.v1+json"
)
