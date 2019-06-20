package middlewares

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/blobquota"
	"github.com/goharbor/harbor/src/core/middlewares/contenttrust"
	"github.com/goharbor/harbor/src/core/middlewares/listrepo"
	"github.com/goharbor/harbor/src/core/middlewares/multiplmanifest"
	"github.com/goharbor/harbor/src/core/middlewares/readonly"
	"github.com/goharbor/harbor/src/core/middlewares/regquota"
	"github.com/goharbor/harbor/src/core/middlewares/url"
	"github.com/goharbor/harbor/src/core/middlewares/vulnerable"
	"github.com/justinas/alice"
	"net/http"
)

type DefaultCreator struct {
	middlewares []string
}

func New(middlewares []string) *DefaultCreator {
	return &DefaultCreator{
		middlewares: middlewares,
	}
}

// CreateChain ...
func (b *DefaultCreator) Create() *alice.Chain {
	chain := alice.New()
	for _, mName := range b.middlewares {
		middlewareName := mName
		chain = chain.Append(func(next http.Handler) http.Handler {
			constructor := b.getMiddleware(middlewareName)
			if constructor == nil {
				log.Errorf("cannot init middle %s", middlewareName)
				return nil
			}
			return constructor(next)
		})
	}
	return &chain
}

func (b *DefaultCreator) getMiddleware(mName string) alice.Constructor {
	var middleware alice.Constructor

	if mName == READONLY {
		middleware = func(next http.Handler) http.Handler {
			return readonly.New(next)
		}
	}
	if mName == URL {
		if middleware != nil {
			return nil
		}
		middleware = func(next http.Handler) http.Handler {
			return url.New(next)
		}
	}
	if mName == MUITIPLEMANIFEST {
		if middleware != nil {
			return nil
		}
		middleware = func(next http.Handler) http.Handler {
			return multiplmanifest.New(next)
		}
	}
	if mName == LISTREPO {
		if middleware != nil {
			return nil
		}
		middleware = func(next http.Handler) http.Handler {
			return listrepo.New(next)
		}
	}
	if mName == CONTENTTRUST {
		if middleware != nil {
			return nil
		}
		middleware = func(next http.Handler) http.Handler {
			return contenttrust.New(next)
		}
	}
	if mName == VULNERABLE {
		if middleware != nil {
			return nil
		}
		middleware = func(next http.Handler) http.Handler {
			return vulnerable.New(next)
		}
	}
	if mName == REGQUOTA {
		if middleware != nil {
			return nil
		}
		middleware = func(next http.Handler) http.Handler {
			return regquota.New(next)
		}
	}
	if mName == BLOBQUOTA {
		if middleware != nil {
			return nil
		}
		middleware = func(next http.Handler) http.Handler {
			return blobquota.New(next)
		}
	}

	return middleware
}
