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

package filter

import (
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/reg/util"
	"strings"
)

// DoFilterArtifacts filter the artifacts according to the filters
func DoFilterArtifacts(artifacts []*model.Artifact, filters []*model.Filter) ([]*model.Artifact, error) {
	fl, err := BuildArtifactFilters(filters)
	if err != nil {
		return nil, err
	}
	return fl.Filter(artifacts)
}

// BuildArtifactFilters from the defined filters
func BuildArtifactFilters(filters []*model.Filter) (ArtifactFilters, error) {
	var fs ArtifactFilters
	for _, filter := range filters {
		var f ArtifactFilter
		switch filter.Type {
		case model.FilterTypeLabel:
			f = &artifactLabelFilter{
				labels:     filter.Value.([]string),
				decoration: filter.Decoration,
			}
		case model.FilterTypeTag:
			f = &artifactTagFilter{
				pattern:    filter.Value.(string),
				decoration: filter.Decoration,
			}
		case model.FilterTypeResource:
			v := filter.Value.(string)
			if v != model.ResourceTypeArtifact && v != model.ResourceTypeChart {
				f = &artifactTypeFilter{
					types: []string{v},
				}
			}
		}
		if f != nil {
			fs = append(fs, f)
		}
	}
	return fs, nil
}

// ArtifactFilter filter the artifacts
type ArtifactFilter interface {
	Filter([]*model.Artifact) ([]*model.Artifact, error)
}

// ArtifactFilters is an array of artifact filter
type ArtifactFilters []ArtifactFilter

// Filter artifacts
func (a ArtifactFilters) Filter(artifacts []*model.Artifact) ([]*model.Artifact, error) {
	var err error
	for _, filter := range a {
		artifacts, err = filter.Filter(artifacts)
		if err != nil {
			return nil, err
		}
	}
	return artifacts, nil
}

type artifactTypeFilter struct {
	types []string
}

func (a *artifactTypeFilter) Filter(artifacts []*model.Artifact) ([]*model.Artifact, error) {
	if len(a.types) == 0 {
		return artifacts, nil
	}
	var result []*model.Artifact
	for _, artifact := range artifacts {
		for _, t := range a.types {
			if strings.EqualFold(strings.ToLower(artifact.Type), strings.ToLower(t)) {
				result = append(result, artifact)
				continue
			}
		}
	}
	return result, nil
}

// filter the artifacts according to the labels. Only the artifact contains all labels defined
// in the filter is the valid one
type artifactLabelFilter struct {
	labels []string
	// "matches", "excludes"
	decoration string
}

func (a *artifactLabelFilter) Filter(artifacts []*model.Artifact) ([]*model.Artifact, error) {
	if len(a.labels) == 0 {
		return artifacts, nil
	}
	var result []*model.Artifact
	for _, artifact := range artifacts {
		labels := map[string]struct{}{}
		for _, label := range artifact.Labels {
			labels[label] = struct{}{}
		}
		match := true
		for _, label := range a.labels {
			if _, exist := labels[label]; !exist {
				match = false
				break
			}
		}
		// add the artifact to the result list if it contains all labels defined for the filter
		if a.decoration == model.Excludes {
			if !match {
				result = append(result, artifact)
			}
		} else {
			if match {
				result = append(result, artifact)
			}
		}
	}
	return result, nil
}

// filter artifacts according to whether the artifact is tagged or untagged artifact
type artifactTaggedFilter struct {
	tagged bool
}

func (a *artifactTaggedFilter) Filter(artifacts []*model.Artifact) ([]*model.Artifact, error) {
	var result []*model.Artifact
	for _, artifact := range artifacts {
		if a.tagged && len(artifact.Tags) > 0 ||
			!a.tagged && len(artifact.Tags) == 0 {
			result = append(result, artifact)
		}
	}
	return result, nil
}

type artifactTagFilter struct {
	pattern string
	// "matches", "excludes"
	decoration string
}

func (a *artifactTagFilter) Filter(artifacts []*model.Artifact) ([]*model.Artifact, error) {
	if len(a.pattern) == 0 {
		return artifacts, nil
	}
	var result []*model.Artifact
	for _, artifact := range artifacts {
		// for individual artifact, use its own tags to match, reserve the matched tags.
		// for accessory artifact, use the parent tags to match,
		var tagsForMatching []string
		if artifact.IsAcc {
			tagsForMatching = append(tagsForMatching, artifact.ParentTags...)
		} else {
			tagsForMatching = append(tagsForMatching, artifact.Tags...)
		}

		// untagged artifact
		if len(tagsForMatching) == 0 {
			match, err := util.Match(a.pattern, "")
			if err != nil {
				return nil, err
			}
			if a.decoration == model.Excludes {
				if !match {
					result = append(result, artifact)
				}
			} else {
				if match {
					result = append(result, artifact)
				}
			}
			continue
		}

		// tagged artifact
		var tags []string
		for _, tag := range tagsForMatching {
			match, err := util.Match(a.pattern, tag)
			if err != nil {
				return nil, err
			}
			if a.decoration == model.Excludes {
				if !match {
					tags = append(tags, tag)
				}
			} else {
				if match {
					tags = append(tags, tag)
				}
			}
		}
		if len(tags) == 0 {
			continue
		}
		// copy a new artifact here to avoid changing the original one
		if artifact.IsAcc {
			result = append(result, &model.Artifact{
				Type:   artifact.Type,
				Digest: artifact.Digest,
				Labels: artifact.Labels,
				Tags:   artifact.Tags, // use its own tags to replicate
			})
		} else {
			result = append(result, &model.Artifact{
				Type:   artifact.Type,
				Digest: artifact.Digest,
				Labels: artifact.Labels,
				Tags:   tags, // only replicate the matched tags
			})
		}
	}
	return result, nil
}
