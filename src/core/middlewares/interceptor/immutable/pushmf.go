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

// NewPushMFInteceptor ....
func NewPushMFInteceptor(mf *util.ManifestInfo) interceptor.Interceptor {
	return &pushmfInterceptor{
		mf: mf,
	}
}

type pushmfInterceptor struct {
	mf *util.ManifestInfo
}

// HandleRequest ...
func (pmf *pushmfInterceptor) HandleRequest(req *http.Request) (err error) {

	_, repoName := common_util.ParseRepository(pmf.mf.Repository)
	var matched bool
	matched, err = rule.NewRuleMatcher(pmf.mf.ProjectID).Match(art.Candidate{
		Repository:  repoName,
		Tag:         pmf.mf.Tag,
		NamespaceID: pmf.mf.ProjectID,
	})
	if err != nil {
		log.Error(err)
		return
	}
	if !matched {
		return
	}

	artifactQuery := &models.ArtifactQuery{
		PID:  pmf.mf.ProjectID,
		Repo: pmf.mf.Repository,
		Tag:  pmf.mf.Tag,
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

	return middlerware_err.NewErrImmutable(repoName, pmf.mf.Tag)
}

// HandleRequest ...
func (pmf *pushmfInterceptor) HandleResponse(w http.ResponseWriter, r *http.Request) {
}
