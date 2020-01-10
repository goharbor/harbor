package blob

import (
	"errors"
	"github.com/docker/distribution/registry/storage"
	"github.com/goharbor/harbor/src/registryctl/api"
	regConf "github.com/goharbor/harbor/src/registryctl/config/registry"
	"github.com/gorilla/mux"
	"net/http"
)

// NewHandler returns the handler to handler blob request
func NewHandler() http.Handler {
	return &handler{}
}

type handler struct{}

// ServeHTTP ...
func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		h.delete(w, req)
	default:
		api.HandleForbidden(w, req)
	}
}

// DeleteBlob ...
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	digest := mux.Vars(r)["reference"]
	if digest == "" {
		api.HandleBadRequest(w, r, errors.New("no reference specified"))
		return
	}
	cleaner := storage.NewVacuum(r.Context(), regConf.StorageDriver)
	if err := cleaner.RemoveBlob(digest); err != nil {
		api.HandleInternalServerError(w, r)
		return
	}
}
