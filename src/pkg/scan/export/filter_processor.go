package export

import (
	"context"
	"github.com/goharbor/harbor/src/pkg"

	"github.com/goharbor/harbor/src/common"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/lib/selector/selectors/doublestar"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/tag"
	"github.com/goharbor/harbor/src/pkg/user"
)

type FilterProcessor interface {
	ProcessProjectFilter(ctx context.Context, userName string, projectsToFilter []int64) ([]int64, error)
	ProcessRepositoryFilter(ctx context.Context, filter string, projectIds []int64) ([]*selector.Candidate, error)
	ProcessTagFilter(ctx context.Context, filter string, repositoryIds []int64) ([]*selector.Candidate, error)
}

type DefaultFilterProcessor struct {
	repoMgr    repository.Manager
	tagMgr     tag.Manager
	usrMgr     user.Manager
	projectMgr project.Manager
}

// NewFilterProcessor constructs an instance of a FilterProcessor
func NewFilterProcessor() FilterProcessor {
	return &DefaultFilterProcessor{
		repoMgr:    pkg.RepositoryMgr,
		tagMgr:     tag.Mgr,
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

func (dfp *DefaultFilterProcessor) ProcessRepositoryFilter(ctx context.Context, filter string, projectIds []int64) ([]*selector.Candidate, error) {
	sel := doublestar.New(doublestar.RepoMatches, filter, "")
	candidates := make([]*selector.Candidate, 0)

	for _, projectID := range projectIds {
		query := q.New(q.KeyWords{"ProjectID": projectID})
		allRepos, err := dfp.repoMgr.List(ctx, query)
		if err != nil {
			return nil, err
		}
		for _, repository := range allRepos {
			namespace, repo := utils.ParseRepository(repository.Name)
			candidates = append(candidates, &selector.Candidate{NamespaceID: repository.RepositoryID, Namespace: namespace, Repository: repo, Kind: "image"})
		}
	}
	// no repo filter specified then return all repos across all projects
	if filter == "" {
		return candidates, nil
	}
	return sel.Select(candidates)
}

func (dfp *DefaultFilterProcessor) ProcessTagFilter(ctx context.Context, filter string, repositoryIds []int64) ([]*selector.Candidate, error) {
	sel := doublestar.New(doublestar.Matches, filter, "")
	candidates := make([]*selector.Candidate, 0)

	for _, repoID := range repositoryIds {
		query := q.New(q.KeyWords{"RepositoryID": repoID})
		allTags, err := dfp.tagMgr.List(ctx, query)
		if err != nil {
			return nil, err
		}
		cand := &selector.Candidate{NamespaceID: repoID, Kind: "image"}
		for _, tag := range allTags {
			cand.Tags = append(cand.Tags, tag.Name)
		}
		candidates = append(candidates, cand)
	}
	// no tags specified then simply return all the candidates
	if filter == "" {
		return candidates, nil
	}
	return sel.Select(candidates)
}

func (dfp *DefaultFilterProcessor) getProjectQueryFilter(user *commonmodels.User) *q.Query {
	secContext := local.NewSecurityContext(user)
	if secContext.IsSysAdmin() {
		logger.Infof("User %v is sys admin. Selecting all projects for export.", user.Username)
		return q.New(q.KeyWords{})
	}
	logger.Infof("User %v is not sys admin. Selecting projects with admin roles for export.", user.Username)
	return q.New(q.KeyWords{"member": &models.MemberQuery{UserID: user.UserID, Role: common.RoleProjectAdmin}})
}
