package native

import (
	"errors"
	"strings"

	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/util"
)

var _ adp.ImageRegistry = native{}

// TODO: support other filters
// MUST have filters and name filter
// NOT support namespaces
func (n native) FetchImages(namespaces []string, filters []*model.Filter) ([]*model.Resource, error) {
	if len(namespaces) > 0 {
		return nil, errors.New("native registry adapter not support namespace")
	}

	if len(filters) < 1 {
		return nil, errors.New("no any repository filter")
	}

	var resources = []*model.Resource{}
	var nameFilter, tagFilter *model.Filter
	for i, filter := range filters {
		switch filter.Type {
		case model.FilterTypeName:
			nameFilter = filters[i]
		case model.FilterTypeTag:
			tagFilter = filters[i]
		}
	}

	repositories, err := n.filterRepoistory(nameFilter)
	if err != nil {
		return nil, err
	}

	var tagFilterTerm string
	var haveWildcardChars bool

	if tagFilter != nil {
		tagFilterTerm = tagFilter.Value.(string)
	}

	if strings.ContainsAny(tagFilterTerm, "?*") {
		haveWildcardChars = true
	}

	if haveWildcardChars || tagFilterTerm == "" {
		// need call list tag api
		for _, repository := range repositories {
			var tags []string
			resp, err := n.DefaultImageRegistry.ListTag(repository)
			if err != nil {
				return nil, err
			}

			if haveWildcardChars {
				for _, tag := range resp {
					if m, _ := util.Match(tagFilterTerm, tag); m {
						tags = append(tags, tag)
					}
				}
			} else {
				tags = resp
			}

			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeRepository,
				Registry: n.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repository,
					},
					Vtags: tags,
				},
			})
		}
	} else if tagFilterTerm != "" {
		for _, repository := range repositories {
			resources = append(resources, &model.Resource{
				Type:     model.ResourceTypeRepository,
				Registry: n.registry,
				Metadata: &model.ResourceMetadata{
					Repository: &model.Repository{
						Name: repository,
					},
					Vtags: []string{tagFilterTerm},
				},
			})
		}
	}

	return resources, nil
}

func (n native) filterRepoistory(nameFilter *model.Filter) (repositories []string, err error) {
	if nameFilter == nil {
		return nil, errors.New("native registry adapter must have repository filter")
	}
	var nameFilterTerm = nameFilter.Value.(string)

	// search repoistories from catalog api
	if strings.ContainsAny(nameFilterTerm, "*?") {
		repos, err := n.DefaultImageRegistry.Catalog()
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			m, err := util.Match(nameFilterTerm, repo)
			if err != nil {
				return nil, err
			}
			if m {
				repositories = append(repositories, repo)
			}
		}
	} else if nameFilterTerm != "" {
		// only single repository
		repositories = []string{nameFilterTerm}
	}

	return
}
