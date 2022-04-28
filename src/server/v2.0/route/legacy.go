// Copyright 2018 Project Harbor Authors
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

package route

import (
	"github.com/beego/beego"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/lib/config"
)

// RegisterRoutes for Harbor legacy APIs
func registerLegacyRoutes() {
	version := APIVersion
	beego.Router("/api/"+version+"/email/ping", &api.EmailAPI{}, "post:Ping")

	// APIs for chart repository
	if config.WithChartMuseum() {
		// Labels for chart
		chartLabelAPIType := &api.ChartLabelAPI{}
		beego.Router("/api/"+version+"/chartrepo/:repo/charts/:name/:version/labels", chartLabelAPIType, "get:GetLabels;post:MarkLabel")
		beego.Router("/api/"+version+"/chartrepo/:repo/charts/:name/:version/labels/:id([0-9]+)", chartLabelAPIType, "delete:RemoveLabel")
	}
}
