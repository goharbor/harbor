package purge

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/lib/log"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
	"github.com/goharbor/harbor/src/registryctl/api"

	"github.com/docker/distribution/registry/storage"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "goharbor/harbor/src/registryctl/api/registry/purge"

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
	case http.MethodPost:
		h.purgeUploads(w, req)
	default:
		api.HandleNotMethodAllowed(w)
	}
}

type PurgeReq struct {
	OlderThan int64 // clean the file that older than, in hours.
	DryRun    bool
	Async     bool
	LogOut    bool
}

// purgeUploads ...
func (h *handler) purgeUploads(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracelib.StartTrace(r.Context(), tracerName, "purge", trace.WithAttributes(attribute.Key("method").String(r.Method)))
	defer span.End()

	var p PurgeReq
	err := json.NewDecoder(r.Body).Decode(&p)
	if err != nil {
		tracelib.RecordError(span, err, "bad purge request")
		api.HandleBadRequest(w, err)
		return
	}

	exePurge := func(ctx context.Context, driver storagedriver.StorageDriver, olderThan int64, dryRun bool, logOut bool) {
		log.Info("Starting upload purge...")
		deleted, errors := storage.PurgeUploads(ctx, driver, time.Now().Add(-time.Duration(olderThan)), !dryRun)
		log.Infof("Purge uploads finished.  Num deleted=%d, num errors=%d", len(deleted), len(errors))
		if logOut {
			if len(errors) != 0 {
				for _, e := range errors {
					log.Errorf("encountered error during purge: %v", e)
				}
			}
			if len(deleted) != 0 {
				for _, d := range deleted {
					log.Infof("purge removed the dir: %s", d)
				}
			}
		}
	}

	if p.Async {
		go func() {
			exePurge(ctx, h.storageDriver, p.OlderThan, p.DryRun, p.LogOut)
		}()
	} else {
		exePurge(ctx, h.storageDriver, p.OlderThan, p.DryRun, p.LogOut)
	}
	return
}
