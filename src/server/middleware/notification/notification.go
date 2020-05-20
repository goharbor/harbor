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
	"net/http"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/server/middleware"
)

// Middleware sends the notification after transaction success
func Middleware(skippers ...middleware.Skipper) func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		res := lib.NewResponseRecorder(w)
		evc := notification.NewEventCtx()
		next.ServeHTTP(res, r.WithContext(notification.NewContext(r.Context(), evc)))
		if res.Success() || evc.MustNotify {
			for e := evc.Events.Front(); e != nil; e = e.Next() {
				event.BuildAndPublish(e.Value.(event.Metadata))
			}
		}
	}, skippers...)
}
