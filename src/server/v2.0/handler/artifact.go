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

package handler

import (
	"context"

	"github.com/go-openapi/runtime/middleware"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/artifact"
)

// ArtifactAPI the api implemention of artifacts
type ArtifactAPI struct {
	BaseAPI
}

// DeleteArtifact ...
func (api *ArtifactAPI) DeleteArtifact(ctx context.Context, params operation.DeleteArtifactParams) middleware.Responder {
	return operation.NewDeleteArtifactOK()
}

// ListArtifacts ...
func (api *ArtifactAPI) ListArtifacts(ctx context.Context, params operation.ListArtifactsParams) middleware.Responder {
	return operation.NewListArtifactsOK()
}

// ReadArtifact ...
func (api *ArtifactAPI) ReadArtifact(ctx context.Context, params operation.ReadArtifactParams) middleware.Responder {
	return operation.NewReadArtifactOK()
}

// NewArtifactAPI returns API of artifacts
func NewArtifactAPI() *ArtifactAPI {
	return &ArtifactAPI{}
}
