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

package filter

import (
	"net/http"
	"strings"

	beegoctx "github.com/astaxie/beego/context"

	hlog "github.com/goharbor/harbor/src/common/utils/log"
)

// MediaTypeFilter filters the POST request, it returns 415 if the content type of the request
// doesn't match the preset ones.
func MediaTypeFilter(mediaType ...string) func(*beegoctx.Context) {
	return func(ctx *beegoctx.Context) {
		filterContentType(ctx.Request, ctx.ResponseWriter, mediaType...)
	}
}

func filterContentType(req *http.Request, resp http.ResponseWriter, mediaType ...string) {
	if req.Method != http.MethodPost {
		return
	}
	v := req.Header.Get("Content-Type")
	mimeType := strings.Split(v, ";")[0]
	hlog.Debugf("Mimetype of incoming request %s: %s", req.RequestURI, mimeType)

	for _, t := range mediaType {
		if t == mimeType {
			return
		}
	}
	resp.WriteHeader(http.StatusUnsupportedMediaType)
}
