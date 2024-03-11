// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package export

import (
	"context"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/lib/selector/selectors/doublestar"
	"github.com/goharbor/harbor/src/pkg"
	artpkg "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/user"
)

type FilterProcessor interface {
	ProcessRepositoryFilter(ctx context.Context, filter string, projectIDs []int64) ([]int64, error)
	ProcessTagFilter(ctx context.Context, filter string, repositoryIDs []int64) ([]*artifact.Artifact, error)
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

func (dfp *DefaultFilterProcessor) ProcessRepositoryFilter(ctx context.Context, filter string, projectIDs []int64) ([]int64, error) {
	sel := doublestar.New(doublestar.RepoMatches, filter, "")
	candidates := make([]*selector.Candidate, 0)
	allRepoIDs := make([]int64, 0)

	for _, projectID := range projectIDs {
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

func (dfp *DefaultFilterProcessor) ProcessTagFilter(ctx context.Context, filter string, repositoryIDs []int64) ([]*artifact.Artifact, error) {
	arts := make([]*artifact.Artifact, 0)
	opts := &artifact.Option{
		WithTag:   true,
		WithLabel: true,
		// if accessory support scan in the future, just add withAccessory here.
		// WithAccessory: true
	}
	// list all artifacts by repository id
	for _, repoID := range repositoryIDs {
		repoArts, err := dfp.artCtl.List(ctx, q.New(q.KeyWords{"RepositoryID": repoID}), opts)
		if err != nil {
			return nil, err
		}

		for _, art := range repoArts {
			if art.IsImageIndex() {
				for _, ref := range art.References {
					arts = append(arts, &artifact.Artifact{
						Artifact: artpkg.Artifact{
							ID:     ref.ChildID,
							Digest: ref.ChildDigest,
						},
						Tags:   art.Tags,
						Labels: art.Labels,
					})
				}
			}

			arts = append(arts, art)
		}
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

func (dfp *DefaultFilterProcessor) ProcessLabelFilter(_ context.Context, labelIDs []int64, arts []*artifact.Artifact) ([]*artifact.Artifact, error) {
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
