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

package filter

import (
	"github.com/astaxie/beego/context"
	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/internal/orm"
)

// OrmFilter set orm.Ormer instance to the context of the http.Request
func OrmFilter(ctx *context.Context) {
	if ctx == nil || ctx.Request == nil {
		return
	}

	ctx.Request = ctx.Request.WithContext(orm.NewContext(ctx.Request.Context(), o.NewOrm()))
}
