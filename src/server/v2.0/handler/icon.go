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
	"github.com/goharbor/harbor/src/controller/icon"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/icon"
)

func newIconAPI() *iconAPI {
	return &iconAPI{
		ctl: icon.Ctl,
	}
}

type iconAPI struct {
	BaseAPI
	ctl icon.Controller
}

func (i *iconAPI) GetIcon(ctx context.Context, params operation.GetIconParams) middleware.Responder {
	icon, err := i.ctl.Get(ctx, params.Digest)
	if err != nil {
		return i.SendError(ctx, err)
	}

	return operation.NewGetIconOK().WithPayload(&models.Icon{
		Content:     icon.Content,
		ContentType: icon.ContentType,
	})
}
