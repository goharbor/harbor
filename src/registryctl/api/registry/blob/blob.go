package blob

import (
	"errors"
	"net/http"

	"github.com/goharbor/harbor/src/lib/log"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
	"github.com/goharbor/harbor/src/registryctl/api"

	"github.com/docker/distribution/registry/storage"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "goharbor/harbor/src/registryctl/api/registry/blob"

// NewHandler returns the handler to handler blob request
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

// DeleteBlob ...
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracelib.StartTrace(r.Context(), tracerName, "delete-blob", trace.WithAttributes(attribute.Key("method").String(r.Method)))
	defer span.End()
	ref := mux.Vars(r)["reference"]
	if ref == "" {
		err := errors.New("no reference specified")
		tracelib.RecordError(span, err, "no reference specified")
		api.HandleBadRequest(w, err)
		return
	}
	// don't parse the reference here as RemoveBlob does.
	cleaner := storage.NewVacuum(ctx, h.storageDriver)
	if err := cleaner.RemoveBlob(ref); err != nil {
		tracelib.RecordError(span, err, "failed to remove blob")
		log.Infof("failed to remove blob: %s, with error:%v", ref, err)
		api.HandleError(w, err)
		return
	}
}
