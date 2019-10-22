package immutable

import (
	"fmt"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor/immutable"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
)

var (
	defaultBuilders = []interceptor.Builder{
		&manifestDeletionBuilder{},
		&manifestCreationBuilder{},
	}
)

type manifestDeletionBuilder struct{}

func (*manifestDeletionBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	if match, _, _ := util.MatchDeleteManifest(req); !match {
		return nil, nil
	}

	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		var err error
		info, err = util.ParseManifestInfoFromPath(req)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest, error %v", err)
		}
	}

	return immutable.NewDeleteMFInteceptor(info), nil
}

type manifestCreationBuilder struct{}

func (*manifestCreationBuilder) Build(req *http.Request) (interceptor.Interceptor, error) {
	if match, _, _ := util.MatchPushManifest(req); !match {
		return nil, nil
	}

	info, ok := util.ManifestInfoFromContext(req.Context())
	if !ok {
		var err error
		info, err = util.ParseManifestInfoFromReq(req)
		if err != nil {
			return nil, fmt.Errorf("failed to parse manifest, error %v", err)
		}
	}

	return immutable.NewPushMFInteceptor(info), nil
}
