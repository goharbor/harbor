package export

import (
	"context"

	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/lib/selector/selectors/doublestar"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/user"
)

type FilterProcessor interface {
	ProcessProjectFilter(ctx context.Context, userName string, projectsToFilter []int64) ([]int64, error)
	ProcessRepositoryFilter(ctx context.Context, filter string, projectIds []int64) ([]int64, error)
	ProcessTagFilter(ctx context.Context, filter string, repositoryIds []int64) ([]*artifact.Artifact, error)
	ProcessLabelFilter(ctx context.Context, labelIDs []int64, arts []*artifact.Artifact) ([]*artifact.Artifact, error)
}

type DefaultFilterProcessor struct {
	artCtl     artifact.Controller
	repoMgr    repository.Manager
	usrMgr     user.Manager
	projectMgr project.Manager
}

// NewFilterProcessor constructs an instance of a FilterProcessor
func NewFilterProcessor() FilterProcessor {
	return &DefaultFilterProcessor{
		artCtl:     artifact.Ctl,
		repoMgr:    pkg.RepositoryMgr,
		usrMgr:     user.Mgr,
		projectMgr: pkg.ProjectMgr,
	}
}

func (dfp *DefaultFilterProcessor) ProcessProjectFilter(ctx context.Context, userName string, projectIdsToFilter []int64) ([]int64, error) {
	// get the user id of the current user

	usr, err := dfp.usrMgr.GetByName(ctx, userName)
	if err != nil {
		return nil, err
	}
	logger.Infof("Retrieved user id :%d for user name : %s", usr.UserID, userName)

	if err != nil {
		return nil, err
	}

	query := dfp.getProjectQueryFilter(usr)
	projects, err := dfp.projectMgr.List(ctx, query)

	if err != nil {
		return nil, err
	}
	logger.Infof("Selected %d projects administered by user %s ", len(projects), userName)
	projectIds := make([]int64, 0)
	for _, proj := range projects {
		projectIds = append(projectIds, proj.ProjectID)
	}

	// check if the project ids specified in the filter are present in the list
	// of projects of which the current user is a project admin
	if len(projectIdsToFilter) == 0 {
		return projectIds, nil
	}
	m := make(map[int64]bool)
	for _, projectID := range projectIds {
		m[projectID] = true
	}
	filtered := make([]int64, 0)

	for _, filteredProjID := range projectIdsToFilter {
		if m[filteredProjID] {
			filtered = append(filtered, filteredProjID)
		}
	}
	return filtered, nil
}

func (dfp *DefaultFilterProcessor) ProcessRepositoryFilter(ctx context.Context, filter string, projectIds []int64) ([]int64, error) {
	sel := doublestar.New(doublestar.RepoMatches, filter, "")
	candidates := make([]*selector.Candidate, 0)
	allRepoIDs := make([]int64, 0)

	for _, projectID := range projectIds {
		query := q.New(q.KeyWords{"ProjectID": projectID})
		allRepos, err := dfp.repoMgr.List(ctx, query)
		if err != nil {
			return nil, err
		}
		for _, repository := range allRepos {
			allRepoIDs = append(allRepoIDs, repository.RepositoryID)
			namespace, repo := utils.ParseRepository(repository.Name)
			candidates = append(candidates, &selector.Candidate{NamespaceID: repository.RepositoryID, Namespace: namespace, Repository: repo, Kind: "image"})
		}
	}
	// no repo filter specified then return all repos across all projects
	if filter == "" {
		return allRepoIDs, nil
	}
	// select candidates by filter
	candidates, err := sel.Select(candidates)
	if err != nil {
		return nil, err
	}
	// extract repository id from candidate
	repoIDs := make([]int64, 0)
	for _, c := range candidates {
		repoIDs = append(repoIDs, c.NamespaceID)
	}

	return repoIDs, nil
}

func (dfp *DefaultFilterProcessor) ProcessTagFilter(ctx context.Context, filter string, repositoryIds []int64) ([]*artifact.Artifact, error) {
	arts := make([]*artifact.Artifact, 0)
	opts := &artifact.Option{
		WithTag:   true,
		WithLabel: true,
		// if accessory support scan in the future, just add withAccessory here.
		// WithAccessory: true
	}
	// list all artifacts by repository id
	for _, repoID := range repositoryIds {
		repoArts, err := dfp.artCtl.List(ctx, q.New(q.KeyWords{"RepositoryID": repoID}), opts)
		if err != nil {
			return nil, err
		}

		arts = append(arts, repoArts...)
	}
	// return earlier if no tag filter
	if filter == "" {
		return arts, nil
	}

	// filter by tag
	sel := doublestar.New(doublestar.Matches, filter, "")
	candidates := make([]*selector.Candidate, 0)
	for _, art := range arts {
		tags := make([]string, 0, len(art.Tags))
		for _, tag := range art.Tags {
			tags = append(tags, tag.Name)
		}
		candidates = append(candidates, &selector.Candidate{
			Kind: selector.Image,
			Tags: tags,
			// keep digest for later match
			Digest: art.Digest,
		})
	}

	candidates, err := sel.Select(candidates)
	if err != nil {
		return nil, err
	}

	candidateDigests := make(map[string]bool)
	for _, c := range candidates {
		candidateDigests[c.Digest] = true
	}

	filteredArts := make([]*artifact.Artifact, 0, len(candidateDigests))
	for _, art := range arts {
		if candidateDigests[art.Digest] {
			filteredArts = append(filteredArts, art)
		}
	}

	return filteredArts, nil
}

func (dfp *DefaultFilterProcessor) ProcessLabelFilter(ctx context.Context, labelIDs []int64, arts []*artifact.Artifact) ([]*artifact.Artifact, error) {
	// return all artifacts if no label need to be filtered
	if len(labelIDs) == 0 {
		return arts, nil
	}
	// matchLabel check whether the artifact match the label filter
	matchLabel := func(art *artifact.Artifact) bool {
		// TODO (as now there should not have many labels, so here just use
		// for^2, we can convert to use map to reduce the time complex if needed. )
		for _, label := range art.Labels {
			for _, labelID := range labelIDs {
				if labelID == label.ID {
					return true
				}
			}
		}
		return false
	}

	filteredArts := make([]*artifact.Artifact, 0)
	for _, art := range arts {
		if matchLabel(art) {
			filteredArts = append(filteredArts, art)
		}
	}

	return filteredArts, nil
}

func (dfp *DefaultFilterProcessor) getProjectQueryFilter(user *commonmodels.User) *q.Query {
	secContext := local.NewSecurityContext(user)
	if secContext.IsSysAdmin() {
		logger.Infof("User %v is sys admin. Selecting all projects for export.", user.Username)
		return q.New(q.KeyWords{})
	}

	logger.Infof("User %v is not sys admin. Selecting projects with admin roles for export.", user.Username)
	return q.New(q.KeyWords{"member": &models.MemberQuery{UserID: user.UserID, GroupIDs: user.GroupIDs}})
}
