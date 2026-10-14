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
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/hub"
)

var (
	invalidTagChars = regexp.MustCompile(`[^\w.-]`)
	tagRe           = regexp.MustCompile(`^[\w][\w.-]{0,127}$`)
)

// repositoryOf returns the repository of a model ID. Some model IDs stay invalid OCI names after
// lowercasing (e.g. "_-", "._" or a trailing "." or "-"); those return false.
func repositoryOf(modelID string) (string, bool) {
	repository := strings.ToLower(modelID)
	if strings.Count(repository, "/") != 1 || !lib.RepositoryNameRe.MatchString(repository) {
		return "", false
	}
	return repository, true
}

// tagOf returns the tag of a ref. The mapping is one-way: "/" and other characters invalid in
// a tag become "-". Refs that cannot become a tag, or would look like a commit, return false.
func tagOf(ref string) (string, bool) {
	tag := invalidTagChars.ReplaceAllString(ref, "-")
	if !tagRe.MatchString(tag) || hub.IsCommit(tag) {
		return "", false
	}
	return tag, true
}

// refIndex maps the tags of a model's refs to commits.
type refIndex struct {
	commits map[string]string
	// collisions holds tags that more than one ref sanitizes to.
	collisions map[string][]string
}

func newRefIndex(refs []hub.Ref) *refIndex {
	idx := &refIndex{commits: map[string]string{}, collisions: map[string][]string{}}
	names := map[string][]string{}
	for _, r := range refs {
		tag, ok := tagOf(r.Name)
		if !ok {
			continue
		}
		names[tag] = append(names[tag], r.Name)
		idx.commits[tag] = r.Commit
	}
	for tag, n := range names {
		if len(n) > 1 {
			sort.Strings(n)
			idx.collisions[tag] = n
			delete(idx.commits, tag)
		}
	}
	return idx
}

// commit returns the commit a tag points at.
func (r *refIndex) commit(tag string) (string, error) {
	if refs, ok := r.collisions[tag]; ok {
		return "", errors.New(nil).WithCode(errors.ConflictCode).
			WithMessagef("tag %s is ambiguous, refs %s all map to it", tag, strings.Join(refs, ", "))
	}
	commit, ok := r.commits[tag]
	if !ok {
		return "", errors.NotFoundError(fmt.Errorf("tag %s not found", tag))
	}
	return commit, nil
}

// tags returns the valid, unambiguous tags, sorted.
func (r *refIndex) tags() []string {
	tags := make([]string, 0, len(r.commits))
	for t := range r.commits {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	return tags
}

// tagsByCommit groups the valid tags by the commit they point at.
func (r *refIndex) tagsByCommit() map[string][]string {
	out := map[string][]string{}
	for _, t := range r.tags() {
		c := r.commits[t]
		out[c] = append(out[c], t)
	}
	return out
}
