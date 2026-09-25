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
	"strings"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/filter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

// fetchConcurrency stays low because the anonymous Hub API budget is 500 calls per 5 minutes.
const fetchConcurrency = 4

// FetchArtifacts returns one artifact per distinct commit of each matching model, tagged with
// the tags of all refs that point at it and match the tag filter. The name filter must pin the
// author, e.g. "org/**" or "org/name", and is matched case-insensitively.
func (a *adapter) FetchArtifacts(filters []*model.Filter) ([]*model.Resource, error) {
	ctx := context.Background()
	pattern, err := namePattern(filters)
	if err != nil {
		return nil, err
	}
	modelIDs, err := a.listModelIDs(ctx, pattern)
	if err != nil {
		return nil, err
	}

	resources := make([]*model.Resource, len(modelIDs))
	runner := utils.NewLimitedConcurrentRunner(fetchConcurrency)
	for i, modelID := range modelIDs {
		runner.AddTask(func() error {
			repository, ok := repositoryOf(modelID)
			if !ok {
				log.Warningf("skip hugging face model %s: its lowercase form is not a valid repository name", modelID)
				return nil
			}
			artifacts, err := a.listArtifacts(ctx, modelID, filters)
			if err != nil {
				return err
			}
			if len(artifacts) == 0 {
				return nil
			}
			resources[i] = &model.Resource{
				Type:     model.ResourceTypeArtifact,
				Registry: a.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{Name: repository},
					Artifacts:  artifacts,
				},
			}
			return nil
		})
	}
	// A partial listing is never returned: a silently shorter artifact list is worse than a failed run.
	if err := runner.Wait(); err != nil {
		return nil, err
	}
	var result []*model.Resource
	for _, r := range resources {
		if r != nil {
			result = append(result, r)
		}
	}
	return result, nil
}

func namePattern(filters []*model.Filter) (string, error) {
	for _, f := range filters {
		if f.Type != model.FilterTypeName {
			continue
		}
		pattern, ok := f.Value.(string)
		if !ok {
			return "", errors.BadRequestError(nil).WithMessage("invalid value of the name filter")
		}
		if strings.Count(pattern, "/") >= 1 {
			if _, ok := util.IsSpecificPathComponent(strings.SplitN(pattern, "/", 2)[0]); ok {
				return strings.ToLower(pattern), nil
			}
		}
		break
	}
	return "", errors.BadRequestError(nil).
		WithMessage(`the hugging face registry requires a name filter that pins the author, e.g. "org/**" or "org/name"`)
}

// listModelIDs returns the model IDs matching the lowercase pattern. A fully specific pattern is
// used as is (the Hub resolves model IDs case-insensitively); otherwise the models of each
// author are listed.
func (a *adapter) listModelIDs(ctx context.Context, pattern string) ([]string, error) {
	if ids, ok := util.IsSpecificPath(pattern); ok {
		sort.Strings(ids)
		return ids, nil
	}
	authors, _ := util.IsSpecificPathComponent(strings.SplitN(pattern, "/", 2)[0])
	var ids []string
	for _, author := range authors {
		// The listing matches the author case-sensitively.
		canonical, err := a.hub.Author(ctx, author)
		if err != nil {
			return nil, err
		}
		listed, err := a.hub.ListModels(ctx, canonical)
		if err != nil {
			return nil, err
		}
		for _, id := range listed {
			match, err := util.Match(pattern, strings.ToLower(id))
			if err != nil {
				return nil, err
			}
			if match {
				ids = append(ids, id)
			}
		}
	}
	sort.Strings(ids)
	return ids, nil
}

func (a *adapter) listArtifacts(ctx context.Context, modelID string, filters []*model.Filter) ([]*model.Artifact, error) {
	refs, err := a.hub.Refs(ctx, modelID)
	if err != nil {
		return nil, err
	}
	idx := newRefIndex(refs)
	for tag, names := range idx.collisions {
		log.Warningf("skip tag %s of hugging face model %s: refs %s all map to it", tag, modelID, strings.Join(names, ", "))
	}
	byCommit := idx.tagsByCommit()
	commits := make([]string, 0, len(byCommit))
	for c := range byCommit {
		commits = append(commits, c)
	}
	sort.Strings(commits)
	artifacts := make([]*model.Artifact, 0, len(commits))
	for _, c := range commits {
		artifacts = append(artifacts, &model.Artifact{Tags: byCommit[c]})
	}
	return filter.DoFilterArtifacts(artifacts, filters)
}
