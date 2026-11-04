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

package huggingface

import (
	"context"
	"sort"

	"github.com/goharbor/harbor/src/lib/errors"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/synth"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var _ adp.ArtifactResolver = (*adapter)(nil)

// Resolve previews the copy of a model at a reference. repository is a model ID in any case;
// reference is a tag, a 40-hex commit, or empty for the default branch. It costs the Hub API
// calls of a manifest pull: the refs (unless reference is a commit), the revision listing,
// and the download of every selected non-LFS file whose hash is not cached yet. The resulting
// digests are indexed, so a following copy or proxy pull reuses them.
func (a *adapter) Resolve(repository, reference string, filter *model.FileFilter) (*model.ResolveResult, error) {
	ctx := context.Background()
	if err := filter.Validate(); err != nil {
		return nil, err
	}
	repo, ok := repositoryOf(repository)
	if !ok {
		return nil, errors.BadRequestError(nil).WithMessagef("invalid hugging face model %q, expected org/name", repository)
	}
	if reference == "" {
		reference = defaultBranch
	}
	commit := reference
	if !hub.IsCommit(reference) {
		var err error
		if commit, err = a.commitOf(ctx, repo, reference); err != nil {
			return nil, err
		}
	}
	s, err := a.snapshot(ctx, repo, commit)
	if err != nil {
		return nil, err
	}
	refs, err := a.refs(ctx, repo)
	if err != nil {
		return nil, err
	}
	tags := newRefIndex(refs).tagsByCommit()[s.Commit]
	if tags == nil {
		tags = []string{}
	}

	result := &model.ResolveResult{
		Repository: repo,
		SourceID:   s.ModelID,
		Reference:  reference,
		Revision:   s.Commit,
		SourceURL:  hub.DefaultEndpoint + "/" + s.ModelID,
		Tags:       tags,
		Licenses:   s.Licenses,
		Task:       s.PipelineTag,
		Files:      []*model.ResolvedFile{},
	}
	if !s.LastModified.IsZero() {
		lastModified := s.LastModified.UTC()
		result.LastModified = &lastModified
	}
	for _, f := range s.Files {
		if !synth.Included(f.Path) {
			continue
		}
		file := &model.ResolvedFile{
			Path:    f.Path,
			Size:    f.Size,
			Class:   synth.Class(f.Path, f.Size),
			SHA256:  f.SHA256,
			Matched: filter.Match(f.Path),
		}
		result.TotalFiles++
		result.TotalSize += f.Size
		if file.Matched {
			result.MatchedFiles++
			result.MatchedSize += f.Size
		}
		result.Files = append(result.Files, file)
	}
	sort.Slice(result.Files, func(i, j int) bool { return result.Files[i].Path < result.Files[j].Path })
	if result.MatchedFiles == 0 {
		return result, nil
	}

	digests, err := a.contentDigests(ctx, s, filter)
	if err != nil {
		return nil, err
	}
	art, err := synth.BuildFiltered(s, digests, filter)
	if err != nil {
		return nil, err
	}
	a.locator.index(ctx, repo, s.ModelID, s.Commit, filter, art)
	result.Digest = art.Digest.String()
	for _, f := range result.Files {
		if d, ok := digests[f.Path]; ok {
			f.SHA256 = d.Encoded()
		}
	}
	return result, nil
}
