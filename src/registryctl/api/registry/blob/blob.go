package blob

import (
	"errors"
	"github.com/docker/distribution/registry/storage"
	"github.com/goharbor/harbor/src/lib/log"
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
		api.HandleNotMethodAllowed(w)
	}
}

// DeleteBlob ...
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	ref := mux.Vars(r)["reference"]
	if ref == "" {
		api.HandleBadRequest(w, errors.New("no reference specified"))
		return
	}
	// don't parse the reference here as RemoveBlob does.
	cleaner := storage.NewVacuum(r.Context(), regConf.StorageDriver)
	if err := cleaner.RemoveBlob(ref); err != nil {
		log.Infof("failed to remove blob: %s, with error:%v", ref, err)
		api.HandleError(w, err)
		return
	}
}
