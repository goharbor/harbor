package config

import (
	libCfg "github.com/goharbor/harbor/src/lib/config"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

// Middleware returns a middleware that set the config manager into the context
func Middleware() func(http.Handler) http.Handler {
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		cfgMgr, err := libCfg.GetManager(libCfg.DefaultCfgManager)
		if err != nil {
			lib_http.SendError(rw, err)
			return
		}
		ctx := libCfg.NewContext(req.Context(), cfgMgr)
		next.ServeHTTP(rw, req.WithContext(ctx))
	})
}
