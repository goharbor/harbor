package mainfest

import (
	"errors"
	"github.com/docker/distribution/registry/storage"
	"github.com/goharbor/harbor/src/registryctl/api"
	regConf "github.com/goharbor/harbor/src/registryctl/config/registry"
	"github.com/gorilla/mux"
	"github.com/opencontainers/go-digest"
	"net/http"
	"strings"
)

// NewHandler returns the handler to handler manifest request
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

// delete deletes manifest ...
func (h *handler) delete(w http.ResponseWriter, r *http.Request) {
	dgst, err := digest.Parse(mux.Vars(r)["reference"])
	if err != nil {
		api.HandleBadRequest(w, r, err)
		return
	}
	if dgst == "" {
		api.HandleBadRequest(w, r, errors.New("no reference specified"))
		return
	}
	repoName := mux.Vars(r)["name"]
	if repoName == "" {
		api.HandleBadRequest(w, r, errors.New("no name specified"))
		return
	}
	var tags []string
	v := r.URL.Query()
	queryTags := v.Get("tags")
	tags = strings.Split(queryTags, ",")

	cleaner := storage.NewVacuum(r.Context(), regConf.StorageDriver)
	if err := cleaner.RemoveManifest(repoName, dgst, tags); err != nil {
		api.HandleInternalServerError(w, r)
		return
	}
}
