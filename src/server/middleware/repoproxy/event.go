package repoproxy

import (
	"github.com/goharbor/harbor/src/lib/orm"
	"net/http"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/controller/proxy"
	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	httpLib "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	reg "github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/router"
)

func canAutoSync(p *models.Project) bool {
	if p.RegistryID < 1 {
		return false
	}
	reg, err := registry.Ctl.Get(orm.Context(), p.RegistryID)
	if err != nil {
		log.Errorf("failed to get registry, error:%v", err)
		return false
	}
	if reg.Status != model.Healthy {
		log.Errorf("current registry is unhealthy, regID:%v, Name:%v, Status: %v", reg.ID, reg.Name, reg.Status)
	}
	return reg.Status == model.Healthy && p.AutoSynced
}

// PullEventMiddleware middleware handle request for get or head manifest
func PullEventMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if err := handlePullEvent(w, r, next); err != nil {
			httpLib.SendError(w, err)
		}
	})
}

func handlePullEvent(w http.ResponseWriter, r *http.Request, next http.Handler) error {

	ctx := r.Context()
	ctx = lib.WithProxyPullMode(ctx, true)

	art, p, proxyCtl, err := preCheck(ctx)
	if err != nil {
		return err
	}

	if !canAutoSync(p) {
		next.ServeHTTP(w, r)
		return nil
	}

	remote, err := proxy.NewRemoteHelper(p.RegistryID)
	if err != nil {
		return err
	}
	useLocal, _, err := proxyCtl.UseLocalManifest(ctx, art, remote)
	if err != nil {
		return err
	}
	if useLocal {
		next.ServeHTTP(w, r)
		return nil
	}
	log.Debugf("the tag is %v, digest is %v", art.Tag, art.Digest)

	if r.Method == http.MethodHead {
		err = proxyManifestHead(ctx, w, proxyCtl, p, art, remote)
	} else if r.Method == http.MethodGet {
		err = proxyManifestGet(ctx, w, proxyCtl, p, art, remote)
	}

	if err != nil {
		if errors.IsNotFoundErr(err) {
			return err
		}
		log.Warningf("Proxy to remote failed, fallback to local repo, error: %v", err)
		next.ServeHTTP(w, r)
	}

	if r.UserAgent() == reg.UserAgent {
		return nil
	}

	return addPullArtifactEvent(r, art, p.ProjectID, remote)
}

func addPullArtifactEvent(r *http.Request, art lib.ArtifactInfo, projectID int64, remote proxy.RemoteInterface) error {
	var ref string
	if len(art.Tag) > 0 {
		ref = art.Tag
	} else {
		ref = art.Digest
	}

	_, dig, err := remote.Manifest(art.Repository, ref)
	if err != nil {
		log.Error("repo %s:%s not found", art.Repository, art.Tag)
		return err
	}

	var a = &artifact.Artifact{}
	a.Type = image.ArtifactTypeImage
	a.RepositoryName = router.Param(r.Context(), ":splat")
	a.ProjectID = projectID
	a.Digest = dig

	proxy.SendPullEvent(a, art.Tag, operator.FromContext(r.Context()))
	return nil
}

// PullEventBlobGetMiddleware handle get blob request
func PullEventBlobGetMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {

		if err := handlePullEventBlob(w, r, next); err != nil {
			httpLib.SendError(w, err)
		}
	})
}

func handlePullEventBlob(w http.ResponseWriter, r *http.Request, next http.Handler) error {
	ctx := r.Context()
	_, p, _, err := preCheck(ctx)
	if err != nil {
		return err
	}

	if p.AutoSynced {
		r = r.WithContext(lib.WithProxyPullMode(ctx, true))
	}

	next.ServeHTTP(w, r)
	return nil
}
