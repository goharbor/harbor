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

package manifest

import (
	"net/http"

	"github.com/distribution/distribution/v3/registry/storage"
	storagedriver "github.com/distribution/distribution/v3/registry/storage/driver"
	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
	"github.com/goharbor/harbor/src/registryctl/api"
)

const tracerName = "goharbor/harbor/src/registryctl/api/registry/manifest"

// NewHandler returns the handler to handler manifest request
func NewHandler(storageDriver storagedriver.StorageDriver) http.Handler {
	return &handler{
		storageDriver: storageDriver,
	}
}

type handler struct {
	storageDriver storagedriver.StorageDriver
}

// ServeHTTP ...
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		h.delete(w, req)
	default:
		api.HandleNotMethodAllowed(w)
	}
}

// delete deletes manifest ...
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	var span trace.Span
	ctx := r.Context()
	ref := mux.Vars(r)["reference"]
	if tracelib.Enabled() {
		ctx, span = tracelib.StartTrace(ctx, tracerName, "delete-manifest", trace.WithAttributes(attribute.Key("method").String(r.Method)))
		defer span.End()
	}
	if ref == "" {
		err := errors.New("no reference specified")
		tracelib.RecordError(span, err, "no reference specified ")
		api.HandleBadRequest(w, err)
		return
	}
	dgst, err := digest.Parse(ref)
	if err != nil {
		tracelib.RecordError(span, err, "invalid reference")
		api.HandleBadRequest(w, errors.Wrap(err, "not supported reference"))
		return
	}
	repoName := mux.Vars(r)["name"]
	if repoName == "" {
		err := errors.New("no repository name specified")
		tracelib.RecordError(span, err, "no repository name specified")
		api.HandleBadRequest(w, err)
		return
	}
	// let the tags as empty here, as it non-blocking GC. The tags deletion will be handled via DELETE /v2/manifest
	var tags []string
	cleaner := storage.NewVacuum(ctx, h.storageDriver)
	if err := cleaner.RemoveManifest(repoName, dgst, tags); err != nil {
		tracelib.RecordError(span, err, "failed to remove manifest")
		log.Infof("failed to remove manifest: %s, with error:%v", ref, err)
		api.HandleError(w, err)
		return
	}
}
