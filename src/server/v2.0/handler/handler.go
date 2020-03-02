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
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	"log"
	"net/http"
)

// New returns http handler for API V2.0
func New() http.Handler {
	h, api, err := restapi.HandlerAPI(restapi.Config{
		ArtifactAPI:   newArtifactAPI(),
		RepositoryAPI: newRepositoryAPI(),
	})
	if err != nil {
		log.Fatal(err)
	}

	api.ServeError = serveError

	return h
}

// Before executing operation handler, go-swagger will bind a parameters object to a request and validate the request,
// it will return directly when bind and validate failed.
// The response format of the default ServeError implementation does not match the internal error response format.
// So we needed to convert the format to the internal error response format.
func serveError(rw http.ResponseWriter, r *http.Request, err error) {
	serror.SendError(rw, err)
}
