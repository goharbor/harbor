package immutable

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	common_util "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/interceptor"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	middlerware_err "github.com/goharbor/harbor/src/core/middlewares/util/error"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
	"net/http"
)

// NewDeleteMFInteceptor ....
func NewDeleteMFInteceptor(mf *util.ManifestInfo) interceptor.Interceptor {
	return &delmfInterceptor{
		mf: mf,
	}
}

type delmfInterceptor struct {
	mf *util.ManifestInfo
}

// HandleRequest ...
func (dmf *delmfInterceptor) HandleRequest(req *http.Request) (err error) {

	artifactQuery := &models.ArtifactQuery{
		Digest: dmf.mf.Digest,
		Repo:   dmf.mf.Repository,
		PID:    dmf.mf.ProjectID,
	}
	var afs []*models.Artifact
	afs, err = dao.ListArtifacts(artifactQuery)
	if err != nil {
		log.Error(err)
		return
	}
	if len(afs) == 0 {
		return
	}

	for _, af := range afs {
		_, repoName := common_util.ParseRepository(dmf.mf.Repository)
		var matched bool
		matched, err = rule.NewRuleMatcher().Match(dmf.mf.ProjectID, art.Candidate{
			Repository:  repoName,
			Tags:        []string{af.Tag},
			NamespaceID: dmf.mf.ProjectID,
		})
		if err != nil {
			log.Error(err)
			return
		}
		if matched {
			return middlerware_err.NewErrImmutable(repoName, af.Tag)
		}
	}

	return
}

// HandleRequest ...
func (dmf *delmfInterceptor) HandleResponse(w http.ResponseWriter, r *http.Request) {
}
