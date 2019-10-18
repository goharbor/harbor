package regtoken

import (
	"context"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	pkg_token "github.com/goharbor/harbor/src/pkg/token"
	"github.com/goharbor/harbor/src/pkg/token/claim"
	"net/http"
	"strings"
)

type regTokenHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &regTokenHandler{
		next: next,
	}
}

// ServeHTTP ...
func (r *regTokenHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(util.ImageInfoCtxKey)
	if imgRaw == nil || !config.WithClair() {
		r.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(util.ImageInfoCtxKey).(util.ImageInfo)
	if img.Digest == "" {
		r.next.ServeHTTP(rw, req)
		return
	}

	parts := strings.Split(req.Header.Get("Authorization"), " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		r.next.ServeHTTP(rw, req)
		return
	}
	rawToken := parts[1]
	opt := pkg_token.DefaultTokenOptions()
	rClaims := &claim.Registry{}
	rtk, err := pkg_token.Parse(opt, rawToken, rClaims)
	if err != nil {
		log.Debug("failed to decode reg token: %v, the error is skipped and round the request to native registry.", err)
		r.next.ServeHTTP(rw, req)
		return
	}
	ctx := context.WithValue(req.Context(), util.PolicyCheckCtxKey, rtk.Claims.(*claim.Registry).PolicyCheck)
	req = req.WithContext(ctx)
	r.next.ServeHTTP(rw, req)
}
