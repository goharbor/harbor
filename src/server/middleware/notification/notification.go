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

package notification

import (
	"container/list"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"

	"github.com/goharbor/harbor/src/internal"
	evt "github.com/goharbor/harbor/src/pkg/notifier/event"
)

// publishEvent publishes the events in the context, it ensures publish happens after transaction success.
func publishEvent(es *list.List) {
	if es == nil {
		return
	}
	for e := es.Front(); e != nil; e = e.Next() {
		evt.BuildAndPublish(e.Value.(evt.Metadata))
	}
	return
}

// Middleware sends the notification after transaction success
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		res := internal.NewResponseRecorder(w)
		eveCtx := &notification.EventCtx{
			Events:     list.New(),
			MustNotify: false,
		}
		ctx := notification.NewContext(r.Context(), eveCtx)
		next.ServeHTTP(res, r.WithContext(ctx))
		if res.Success() || eveCtx.MustNotify {
			publishEvent(eveCtx.Events)
		}
	}, skippers...)
}
